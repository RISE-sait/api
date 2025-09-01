import logging
import asyncio
from datetime import datetime, timedelta
from typing import List, Tuple
from database.connection import get_db_conn
from services.square_client import SquareAPIClient
from webhooks.handlers import WebhookHandler

logger = logging.getLogger(__name__)

class AutoChargingRetryService:
    def __init__(self, square_client: SquareAPIClient):
        self.square_client = square_client
        self.webhook_handler = WebhookHandler(square_client)
    
    async def retry_failed_auto_charging_setups(self, max_retries: int = 3, retry_age_hours: int = 24) -> dict:
        """Retry auto-charging setup for subscriptions that failed in the past 24 hours."""
        logger.info(f"[RETRY] Starting auto-charging retry job, max_retries={max_retries}, retry_age_hours={retry_age_hours}")
        
        failed_setups = await self._get_failed_auto_charging_setups(retry_age_hours)
        
        if not failed_setups:
            logger.info("[RETRY] No failed auto-charging setups found to retry")
            return {"retried": 0, "succeeded": 0, "failed": 0}
        
        logger.info(f"[RETRY] Found {len(failed_setups)} failed auto-charging setups to retry")
        
        succeeded = 0
        failed = 0
        
        for user_id, subscription_id, payment_id, customer_id, retry_count, error_type in failed_setups:
            if retry_count >= max_retries:
                logger.info(f"[RETRY] Skipping user {user_id} - max retries ({max_retries}) reached")
                await self._mark_auto_charging_permanently_failed(user_id)
                continue
            
            logger.info(f"[RETRY] Attempting auto-charging setup for user {user_id}, attempt {retry_count + 1}/{max_retries}")
            
            try:
                # Use the same webhook handler logic to retry auto-charging setup
                success = await self.webhook_handler._update_subscription_with_card_id(user_id, payment_id, customer_id)
                
                if success:
                    succeeded += 1
                    logger.info(f"[RETRY] ✅ Successfully set up auto-charging for user {user_id} on retry")
                else:
                    failed += 1
                    await self._increment_retry_count(user_id)
                    logger.warning(f"[RETRY] ❌ Failed to set up auto-charging for user {user_id} on retry")
                
                # Add delay between retries to avoid rate limiting
                await asyncio.sleep(2)
                
            except Exception as e:
                failed += 1
                logger.error(f"[RETRY] Exception during auto-charging retry for user {user_id}: {e}")
                await self._increment_retry_count(user_id)
        
        logger.info(f"[RETRY] Completed auto-charging retry job: succeeded={succeeded}, failed={failed}")
        return {"retried": len(failed_setups), "succeeded": succeeded, "failed": failed}
    
    async def _get_failed_auto_charging_setups(self, retry_age_hours: int) -> List[Tuple]:
        """Get failed auto-charging setups that are eligible for retry."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT DISTINCT
                        cmp.customer_id,
                        sac.square_subscription_id,
                        COALESCE(sac.last_payment_id, '') as payment_id,
                        u.square_customer_id,
                        sac.retry_count,
                        sac.error_type
                    FROM users.subscription_auto_charging sac
                    JOIN users.customer_membership_plans cmp ON sac.customer_membership_plan_id = cmp.id
                    JOIN users.users u ON cmp.customer_id = u.id
                    WHERE sac.enabled = false
                    AND sac.error_type IS NOT NULL
                    AND cmp.status = 'active'::membership.membership_status
                    AND sac.updated_at > CURRENT_TIMESTAMP - INTERVAL '%s hours'
                    AND sac.retry_count < 3
                    AND sac.permanently_failed = false
                    ORDER BY sac.updated_at DESC
                    """,
                    (retry_age_hours,)
                )
                return cur.fetchall()
        except Exception as e:
            logger.error(f"[RETRY] Error getting failed auto-charging setups: {e}")
            return []
    
    async def _increment_retry_count(self, user_id: str) -> None:
        """Increment the retry count for a user's auto-charging setup."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    UPDATE users.subscription_auto_charging 
                    SET retry_count = retry_count + 1,
                        updated_at = CURRENT_TIMESTAMP
                    FROM users.customer_membership_plans cmp
                    WHERE users.subscription_auto_charging.customer_membership_plan_id = cmp.id
                    AND cmp.customer_id = %s 
                    AND cmp.status = 'active'::membership.membership_status
                    """,
                    (user_id,)
                )
                conn.commit()
        except Exception as e:
            logger.error(f"[RETRY] Error incrementing retry count for user {user_id}: {e}")
    
    async def _mark_auto_charging_permanently_failed(self, user_id: str) -> None:
        """Mark auto-charging as permanently failed for a user."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    UPDATE users.subscription_auto_charging 
                    SET permanently_failed = true,
                        updated_at = CURRENT_TIMESTAMP
                    FROM users.customer_membership_plans cmp
                    WHERE users.subscription_auto_charging.customer_membership_plan_id = cmp.id
                    AND cmp.customer_id = %s 
                    AND cmp.status = 'active'::membership.membership_status
                    """,
                    (user_id,)
                )
                conn.commit()
                logger.warning(f"[RETRY] Marked auto-charging as permanently failed for user {user_id}")
        except Exception as e:
            logger.error(f"[RETRY] Error marking auto-charging as permanently failed for user {user_id}: {e}")
    
    async def get_auto_charging_status_report(self) -> dict:
        """Get a report of auto-charging status across all active subscriptions."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT 
                        COUNT(*) FILTER (WHERE sac.enabled = true) as enabled,
                        COUNT(*) FILTER (WHERE sac.enabled = false AND sac.error_type IS NOT NULL) as failed,
                        COUNT(*) FILTER (WHERE sac.enabled IS NULL) as unknown,
                        COUNT(*) FILTER (WHERE sac.permanently_failed = true) as permanently_failed,
                        COUNT(cmp.*) as total
                    FROM users.customer_membership_plans cmp
                    LEFT JOIN users.subscription_auto_charging sac ON cmp.id = sac.customer_membership_plan_id
                    WHERE cmp.status = 'active'::membership.membership_status
                    AND cmp.subscription_source = 'subscription'
                    """
                )
                result = cur.fetchone()
                
                if result:
                    enabled, failed, unknown, permanently_failed, total = result
                    return {
                        "total_active_subscriptions": total,
                        "auto_charging_enabled": enabled,
                        "auto_charging_failed": failed,
                        "auto_charging_unknown": unknown,
                        "permanently_failed": permanently_failed,
                        "success_rate": f"{(enabled/total*100):.1f}%" if total > 0 else "0%"
                    }
                else:
                    return {"error": "No data found"}
        except Exception as e:
            logger.error(f"[RETRY] Error getting auto-charging status report: {e}")
            return {"error": str(e)}