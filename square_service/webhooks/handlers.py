import logging
import json
import asyncio
from typing import Dict, Any
from fastapi import HTTPException
from database.connection import get_db_conn
from services.square_client import SquareAPIClient
from utils.billing import calculate_next_billing_date
from utils.duplicate_prevention import is_duplicate_payment_webhook

logger = logging.getLogger(__name__)

class WebhookHandler:
    def __init__(self, square_client: SquareAPIClient):
        self.square_client = square_client
    
    async def handle_payment_webhook(self, event: Dict[str, Any]) -> Dict[str, str]:
        """Handle payment.updated webhook events."""
        payment = event.get("data", {}).get("object", {}).get("payment") or {}
        status = payment.get("status")
        payment_id = payment.get("id")
        order_id = payment.get("order_id")
        
        # Prevent duplicate processing
        if is_duplicate_payment_webhook(payment_id):
            logger.warning(f"[DUPLICATE_PREVENTION] Duplicate payment webhook ignored payment_id={payment_id}")
            return {"status": "duplicate_ignored"}
        
        amount_money = payment.get("amount_money") or {}
        logger.info(
            "[WEBHOOK] payment.updated status=%s order_id=%s payment_id=%s amount=%s currency=%s",
            status, order_id, payment.get("id"), amount_money.get("amount"), amount_money.get("currency")
        )

        if status == "COMPLETED" and order_id:
            return await self._process_completed_payment(order_id, payment_id)
        
        return {"status": "ok"}
    
    async def _process_completed_payment(self, order_id: str, payment_id: str) -> Dict[str, str]:
        """Process completed payment and create subscription if needed."""
        try:
            # Get order details from Square
            order_response = await self.square_client.make_request("GET", f"/v2/orders/{order_id}")
            order_data = order_response.json().get("order", {})
            
            # Extract metadata from order
            metadata_source = "order"
            metadata = order_data.get("metadata", {})
            
            if not metadata:
                logger.warning(f"[WEBHOOK] No metadata found in order {order_id}")
                return {"status": "no_metadata"}
            
            logger.info(f"[WEBHOOK] order_id={order_id} metadata_source={metadata_source} metadata={metadata}")
            
            user_id = metadata.get("user_id")
            membership_plan_id = metadata.get("membership_plan_id")
            
            if not user_id or not membership_plan_id:
                logger.error(f"[WEBHOOK] Missing user_id or membership_plan_id in metadata: {metadata}")
                return {"status": "invalid_metadata"}
            
            return await self._activate_membership_and_create_subscription(user_id, membership_plan_id)
            
        except Exception as e:
            logger.exception(f"[WEBHOOK] Error processing payment {payment_id}: {e}")
            raise HTTPException(status_code=500, detail="Payment processing error")
    
    async def _activate_membership_and_create_subscription(self, user_id: str, membership_plan_id: str) -> Dict[str, str]:
        """Activate membership and create Square subscription."""
        logger.info(f"[WEBHOOK] attempting membership activation user={user_id} plan={membership_plan_id}")
        
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                # Update membership status to active
                cur.execute(
                    """
                    UPDATE users.customer_membership_plans
                       SET status='active'
                     WHERE customer_id=%s
                       AND membership_plan_id=%s
                       AND status != 'active'
                    """,
                    (user_id, membership_plan_id)
                )
                
                rowcount = cur.rowcount
                conn.commit()
                logger.info(f"[WEBHOOK] membership UPDATE → active rowcount={rowcount} user={user_id} plan={membership_plan_id}")
                
                if rowcount == 0:
                    logger.info(f"[WEBHOOK] membership already ACTIVE user={user_id} plan={membership_plan_id}")
                    return {"status": "already_active"}
                
                # Check if we need to create a subscription
                cur.execute(
                    """
                    SELECT square_subscription_id, subscription_source 
                    FROM users.customer_membership_plans 
                    WHERE customer_id = %s AND membership_plan_id = %s
                    """,
                    (user_id, membership_plan_id)
                )
                result = cur.fetchone()
                
                if result:
                    sub_id, source = result[0], result[1]
                    
                    if sub_id is None and source == "subscription":
                        logger.info(f"[WEBHOOK] Need to create subscription - current: sub_id={sub_id} source={source}")
                        await self._create_square_subscription(user_id, membership_plan_id, cur, conn)
                
                return {"status": "ok", "did_anything": True}
                
        except Exception as e:
            logger.exception(f"[WEBHOOK] Error activating membership: {e}")
            raise
    
    async def _create_square_subscription(self, user_id: str, membership_plan_id: str, cur, conn):
        """Create Square subscription for the user."""
        logger.info(f"[WEBHOOK] Creating Square subscription for user={user_id} plan={membership_plan_id}")
        
        # Get user details
        cur.execute("SELECT email, first_name, last_name, square_customer_id FROM users.users WHERE id = %s", (user_id,))
        user_result = cur.fetchone()
        
        if not user_result:
            logger.error(f"[WEBHOOK] User not found: {user_id}")
            return
        
        user_email, first_name, last_name, square_customer_id = user_result
        
        # Create Square customer if needed
        if not square_customer_id:
            customer_data = {
                "idempotency_key": f"customer-{user_id}",
                "given_name": first_name or "",
                "family_name": last_name or "",
                "email_address": user_email
            }
            
            customer_response = await self.square_client.create_customer(user_id, user_email, 
                                                                       first_name=first_name, 
                                                                       last_name=last_name)
            
            if customer_response.get("customer"):
                square_customer_id = customer_response["customer"]["id"]
                cur.execute("UPDATE users.users SET square_customer_id = %s WHERE id = %s", 
                           (square_customer_id, user_id))
                logger.info(f"[WEBHOOK] Created Square customer {square_customer_id} for user {user_id}")
        
        # Get subscription plan details
        cur.execute(
            """SELECT stripe_price_id, unit_amount FROM membership.membership_plans 
               WHERE id = %s""",
            (membership_plan_id,)
        )
        plan_result = cur.fetchone()
        
        if plan_result and plan_result[0] and square_customer_id:
            variation_id, plan_price_cents = plan_result[0], plan_result[1]
            
            try:
                square_plan_details = await self.square_client.get_subscription_plan(variation_id)
                plan_phases = square_plan_details.get("object", {}).get("subscription_plan_variation_data", {}).get("phases", [])
                has_relative_pricing = any(phase.get("pricing", {}).get("type") == "RELATIVE" for phase in plan_phases)
                
                logger.info(f"[WEBHOOK] Plan has relative pricing: {has_relative_pricing}")
                
                from config.settings import SQUARE_LOCATION_ID
                subscription_data = {
                    "idempotency_key": f"subscription-{user_id}-{membership_plan_id}",
                    "location_id": SQUARE_LOCATION_ID,
                    "plan_variation_id": variation_id,
                    "customer_id": square_customer_id
                }
                
                if has_relative_pricing:
                    logger.info("[WEBHOOK] Creating order template for relative pricing")
                    order_template_response = await self.square_client.create_order_template(SQUARE_LOCATION_ID)
                    
                    if order_template_response.get("order"):
                        order_template_id = order_template_response["order"]["id"]
                        logger.info(f"[WEBHOOK] Created order template: {order_template_id}")
                        
                        subscription_data["phases"] = [{
                            "ordinal": 0,
                            "order_template_id": order_template_id
                        }]
                        logger.info(f"[WEBHOOK] Added order template {order_template_id} to subscription")
                
                # Create subscription
                subscription_response = await self.square_client.create_subscription(subscription_data)
                
                if subscription_response.get("subscription"):
                    square_subscription = subscription_response["subscription"]
                    square_subscription_id = square_subscription["id"]
                    subscription_status = square_subscription.get("status", "UNKNOWN")
                    subscription_created_at = square_subscription.get("created_at")
                    
                    # Get detailed subscription info
                    detailed_subscription = await self.square_client.get_subscription(square_subscription_id)
                    sub_details = detailed_subscription.get("subscription", {})
                    logger.info(f"[WEBHOOK] Detailed subscription from API: {sub_details}")
                    
                    next_billing_date = (sub_details.get("charged_through_date") or 
                                       sub_details.get("next_billing_date") or
                                       sub_details.get("billing_anchor_date"))
                    
                    # Update database with subscription info
                    cur.execute(
                        """
                        UPDATE users.customer_membership_plans 
                        SET square_subscription_id = %s,
                            subscription_status = %s,
                            next_billing_date = %s,
                            subscription_created_at = %s
                        WHERE customer_id = %s AND membership_plan_id = %s
                        """,
                        (square_subscription_id, subscription_status, next_billing_date, subscription_created_at, user_id, membership_plan_id)
                    )
                    
                    logger.info(f"[WEBHOOK] Created Square subscription {square_subscription_id} status={subscription_status} next_billing={next_billing_date} created_at={subscription_created_at}")
                    
                    # Schedule billing date sync if needed
                    if not next_billing_date:
                        logger.info(f"[WEBHOOK] Scheduling billing date sync for subscription {square_subscription_id}")
                        # Note: This would need to be implemented with proper async task scheduling
                
                else:
                    logger.error(f"[WEBHOOK] Failed to create Square subscription: {subscription_response}")
                    
            except Exception as e:
                logger.error(f"[WEBHOOK] Error creating Square subscription: {e}")
    
    async def handle_subscription_webhook(self, event: Dict[str, Any]) -> Dict[str, str]:
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
                # For subscription.created, we need to find the record by customer and update it with Square details
                if evt_type == "subscription.created":
                    # Get subscription details from Square API
                    try:
                        detailed_subscription = await self.square_client.get_subscription(subscription_id)
                        sub_details = detailed_subscription.get("subscription", {})
                        
                        # Extract billing and creation dates
                        next_billing_date = (sub_details.get("charged_through_date") or 
                                           sub_details.get("next_billing_date") or
                                           sub_details.get("billing_anchor_date"))
                        subscription_created_at = sub_details.get("created_at")
                        
                        # Find the membership record by customer_id and update with Square subscription details
                        # We'll need to get the user_id from the Square customer_id
                        cur.execute(
                            "SELECT id FROM users.users WHERE square_customer_id = %s",
                            (customer_id,)
                        )
                        user_result = cur.fetchone()
                        
                        if user_result:
                            user_id = user_result[0]
                            
                            # Update the most recent inactive subscription record for this user
                            cur.execute(
                                """
                                UPDATE users.customer_membership_plans 
                                SET square_subscription_id = %s,
                                    subscription_status = %s,
                                    subscription_source = 'subscription',
                                    next_billing_date = %s,
                                    subscription_created_at = %s,
                                    status = CASE 
                                        WHEN %s = 'ACTIVE' THEN 'active'::membership.membership_status
                                        ELSE 'inactive'::membership.membership_status
                                    END
                                WHERE customer_id = %s 
                                AND subscription_source = 'subscription'
                                AND square_subscription_id IS NULL
                                """,
                                (subscription_id, status, next_billing_date, subscription_created_at, status, user_id)
                            )
                            
                            updated_rows = cur.rowcount
                            if updated_rows > 0:
                                logger.info(f"[WEBHOOK] Created Square subscription {subscription_id} status={status} next_billing={next_billing_date} created_at={subscription_created_at}")
                            else:
                                logger.warning(f"[WEBHOOK] No matching subscription record found for user {user_id}")
                        else:
                            logger.warning(f"[WEBHOOK] No user found for Square customer {customer_id}")
                            
                    except Exception as e:
                        logger.error(f"[WEBHOOK] Error getting subscription details: {e}")
                        
                else:
                    # For other subscription events, update by subscription_id
                    cur.execute(
                        """
                        UPDATE users.customer_membership_plans 
                        SET subscription_status = %s,
                            status = CASE 
                                WHEN %s = 'ACTIVE' THEN 'active'::membership.membership_status
                                WHEN %s = 'CANCELED' THEN 'inactive'::membership.membership_status
                                WHEN %s = 'PAUSED' THEN 'inactive'::membership.membership_status
                                ELSE status
                            END
                        WHERE square_subscription_id = %s
                        """,
                        (status, status, status, status, subscription_id)
                    )
                    
                    updated_rows = cur.rowcount
                    if updated_rows > 0:
                        logger.info(f"[WEBHOOK] Updated subscription {subscription_id} to status {status}")
                    else:
                        logger.warning(f"[WEBHOOK] No subscription found in database for {subscription_id}")
                
                conn.commit()
            
            return {"status": "ok"}
            
        except Exception as e:
            logger.exception(f"[WEBHOOK] Error processing subscription webhook: {e}")
            raise HTTPException(status_code=500, detail="Webhook processing error")
    
    async def handle_invoice_webhook(self, event: Dict[str, Any]) -> Dict[str, str]:
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
            return {"status": "ignored"}
        
        logger.info(
            f"[WEBHOOK] Processing {evt_type} invoice_id={invoice_id} subscription_id={subscription_id} status={status}"
        )
        
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                if evt_type == "invoice.payment_made":
                    # Payment successful - activate membership and advance billing date
                    try:
                        subscription_response = await self.square_client.get_subscription(subscription_id)
                        sub_data = subscription_response.get("subscription", {})
                        
                        # Get plan details to determine cadence
                        plan_variation_id = sub_data.get("plan_variation_id")
                        cadence = "MONTHLY"  # Default fallback
                        anchor_day = sub_data.get("monthly_billing_anchor_date", 1)
                        
                        if plan_variation_id:
                            try:
                                plan_response = await self.square_client.get_subscription_plan(plan_variation_id)
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
                            SET status = 'active'::membership.membership_status,
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
                            SET status = 'active'::membership.membership_status,
                                subscription_status = 'ACTIVE'
                            WHERE square_subscription_id = %s
                            """,
                            (subscription_id,)
                        )
                
                elif evt_type == "invoice.scheduled_charge_failed":
                    # Payment failed - mark as delinquent
                    cur.execute(
                        """
                        UPDATE users.customer_membership_plans 
                        SET status = 'inactive'::membership.membership_status,
                            subscription_status = 'DELINQUENT'
                        WHERE square_subscription_id = %s
                        """,
                        (subscription_id,)
                    )
                
                conn.commit()
            
            return {"status": "ok"}
            
        except Exception as e:
            logger.exception(f"[WEBHOOK] Error processing invoice webhook: {e}")
            raise HTTPException(status_code=500, detail="Webhook processing error")