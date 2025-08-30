import time
import logging
from database.connection import get_db_conn

logger = logging.getLogger(__name__)

# In-memory caches for duplicate prevention
payment_cache = {}
checkout_cache = {}

def generate_idempotency_key(user_id: str, plan_id: str) -> str:
    """Generate unique idempotency key for user+plan combination."""
    timestamp = str(int(time.time()))
    return f"checkout-{user_id}-{plan_id}-{timestamp}"

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

def is_duplicate_payment_webhook(payment_id: str) -> bool:
    """Check if payment webhook has already been processed successfully."""
    webhook_cache_key = f"payment-{payment_id}"
    current_time = time.time()
    
    if webhook_cache_key in payment_cache:
        return True
    
    # Don't cache yet - cache only after successful processing
    return False

def mark_payment_webhook_processed(payment_id: str) -> None:
    """Mark payment webhook as successfully processed."""
    webhook_cache_key = f"payment-{payment_id}"
    current_time = time.time()
    payment_cache[webhook_cache_key] = current_time

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