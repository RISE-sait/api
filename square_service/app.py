from fastapi import FastAPI, HTTPException, Request, Depends
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import os
import psycopg2
import httpx
from typing import Optional
from dotenv import load_dotenv
from datetime import datetime
from dateutil.relativedelta import relativedelta
import jwt
import hmac
import hashlib
import base64
from contextlib import closing
import logging
from slowapi import Limiter
from slowapi.middleware import SlowAPIMiddleware
from slowapi.util import get_remote_address

# Load + enforce required env vars

load_dotenv()
REQUIRED_ENVS = [
    "DATABASE_URL",
    "JWT_SECRET",
    "SQUARE_ACCESS_TOKEN",
    "SQUARE_LOCATION_ID",
    "SQUARE_WEBHOOK_SIGNATURE_KEY",
    "FRONTEND_URL",
    "WEBHOOK_URL",
]
_missing = [v for v in REQUIRED_ENVS if not os.getenv(v)]
if _missing:
    raise RuntimeError(f"Missing required env vars: {', '.join(_missing)}")


# Logging: DEBUG by default in non-production; configurable via LOG_LEVEL

DEFAULT_LEVEL = "INFO" if os.getenv("SQUARE_ENV") == "production" else "DEBUG"
LOG_LEVEL = os.getenv("LOG_LEVEL", DEFAULT_LEVEL).upper()
logging.basicConfig(level=getattr(logging, LOG_LEVEL, logging.INFO))
logger = logging.getLogger(__name__)
logger.debug("[BOOT] log level=%s", LOG_LEVEL)


# FastAPI app & middleware

app = FastAPI()

# CORS: only your frontend
app.add_middleware(
    CORSMiddleware,
    allow_origins=[os.getenv("FRONTEND_URL")],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Rate limiting
limiter = Limiter(key_func=get_remote_address)
app.state.limiter = limiter
app.add_middleware(SlowAPIMiddleware)

# Security headers
@app.middleware("http")
async def security_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["Strict-Transport-Security"] = "max-age=63072000; includeSubDomains"
    response.headers["X-Frame-Options"] = "DENY"
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["Referrer-Policy"] = "no-referrer"
    response.headers["Permissions-Policy"] = "geolocation=()"
    return response


# Square API configuration

SQUARE_BASE_URL = (
    "https://connect.squareupsandbox.com"
    if os.getenv("SQUARE_ENV") != "production"
    else "https://connect.squareup.com"
)
HEADERS = {
    "Square-Version": os.getenv("SQUARE_VERSION", "2025-07-16"),
    "Authorization": f"Bearer {os.getenv('SQUARE_ACCESS_TOKEN')}",
    "Content-Type": "application/json",
}

# JWT & Webhook keys
JWT_SECRET = os.getenv("JWT_SECRET")
JWT_ALGORITHM = "HS256"
WEBHOOK_SIGNATURE_KEY = os.getenv("SQUARE_WEBHOOK_SIGNATURE_KEY")
WEBHOOK_URL = os.getenv("WEBHOOK_URL")


# Models

class CheckoutRequest(BaseModel):
    membership_plan_id: str
    discount_code: Optional[str] = None

class ProgramCheckoutRequest(BaseModel):
    program_id: str

class EventCheckoutRequest(BaseModel):
    event_id: str


# DB helper

def get_db_conn():
    return psycopg2.connect(os.getenv("DATABASE_URL"))

def _warn_if_db_superuser():
    """Best-effort warning if connected as a superuser (least-privilege reminder)."""
    try:
        with closing(get_db_conn()) as conn, conn.cursor() as cur:
            cur.execute("SELECT current_user")
            user = cur.fetchone()[0]
            if user == "postgres":
                logger.warning("[SECURITY] Connected as DB superuser 'postgres'. "
                               "Use a least-privileged DB user in production.")
            else:
                logger.debug("[SECURITY] Connected DB user=%s", user)
    except Exception as e:
        logger.debug("[SECURITY] DB user check skipped: %s", e)

_warn_if_db_superuser()


# JWT decode helper

def get_user_id(creds: HTTPAuthorizationCredentials) -> str:
    token = creds.credentials
    try:
        payload = jwt.decode(
            token,
            JWT_SECRET,
            algorithms=[JWT_ALGORITHM],
            options={"require": ["exp"]},  # require expiration
        )
        return payload.get("user_id") or payload.get("sub")
    except jwt.ExpiredSignatureError:
        logger.warning("JWT expired")
        raise HTTPException(status_code=401, detail="Token expired")
    except jwt.InvalidTokenError as e:
        logger.error(f"JWT invalid: {e}")
        raise HTTPException(status_code=401, detail="Invalid token")


# Signature verification

def verify_signature(signature_header: str, body: bytes) -> bool:
    # Square signs: HMAC_SHA256
    url = (WEBHOOK_URL or "").encode("utf-8")
    computed = base64.b64encode(
        hmac.new(WEBHOOK_SIGNATURE_KEY.encode("utf-8"), url + body, hashlib.sha256).digest()
    ).decode("utf-8")

    # Only log sensitive details at DEBUG
    logger.debug("[WEBHOOK] env_url=%r body_len=%s", WEBHOOK_URL, len(body))
    logger.debug("[WEBHOOK] header_sig=%r", signature_header)
    logger.debug("[WEBHOOK] computed_sig=%r", computed)

    ok = hmac.compare_digest(signature_header or "", computed)
    if not ok:
        logger.warning("[WEBHOOK] signature mismatch — check WEBHOOK_URL & SQUARE_WEBHOOK_SIGNATURE_KEY")
    return ok


# Payload generator

def generate_checkout_payload(location_id, catalog_ids, metadata):
    line_items = [{"catalog_object_id": cid, "quantity": "1"} for cid in catalog_ids]
    if line_items:
        line_items[0]["metadata"] = metadata

    payload = {
        "idempotency_key": os.urandom(16).hex(),
        "order": {
            "location_id": location_id,
            "line_items": line_items,
        },
        "ask_for_shipping_address": False,
    }
    return payload

# Membership checkout endpoint

@app.post("/checkout/membership")
@limiter.limit("10/minute")
async def checkout_membership(
    request: Request,
    req: CheckoutRequest,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer()),
):
    customer_id = get_user_id(creds)
    logger.info(f"Received membership checkout for user: {customer_id}, plan: {req.membership_plan_id}")

    with closing(get_db_conn()) as conn, conn.cursor() as cur:
        cur.execute(
            "SELECT stripe_price_id, stripe_joining_fee_id, amt_periods "
            "FROM membership.membership_plans WHERE id = %s",
            (req.membership_plan_id,),
        )
        row = cur.fetchone()
        if not row:
            raise HTTPException(status_code=404, detail="membership plan not found")
        price_id, joining_fee_id, amt_periods = row

        renewal_date = None
        if amt_periods is not None:
            renewal_date = datetime.utcnow() + relativedelta(months=amt_periods)

        # webhook sets to 'active' on payment completion.
        cur.execute(
            "INSERT INTO users.customer_membership_plans "
            "(customer_id, membership_plan_id, renewal_date, status) "
            "VALUES (%s,%s,%s,'inactive')",
            (customer_id, req.membership_plan_id, renewal_date),
        )
        conn.commit()

    metadata = {"user_id": customer_id, "membership_plan_id": req.membership_plan_id}
    if amt_periods is not None:
        metadata["amt_periods"] = str(amt_periods)

    catalog_ids = [price_id] + ([joining_fee_id] if joining_fee_id else [])
    payload = generate_checkout_payload(os.getenv("SQUARE_LOCATION_ID"), catalog_ids, metadata)

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.post(
                f"{SQUARE_BASE_URL}/v2/online-checkout/payment-links",
                headers=HEADERS,
                json=payload,
            )
            if resp.status_code >= 400:
                logger.error("[SQUARE] checkout_membership error %s: %s", resp.status_code, resp.text)
                raise HTTPException(status_code=502, detail="Payment provider error")
            return {"checkout_url": resp.json()["payment_link"]["url"]}
    except httpx.RequestError as e:
        logger.exception("[SQUARE] checkout_membership request failed: %s", e)
        raise HTTPException(status_code=502, detail="Payment provider error")


# Event checkout endpoint
@app.post("/checkout/event")
@limiter.limit("10/minute")
async def checkout_event(
    request: Request,
    req: EventCheckoutRequest,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer()),
):
    customer_id = get_user_id(creds)
    logger.info(f"[EVENT CHECKOUT] user={customer_id} event={req.event_id}")

    with closing(get_db_conn()) as conn, conn.cursor() as cur:
        # Fetch price and required membership for the event
        cur.execute(
            "SELECT price_id, required_membership_plan_id FROM events.events WHERE id = %s",
            (req.event_id,),
        )
        row = cur.fetchone()
        if not row:
            raise HTTPException(status_code=404, detail="event not found")
        price_id, required_plan_id = row

        # Upsert enrollment set as pending
        cur.execute(
            """
            INSERT INTO events.customer_enrollment
                (customer_id, event_id, payment_expired_at, payment_status)
            VALUES (%s, %s, CURRENT_TIMESTAMP + interval '10 minute', 'pending')
            ON CONFLICT (customer_id, event_id) DO UPDATE
                SET payment_expired_at = EXCLUDED.payment_expired_at,
                    payment_status     = EXCLUDED.payment_status
            WHERE events.customer_enrollment.payment_status != 'paid'
            """,
            (customer_id, req.event_id),
        )

        # Check membership (JWT first, DB fallback)
        has_membership = False
        if required_plan_id:
            try:
                claims = jwt.decode(
                    creds.credentials,
                    JWT_SECRET,
                    algorithms=[JWT_ALGORITHM],
                    options={"require": ["exp"]},
                )
            except Exception:
                claims = {}

            claimed = set()
            for k in ("membership_plan_ids", "memberships"):
                v = claims.get(k)
                if isinstance(v, (list, tuple)):
                    claimed.update(str(x) for x in v)

            if str(required_plan_id) in claimed:
                has_membership = True
            else:
                # Treat expired memberships as inactive
                cur.execute(
                    """
                    SELECT 1
                      FROM users.customer_membership_plans
                     WHERE customer_id = %s
                       AND membership_plan_id = %s
                       AND status = 'active'
                       AND (renewal_date IS NULL OR renewal_date > NOW())
                     LIMIT 1
                    """,
                    (customer_id, required_plan_id),
                )
                if cur.fetchone():
                    has_membership = True

        # Decide flow
        needs_payment = False
        requires_membership = bool(required_plan_id)

        if has_membership:
            # Covered by membership = free
            cur.execute(
                """
                UPDATE events.customer_enrollment
                   SET payment_status='paid'
                 WHERE customer_id=%s AND event_id=%s
                """,
                (customer_id, req.event_id),
            )
            conn.commit()
            logger.info(f"[EVENT CHECKOUT] covered by membership user={customer_id} event={req.event_id}")
            return {"checkout_url": None}

        if requires_membership and not has_membership:
            # No membership, but membership required pay if we have a price, else block
            if price_id:
                needs_payment = True
            else:
                conn.rollback()
                raise HTTPException(status_code=403, detail="membership required")

        if not requires_membership:
            if not price_id:
                # Free public event
                cur.execute(
                    """
                    UPDATE events.customer_enrollment
                       SET payment_status='paid'
                     WHERE customer_id=%s AND event_id=%s
                    """,
                    (customer_id, req.event_id),
                )
                conn.commit()
                logger.info(f"[EVENT CHECKOUT] free event user={customer_id} event={req.event_id}")
                return {"checkout_url": None}
            else:
                needs_payment = True

        conn.commit()

    # Create a Square payment link using the variation in price_id
    metadata = {"user_id": customer_id, "event_id": req.event_id}
    payload = generate_checkout_payload(os.getenv("SQUARE_LOCATION_ID"), [price_id], metadata)

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.post(
                f"{SQUARE_BASE_URL}/v2/online-checkout/payment-links",
                headers=HEADERS,
                json=payload,
            )
            if resp.status_code >= 400:
                logger.error("[SQUARE] checkout_event error %s: %s", resp.status_code, resp.text)
                raise HTTPException(status_code=502, detail="Payment provider error")

            url = resp.json()["payment_link"]["url"]
            logger.info(
                f"[EVENT CHECKOUT] payment_link created user={customer_id} event={req.event_id} price_id={price_id} url={url}"
            )
            return {"checkout_url": url}
    except httpx.RequestError as e:
        logger.exception("[SQUARE] checkout_event request failed: %s", e)
        raise HTTPException(status_code=502, detail="Payment provider error")


# Webhook endpoint logging in DEBUG
@app.post("/webhook")
@limiter.limit("60/minute")
async def handle_webhook(request: Request):
    signature = request.headers.get("x-square-hmacsha256-signature", "")
    body = await request.body()

    if not verify_signature(signature, body):
        logger.warning("[WEBHOOK] invalid signature")
        raise HTTPException(status_code=400, detail="Invalid signature")

    try:
        event = await request.json()
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid JSON")

    evt_type = event.get("type")
    logger.info(f"[WEBHOOK] received event type={evt_type}")

    if evt_type == "payment.updated":
        payment = event.get("data", {}).get("object", {}).get("payment") or {}
        status = payment.get("status")
        order_id = payment.get("order_id")
        amount_money = payment.get("amount_money") or {}
        logger.info(
            "[WEBHOOK] payment.updated status=%s order_id=%s payment_id=%s amount=%s currency=%s",
            status, order_id, payment.get("id"), amount_money.get("amount"), amount_money.get("currency")
        )

        if status == "COMPLETED":
            if not order_id:
                logger.warning("[WEBHOOK] completed payment without order_id")
                return {"status": "ignored"}

            # Pull order to read metadata
            try:
                async with httpx.AsyncClient(timeout=10.0) as client:
                    resp = await client.get(f"{SQUARE_BASE_URL}/v2/orders/{order_id}", headers=HEADERS)
                    if resp.status_code >= 400:
                        logger.error("[SQUARE] fetch order %s failed: %s %s", order_id, resp.status_code, resp.text)
                        raise HTTPException(status_code=502, detail="Payment provider error")

                    order = resp.json().get("order", {}) or {}
                    metadata = order.get("metadata", {}) or {}
                    metadata_source = "order"
                    if not metadata and order.get("line_items"):
                        for li in order["line_items"]:
                            if li.get("metadata"):
                                metadata = li["metadata"]
                                metadata_source = "line_item"
                                break
            except httpx.RequestError as e:
                logger.exception("[SQUARE] fetch order failed: %s", e)
                raise HTTPException(status_code=502, detail="Payment provider error")

            logger.debug("[WEBHOOK] order_id=%s metadata_source=%s metadata=%s", order_id, metadata_source, metadata)

            user_id = metadata.get("user_id")
            membership_plan_id = metadata.get("membership_plan_id")
            event_id = metadata.get("event_id")
            periods = metadata.get("amt_periods")

            if not user_id:
                logger.warning("[WEBHOOK] missing user_id in metadata for order_id=%s", order_id)
                return {"status": "ignored"}

            try:
                with closing(get_db_conn()) as conn, conn.cursor() as cur:
                    did_anything = False

                    # Membership activation 
                    if membership_plan_id:
                        logger.info(
                            "[WEBHOOK] attempting membership activation user=%s plan=%s periods=%s",
                            user_id, membership_plan_id, periods
                        )

                        cur.execute(
                            """
                            UPDATE users.customer_membership_plans
                               SET status='active'
                             WHERE customer_id=%s
                               AND membership_plan_id=%s
                               AND status != 'active'
                            """,
                            (user_id, membership_plan_id),
                        )
                        rc = cur.rowcount
                        logger.info(
                            "[WEBHOOK] membership UPDATE → active rowcount=%s user=%s plan=%s",
                            rc, user_id, membership_plan_id
                        )
                        if rc == 0:
                            # If no row yet, insert as active (with renewal if provided)
                            if periods:
                                cur.execute("SELECT NOW() + (%s || ' months')::interval", (str(periods),))
                                renewal_date = cur.fetchone()[0]
                                cur.execute(
                                    """
                                    INSERT INTO users.customer_membership_plans
                                        (customer_id, membership_plan_id, renewal_date, status)
                                    VALUES (%s, %s, %s, 'active')
                                    """,
                                    (user_id, membership_plan_id, renewal_date),
                                )
                                logger.info("[WEBHOOK] membership inserted ACTIVE with renewal user=%s plan=%s renewal=%s",
                                            user_id, membership_plan_id, renewal_date)
                            else:
                                cur.execute(
                                    """
                                    INSERT INTO users.customer_membership_plans
                                        (customer_id, membership_plan_id, status)
                                    VALUES (%s, %s, 'active')
                                    """,
                                    (user_id, membership_plan_id),
                                )
                                logger.info("[WEBHOOK] membership inserted ACTIVE (no renewal) user=%s plan=%s",
                                            user_id, membership_plan_id)
                        did_anything = True

                    # --- Event enrollment paid ---
                    if event_id:
                        logger.info("[WEBHOOK] marking event enrollment paid user=%s event=%s", user_id, event_id)
                        cur.execute(
                            """
                            UPDATE events.customer_enrollment
                               SET payment_status='paid'
                             WHERE customer_id=%s AND event_id=%s
                            """,
                            (user_id, event_id),
                        )
                        erc = cur.rowcount
                        logger.info(
                            "[WEBHOOK] event enrollment UPDATE rowcount=%s user=%s event=%s",
                            erc, user_id, event_id
                        )
                        did_anything = did_anything or (erc > 0)

                    conn.commit()

                logger.info("[WEBHOOK] done order_id=%s did_anything=%s", order_id, did_anything)
                return {"status": "ok"}

            except Exception:
                logger.exception("[WEBHOOK] DB error while applying updates for order_id=%s", order_id)
                # Let Square retry if this is transient
                raise HTTPException(status_code=500, detail="DB error")

    return {"status": "ignored"}


# Admin: expire memberships (treat expired as inactive by flipping status)

only_authenticated = [Depends(HTTPBearer())]

@app.post("/admin/memberships/expire", dependencies=only_authenticated)
async def expire_memberships():
    with closing(get_db_conn()) as conn, conn.cursor() as cur:
        cur.execute(
            """
            UPDATE users.customer_membership_plans
               SET status = 'inactive'
             WHERE status = 'active'
               AND renewal_date IS NOT NULL
               AND renewal_date < NOW()
            """
        )
        changed = cur.rowcount
        conn.commit()
    logger.info(f"[ADMIN] expired {changed} memberships")
    return {"expired": changed}
