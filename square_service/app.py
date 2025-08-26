from fastapi import FastAPI, HTTPException, Request, Depends, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, validator, Field
import os
import psycopg2
from psycopg2 import pool
import httpx
from typing import Optional, Dict, Any
from dotenv import load_dotenv
from datetime import datetime
from dateutil.relativedelta import relativedelta
import jwt
import hmac
import hashlib
import base64
from contextlib import closing, contextmanager
import logging
import uuid
from slowapi import Limiter
from slowapi.middleware import SlowAPIMiddleware
from slowapi.util import get_remote_address
import asyncio
from functools import wraps

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


# Database connection pool
db_pool = None

def init_db_pool():
    global db_pool
    db_pool = psycopg2.pool.ThreadedConnectionPool(
        minconn=1,
        maxconn=20,
        dsn=os.getenv("DATABASE_URL")
    )
    logger.info("[DB] Connection pool initialized")

@contextmanager
def get_db_conn():
    """Get database connection from pool with proper error handling."""
    conn = None
    try:
        conn = db_pool.getconn()
        yield conn
    except Exception as e:
        if conn:
            conn.rollback()
        logger.error(f"[DB] Database error: {e}")
        raise
    finally:
        if conn:
            db_pool.putconn(conn)

# FastAPI app & middleware

app = FastAPI(
    title="Rise Square Payment Service",
    description="Production-grade Square payment processing service",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc"
)

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


# Enhanced Models with Validation

class CheckoutRequest(BaseModel):
    membership_plan_id: str = Field(..., min_length=1, max_length=100, description="Valid membership plan UUID")
    discount_code: Optional[str] = Field(None, max_length=50, description="Optional discount code")
    
    @validator('membership_plan_id')
    def validate_membership_plan_id(cls, v):
        try:
            uuid.UUID(v)
            return v
        except ValueError:
            raise ValueError('membership_plan_id must be a valid UUID')

class ProgramCheckoutRequest(BaseModel):
    program_id: str = Field(..., min_length=1, max_length=100, description="Valid program UUID")
    
    @validator('program_id')
    def validate_program_id(cls, v):
        try:
            uuid.UUID(v)
            return v
        except ValueError:
            raise ValueError('program_id must be a valid UUID')

class EventCheckoutRequest(BaseModel):
    event_id: str = Field(..., min_length=1, max_length=100, description="Valid event UUID")
    
    @validator('event_id')
    def validate_event_id(cls, v):
        try:
            uuid.UUID(v)
            return v
        except ValueError:
            raise ValueError('event_id must be a valid UUID')

class ErrorResponse(BaseModel):
    error: str
    detail: Optional[str] = None
    code: Optional[str] = None

class CheckoutResponse(BaseModel):
    checkout_url: Optional[str] = Field(None, description="Square checkout URL or null if free")
    status: str = Field(default="success", description="Transaction status")


# Production DB helpers

def _warn_if_db_superuser():
    """Best-effort warning if connected as a superuser (least-privilege reminder)."""
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute("SELECT current_user")
            user = cur.fetchone()[0]
            if user == "postgres":
                logger.warning("[SECURITY] Connected as DB superuser 'postgres'. "
                               "Use a least-privileged DB user in production.")
            else:
                logger.debug("[SECURITY] Connected DB user=%s", user)
    except Exception as e:
        logger.debug("[SECURITY] DB user check skipped: %s", e)

# Retry decorator for database operations
def db_retry(max_retries: int = 3, delay: float = 0.1):
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            last_exception = None
            for attempt in range(max_retries):
                try:
                    return await func(*args, **kwargs)
                except (psycopg2.OperationalError, psycopg2.InterfaceError) as e:
                    last_exception = e
                    if attempt < max_retries - 1:
                        await asyncio.sleep(delay * (2 ** attempt))  # exponential backoff
                        logger.warning(f"[DB] Retry {attempt + 1}/{max_retries} for {func.__name__}: {e}")
                    continue
                except Exception as e:
                    # Don't retry non-connection errors
                    raise e
            raise last_exception
        return wrapper
    return decorator


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

# Production-grade membership checkout endpoint

@app.post("/checkout/membership", response_model=CheckoutResponse)
@limiter.limit("10/minute")
async def checkout_membership(
    request: Request,
    req: CheckoutRequest,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer()),
) -> CheckoutResponse:
    customer_id = get_user_id(creds)
    logger.info(f"[CHECKOUT] Membership checkout user={customer_id} plan={req.membership_plan_id}")

    try:
        with get_db_conn() as conn:
            with conn.cursor() as cur:
                # Check for existing active membership (prevent duplicates)
                cur.execute(
                    """SELECT 1 FROM users.customer_membership_plans 
                       WHERE customer_id = %s AND membership_plan_id = %s 
                       AND status = 'active' AND (renewal_date IS NULL OR renewal_date > NOW())""",
                    (customer_id, req.membership_plan_id)
                )
                if cur.fetchone():
                    raise HTTPException(
                        status_code=409,
                        detail="Active membership already exists for this plan"
                    )

                # Get membership plan details  
                cur.execute(
                    """SELECT stripe_price_id, stripe_joining_fee_id, amt_periods 
                       FROM membership.membership_plans WHERE id = %s""",
                    (req.membership_plan_id,)
                )
                row = cur.fetchone()
                if not row:
                    raise HTTPException(status_code=404, detail="Membership plan not found")
                
                catalog_id, joining_fee_id, amt_periods = row
                if not catalog_id:
                    raise HTTPException(
                        status_code=400,
                        detail="Membership plan missing Square catalog configuration"
                    )

                # Calculate renewal date safely
                renewal_date = None
                if amt_periods is not None and amt_periods > 0:
                    cur.execute("SELECT NOW() + INTERVAL %s", (f"{amt_periods} months",))
                    renewal_date = cur.fetchone()[0]

                # Upsert membership record (handles race conditions)
                cur.execute(
                    """
                    INSERT INTO users.customer_membership_plans 
                    (customer_id, membership_plan_id, renewal_date, status) 
                    VALUES (%s, %s, %s, 'inactive')
                    ON CONFLICT (customer_id, membership_plan_id) 
                    DO UPDATE SET 
                        renewal_date = EXCLUDED.renewal_date,
                        status = 'inactive'
                    WHERE customer_membership_plans.status != 'active'
                    """,
                    (customer_id, req.membership_plan_id, renewal_date)
                )
                
                conn.commit()
                logger.info(f"[CHECKOUT] Membership record created/updated user={customer_id}")

        # Prepare Square checkout
        metadata = {"user_id": customer_id, "membership_plan_id": req.membership_plan_id}
        if amt_periods is not None:
            metadata["amt_periods"] = str(amt_periods)

        catalog_ids = [catalog_id] + ([joining_fee_id] if joining_fee_id else [])
        payload = generate_checkout_payload(os.getenv("SQUARE_LOCATION_ID"), catalog_ids, metadata)

        # Square API call with proper timeout and error handling
        async with httpx.AsyncClient(timeout=30.0) as client:
            resp = await client.post(
                f"{SQUARE_BASE_URL}/v2/online-checkout/payment-links",
                headers=HEADERS,
                json=payload,
            )
            
            if resp.status_code >= 400:
                error_detail = resp.text
                logger.error(f"[SQUARE] Checkout failed {resp.status_code}: {error_detail}")
                raise HTTPException(status_code=502, detail="Payment provider error")
            
            checkout_url = resp.json().get("payment_link", {}).get("url")
            if not checkout_url:
                logger.error("[SQUARE] Missing checkout URL in response")
                raise HTTPException(status_code=502, detail="Payment provider error")
            
            logger.info(f"[CHECKOUT] Success user={customer_id}")
            return CheckoutResponse(checkout_url=checkout_url)
            
    except HTTPException:
        raise
    except httpx.RequestError as e:
        logger.exception(f"[SQUARE] Network error: {e}")
        raise HTTPException(status_code=502, detail="Payment service temporarily unavailable")
    except Exception as e:
        logger.exception(f"[CHECKOUT] Unexpected error: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")


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
                                cur.execute("SELECT NOW() + INTERVAL %s", (f"{periods} months",))
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


# Health check and startup events

@app.on_event("startup")
async def startup_event():
    """Initialize services on startup."""
    init_db_pool()
    _warn_if_db_superuser()
    logger.info("[STARTUP] Rise Square Payment Service ready")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown."""
    if db_pool:
        db_pool.closeall()
    logger.info("[SHUTDOWN] Rise Square Payment Service stopped")


@app.get("/health", include_in_schema=False)
async def health_check():
    """Health check endpoint for load balancers."""
    try:
        with get_db_conn() as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT 1")
                cur.fetchone()
        return {"status": "healthy", "service": "rise-square-payment", "timestamp": datetime.utcnow().isoformat()}
    except Exception as e:
        logger.error(f"[HEALTH] Health check failed: {e}")
        raise HTTPException(status_code=503, detail="Service unavailable")

@app.get("/", include_in_schema=False)
async def root():
    """Root endpoint."""
    return {"service": "Rise Square Payment Service", "status": "running", "docs": "/docs"}

# Server startup
if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8080"))
    uvicorn.run(
        app, 
        host="0.0.0.0", 
        port=port,
        log_level="info",
        access_log=True
    )
