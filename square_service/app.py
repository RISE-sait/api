from inspect import signature
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

# Security / rate-limit imports
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

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 
#  FastAPI app & middleware
app = FastAPI()

#  CORS: only your frontend
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
    "Content-Type": "application/json"
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


#  DB helper

def get_db_conn():
    return psycopg2.connect(os.getenv("DATABASE_URL"))

# JWT decode helper (with explicit expiry handling)

def get_user_id(creds: HTTPAuthorizationCredentials) -> str:
    token = creds.credentials
    try:
        payload = jwt.decode(
            token,
            JWT_SECRET,
            algorithms=[JWT_ALGORITHM],
            options={"require": ["exp"]},  
        )
        return payload.get("user_id") or payload.get("sub")
    except jwt.ExpiredSignatureError:
        logger.warning("JWT expired")
        raise HTTPException(status_code=401, detail="Token expired")
    except jwt.InvalidTokenError as e:
        logger.error(f"JWT invalid: {e}")
        raise HTTPException(status_code=401, detail="Invalid token")


# Signature verification (use canonical URL)

def verify_signature(signature_header: str, body: bytes) -> bool:
    # use the exact URL Square registered (must match protocol, host, port, path)
    msg = WEBHOOK_URL.encode("utf-8") + body
    computed = base64.b64encode(
        hmac.new(
            WEBHOOK_SIGNATURE_KEY.encode("utf-8"),
            msg,
            hashlib.sha256
        ).digest()
    ).decode("utf-8")
    logger.debug(f"[WEBHOOK] Signature header: {signature_header}")
    logger.debug(f"[WEBHOOK] Computed signature: {computed}")
    return hmac.compare_digest(signature_header, computed)


# Payload generator

def generate_checkout_payload(location_id, catalog_ids, metadata):
    payload = {
        "idempotency_key": os.urandom(16).hex(),
        "order": {
            "location_id": location_id,
            "line_items": [{"catalog_object_id": cid, "quantity": "1"} for cid in catalog_ids],
            "metadata": metadata,
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

    async with httpx.AsyncClient(timeout=10.0) as client:
        resp = await client.post(
            f"{SQUARE_BASE_URL}/v2/online-checkout/payment-links",
            headers=HEADERS,
            json=payload
        )
        if resp.status_code >= 400:
            raise HTTPException(status_code=500, detail=resp.text)
        return {"checkout_url": resp.json()["payment_link"]["url"]}


# Event checkout endpoint

@app.post("/checkout/event")
@limiter.limit("10/minute")
async def checkout_event(
    request: Request,
    req: EventCheckoutRequest,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer()),
):
    customer_id = get_user_id(creds)
    logger.info(
        f"Received event checkout for user: {customer_id}, event: {req.event_id}"
    )

    with closing(get_db_conn()) as conn, conn.cursor() as cur:
        # Get program for the event
        cur.execute("SELECT program_id FROM events.events WHERE id = %s", (req.event_id,))
        row = cur.fetchone()
        if not row:
            raise HTTPException(status_code=404, detail="event not found")
        program_id = row[0]

        # Reserve a seat for the user
        cur.execute(
            """
            INSERT INTO events.customer_enrollment
                (customer_id, event_id, payment_expired_at, payment_status)
            VALUES (%s, %s, CURRENT_TIMESTAMP + interval '10 minute', 'pending')
            ON CONFLICT (customer_id, event_id) DO UPDATE
                SET payment_expired_at = EXCLUDED.payment_expired_at,
                    payment_status = EXCLUDED.payment_status
            WHERE events.customer_enrollment.payment_status != 'paid'
            """,
            (customer_id, req.event_id),
        )

        # Determine price for customer's membership
        cur.execute(
            """
            SELECT f.stripe_price_id, p.pay_per_event
            FROM program.fees f
            JOIN program.programs p ON p.id = f.program_id
            WHERE f.membership_id = (
                SELECT mp.membership_id
                FROM users.customer_membership_plans cmp
                         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
                WHERE customer_id = %s
                  AND status = 'active'
                ORDER BY cmp.start_date DESC
                LIMIT 1
            )
              AND f.program_id = %s
            """,
            (customer_id, program_id),
        )
        row = cur.fetchone()
        if not row:
            raise HTTPException(status_code=404, detail="price not found")

        price_id, pay_per_event = row

        # Free enrollment if membership covers it
        if not pay_per_event or not price_id:
            cur.execute(
                "UPDATE events.customer_enrollment SET payment_status='paid' WHERE customer_id=%s AND event_id=%s",
                (customer_id, req.event_id),
            )
            conn.commit()
            return {"checkout_url": None}

        conn.commit()

    metadata = {"user_id": customer_id, "event_id": req.event_id}
    payload = generate_checkout_payload(os.getenv("SQUARE_LOCATION_ID"), [price_id], metadata)

    async with httpx.AsyncClient(timeout=10.0) as client:
        resp = await client.post(
            f"{SQUARE_BASE_URL}/v2/online-checkout/payment-links",
            headers=HEADERS,
            json=payload,
        )
        if resp.status_code >= 400:
            raise HTTPException(status_code=500, detail=resp.text)
        return {"checkout_url": resp.json()["payment_link"]["url"]}


# Webhook endpoint
@app.post("/webhook")
@limiter.limit("60/minute")
async def handle_webhook(request: Request):
    signature = request.headers.get("x-square-hmacsha256-signature", "")
    body = await request.body()

    if not verify_signature(signature, body):
        logger.warning("Invalid webhook signature")
        raise HTTPException(status_code=400, detail="Invalid signature")

    try:
        event = await request.json()
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid JSON")

    if event.get("type") == "payment.updated":
        payment = event.get("data", {}).get("object", {}).get("payment")
        if payment and payment.get("status") == "COMPLETED":
            order_id = payment.get("order_id")
            async with httpx.AsyncClient(timeout=10.0) as client:
                resp = await client.get(f"{SQUARE_BASE_URL}/v2/orders/{order_id}", headers=HEADERS)
                resp.raise_for_status()
                metadata = resp.json()["order"].get("metadata", {})