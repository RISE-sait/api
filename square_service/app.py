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

# —————————————————————————————
# 1) Load + enforce required env vars
# —————————————————————————————
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

# —————————————————————————————
# 2) FastAPI app & middleware
# —————————————————————————————
app = FastAPI()

# 2a) CORS: only your frontend
app.add_middleware(
    CORSMiddleware,
    allow_origins=[os.getenv("FRONTEND_URL")],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 2b) Rate limiting
limiter = Limiter(key_func=get_remote_address)
app.state.limiter = limiter
app.add_middleware(SlowAPIMiddleware)

# 2c) Security headers
@app.middleware("http")
async def security_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["Strict-Transport-Security"] = "max-age=63072000; includeSubDomains"
    response.headers["X-Frame-Options"] = "DENY"
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["Referrer-Policy"] = "no-referrer"
    response.headers["Permissions-Policy"] = "geolocation=()"
    return response

# —————————————————————————————
# 3) Square API configuration
# —————————————————————————————
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

# —————————————————————————————
# 4) JWT & Webhook keys
# —————————————————————————————
JWT_SECRET = os.getenv("JWT_SECRET")
JWT_ALGORITHM = "HS256"
WEBHOOK_SIGNATURE_KEY = os.getenv("SQUARE_WEBHOOK_SIGNATURE_KEY")
WEBHOOK_URL = os.getenv("WEBHOOK_URL")

# —————————————————————————————
# 5) Models
# —————————————————————————————
class CheckoutRequest(BaseModel):
    membership_plan_id: str
    discount_code: Optional[str] = None

class ProgramCheckoutRequest(BaseModel):
    program_id: str

class EventCheckoutRequest(BaseModel):
    event_id: str

# —————————————————————————————
# 6) DB helper
# —————————————————————————————
def get_db_conn():
    return psycopg2.connect(os.getenv("DATABASE_URL"))

# —————————————————————————————
# 7) JWT decode helper (with explicit expiry handling)
# —————————————————————————————
def get_user_id(creds: HTTPAuthorizationCredentials) -> str:
    token = creds.credentials
    try:
        payload = jwt.decode(
            token,
            JWT_SECRET,
            algorithms=[JWT_ALGORITHM],
            options={"require": ["exp"]},  # ensure exp claim
        )
        return payload.get("user_id") or payload.get("sub")
    except jwt.ExpiredSignatureError:
        logger.warning("JWT expired")
        raise HTTPException(status_code=401, detail="Token expired")
    except jwt.InvalidTokenError as e:
        logger.error(f"JWT invalid: {e}")
        raise HTTPException(status_code=401, detail="Invalid token")

# —————————————————————————————
# 8) Signature verification (use canonical URL)
# —————————————————————————————
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

# —————————————————————————————
# 9) Payload generator
# —————————————————————————————
def generate_checkout_payload(location_id, catalog_ids, metadata):
    return {
        "idempotency_key": os.urandom(16).hex(),
        "order": {
            "location_id": location_id,
            "line_items": [{"catalog_object_id": cid, "quantity": "1"} for cid in catalog_ids],
            "metadata": metadata,
        },
        "ask_for_shipping_address": False,
    }

# —————————————————————————————
# 10) Membership checkout endpoint
# —————————————————————————————
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

# —————————————————————————————
# 11) Webhook endpoint
# —————————————————————————————
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


            customer_id = metadata.get("user_id")
            if not customer_id or customer_id.lower() == "none":
                logger.warning(f"Webhook metadata missing user_id, skipping DB: {metadata}")
                return {"status": "ok"}

            plan_id = metadata.get("membership_plan_id")
            program_id = metadata.get("program_id")
            event_id = metadata.get("event_id")

            with closing(get_db_conn()) as conn, conn.cursor() as cur:
                if customer_id and plan_id:
                    cur.execute("SELECT 1 FROM users.users WHERE id = %s", (customer_id,))
                    if cur.fetchone():
                        renewal_date = None
                        amt = metadata.get("amt_periods")
                        if amt and amt.isdigit():
                            renewal_date = datetime.utcnow() + relativedelta(months=int(amt))
                        cur.execute(
                            "INSERT INTO users.customer_membership_plans "
                            "(customer_id, membership_plan_id, renewal_date) VALUES (%s,%s,%s)",
                            (customer_id, plan_id, renewal_date),
                        )
                    else:
                        logger.warning(f"Unknown customer_id={customer_id}; skipping membership insert")
                elif customer_id and program_id:
                    cur.execute(
                        "UPDATE program.customer_enrollment SET payment_status='paid' "
                        "WHERE customer_id=%s AND program_id=%s",
                        (customer_id, program_id),
                    )
                elif customer_id and event_id:
                    cur.execute(
                        "UPDATE events.customer_enrollment SET payment_status='paid' "
                        "WHERE customer_id=%s AND event_id=%s",
                        (customer_id, event_id),
                    )
                conn.commit()

    return {"status": "ok"}
