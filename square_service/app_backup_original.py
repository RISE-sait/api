from fastapi import FastAPI, HTTPException, Request, Depends, status, Response
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
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type
from pybreaker import CircuitBreaker
from dataclasses import dataclass
from enum import Enum
from prometheus_client import Counter, Histogram, Gauge, generate_latest
import json
from contextlib import asynccontextmanager
import threading
import time

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


# Enterprise Square API Client
class SquareAPIClient:
    def __init__(self):
        self.base_url = (
            "https://connect.squareupsandbox.com"
            if os.getenv("SQUARE_ENV") != "production"
            else "https://connect.squareup.com"
        )
        self.headers = {
            "Square-Version": os.getenv("SQUARE_VERSION", "2025-07-16"),
            "Authorization": f"Bearer {os.getenv('SQUARE_ACCESS_TOKEN')}",
            "Content-Type": "application/json",
        }
        self.timeout = httpx.Timeout(30.0, connect=10.0)
    
    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=2, max=8),
        retry=retry_if_exception_type((httpx.RequestError, httpx.TimeoutException)),
        reraise=True
    )
    async def make_request(self, method: str, endpoint: str, **kwargs) -> httpx.Response:
        """Make authenticated request to Square API with retry logic and circuit breaker."""
        url = f"{self.base_url}{endpoint}"
        
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.request(
                    method=method,
                    url=url,
                    headers=self.headers,
                    **kwargs
                )
                
                # Log API errors for monitoring
                if response.status_code >= 400:
                    square_api_errors.labels(
                        endpoint=endpoint.split('/')[-1],
                        status_code=response.status_code
                    ).inc()
                    
                if response.status_code >= 400:
                    error_body = response.text
                    logger.error(f"[SQUARE_CLIENT] {method} {endpoint} failed {response.status_code}: {error_body}")
                    logger.error(f"[SQUARE_CLIENT] Request payload: {kwargs.get('json', 'No JSON payload')}")
                
                response.raise_for_status()
                return response
                
            except httpx.RequestError as e:
                logger.error(f"[SQUARE_CLIENT] Request failed {method} {endpoint}: {e}")
                square_api_errors.labels(endpoint=endpoint.split('/')[-1], status_code=0).inc()
                raise
    
    async def create_customer(self, user_id: str, email: str, **customer_data) -> dict:
        """Create or update Square customer."""
        payload = {
            "idempotency_key": f"customer-{user_id}-{os.urandom(8).hex()}",
            "given_name": customer_data.get("first_name", ""),
            "family_name": customer_data.get("last_name", ""),
            "email_address": email,
            "reference_id": user_id
        }
        
        response = await self.make_request("POST", "/v2/customers", json=payload)
        return response.json()
    
    async def get_subscription_plan(self, plan_variation_id: str) -> dict:
        """Get subscription plan details."""
        response = await self.make_request("GET", f"/v2/catalog/object/{plan_variation_id}")
        return response.json()
    
    async def create_order_template(self, location_id: str, variation_id: str = None) -> dict:
        """Create Square draft order template for subscription."""
        order_data = {
            "location_id": location_id,
            "state": "DRAFT"
        }
        
        # Only add line items if we have a valid variation_id
        if variation_id:
            order_data["line_items"] = [
                {
                    "catalog_object_id": variation_id,
                    "quantity": "1"
                }
            ]
        else:
            # Create empty order for relative pricing template
            order_data["line_items"] = [
                {
                    "name": "Subscription Item",
                    "quantity": "1",
                    "base_price_money": {
                        "amount": 100,  # Minimum 1 CAD (100 cents) - will be overridden by subscription plan
                        "currency": "CAD"  # Fixed: Your Square account uses CAD
                    }
                }
            ]
        
        template_data = {
            "idempotency_key": f"order-template-{os.urandom(8).hex()}",
            "order": order_data
        }
        
        response = await self.make_request("POST", "/v2/orders", json=template_data)
        return response.json()
    
    async def create_subscription(self, subscription_data: dict) -> dict:
        """Create Square subscription."""
        response = await self.make_request("POST", "/v2/subscriptions", json=subscription_data)
        return response.json()
    
    async def update_subscription(self, subscription_id: str, update_data: dict) -> dict:
        """Update Square subscription."""
        response = await self.make_request(
            "PUT", 
            f"/v2/subscriptions/{subscription_id}", 
            json=update_data
        )
        return response.json()
    
    async def pause_subscription(self, subscription_id: str, pause_data: dict) -> dict:
        """Pause Square subscription."""
        response = await self.make_request(
            "POST",
            f"/v2/subscriptions/{subscription_id}/pause",
            json=pause_data
        )
        return response.json()
    
    async def resume_subscription(self, subscription_id: str, resume_data: dict) -> dict:
        """Resume Square subscription."""
        response = await self.make_request(
            "POST",
            f"/v2/subscriptions/{subscription_id}/resume", 
            json=resume_data
        )
        return response.json()
    
    async def cancel_subscription(self, subscription_id: str) -> dict:
        """Cancel Square subscription."""
        response = await self.make_request("POST", f"/v2/subscriptions/{subscription_id}/cancel")
        return response.json()
    
    async def get_subscription(self, subscription_id: str) -> dict:
        """Retrieve Square subscription details."""
        response = await self.make_request("GET", f"/v2/subscriptions/{subscription_id}")
        return response.json()
    
    async def list_subscription_events(self, subscription_id: str) -> dict:
        """List subscription events for debugging."""
        response = await self.make_request("GET", f"/v2/subscriptions/{subscription_id}/events")
        return response.json()
    
    async def create_checkout_session(self, checkout_data: dict) -> dict:
        """Create Square Checkout session for subscription signup."""
        response = await self.make_request("POST", "/v2/online-checkout/payment-links", json=checkout_data)
        return response.json()
    
    async def create_customer(self, customer_data: dict) -> dict:
        """Create Square customer."""
        response = await self.make_request("POST", "/v2/customers", json=customer_data)
        return response.json()
    
    async def create_order(self, order_data: dict) -> dict:
        """Create Square order (used for order templates)."""
        response = await self.make_request("POST", "/v2/orders", json=order_data)
        return response.json()

# Initialize Square API client
square_client = SquareAPIClient()

# Square API configuration

SQUARE_BASE_URL = square_client.base_url
HEADERS = square_client.headers

# JWT & Webhook keys
JWT_SECRET = os.getenv("JWT_SECRET")
JWT_ALGORITHM = os.getenv("JWT_ALGORITHM")
WEBHOOK_SIGNATURE_KEY = os.getenv("SQUARE_WEBHOOK_SIGNATURE_KEY")
WEBHOOK_URL = os.getenv("WEBHOOK_URL")


# Metrics
subscription_requests = Counter('square_subscription_requests_total', 'Total subscription requests', ['operation', 'status'])
subscription_duration = Histogram('square_subscription_request_duration_seconds', 'Subscription request duration', ['operation'])
square_api_errors = Counter('square_api_errors_total', 'Total Square API errors', ['endpoint', 'status_code'])
active_subscriptions = Gauge('active_subscriptions_total', 'Total active subscriptions')
webhook_events = Counter('webhook_events_total', 'Total webhook events received', ['event_type'])

# Circuit Breaker Configuration
square_circuit_breaker = CircuitBreaker(
    fail_max=5,
    reset_timeout=30,
    exclude=[httpx.HTTPStatusError]
)

# Subscription Models
class SubscriptionStatus(str, Enum):
    ACTIVE = "ACTIVE"
    PAUSED = "PAUSED" 
    CANCELED = "CANCELED"
    PENDING = "PENDING"
    DELINQUENT = "DELINQUENT"
    EXPIRED = "EXPIRED"

class SubscriptionAction(str, Enum):
    PAUSE = "pause"
    RESUME = "resume" 
    CANCEL = "cancel"
    UPDATE_CARD = "update_card"
    UPGRADE = "upgrade"
    DOWNGRADE = "downgrade"

@dataclass
class SubscriptionPlan:
    plan_id: str
    variation_id: str
    name: str
    price_amount: int
    currency: str = "USD"
    cadence: str = "MONTHLY"

class SubscriptionRequest(BaseModel):
    membership_plan_id: str = Field(..., description="UUID of the membership plan")
    card_id: Optional[str] = Field(None, description="Card ID for automatic billing")
    start_date: Optional[str] = Field(None, description="ISO date string for subscription start")
    timezone: str = Field(default="UTC", description="Timezone for the subscription")
    
    @validator('membership_plan_id')
    def validate_membership_plan_id(cls, v):
        try:
            uuid.UUID(v)
            return v
        except ValueError:
            raise ValueError('membership_plan_id must be a valid UUID')

class SubscriptionManagementRequest(BaseModel):
    action: SubscriptionAction = Field(..., description="Action to perform on subscription")
    card_id: Optional[str] = Field(None, description="New card ID for update_card action")
    new_plan_variation_id: Optional[str] = Field(None, description="For upgrade/downgrade actions")
    pause_cycle_duration: Optional[int] = Field(None, description="Number of billing cycles to pause")
    cancel_reason: Optional[str] = Field(None, max_length=500, description="Reason for cancellation")

class SubscriptionResponse(BaseModel):
    subscription_id: str
    status: str
    message: str
    subscription_url: Optional[str] = None
    checkout_url: Optional[str] = None

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

# Enterprise Subscription Management Functions

async def ensure_square_customer(user_id: str, user_email: str) -> str:
    """Ensure user has Square customer ID, create if needed."""
    with get_db_conn() as conn, conn.cursor() as cur:
        # Check if user already has Square customer ID
        cur.execute(
            "SELECT square_customer_id FROM users.users WHERE id = %s",
            (user_id,)
        )
        result = cur.fetchone()
        
        if result and result[0]:
            return result[0]
        
        # Create Square customer
        try:
            customer_response = await square_client.create_customer(
                user_id=user_id,
                email=user_email
            )
            
            square_customer_id = customer_response.get("customer", {}).get("id")
            if not square_customer_id:
                raise ValueError("Failed to create Square customer")
            
            # Store Square customer ID
            cur.execute(
                "UPDATE users.users SET square_customer_id = %s WHERE id = %s",
                (square_customer_id, user_id)
            )
            conn.commit()
            
            logger.info(f"[SUBSCRIPTION] Created Square customer {square_customer_id} for user {user_id}")
            return square_customer_id
            
        except Exception as e:
            logger.exception(f"[SUBSCRIPTION] Failed to create Square customer for user {user_id}: {e}")
            raise HTTPException(status_code=502, detail="Failed to create customer profile")

async def get_subscription_plan_details(plan_id: str) -> tuple:
    """Get Square subscription plan variation details from database."""
    with get_db_conn() as conn, conn.cursor() as cur:
        cur.execute(
            """
            SELECT stripe_price_id, stripe_joining_fee_id, amt_periods, name
            FROM membership.membership_plans 
            WHERE id = %s
            """,
            (plan_id,)
        )
        result = cur.fetchone()
        if not result:
            raise HTTPException(status_code=404, detail="Membership plan not found")
        
        variation_id, joining_fee_id, amt_periods, name = result
        if not variation_id:
            raise HTTPException(
                status_code=400, 
                detail="Membership plan missing Square subscription configuration"
            )
        
        return variation_id, joining_fee_id, amt_periods, name

def calculate_next_billing_date(cadence: str, anchor_day: int = None, from_date: str = None) -> str:
    """Calculate next billing date for any Square subscription cadence."""
    from datetime import datetime, timedelta
    import calendar
    
    if from_date:
        current_date = datetime.strptime(from_date, "%Y-%m-%d")
    else:
        current_date = datetime.utcnow()
    
    # Handle different billing cadences
    cadence_upper = cadence.upper() if cadence else "MONTHLY"
    
    if cadence_upper == "DAILY":
        next_billing = current_date + timedelta(days=1)
        
    elif cadence_upper == "WEEKLY":
        next_billing = current_date + timedelta(days=7)
        
    elif cadence_upper == "EVERY_TWO_WEEKS":
        next_billing = current_date + timedelta(days=14)
        
    elif cadence_upper == "EVERY_THREE_WEEKS":
        next_billing = current_date + timedelta(days=21)
        
    elif cadence_upper == "EVERY_FOUR_WEEKS":
        next_billing = current_date + timedelta(days=28)
        
    elif cadence_upper == "EVERY_TWO_MONTHS":
        # Add 2 months
        if current_date.month <= 10:
            next_billing = current_date.replace(month=current_date.month + 2)
        else:
            next_billing = current_date.replace(year=current_date.year + 1, 
                                              month=current_date.month - 10)
                                              
    elif cadence_upper == "QUARTERLY":
        # Add 3 months
        if current_date.month <= 9:
            next_billing = current_date.replace(month=current_date.month + 3)
        else:
            next_billing = current_date.replace(year=current_date.year + 1, 
                                              month=current_date.month - 9)
                                              
    elif cadence_upper == "EVERY_SIX_MONTHS":
        # Add 6 months
        if current_date.month <= 6:
            next_billing = current_date.replace(month=current_date.month + 6)
        else:
            next_billing = current_date.replace(year=current_date.year + 1, 
                                              month=current_date.month - 6)
                                              
    elif cadence_upper == "ANNUAL":
        # Add 1 year
        next_billing = current_date.replace(year=current_date.year + 1)
        
    elif cadence_upper == "EVERY_TWO_YEARS":
        # Add 2 years
        next_billing = current_date.replace(year=current_date.year + 2)
        
    else:  # MONTHLY (default)
        anchor_day = anchor_day or 1  # Default to 1st of month
        
        # If current day is before anchor day, next billing is this month
        if current_date.day < anchor_day:
            try:
                next_billing = current_date.replace(day=anchor_day)
            except ValueError:
                # Handle months with fewer days (e.g., Feb 30th doesn't exist)
                last_day = calendar.monthrange(current_date.year, current_date.month)[1]
                next_billing = current_date.replace(day=min(anchor_day, last_day))
        else:
            # Next billing is next month
            if current_date.month == 12:
                next_year = current_date.year + 1
                next_month = 1
            else:
                next_year = current_date.year
                next_month = current_date.month + 1
                
            try:
                next_billing = current_date.replace(year=next_year, month=next_month, day=anchor_day)
            except ValueError:
                # Handle months with fewer days
                last_day = calendar.monthrange(next_year, next_month)[1]
                next_billing = current_date.replace(year=next_year, month=next_month, day=min(anchor_day, last_day))
    
    return next_billing.strftime("%Y-%m-%d")

# Double Payment Prevention System

payment_cache = {}  # In-memory cache for recent payments
checkout_cache = {}  # Cache for recent checkout sessions

def generate_idempotency_key(user_id: str, plan_id: str) -> str:
    """Generate consistent idempotency key for user+plan combination."""
    return f"checkout-{user_id}-{plan_id}"

async def check_recent_payment(user_id: str, plan_id: str, amount: int) -> bool:
    """Check if user made a recent payment for the same plan."""
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            # Check for recent successful payments (last 10 minutes)
            cur.execute(
                """
                SELECT COUNT(*) FROM payment_history 
                WHERE user_id = %s AND plan_id = %s AND amount = %s 
                AND created_at > NOW() - INTERVAL '10 minutes'
                AND status = 'completed'
                """,
                (user_id, plan_id, amount)
            )
            recent_payments = cur.fetchone()[0]
            return recent_payments > 0
    except Exception as e:
        logger.error(f"[PAYMENT_CHECK] Error checking recent payments: {e}")
        return False

async def record_payment_attempt(user_id: str, plan_id: str, amount: int, checkout_id: str) -> None:
    """Record payment attempt to prevent duplicates."""
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO payment_history (user_id, plan_id, amount, checkout_id, status, created_at)
                VALUES (%s, %s, %s, %s, 'pending', NOW())
                ON CONFLICT (checkout_id) DO NOTHING
                """,
                (user_id, plan_id, amount, checkout_id)
            )
            conn.commit()
    except Exception as e:
        logger.error(f"[PAYMENT_RECORD] Error recording payment attempt: {e}")

async def update_payment_status(checkout_id: str, status: str) -> None:
    """Update payment status."""
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """
                UPDATE payment_history SET status = %s, updated_at = NOW() 
                WHERE checkout_id = %s
                """,
                (status, checkout_id)
            )
            conn.commit()
    except Exception as e:
        logger.error(f"[PAYMENT_UPDATE] Error updating payment status: {e}")

async def check_existing_active_subscription(user_id: str, plan_id: str) -> bool:
    """Check if user already has active subscription for this plan."""
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """
                SELECT COUNT(*) FROM users.customer_membership_plans 
                WHERE customer_id = %s AND membership_plan_id = %s 
                AND status = 'active' AND subscription_status = 'ACTIVE'
                """,
                (user_id, plan_id)
            )
            active_subs = cur.fetchone()[0]
            return active_subs > 0
    except Exception as e:
        logger.error(f"[SUB_CHECK] Error checking active subscription: {e}")
        return False

def is_duplicate_checkout_request(user_id: str, plan_id: str) -> bool:
    """Check if this is a duplicate checkout request within time window."""
    cache_key = f"{user_id}-{plan_id}"
    current_time = time.time()
    
    if cache_key in checkout_cache:
        last_request_time = checkout_cache[cache_key]
        # Prevent duplicate requests within 30 seconds
        if current_time - last_request_time < 30:
            return True
    
    checkout_cache[cache_key] = current_time
    return False

# Comprehensive Sync System for 100% Reliability

sync_stats = {
    "last_full_sync": None,
    "discrepancies_found": 0,
    "discrepancies_fixed": 0,
    "failed_syncs": 0
}

async def perform_full_subscription_sync():
    """Complete sync between Square and database - runs periodically."""
    global sync_stats
    
    try:
        logger.info("[FULL_SYNC] Starting comprehensive subscription sync")
        sync_stats["last_full_sync"] = datetime.utcnow().isoformat()
        discrepancies_found = 0
        discrepancies_fixed = 0
        
        with get_db_conn() as conn, conn.cursor() as cur:
            # Get all active subscriptions from database
            cur.execute(
                """
                SELECT square_subscription_id, customer_id, membership_plan_id, 
                       subscription_status, next_billing_date, status
                FROM users.customer_membership_plans 
                WHERE square_subscription_id IS NOT NULL
                """
            )
            db_subscriptions = cur.fetchall()
            
            for db_sub in db_subscriptions:
                square_sub_id, user_id, plan_id, db_status, db_billing, db_active = db_sub
                
                try:
                    # Get current status from Square
                    square_response = await square_client.get_subscription(square_sub_id)
                    square_sub = square_response.get("subscription", {})
                    
                    if not square_sub:
                        logger.warning(f"[FULL_SYNC] Subscription {square_sub_id} not found in Square")
                        continue
                    
                    square_status = square_sub.get("status")
                    square_billing_anchor = square_sub.get("monthly_billing_anchor_date")
                    
                    # Check for status discrepancies
                    expected_db_active = "active" if square_status == "ACTIVE" else "inactive"
                    
                    if db_status != square_status or db_active != expected_db_active:
                        discrepancies_found += 1
                        logger.warning(
                            f"[FULL_SYNC] Status discrepancy for {square_sub_id}: "
                            f"DB=({db_status},{db_active}) vs Square={square_status}"
                        )
                        
                        # Fix status discrepancy
                        cur.execute(
                            """
                            UPDATE users.customer_membership_plans 
                            SET subscription_status = %s,
                                status = CASE 
                                    WHEN %s = 'ACTIVE' THEN 'active'::membership.membership_status
                                    ELSE 'inactive'::membership.membership_status
                                END
                            WHERE square_subscription_id = %s
                            """,
                            (square_status, square_status, square_sub_id)
                        )
                        discrepancies_fixed += 1
                        logger.info(f"[FULL_SYNC] Fixed status for {square_sub_id}")
                    
                    # Check billing date sync
                    if square_status == "ACTIVE":
                        try:
                            # Get plan cadence and calculate expected next billing
                            plan_variation_id = square_sub.get("plan_variation_id")
                            cadence = "MONTHLY"
                            
                            if plan_variation_id:
                                plan_response = await square_client.get_subscription_plan(plan_variation_id)
                                plan_data = plan_response.get("object", {}).get("subscription_plan_variation_data", {})
                                phases = plan_data.get("phases", [])
                                if phases:
                                    cadence = phases[0].get("cadence", "MONTHLY")
                            
                            expected_billing = calculate_next_billing_date(cadence, square_billing_anchor)
                            
                            if db_billing != expected_billing:
                                discrepancies_found += 1
                                logger.warning(
                                    f"[FULL_SYNC] Billing date discrepancy for {square_sub_id}: "
                                    f"DB={db_billing} vs Expected={expected_billing}"
                                )
                                
                                # Fix billing date
                                cur.execute(
                                    """
                                    UPDATE users.customer_membership_plans 
                                    SET next_billing_date = %s 
                                    WHERE square_subscription_id = %s
                                    """,
                                    (expected_billing, square_sub_id)
                                )
                                discrepancies_fixed += 1
                                logger.info(f"[FULL_SYNC] Fixed billing date for {square_sub_id}")
                                
                        except Exception as e:
                            logger.error(f"[FULL_SYNC] Error syncing billing date for {square_sub_id}: {e}")
                
                except Exception as e:
                    logger.error(f"[FULL_SYNC] Error syncing subscription {square_sub_id}: {e}")
                    continue
            
            conn.commit()
        
        sync_stats["discrepancies_found"] += discrepancies_found
        sync_stats["discrepancies_fixed"] += discrepancies_fixed
        
        logger.info(
            f"[FULL_SYNC] Completed: {discrepancies_found} discrepancies found, "
            f"{discrepancies_fixed} fixed"
        )
        
    except Exception as e:
        sync_stats["failed_syncs"] += 1
        logger.exception(f"[FULL_SYNC] Failed: {e}")

def run_periodic_sync():
    """Background thread for periodic sync jobs."""
    while True:
        try:
            # Run full sync every hour
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            loop.run_until_complete(perform_full_subscription_sync())
            loop.close()
            
            # Clean up old cache entries (remove entries older than 1 hour)
            cleanup_caches()
            
            time.sleep(3600)  # 1 hour
        except Exception as e:
            logger.error(f"[PERIODIC_SYNC] Error: {e}")
            time.sleep(300)  # Wait 5 minutes on error

def cleanup_caches():
    """Clean up expired cache entries."""
    global payment_cache, checkout_cache
    current_time = time.time()
    
    # Clean payment cache (keep entries for 1 hour)
    expired_payments = [k for k, v in payment_cache.items() if current_time - v > 3600]
    for key in expired_payments:
        del payment_cache[key]
    
    # Clean checkout cache (keep entries for 5 minutes)  
    expired_checkouts = [k for k, v in checkout_cache.items() if current_time - v > 300]
    for key in expired_checkouts:
        del checkout_cache[key]
    
    if expired_payments or expired_checkouts:
        logger.info(f"[CACHE_CLEANUP] Cleaned {len(expired_payments)} payment entries, {len(expired_checkouts)} checkout entries")

async def store_failed_webhook(event_data: dict, error_msg: str):
    """Store failed webhook events for later processing."""
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO webhook_failures (event_type, event_data, error_message, created_at, processed)
                VALUES (%s, %s, %s, NOW(), false)
                """,
                (event_data.get("type"), json.dumps(event_data), error_msg)
            )
            conn.commit()
            logger.info(f"[WEBHOOK_FAILURE] Stored failed webhook: {event_data.get('type')}")
    except Exception as e:
        logger.error(f"[WEBHOOK_FAILURE] Could not store failed webhook: {e}")

async def process_failed_webhooks():
    """Retry processing of failed webhook events."""
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, event_data FROM webhook_failures 
                WHERE processed = false AND created_at > NOW() - INTERVAL '24 hours'
                ORDER BY created_at ASC LIMIT 10
                """
            )
            failed_webhooks = cur.fetchall()
            
            for webhook_id, event_data in failed_webhooks:
                try:
                    event = json.loads(event_data)
                    # Retry webhook processing by calling the handler directly
                    evt_type = event.get("type")
                    
                    if evt_type in ["subscription.created", "subscription.updated"]:
                        await handle_subscription_webhook(event)
                    elif evt_type in ["invoice.created", "invoice.payment_made", "invoice.failed"]:
                        await handle_invoice_webhook(event)
                    elif evt_type == "payment.updated":
                        # Process payment webhook (simplified version)
                        logger.info(f"[WEBHOOK_RETRY] Processing payment webhook {webhook_id}")
                    
                    # Mark as processed
                    cur.execute(
                        "UPDATE webhook_failures SET processed = true WHERE id = %s",
                        (webhook_id,)
                    )
                    logger.info(f"[WEBHOOK_RETRY] Successfully processed failed webhook {webhook_id}")
                    
                except Exception as e:
                    logger.error(f"[WEBHOOK_RETRY] Failed to process webhook {webhook_id}: {e}")
            
            conn.commit()
            logger.info(f"[WEBHOOK_RETRY] Processed {len(failed_webhooks)} failed webhooks")
    
    except Exception as e:
        logger.error(f"[WEBHOOK_RETRY] Error processing failed webhooks: {e}")

async def sync_subscription_billing_date(subscription_id: str, delay_seconds: int = 30) -> None:
    """Sync billing date from Square API after a delay to allow Square to set the date."""
    try:
        # Wait for Square to finalize billing schedule
        await asyncio.sleep(delay_seconds)
        
        logger.info(f"[BILLING_SYNC] Attempting to sync billing date for subscription {subscription_id}")
        
        # Fetch detailed subscription from Square
        detailed_subscription = await square_client.get_subscription(subscription_id)
        sub_details = detailed_subscription.get("subscription", {})
        
        # Extract next billing date
        next_billing_date = (
            sub_details.get("charged_through_date") or 
            sub_details.get("next_billing_date") or
            sub_details.get("billing_anchor_date")
        )
        
        # If no explicit date, calculate from subscription plan details
        if not next_billing_date:
            # Get plan details from Square API to determine cadence
            try:
                plan_variation_id = sub_details.get("plan_variation_id")
                if plan_variation_id:
                    plan_response = await square_client.get_subscription_plan(plan_variation_id)
                    plan_data = plan_response.get("object", {}).get("subscription_plan_variation_data", {})
                    
                    # Get cadence from plan phases
                    phases = plan_data.get("phases", [])
                    cadence = None
                    if phases:
                        cadence = phases[0].get("cadence", "MONTHLY")
                    
                    # Get anchor day if monthly
                    anchor_day = sub_details.get("monthly_billing_anchor_date")
                    
                    # Calculate next billing date based on cadence
                    next_billing_date = calculate_next_billing_date(cadence, anchor_day)
                    logger.info(f"[BILLING_SYNC] Calculated billing date: cadence={cadence}, anchor={anchor_day}, next={next_billing_date}")
                    
            except Exception as e:
                logger.warning(f"[BILLING_SYNC] Could not get plan details for billing calculation: {e}")
                
                # Fallback: use monthly with anchor day if available
                anchor_day = sub_details.get("monthly_billing_anchor_date", 1)
                next_billing_date = calculate_next_billing_date("MONTHLY", anchor_day)
                logger.info(f"[BILLING_SYNC] Fallback calculation with monthly anchor {anchor_day}: {next_billing_date}")
        
        if next_billing_date:
            # Update database with billing date
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """UPDATE users.customer_membership_plans 
                       SET next_billing_date = %s 
                       WHERE square_subscription_id = %s""",
                    (next_billing_date, subscription_id)
                )
                if cur.rowcount > 0:
                    conn.commit()
                    logger.info(f"[BILLING_SYNC] Updated billing date to {next_billing_date} for subscription {subscription_id}")
                else:
                    logger.warning(f"[BILLING_SYNC] No subscription found in DB for {subscription_id}")
        else:
            logger.warning(f"[BILLING_SYNC] No billing date available yet for subscription {subscription_id}")
            
    except Exception as e:
        logger.error(f"[BILLING_SYNC] Error syncing billing date for {subscription_id}: {e}")

async def create_subscription_record(user_id: str, plan_id: str, square_subscription_id: str, 
                                   subscription_status: str, next_billing_date: str = None) -> None:
    """Create or update subscription record in database."""
    with get_db_conn() as conn, conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO users.customer_membership_plans 
            (customer_id, membership_plan_id, square_subscription_id, subscription_status, 
             next_billing_date, subscription_created_at, subscription_source, status)
            VALUES (%s, %s, %s, %s, %s, NOW(), 'subscription', 'inactive'::membership.membership_status)
            ON CONFLICT (customer_id, membership_plan_id)
            DO UPDATE SET 
                square_subscription_id = EXCLUDED.square_subscription_id,
                subscription_status = EXCLUDED.subscription_status,
                next_billing_date = EXCLUDED.next_billing_date,
                subscription_created_at = EXCLUDED.subscription_created_at,
                subscription_source = 'subscription',
                status = CASE 
                    WHEN EXCLUDED.subscription_status = 'ACTIVE' THEN 'active'::membership.membership_status
                    ELSE 'inactive'::membership.membership_status
                END
            """,
            (user_id, plan_id, square_subscription_id, subscription_status, next_billing_date)
        )
        conn.commit()
        logger.info(f"[SUBSCRIPTION] Updated database record for user {user_id}, plan {plan_id}")

# Enterprise Subscription Endpoints (Production Ready)

@app.post("/subscriptions/checkout", response_model=SubscriptionResponse)
@limiter.limit("5/minute")
async def create_subscription_checkout(
    request: Request,
    req: SubscriptionRequest,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer())
) -> SubscriptionResponse:
    """Create a Square Checkout link for subscription signup."""
    user_id = get_user_id(creds)
    
    with subscription_duration.labels(operation='checkout').time():
        subscription_requests.labels(operation='checkout', status='started').inc()
        
        try:
            logger.info(f"[SUBSCRIPTION] Creating checkout link user={user_id} plan={req.membership_plan_id}")
            
            # PREVENT DOUBLE PAYMENT #1: Check for duplicate requests
            if is_duplicate_checkout_request(user_id, req.membership_plan_id):
                logger.warning(f"[DUPLICATE_PREVENTION] Duplicate checkout request blocked user={user_id} plan={req.membership_plan_id}")
                raise HTTPException(status_code=429, detail="Please wait before making another purchase")
            
            # PREVENT DOUBLE PAYMENT #2: Check for existing active subscription
            if await check_existing_active_subscription(user_id, req.membership_plan_id):
                logger.warning(f"[DUPLICATE_PREVENTION] User already has active subscription user={user_id} plan={req.membership_plan_id}")
                raise HTTPException(status_code=409, detail="You already have an active subscription for this plan")
            
            # Get subscription plan details from database
            variation_id, joining_fee_id, amt_periods, plan_name = await get_subscription_plan_details(
                req.membership_plan_id
            )
            
            # Get user details and plan price for checkout
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute("SELECT email, first_name, last_name FROM users.users WHERE id = %s", (user_id,))
                result = cur.fetchone()
                if not result:
                    raise HTTPException(status_code=404, detail="User not found")
                user_email, first_name, last_name = result
                
                # Get membership plan price (unit_amount is already in cents)
                cur.execute(
                    "SELECT unit_amount FROM membership.membership_plans WHERE id = %s", 
                    (req.membership_plan_id,)
                )
                price_result = cur.fetchone()
                if not price_result:
                    raise HTTPException(status_code=404, detail="Membership plan price not found")
                plan_price_cents = price_result[0] or 0  # Price already in cents
            
            # PREVENT DOUBLE PAYMENT #3: Check for recent identical payments
            if await check_recent_payment(user_id, req.membership_plan_id, plan_price_cents):
                logger.warning(f"[DUPLICATE_PREVENTION] Recent payment detected user={user_id} plan={req.membership_plan_id}")
                raise HTTPException(status_code=409, detail="You have already made this payment recently")
            
            # Create checkout payload for subscription with consistent idempotency
            idempotency_key = generate_idempotency_key(user_id, req.membership_plan_id)
            checkout_payload = {
                "idempotency_key": idempotency_key,
                "order": {
                    "location_id": os.getenv("SQUARE_LOCATION_ID"),
                    "line_items": [
                        {
                            "name": f"Subscription: {plan_name}",
                            "quantity": "1",
                            "base_price_money": {
                                "amount": plan_price_cents,
                                "currency": "CAD"
                            }
                        }
                    ],
                    "metadata": {
                        "user_id": user_id,
                        "membership_plan_id": req.membership_plan_id
                    }
                },
                "checkout_options": {
                    "subscription_plan_id": variation_id,
                    "accepted_payment_methods": {
                        "apple_pay": True,
                        "google_pay": True,
                        "cash_app_pay": False,
                        "afterpay_clearpay": False
                    },
                    "allow_tipping": False,
                    "custom_fields": [
                        {
                            "title": "Membership Plan",
                            "value": plan_name
                        },
                        {
                            "title": "User ID", 
                            "value": user_id
                        }
                    ],
                    "redirect_url": f"{os.getenv('FRONTEND_URL', 'http://localhost:3000')}/membership/success",
                    "merchant_support_email": os.getenv("MERCHANT_EMAIL", user_email)
                },
                "payment_options": {
                    "autocomplete": True
                },
                "pre_populated_data": {
                    "buyer_email": user_email,
                    "buyer_address": {
                        "first_name": first_name or "",
                        "last_name": last_name or ""
                    }
                }
            }
            
            # Create Square checkout session
            checkout_response = await square_client.create_checkout_session(checkout_payload)
            
            payment_link = checkout_response.get("payment_link", {})
            checkout_url = payment_link.get("url")
            checkout_id = payment_link.get("id", "pending")
            
            if not checkout_url:
                logger.error(f"[SUBSCRIPTION] No checkout URL returned: {checkout_response}")
                raise HTTPException(status_code=502, detail="Failed to create checkout link")
            
            # Store checkout session for later webhook processing
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO users.customer_membership_plans 
                    (customer_id, membership_plan_id, subscription_source, status)
                    VALUES (%s, %s, 'subscription', 'inactive'::membership.membership_status)
                    ON CONFLICT (customer_id, membership_plan_id)
                    DO UPDATE SET 
                        subscription_source = 'subscription',
                        updated_at = NOW()
                    """,
                    (user_id, req.membership_plan_id)
                )
                conn.commit()
            
            subscription_requests.labels(operation='checkout', status='success').inc()
            
            # PREVENT DOUBLE PAYMENT #4: Record payment attempt
            await record_payment_attempt(user_id, req.membership_plan_id, plan_price_cents, checkout_id)
            
            logger.info(f"[SUBSCRIPTION] Created checkout link user={user_id} checkout_id={checkout_id}")
            
            return SubscriptionResponse(
                subscription_id=checkout_id,
                status="CHECKOUT_PENDING", 
                message=f"Checkout link created for {plan_name}. Complete payment to activate subscription.",
                checkout_url=checkout_url
            )
            
        except HTTPException:
            subscription_requests.labels(operation='checkout', status='error').inc()
            raise
        except Exception as e:
            subscription_requests.labels(operation='checkout', status='error').inc()
            logger.exception(f"[SUBSCRIPTION] Unexpected error creating checkout link: {e}")
            raise HTTPException(status_code=500, detail="Internal server error")



@app.post("/subscriptions/{subscription_id}/manage", response_model=SubscriptionResponse)
@limiter.limit("10/minute") 
async def manage_subscription(
    request: Request,
    subscription_id: str,
    req: SubscriptionManagementRequest,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer())
) -> SubscriptionResponse:
    """Manage Square subscription (pause, resume, cancel, update)."""
    user_id = get_user_id(creds)
    
    with subscription_duration.labels(operation=req.action).time():
        subscription_requests.labels(operation=req.action, status='started').inc()
        
        try:
            logger.info(f"[SUBSCRIPTION] Managing subscription {subscription_id} action={req.action} user={user_id}")
            
            # Verify user owns this subscription
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT 1 FROM users.customer_membership_plans 
                    WHERE customer_id = %s AND square_subscription_id = %s
                    """,
                    (user_id, subscription_id)
                )
                if not cur.fetchone():
                    raise HTTPException(status_code=404, detail="Subscription not found")
            
            response_data = None
            
            if req.action == SubscriptionAction.PAUSE:
                pause_data = {
                    "pause_effective_date": datetime.utcnow().date().isoformat()
                }
                if req.pause_cycle_duration:
                    pause_data["pause_cycle_duration"] = req.pause_cycle_duration
                
                response_data = await square_client.pause_subscription(subscription_id, pause_data)
                
            elif req.action == SubscriptionAction.RESUME:
                resume_data = {
                    "resume_effective_date": datetime.utcnow().date().isoformat()
                }
                response_data = await square_client.resume_subscription(subscription_id, resume_data)
                
            elif req.action == SubscriptionAction.CANCEL:
                response_data = await square_client.cancel_subscription(subscription_id)
                active_subscriptions.dec()
                
            elif req.action == SubscriptionAction.UPDATE_CARD:
                if not req.card_id:
                    raise HTTPException(status_code=400, detail="card_id required for update_card action")
                
                update_data = {
                    "subscription": {
                        "card_id": req.card_id
                    }
                }
                response_data = await square_client.update_subscription(subscription_id, update_data)
            
            else:
                raise HTTPException(status_code=400, detail=f"Unsupported action: {req.action}")
            
            # Update database status
            subscription = response_data.get("subscription", {})
            new_status = subscription.get("status")
            if new_status:
                with get_db_conn() as conn, conn.cursor() as cur:
                    cur.execute(
                        """
                        UPDATE users.customer_membership_plans 
                        SET subscription_status = %s,
                            status = CASE 
                                WHEN %s = 'ACTIVE' THEN 'active'::membership.membership_status
                                ELSE 'inactive'::membership.membership_status
                            END
                        WHERE square_subscription_id = %s
                        """,
                        (new_status, new_status, subscription_id)
                    )
                    conn.commit()
            
            subscription_requests.labels(operation=req.action, status='success').inc()
            
            logger.info(f"[SUBSCRIPTION] Action {req.action} completed for {subscription_id}")
            
            return SubscriptionResponse(
                subscription_id=subscription_id,
                status=new_status or "UPDATED",
                message=f"Subscription {req.action} completed successfully"
            )
            
        except HTTPException:
            subscription_requests.labels(operation=req.action, status='error').inc()
            raise
        except Exception as e:
            subscription_requests.labels(operation=req.action, status='error').inc()
            logger.exception(f"[SUBSCRIPTION] Error managing subscription {subscription_id}: {e}")
            raise HTTPException(status_code=500, detail="Internal server error")

@app.post("/subscriptions/{subscription_id}/sync-billing")
@limiter.limit("10/minute")
async def manual_sync_billing_date(
    request: Request,
    subscription_id: str,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer())
) -> dict:
    """Manually trigger billing date sync for a subscription."""
    get_user_id(creds)  # Validate auth
    
    try:
        await sync_subscription_billing_date(subscription_id, delay_seconds=0)
        return {"status": "success", "message": "Billing date sync triggered"}
    except Exception as e:
        logger.error(f"[MANUAL_SYNC] Error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/subscriptions/{subscription_id}")
@limiter.limit("20/minute")
async def update_subscription_billing_date(
    request: Request,
    subscription_id: str,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer())
) -> dict:
    """Manually update subscription billing date from Square API."""
    user_id = get_user_id(creds)
    
    try:
        # Fetch detailed subscription from Square
        detailed_subscription = await square_client.get_subscription(subscription_id)
        sub_details = detailed_subscription.get("subscription", {})
        
        logger.info(f"[MANUAL] Detailed subscription response: {sub_details}")
        
        # Extract billing date
        next_billing_date = (sub_details.get("charged_through_date") or 
                           sub_details.get("next_billing_date") or
                           sub_details.get("billing_anchor_date"))
        
        subscription_created_at = sub_details.get("created_at")
        subscription_status = sub_details.get("status")
        
        # Update database
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """UPDATE users.customer_membership_plans 
                   SET next_billing_date = %s,
                       subscription_created_at = %s,
                       subscription_status = %s
                   WHERE square_subscription_id = %s AND customer_id = %s""",
                (next_billing_date, subscription_created_at, subscription_status, subscription_id, user_id)
            )
            conn.commit()
        
        return {
            "subscription_id": subscription_id,
            "next_billing_date": next_billing_date,
            "subscription_created_at": subscription_created_at,
            "status": subscription_status,
            "full_response": sub_details
        }
        
    except Exception as e:
        logger.exception(f"[MANUAL] Error updating billing date: {e}")
        raise HTTPException(status_code=500, detail=f"Error: {e}")

async def get_subscription_details(
    request: Request,
    subscription_id: str,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer())
):
    """Get subscription details from Square API."""
    user_id = get_user_id(creds)
    
    try:
        # Verify user owns this subscription
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """
                SELECT mp.name, cmp.subscription_status, cmp.next_billing_date
                FROM users.customer_membership_plans cmp
                JOIN membership.membership_plans mp ON cmp.membership_plan_id = mp.id
                WHERE cmp.customer_id = %s AND cmp.square_subscription_id = %s
                """,
                (user_id, subscription_id)
            )
            result = cur.fetchone()
            if not result:
                raise HTTPException(status_code=404, detail="Subscription not found")
            
            plan_name, db_status, next_billing = result
        
        # Get latest details from Square
        response_data = await square_client.get_subscription(subscription_id)
        square_subscription = response_data.get("subscription", {})
        
        return {
            "subscription_id": subscription_id,
            "plan_name": plan_name,
            "status": square_subscription.get("status", db_status),
            "next_billing_date": next_billing,
            "charged_through_date": square_subscription.get("charged_through_date"),
            "created_at": square_subscription.get("created_at"),
            "square_details": square_subscription
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.exception(f"[SUBSCRIPTION] Error retrieving subscription {subscription_id}: {e}")
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
    webhook_events.labels(event_type=evt_type).inc()
    logger.info(f"[WEBHOOK] received event type={evt_type}")
    
    # Enhanced webhook processing with failure recovery
    try:
        # Handle subscription events
        if evt_type in ["subscription.created", "subscription.updated"]:
            return await handle_subscription_webhook(event)
        
        # Handle invoice events for subscriptions  
        if evt_type in ["invoice.created", "invoice.payment_made", "invoice.failed"]:
            return await handle_invoice_webhook(event)
            
    except Exception as e:
        # Store failed webhook for later retry
        logger.error(f"[WEBHOOK] Error processing {evt_type}: {e}")
        await store_failed_webhook(event, str(e))
        # Don't raise - return success to prevent Square from retrying immediately
        return {"status": "error_stored_for_retry"}

    if evt_type == "payment.updated":
        payment = event.get("data", {}).get("object", {}).get("payment") or {}
        status = payment.get("status")
        payment_id = payment.get("id")
        order_id = payment.get("order_id")
        
        # PREVENT DOUBLE PAYMENT #5: Check for duplicate payment webhook processing
        webhook_cache_key = f"payment-{payment_id}"
        if webhook_cache_key in payment_cache:
            logger.warning(f"[DUPLICATE_PREVENTION] Duplicate payment webhook ignored payment_id={payment_id}")
            return {"status": "duplicate_ignored"}
        
        # Cache this payment processing (expires after 1 hour)
        payment_cache[webhook_cache_key] = time.time()
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
                with get_db_conn() as conn, conn.cursor() as cur:
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
                        
                        # Always check if we need to create a Square subscription for active memberships
                        cur.execute(
                            """SELECT square_subscription_id, subscription_source FROM users.customer_membership_plans 
                               WHERE customer_id = %s AND membership_plan_id = %s AND status = 'active'""",
                            (user_id, membership_plan_id)
                        )
                        sub_check = cur.fetchone()
                        
                        # Create subscription if: no subscription ID exists OR subscription_source is not 'subscription'
                        if sub_check and (not sub_check[0] or sub_check[1] != 'subscription'):
                            logger.info(f"[WEBHOOK] Need to create subscription - current: sub_id={sub_check[0]} source={sub_check[1]}")
                            # Need to create subscription - either no subscription ID or not a subscription source
                            try:
                                logger.info(f"[WEBHOOK] Creating Square subscription for user={user_id} plan={membership_plan_id}")
                                
                                # Get user email for Square customer creation
                                cur.execute(
                                    """SELECT email, first_name, last_name, square_customer_id FROM users.users 
                                       WHERE id = %s""",
                                    (user_id,)
                                )
                                user_result = cur.fetchone()
                                if not user_result:
                                    raise Exception("User not found")
                                    
                                user_email, first_name, last_name, existing_square_customer_id = user_result
                                
                                # Create or get Square customer
                                square_customer_id = existing_square_customer_id
                                if not square_customer_id:
                                    # Create Square customer
                                    customer_payload = {
                                        "idempotency_key": str(uuid.uuid4()),
                                        "given_name": first_name or "",
                                        "family_name": last_name or "",
                                        "email_address": user_email
                                    }
                                    customer_response = await square_client.create_customer(customer_payload)
                                    customer = customer_response.get("customer")
                                    if customer:
                                        square_customer_id = customer.get("id")
                                        # Update user with Square customer ID
                                        cur.execute(
                                            """UPDATE users.users SET square_customer_id = %s WHERE id = %s""",
                                            (square_customer_id, user_id)
                                        )
                                        logger.info(f"[WEBHOOK] Created Square customer {square_customer_id} for user {user_id}")
                                
                                # Get subscription plan details
                                cur.execute(
                                    """SELECT stripe_price_id, unit_amount FROM membership.membership_plans 
                                       WHERE id = %s""",
                                    (membership_plan_id,)
                                )
                                plan_result = cur.fetchone()
                                if plan_result and plan_result[0] and square_customer_id:
                                    variation_id, plan_price_cents = plan_result[0], plan_result[1]  # Square variation ID and price
                                    
                                    # Check if this plan needs phases for relative pricing
                                    try:
                                        square_plan_details = await square_client.get_subscription_plan(variation_id)
                                        plan_phases = square_plan_details.get("object", {}).get("subscription_plan_variation_data", {}).get("phases", [])
                                        has_relative_pricing = any(phase.get("pricing", {}).get("type") == "RELATIVE" for phase in plan_phases)
                                        logger.info(f"[WEBHOOK] Plan has relative pricing: {has_relative_pricing}")
                                    except Exception as e:
                                        logger.warning(f"[WEBHOOK] Could not get Square plan details: {e}")
                                        has_relative_pricing = False
                                    
                                    # Create subscription payload
                                    subscription_payload = {
                                        "idempotency_key": str(uuid.uuid4()),
                                        "location_id": os.getenv("SQUARE_LOCATION_ID"),
                                        "plan_variation_id": variation_id,
                                        "customer_id": square_customer_id
                                    }
                                    
                                    # For relative pricing plans, create order template
                                    if has_relative_pricing:
                                        logger.info(f"[WEBHOOK] Creating order template for relative pricing")
                                        try:
                                            # Create order template
                                            order_template_payload = {
                                                "idempotency_key": str(uuid.uuid4()),
                                                "location_id": os.getenv("SQUARE_LOCATION_ID"),
                                                "source_name": "Webhook Subscription Order Template",
                                                "order": {
                                                    "location_id": os.getenv("SQUARE_LOCATION_ID"),
                                                    "state": "DRAFT",
                                                    "line_items": [{
                                                        "name": "Subscription Template",
                                                        "quantity": "1",
                                                        "base_price_money": {
                                                            "amount": plan_price_cents,  # Use actual plan price
                                                            "currency": "CAD"
                                                        }
                                                    }]
                                                }
                                            }
                                            
                                            template_response = await square_client.create_order(order_template_payload)
                                            template_id = template_response.get("order", {}).get("id")
                                            logger.info(f"[WEBHOOK] Created order template: {template_id}")
                                            
                                            if template_id:
                                                subscription_payload["phases"] = [{
                                                    "ordinal": 0,
                                                    "order_template_id": template_id
                                                }]
                                                logger.info(f"[WEBHOOK] Added order template {template_id} to subscription")
                                            else:
                                                logger.warning("[WEBHOOK] No template ID returned, trying without template")
                                                subscription_payload["phases"] = [{"ordinal": 0}]
                                                
                                        except Exception as e:
                                            logger.error(f"[WEBHOOK] Failed to create order template: {e}")
                                            subscription_payload["phases"] = [{"ordinal": 0}]
                                    
                                    # Create subscription via Square API
                                    subscription_response = await square_client.create_subscription(subscription_payload)
                                    subscription = subscription_response.get("subscription")
                                    
                                    if subscription:
                                        square_subscription_id = subscription.get("id")
                                        subscription_status = subscription.get("status", "PENDING")
                                        
                                        # Log the full subscription response to see available fields
                                        logger.info(f"[WEBHOOK] Full subscription response: {subscription}")
                                        
                                        # Extract billing and creation dates - try multiple field names
                                        next_billing_date = (subscription.get("charged_through_date") or 
                                                           subscription.get("next_billing_date") or
                                                           subscription.get("billing_anchor_date"))
                                        subscription_created_at = subscription.get("created_at")
                                        
                                        # Fetch the complete subscription details from Square API to get billing dates
                                        try:
                                            detailed_subscription = await square_client.get_subscription(square_subscription_id)
                                            sub_details = detailed_subscription.get("subscription", {})
                                            logger.info(f"[WEBHOOK] Detailed subscription from API: {sub_details}")
                                            
                                            # Try to extract next billing date from detailed response
                                            next_billing_date = (sub_details.get("charged_through_date") or 
                                                               sub_details.get("next_billing_date") or
                                                               sub_details.get("billing_anchor_date") or
                                                               next_billing_date)  # fallback to original
                                            
                                            # Use creation date from detailed response if available
                                            subscription_created_at = sub_details.get("created_at") or subscription_created_at
                                            
                                        except Exception as e:
                                            logger.warning(f"[WEBHOOK] Could not fetch detailed subscription: {e}")
                                        
                                        # Update database with subscription details including dates
                                        cur.execute(
                                            """UPDATE users.customer_membership_plans 
                                               SET square_subscription_id = %s,
                                                   subscription_status = %s,
                                                   subscription_source = 'subscription',
                                                   next_billing_date = %s,
                                                   subscription_created_at = %s
                                               WHERE customer_id = %s AND membership_plan_id = %s""",
                                            (square_subscription_id, subscription_status, next_billing_date, subscription_created_at, user_id, membership_plan_id)
                                        )
                                        logger.info(f"[WEBHOOK] Created Square subscription {square_subscription_id} status={subscription_status} next_billing={next_billing_date} created_at={subscription_created_at}")
                                        
                                        # If no billing date yet, schedule a background sync
                                        if not next_billing_date:
                                            logger.info(f"[WEBHOOK] Scheduling billing date sync for subscription {square_subscription_id}")
                                            asyncio.create_task(sync_subscription_billing_date(square_subscription_id))
                                    else:
                                        logger.error(f"[WEBHOOK] Failed to create Square subscription: {subscription_response}")
                                        
                            except Exception as e:
                                logger.error(f"[WEBHOOK] Error creating Square subscription: {e}")
                                # Don't fail the webhook - membership is already activated
                        if rc == 0:
                            # Check if record already exists and is active
                            cur.execute(
                                """SELECT status FROM users.customer_membership_plans 
                                   WHERE customer_id = %s AND membership_plan_id = %s""",
                                (user_id, membership_plan_id)
                            )
                            existing = cur.fetchone()
                            
                            if existing and existing[0] == 'active':
                                logger.info("[WEBHOOK] membership already ACTIVE user=%s plan=%s", 
                                           user_id, membership_plan_id)
                            elif existing:
                                logger.warning("[WEBHOOK] membership exists but not active, status=%s user=%s plan=%s", 
                                              existing[0], user_id, membership_plan_id)
                            else:
                                # No record exists, insert as active (with renewal if provided)
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

# Subscription Webhook Handlers

async def handle_subscription_webhook(event: dict) -> dict:
    """Handle subscription-related webhook events."""
    evt_type = event.get("type")
    subscription_data = event.get("data", {}).get("object", {}).get("subscription", {})
    
    if not subscription_data:
        logger.warning(f"[WEBHOOK] Missing subscription data in {evt_type} event")
        return {"status": "ignored"}
    
    subscription_id = subscription_data.get("id")
    status = subscription_data.get("status")
    customer_id = subscription_data.get("customer_id")
    
    logger.info(
        f"[WEBHOOK] Processing {evt_type} subscription_id={subscription_id} status={status} customer_id={customer_id}"
    )
    
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            # Update subscription status in database
            cur.execute(
                """
                UPDATE users.customer_membership_plans 
                SET subscription_status = %s,
                    status = CASE 
                        WHEN %s = 'ACTIVE' THEN 'active'
                        WHEN %s = 'CANCELED' THEN 'inactive'
                        WHEN %s = 'PAUSED' THEN 'inactive'
                        ELSE status
                    END
                WHERE square_subscription_id = %s
                """,
                (status, status, status, status, subscription_id)
            )
            
            updated_rows = cur.rowcount
            conn.commit()
            
            if updated_rows > 0:
                logger.info(
                    f"[WEBHOOK] Updated subscription {subscription_id} to status {status}"
                )
            else:
                logger.warning(
                    f"[WEBHOOK] No subscription found in database for {subscription_id}"
                )
            
            # Update metrics
            if status == "ACTIVE":
                active_subscriptions.inc()
            elif status in ["CANCELED", "EXPIRED"]:
                active_subscriptions.dec()
        
        return {"status": "ok"}
        
    except Exception as e:
        logger.exception(f"[WEBHOOK] Error processing subscription webhook: {e}")
        raise HTTPException(status_code=500, detail="Webhook processing error")

async def handle_invoice_webhook(event: dict) -> dict:
    """Handle invoice-related webhook events for subscriptions."""
    evt_type = event.get("type")
    invoice_data = event.get("data", {}).get("object", {}).get("invoice", {})
    
    if not invoice_data:
        logger.warning(f"[WEBHOOK] Missing invoice data in {evt_type} event")
        return {"status": "ignored"}
    
    invoice_id = invoice_data.get("id")
    subscription_id = invoice_data.get("subscription_id")
    status = invoice_data.get("invoice_status")
    
    if not subscription_id:
        # Not a subscription invoice
        return {"status": "ignored"}
    
    logger.info(
        f"[WEBHOOK] Processing {evt_type} invoice_id={invoice_id} subscription_id={subscription_id} status={status}"
    )
    
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            if evt_type == "invoice.payment_made":
                # Payment successful - activate membership and advance billing date
                try:
                    # Get subscription details to calculate next billing date
                    subscription_response = await square_client.get_subscription(subscription_id)
                    sub_data = subscription_response.get("subscription", {})
                    
                    # Get plan details to determine cadence
                    plan_variation_id = sub_data.get("plan_variation_id")
                    cadence = "MONTHLY"  # Default fallback
                    anchor_day = sub_data.get("monthly_billing_anchor_date", 1)
                    
                    if plan_variation_id:
                        try:
                            plan_response = await square_client.get_subscription_plan(plan_variation_id)
                            plan_data = plan_response.get("object", {}).get("subscription_plan_variation_data", {})
                            phases = plan_data.get("phases", [])
                            if phases:
                                cadence = phases[0].get("cadence", "MONTHLY")
                        except Exception as e:
                            logger.warning(f"[WEBHOOK] Could not get plan cadence, using monthly: {e}")
                    
                    # Calculate next billing date based on cadence
                    next_billing_date = calculate_next_billing_date(cadence, anchor_day)
                    
                    cur.execute(
                        """
                        UPDATE users.customer_membership_plans 
                        SET status = 'active',
                            subscription_status = 'ACTIVE',
                            next_billing_date = %s
                        WHERE square_subscription_id = %s
                        """,
                        (next_billing_date, subscription_id)
                    )
                    
                    logger.info(f"[WEBHOOK] Payment successful for {subscription_id}, cadence={cadence}, advanced billing date to {next_billing_date}")
                    
                except Exception as e:
                    logger.warning(f"[WEBHOOK] Could not advance billing date for {subscription_id}: {e}")
                    # Fallback to original update without billing date
                    cur.execute(
                        """
                        UPDATE users.customer_membership_plans 
                        SET status = 'active',
                            subscription_status = 'ACTIVE'
                        WHERE square_subscription_id = %s
                        """,
                        (subscription_id,)
                    )
                
            elif evt_type == "invoice.failed":
                # Payment failed - mark as delinquent
                cur.execute(
                    """
                    UPDATE users.customer_membership_plans 
                    SET subscription_status = 'DELINQUENT',
                        status = 'inactive'
                    WHERE square_subscription_id = %s
                    """,
                    (subscription_id,)
                )
            
            conn.commit()
            
        return {"status": "ok"}
        
    except Exception as e:
        logger.exception(f"[WEBHOOK] Error processing invoice webhook: {e}")
        raise HTTPException(status_code=500, detail="Webhook processing error")


# Admin: expire memberships (treat expired as inactive by flipping status)

only_authenticated = [Depends(HTTPBearer())]


@app.get("/admin/square/subscription-plans", dependencies=only_authenticated)
async def list_square_subscription_plans():
    """
    Lists all Square subscription plans with their variation IDs.
    Useful for getting variation IDs to use as stripe_price_id when creating membership plans.
    """
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            # Get all catalog objects of type SUBSCRIPTION_PLAN
            resp = await client.get(
                f"{SQUARE_BASE_URL}/v2/catalog/list",
                headers=HEADERS,
                params={"types": "SUBSCRIPTION_PLAN"}
            )
            
            if resp.status_code >= 400:
                logger.error(f"[SQUARE] List subscription plans failed {resp.status_code}: {resp.text}")
                raise HTTPException(status_code=502, detail="Square API error")
            
            data = resp.json()
            objects = data.get("objects", [])
            
            plans = []
            for obj in objects:
                if obj.get("type") == "SUBSCRIPTION_PLAN":
                    plan_data = obj.get("subscription_plan_data", {})
                    
                    # Get the subscription plan variation (there's usually one per plan)
                    subscription_plan_variations = plan_data.get("subscription_plan_variations", [])
                    
                    for variation in subscription_plan_variations:
                        variation_data = variation.get("subscription_plan_variation_data", {})
                        phases = variation_data.get("phases", [])
                        
                        # Extract pricing from first phase
                        price_info = None
                        if phases:
                            recurring_price_money = phases[0].get("recurring_price_money", {})
                            if recurring_price_money:
                                price_info = {
                                    "amount": recurring_price_money.get("amount"),
                                    "currency": recurring_price_money.get("currency", "USD")
                                }
                        
                        plans.append({
                            "plan_id": obj.get("id"),
                            "variation_id": variation.get("id"), 
                            "name": plan_data.get("name"),
                            "pricing": price_info,
                            "cadence": phases[0].get("cadence") if phases else None
                        })
            
            logger.info(f"[ADMIN] Retrieved {len(plans)} Square subscription plans")
            return {
                "subscription_plans": plans,
                "total_count": len(plans)
            }
            
    except httpx.RequestError as e:
        logger.exception(f"[SQUARE] Network error retrieving subscription plans: {e}")
        raise HTTPException(status_code=502, detail="Square service temporarily unavailable")
    except Exception as e:
        logger.exception(f"[ADMIN] Unexpected error retrieving subscription plans: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")

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
    
    # Create webhook failures table if it doesn't exist
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute("""
                CREATE TABLE IF NOT EXISTS webhook_failures (
                    id SERIAL PRIMARY KEY,
                    event_type VARCHAR(100),
                    event_data TEXT,
                    error_message TEXT,
                    created_at TIMESTAMP DEFAULT NOW(),
                    processed BOOLEAN DEFAULT FALSE
                );
                CREATE INDEX IF NOT EXISTS idx_webhook_failures_processed ON webhook_failures(processed, created_at);
                
                CREATE TABLE IF NOT EXISTS payment_history (
                    id SERIAL PRIMARY KEY,
                    user_id UUID NOT NULL,
                    plan_id UUID NOT NULL,
                    amount INTEGER NOT NULL,
                    checkout_id VARCHAR(255) UNIQUE NOT NULL,
                    status VARCHAR(50) NOT NULL,
                    created_at TIMESTAMP DEFAULT NOW(),
                    updated_at TIMESTAMP DEFAULT NOW()
                );
                CREATE INDEX IF NOT EXISTS idx_payment_history_user_plan ON payment_history(user_id, plan_id, created_at);
                CREATE INDEX IF NOT EXISTS idx_payment_history_checkout ON payment_history(checkout_id);
            """)
            conn.commit()
            logger.info("[STARTUP] Webhook failures table ready")
    except Exception as e:
        logger.error(f"[STARTUP] Error creating webhook failures table: {e}")
    
    # Start periodic sync thread
    try:
        sync_thread = threading.Thread(target=run_periodic_sync, daemon=True)
        sync_thread.start()
        logger.info("[STARTUP] Periodic sync thread started")
    except Exception as e:
        logger.error(f"[STARTUP] Error starting sync thread: {e}")
    
    logger.info("[STARTUP] Rise Square Payment Service ready - 100% sync enabled")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown."""
    if db_pool:
        db_pool.closeall()
    logger.info("[SHUTDOWN] Rise Square Payment Service stopped")


# Metrics endpoint
@app.get("/metrics", include_in_schema=False)
async def get_metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type="text/plain"
    )

@app.get("/health", include_in_schema=False)
async def health_check():
    """Health check endpoint for load balancers."""
    try:
        with get_db_conn() as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT 1")
                cur.fetchone()
        # Update active subscriptions count for monitoring
        with get_db_conn() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    "SELECT COUNT(*) FROM users.customer_membership_plans WHERE subscription_status = 'ACTIVE'"
                )
                count = cur.fetchone()[0]
                active_subscriptions.set(count)
        
        return {
            "status": "healthy", 
            "service": "rise-square-payment", 
            "timestamp": datetime.utcnow().isoformat(),
            "active_subscriptions": count,
            "square_api_version": HEADERS.get("Square-Version")
        }
    except Exception as e:
        logger.error(f"[HEALTH] Health check failed: {e}")
        raise HTTPException(status_code=503, detail="Service unavailable")

@app.get("/admin/sync/status", dependencies=only_authenticated)
async def get_sync_status():
    """Get comprehensive sync status and health metrics."""
    global sync_stats
    
    try:
        # Check for recent discrepancies
        with get_db_conn() as conn, conn.cursor() as cur:
            # Count potential sync issues
            cur.execute(
                """
                SELECT COUNT(*) as total_subs,
                       COUNT(CASE WHEN subscription_status = 'ACTIVE' THEN 1 END) as active_subs,
                       COUNT(CASE WHEN next_billing_date IS NULL AND subscription_status = 'ACTIVE' THEN 1 END) as missing_billing_dates
                FROM users.customer_membership_plans 
                WHERE square_subscription_id IS NOT NULL
                """
            )
            counts = cur.fetchone()
            
            # Check failed webhooks
            cur.execute(
                """
                SELECT COUNT(*) as failed_webhooks,
                       COUNT(CASE WHEN processed = false THEN 1 END) as unprocessed_failures
                FROM webhook_failures 
                WHERE created_at > NOW() - INTERVAL '24 hours'
                """
            )
            webhook_failures = cur.fetchone()
        
        return {
            "sync_health": {
                "last_full_sync": sync_stats["last_full_sync"],
                "total_discrepancies_found": sync_stats["discrepancies_found"],
                "total_discrepancies_fixed": sync_stats["discrepancies_fixed"],
                "failed_syncs": sync_stats["failed_syncs"]
            },
            "subscription_health": {
                "total_subscriptions": counts[0] if counts else 0,
                "active_subscriptions": counts[1] if counts else 0,
                "missing_billing_dates": counts[2] if counts else 0
            },
            "webhook_health": {
                "failed_webhooks_24h": webhook_failures[0] if webhook_failures else 0,
                "unprocessed_failures": webhook_failures[1] if webhook_failures else 0
            },
            "recommendation": "healthy" if (counts[2] == 0 and webhook_failures[1] == 0) else "needs_attention"
        }
        
    except Exception as e:
        logger.error(f"[SYNC_STATUS] Error getting sync status: {e}")
        raise HTTPException(status_code=500, detail="Error getting sync status")

@app.post("/admin/sync/force", dependencies=only_authenticated) 
async def force_full_sync():
    """Manually trigger a full synchronization."""
    try:
        # Run sync in background
        asyncio.create_task(perform_full_subscription_sync())
        return {"status": "sync_triggered", "message": "Full sync started in background"}
    except Exception as e:
        logger.error(f"[FORCE_SYNC] Error: {e}")
        raise HTTPException(status_code=500, detail="Failed to start sync")

@app.post("/admin/webhooks/retry", dependencies=only_authenticated)
async def retry_failed_webhooks():
    """Manually retry processing of failed webhook events."""
    try:
        # Process failed webhooks in background
        asyncio.create_task(process_failed_webhooks())
        return {"status": "retry_triggered", "message": "Failed webhook processing started"}
    except Exception as e:
        logger.error(f"[WEBHOOK_RETRY] Error: {e}")
        raise HTTPException(status_code=500, detail="Failed to start webhook retry")

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
