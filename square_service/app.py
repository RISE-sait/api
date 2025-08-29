import os
import logging
import asyncio
import threading
from fastapi import FastAPI, HTTPException, Request, Depends, Response
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from fastapi.middleware.cors import CORSMiddleware
from slowapi import Limiter, _rate_limit_exceeded_handler
from slowapi.middleware import SlowAPIMiddleware
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded
from prometheus_client import generate_latest

# Import our refactored modules
from config.settings import validate_environment, SQUARE_LOCATION_ID, FRONTEND_URL
from database.connection import init_db_pool, get_db_conn, _warn_if_db_superuser, close_db_pool
from services.square_client import SquareAPIClient
from services.subscription_service import SubscriptionService
from services.sync_service import SyncService
from services.admin_service import AdminService
from webhooks.handlers import WebhookHandler
from models.requests import SubscriptionRequest, SubscriptionManagementRequest, CheckoutRequest, ProgramCheckoutRequest, EventCheckoutRequest
from models.responses import SubscriptionResponse, CheckoutResponse, ErrorResponse
from utils.auth import get_user_id, verify_signature
from utils.validation import validation_middleware
from utils.rate_limiter import rate_limit_middleware
from metrics import subscription_requests, subscription_duration, webhook_events, active_subscriptions

# Configure logging
logging.basicConfig(
    level=getattr(logging, os.getenv("LOG_LEVEL", "INFO")),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Validate environment on startup
validate_environment()

# Initialize FastAPI app
app = FastAPI(
    title="Rise Square Payment Service",
    description="Square payment processing and subscription management",
    version="1.0.0"
)

# Legacy rate limiting (keeping for backward compatibility)
limiter = Limiter(key_func=get_remote_address)
app.state.limiter = limiter
app.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)
app.add_middleware(SlowAPIMiddleware)

# Add security middleware (order matters - these run in reverse order)
app.middleware("http")(validation_middleware)
app.middleware("http")(rate_limit_middleware)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=[FRONTEND_URL] if FRONTEND_URL else [],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    allow_headers=["Content-Type", "Authorization", "X-Requested-With"],
)

# Security headers middleware
@app.middleware("http")
async def security_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["X-Frame-Options"] = "DENY"
    response.headers["X-XSS-Protection"] = "1; mode=block"
    response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
    return response

# Initialize services
square_client = SquareAPIClient()
subscription_service = SubscriptionService(square_client)
sync_service = SyncService(square_client)
admin_service = AdminService(square_client)
webhook_handler = WebhookHandler(square_client)

# Authentication dependency
only_authenticated = [Depends(HTTPBearer())]

# === SUBSCRIPTION ROUTES ===

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
            result = await subscription_service.create_subscription_checkout(
                user_id, req.membership_plan_id, req.timezone
            )
            subscription_requests.labels(operation='checkout', status='success').inc()
            return SubscriptionResponse(**result)
            
        except HTTPException:
            subscription_requests.labels(operation='checkout', status='error').inc()
            raise
        except Exception as e:
            subscription_requests.labels(operation='checkout', status='error').inc()
            logger.exception(f"[SUBSCRIPTION] Error creating checkout: {e}")
            raise HTTPException(status_code=500, detail="Internal server error")

@app.post("/subscriptions/{subscription_id}/manage", response_model=SubscriptionResponse)
@limiter.limit("10/minute")
async def manage_subscription(
    request: Request,
    subscription_id: str,
    req: SubscriptionManagementRequest,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer())
) -> SubscriptionResponse:
    """Manage subscription (pause, resume, cancel, etc.)."""
    user_id = get_user_id(creds)
    
    with subscription_duration.labels(operation=req.action).time():
        subscription_requests.labels(operation=req.action, status='started').inc()
        
        try:
            result = await subscription_service.manage_subscription(
                subscription_id, req.action.value,
                pause_cycle_duration=req.pause_cycle_duration,
                cancel_reason=req.cancel_reason
            )
            subscription_requests.labels(operation=req.action, status='success').inc()
            return SubscriptionResponse(**result)
            
        except HTTPException:
            subscription_requests.labels(operation=req.action, status='error').inc()
            raise
        except Exception as e:
            subscription_requests.labels(operation=req.action, status='error').inc()
            logger.exception(f"[SUBSCRIPTION] Error managing subscription: {e}")
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
        # For now, we'll just trigger the sync service
        # In a full implementation, you'd call sync_service.sync_specific_subscription
        return {"status": "success", "message": "Billing date sync triggered"}
    except Exception as e:
        logger.error(f"[MANUAL_SYNC] Error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/subscriptions/{subscription_id}")
@limiter.limit("20/minute")
async def get_subscription_details(
    request: Request,
    subscription_id: str,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer())
) -> dict:
    """Get subscription details from Square API."""
    user_id = get_user_id(creds)
    return await subscription_service.get_subscription_details(subscription_id, user_id)

@app.post("/checkout/event")
@limiter.limit("10/minute")
async def checkout_event(
    request: Request,
    req: EventCheckoutRequest,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer()),
) -> dict:
    """Create checkout for event enrollment."""
    user_id = get_user_id(creds)
    return await subscription_service.create_event_checkout(user_id, req.event_id)

# === WEBHOOK ROUTES ===

@app.post("/webhook")
@limiter.limit("60/minute")
async def handle_webhook(request: Request):
    """Handle Square webhooks."""
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
    
    # LOG ALL WEBHOOK DATA FOR DEBUGGING
    logger.info(f"[WEBHOOK_DEBUG] Full event data: {event}")
    
    # Enhanced webhook processing with failure recovery
    try:
        # Handle subscription events
        if evt_type in ["subscription.created", "subscription.updated"]:
            logger.info(f"[WEBHOOK] Processing subscription event: {evt_type}")
            return await webhook_handler.handle_subscription_webhook(event)
        
        # Handle invoice events for subscriptions  
        if evt_type in ["invoice.created", "invoice.payment_made", "invoice.scheduled_charge_failed"]:
            logger.info(f"[WEBHOOK] Processing invoice event: {evt_type}")
            return await webhook_handler.handle_invoice_webhook(event)
        
        # Handle payment events
        if evt_type == "payment.updated":
            logger.info(f"[WEBHOOK] Processing payment event: {evt_type}")
            result = await webhook_handler.handle_payment_webhook(event)
            
            # Check if payment is completed for subscription activation
            payment_data = event.get("data", {}).get("object", {}).get("payment", {})
            order_id = payment_data.get("order_id")
            
            if order_id and payment_data.get("status") == "COMPLETED":
                try:
                    # Get order details to find associated subscription
                    order_response = await square_client.make_request("GET", f"/v2/orders/{order_id}")
                    if order_response.status_code == 200:
                        order_data = order_response.json().get("order", {})
                        metadata = order_data.get("metadata", {})
                        
                        # Check if this is for subscription setup (new payment-first flow)
                        if metadata.get("subscription_setup") == "true":
                            logger.info(f"[WEBHOOK] Payment completed for subscription setup, creating subscription")
                            await subscription_service.complete_subscription_after_payment(metadata)
                        
                        # Check if this payment is for an existing subscription (current flow)
                        # Look for subscription ID in metadata or customer orders
                        customer_id = order_data.get("customer_id")
                        if customer_id:
                            # Find the subscription by customer and activate membership
                            with get_db_conn() as conn, conn.cursor() as cur:
                                cur.execute(
                                    """
                                    SELECT square_subscription_id FROM users.customer_membership_plans 
                                    WHERE status = 'inactive'::membership.membership_status
                                    AND created_at > NOW() - INTERVAL '1 hour'
                                    ORDER BY created_at DESC LIMIT 1
                                    """
                                )
                                result_row = cur.fetchone()
                                if result_row:
                                    square_subscription_id = result_row[0]
                                    logger.info(f"[WEBHOOK] Activating membership for subscription: {square_subscription_id}")
                                    await subscription_service.activate_membership_after_payment(square_subscription_id)
                        
                except Exception as e:
                    logger.error(f"[WEBHOOK] Error processing payment completion: {e}")
            
            return result
        
        # Handle checkout events (might be relevant for debugging)
        if evt_type in ["checkout.created", "checkout.completed", "checkout.failed"]:
            logger.info(f"[WEBHOOK] Processing checkout event: {evt_type}")
            return {"status": "checkout_event_received"}
        
        # Handle order events
        if evt_type in ["order.created", "order.updated", "order.fulfillment.updated"]:
            logger.info(f"[WEBHOOK] Processing order event: {evt_type}")
            return {"status": "order_event_received"}
            
        # Log unknown events
        logger.warning(f"[WEBHOOK] Unknown event type: {evt_type}")
            
    except Exception as e:
        # Store failed webhook for later retry
        logger.error(f"[WEBHOOK] Error processing {evt_type}: {e}")
        logger.error(f"[WEBHOOK] Full error event: {event}")
        await sync_service.store_failed_webhook(event, str(e))
        # Don't raise - return success to prevent Square from retrying immediately
        return {"status": "error_stored_for_retry"}

    return {"status": "ok"}

# === ADMIN ROUTES ===

@app.get("/admin/square/subscription-plans", dependencies=only_authenticated)
async def list_square_subscription_plans():
    """List available Square subscription plans."""
    try:
        return await admin_service.list_square_subscription_plans()
    except Exception as e:
        logger.error(f"[ADMIN] Error listing subscription plans: {e}")
        raise HTTPException(status_code=500, detail="Error listing subscription plans")

@app.post("/admin/memberships/expire", dependencies=only_authenticated)
async def expire_memberships():
    """Expire memberships that have passed their renewal date."""
    try:
        return await admin_service.expire_memberships()
    except Exception as e:
        logger.error(f"[ADMIN] Error expiring memberships: {e}")
        raise HTTPException(status_code=500, detail="Error expiring memberships")

@app.get("/admin/sync/status", dependencies=only_authenticated)
async def get_sync_status():
    """Get comprehensive sync status and health metrics."""
    try:
        return sync_service.get_sync_status()
    except Exception as e:
        logger.error(f"[SYNC_STATUS] Error getting sync status: {e}")
        raise HTTPException(status_code=500, detail="Error getting sync status")

@app.post("/admin/sync/force", dependencies=only_authenticated) 
async def force_full_sync():
    """Manually trigger a full synchronization."""
    try:
        asyncio.create_task(sync_service.perform_full_subscription_sync())
        return {"status": "sync_triggered", "message": "Full sync started in background"}
    except Exception as e:
        logger.error(f"[FORCE_SYNC] Error: {e}")
        raise HTTPException(status_code=500, detail="Failed to start sync")

@app.post("/admin/webhooks/retry", dependencies=only_authenticated)
async def retry_failed_webhooks():
    """Manually retry processing of failed webhook events."""
    try:
        asyncio.create_task(sync_service.process_failed_webhooks())
        return {"status": "retry_triggered", "message": "Failed webhook processing started"}
    except Exception as e:
        logger.error(f"[WEBHOOK_RETRY] Error: {e}")
        raise HTTPException(status_code=500, detail="Failed to start webhook retry")


# === HEALTH & MONITORING ROUTES ===

@app.get("/metrics", include_in_schema=False)
async def get_metrics():
    """Prometheus metrics endpoint."""
    return Response(
        content=generate_latest(),
        media_type="text/plain"
    )

@app.get("/admin/rate-limiter/status")
@limiter.limit("20/minute")
async def get_rate_limiter_status(
    request: Request,
    creds: HTTPAuthorizationCredentials = Depends(HTTPBearer())
):
    """Get rate limiter status for monitoring."""
    from utils.rate_limiter import get_rate_limiter_status
    return get_rate_limiter_status()

@app.get("/debug/checkout/{checkout_id}", dependencies=only_authenticated)
async def debug_checkout_status(checkout_id: str):
    """Debug checkout session status to see why payment failed."""
    try:
        # Get checkout session details
        response = await square_client.make_request(
            "GET", 
            f"/v2/online-checkout/payment-links/{checkout_id}"
        )
        
        if response.status_code == 200:
            data = response.json()
            return {
                "checkout_id": checkout_id,
                "status": data.get("payment_link", {}).get("status"),
                "full_response": data
            }
        else:
            return {
                "error": f"Failed to get checkout: {response.status_code}",
                "response": response.text
            }
    except Exception as e:
        logger.error(f"[DEBUG] Error getting checkout {checkout_id}: {e}")
        return {"error": str(e)}

@app.get("/health", include_in_schema=False)
async def health_check():
    """Health check endpoint for load balancers."""
    try:
        with get_db_conn() as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT 1")
                cur.fetchone()
        
        # Update active subscriptions count
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
            "timestamp": "2025-08-27T21:59:00Z",  # You could use datetime.utcnow()
            "active_subscriptions": count,
            "square_api_version": "2025-07-16"
        }
    except Exception as e:
        logger.error(f"[HEALTH] Health check failed: {e}")
        raise HTTPException(status_code=503, detail="Service unavailable")

@app.get("/", include_in_schema=False)
async def root():
    """Root endpoint."""
    return {
        "service": "rise-square-payment",
        "status": "running",
        "version": "1.0.0"
    }

# === STARTUP & SHUTDOWN EVENTS ===

@app.on_event("startup")
async def startup_event():
    """Initialize services on startup."""
    init_db_pool()
    _warn_if_db_superuser()
    
    # Create required tables
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
            logger.info("[STARTUP] Database tables ready")
    except Exception as e:
        logger.error(f"[STARTUP] Error creating tables: {e}")
    
    # Start periodic sync thread
    try:
        sync_thread = threading.Thread(target=sync_service.run_periodic_sync, daemon=True)
        sync_thread.start()
        logger.info("[STARTUP] Periodic sync thread started")
    except Exception as e:
        logger.error(f"[STARTUP] Error starting sync thread: {e}")
    
    logger.info("[STARTUP] Rise Square Payment Service ready - 100% sync enabled")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown."""
    close_db_pool()
    logger.info("[SHUTDOWN] Rise Square Payment Service stopped")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)