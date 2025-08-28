import asyncio
import logging
import threading
import time
import json
from datetime import datetime
from typing import Dict, Any
from database.connection import get_db_conn
from services.square_client import SquareAPIClient
from utils.billing import calculate_next_billing_date
from utils.duplicate_prevention import cleanup_caches

logger = logging.getLogger(__name__)

class SyncService:
    def __init__(self, square_client: SquareAPIClient):
        self.square_client = square_client
        self.sync_stats = {
            "last_full_sync": None,
            "discrepancies_found": 0,
            "discrepancies_fixed": 0,
            "failed_syncs": 0
        }
    
    async def perform_full_subscription_sync(self):
        """Complete sync between Square and database - runs periodically."""
        try:
            logger.info("[FULL_SYNC] Starting comprehensive subscription sync")
            self.sync_stats["last_full_sync"] = datetime.utcnow().isoformat()
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
                        square_response = await self.square_client.get_subscription(square_sub_id)
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
                                    plan_response = await self.square_client.get_subscription_plan(plan_variation_id)
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
            
            self.sync_stats["discrepancies_found"] += discrepancies_found
            self.sync_stats["discrepancies_fixed"] += discrepancies_fixed
            
            logger.info(
                f"[FULL_SYNC] Completed: {discrepancies_found} discrepancies found, "
                f"{discrepancies_fixed} fixed"
            )
            
        except Exception as e:
            self.sync_stats["failed_syncs"] += 1
            logger.exception(f"[FULL_SYNC] Failed: {e}")
    
    def run_periodic_sync(self):
        """Background thread for periodic sync jobs."""
        while True:
            try:
                # Run full sync every hour
                loop = asyncio.new_event_loop()
                asyncio.set_event_loop(loop)
                loop.run_until_complete(self.perform_full_subscription_sync())
                loop.close()
                
                # Clean up old cache entries
                cleanup_caches()
                
                time.sleep(3600)  # 1 hour
            except Exception as e:
                logger.error(f"[PERIODIC_SYNC] Error: {e}")
                time.sleep(300)  # Wait 5 minutes on error
    
    async def store_failed_webhook(self, event_data: Dict[str, Any], error_msg: str):
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
    
    async def process_failed_webhooks(self):
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
                
                from webhooks.handlers import WebhookHandler
                webhook_handler = WebhookHandler(self.square_client)
                
                for webhook_id, event_data in failed_webhooks:
                    try:
                        event = json.loads(event_data)
                        evt_type = event.get("type")
                        
                        if evt_type in ["subscription.created", "subscription.updated"]:
                            await webhook_handler.handle_subscription_webhook(event)
                        elif evt_type in ["invoice.created", "invoice.payment_made", "invoice.failed"]:
                            await webhook_handler.handle_invoice_webhook(event)
                        elif evt_type == "payment.updated":
                            await webhook_handler.handle_payment_webhook(event)
                        
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
    
    def get_sync_status(self) -> Dict[str, Any]:
        """Get current sync status and health metrics."""
        try:
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
                    "last_full_sync": self.sync_stats["last_full_sync"],
                    "total_discrepancies_found": self.sync_stats["discrepancies_found"],
                    "total_discrepancies_fixed": self.sync_stats["discrepancies_fixed"],
                    "failed_syncs": self.sync_stats["failed_syncs"]
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
            return {"error": str(e)}