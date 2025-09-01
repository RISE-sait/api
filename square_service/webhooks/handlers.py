import logging
import json
import asyncio
from typing import Dict, Any
from fastapi import HTTPException
from database.connection import get_db_conn
from services.square_client import SquareAPIClient
from utils.billing import calculate_next_billing_date
from utils.duplicate_prevention import is_duplicate_payment_webhook, mark_payment_webhook_processed

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
        
        # Check for duplicate processing but don't block auto-charging setup
        is_duplicate = is_duplicate_payment_webhook(payment_id)
        if is_duplicate:
            logger.info(f"[DUPLICATE_PREVENTION] Payment {payment_id} already processed, but checking auto-charging setup")
            # Continue processing to ensure auto-charging is set up properly
        
        amount_money = payment.get("amount_money") or {}
        logger.info(
            "[WEBHOOK] payment.updated status=%s order_id=%s payment_id=%s amount=%s currency=%s",
            status, order_id, payment.get("id"), amount_money.get("amount"), amount_money.get("currency")
        )

        if status == "COMPLETED" and order_id:
            return await self._process_completed_payment(order_id, payment_id)
        elif status in ["FAILED", "CANCELED"] and order_id:
            return await self._process_failed_payment(order_id, payment_id, status)
        
        return {"status": "ok"}
    
    async def _process_completed_payment(self, order_id: str, payment_id: str) -> Dict[str, str]:
        """Process completed payment and activate subscription if needed."""
        try:
            # Get order details from Square  
            order_response = await self.square_client.make_request("GET", f"/v2/orders/{order_id}")
            order_data = order_response.json().get("order", {})
            
            # Extract metadata from order (legacy flow)
            metadata = order_data.get("metadata", {})
            
            if metadata:
                # Legacy metadata-based flow
                user_id = metadata.get("user_id")
                membership_plan_id = metadata.get("membership_plan_id")
                
                if user_id and membership_plan_id:
                    logger.info(f"[WEBHOOK] Processing metadata payment user={user_id} plan={membership_plan_id}")
                    return await self._activate_membership_and_create_subscription(user_id, membership_plan_id)
            
            # New subscription flow - find subscription by customer_id
            logger.warning(f"[WEBHOOK] No metadata found in order {order_id}, checking for subscription payment")
            return await self._handle_subscription_payment_activation(payment_id)
            
        except Exception as e:
            logger.exception(f"[WEBHOOK] Error processing payment {payment_id}: {e}")
            raise HTTPException(status_code=500, detail="Payment processing error")
    
    async def _handle_subscription_payment_activation(self, payment_id: str) -> Dict[str, str]:
        """Handle subscription payment activation by customer lookup."""
        try:
            # Get payment details to find customer_id
            payment_response = await self.square_client.make_request("GET", f"/v2/payments/{payment_id}")
            payment_data = payment_response.json().get("payment", {})
            customer_id = payment_data.get("customer_id")
            
            if not customer_id:
                logger.warning(f"[WEBHOOK] No customer_id found in payment {payment_id}")
                return {"status": "no_customer"}
            
            # Find user by square_customer_id or by email if customer ID doesn't match
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    "SELECT id FROM users.users WHERE square_customer_id = %s",
                    (customer_id,)
                )
                user_result = cur.fetchone()
                
                if not user_result:
                    # Fallback: get customer email from Square and find user by email
                    logger.info(f"[WEBHOOK] Customer ID {customer_id} not found, trying email lookup")
                    try:
                        customer_response = await self.square_client.make_request("GET", f"/v2/customers/{customer_id}")
                        customer_data = customer_response.json().get("customer", {})
                        email = customer_data.get("email_address")
                        
                        if email:
                            cur.execute(
                                "SELECT id FROM users.users WHERE email = %s",
                                (email,)
                            )
                            user_result = cur.fetchone()
                            
                            if user_result:
                                # Update the user's square_customer_id to the new one
                                logger.info(f"[WEBHOOK] Found user by email {email}, updating customer ID to {customer_id}")
                                cur.execute(
                                    "UPDATE users.users SET square_customer_id = %s WHERE email = %s",
                                    (customer_id, email)
                                )
                                conn.commit()
                            else:
                                logger.warning(f"[WEBHOOK] No user found for email {email}")
                                return {"status": "no_user"}
                        else:
                            logger.warning(f"[WEBHOOK] No email found for Square customer {customer_id}")
                            return {"status": "no_email"}
                    except Exception as e:
                        logger.error(f"[WEBHOOK] Error looking up customer by email: {e}")
                        return {"status": "lookup_error"}
                
                if not user_result:
                    logger.warning(f"[WEBHOOK] No user found for Square customer {customer_id} even after email lookup")
                    return {"status": "no_user"}
                
                user_id = user_result[0]
                
                # Activate subscription for this user
                cur.execute(
                    """
                    UPDATE users.customer_membership_plans 
                    SET status = 'active'::membership.membership_status
                    WHERE customer_id = %s 
                    AND subscription_source = 'subscription'
                    AND status = 'inactive'::membership.membership_status
                    """,
                    (user_id,)
                )
                
                updated_rows = cur.rowcount
                conn.commit()
                
                if updated_rows > 0:
                    # Set up auto-charging by creating card-on-file and updating subscription
                    auto_charging_success = await self._update_subscription_with_card_id(user_id, payment_id, customer_id)
                    
                    if auto_charging_success:
                        logger.info(f"[WEBHOOK] ✅ Activated subscription for user {user_id} with auto-charging enabled")
                        mark_payment_webhook_processed(payment_id)
                        return {"status": "activated", "auto_charging": "enabled"}
                    else:
                        logger.warning(f"[WEBHOOK] ⚠️  Activated subscription for user {user_id} but auto-charging FAILED - user will receive invoices")
                        mark_payment_webhook_processed(payment_id)
                        return {"status": "activated", "auto_charging": "failed"}
                else:
                    logger.warning(f"[WEBHOOK] No inactive subscription found for user {user_id}")
                    return {"status": "no_subscription"}
                    
        except Exception as e:
            logger.exception(f"[WEBHOOK] Error handling subscription payment activation: {e}")
            return {"status": "error"}
    
    async def _update_subscription_with_card_id(self, user_id: str, payment_id: str, customer_id: str) -> bool:
        """Update Square subscription with card_id from payment to enable auto-charging.
        
        Returns True if auto-charging was successfully set up, False otherwise.
        """
        try:
            logger.info(f"[WEBHOOK] Starting auto-charging setup for payment {payment_id}, customer {customer_id}")
            
            # Get payment details to extract source_id for card creation
            payment_response = await self.square_client.make_request("GET", f"/v2/payments/{payment_id}")
            payment_data = payment_response.json().get("payment", {})
            logger.info(f"[WEBHOOK] Retrieved payment data for {payment_id}")
            source_id = payment_data.get("source_id")
            
            if not source_id:
                logger.error(f"[WEBHOOK] CRITICAL: No source_id found in payment {payment_id} - auto-charging will not work")
                await self._store_auto_charging_failure(user_id, "NO_SOURCE_ID", f"Payment {payment_id} missing source_id")
                return False
            
            # Create card-on-file for future auto-charging
            logger.info(f"[WEBHOOK] Creating card-on-file for customer {customer_id} from payment {payment_id}")
            card_response = await self.square_client.create_card_on_file(customer_id, source_id)
            
            if "card" not in card_response:
                error_details = card_response.get("errors", [])
                logger.error(f"[WEBHOOK] CRITICAL: Failed to create card-on-file: {card_response}")
                await self._store_auto_charging_failure(user_id, "CARD_CREATION_FAILED", f"Card creation failed: {error_details}")
                return False
                
            card_id = card_response["card"]["id"]
            logger.info(f"[WEBHOOK] Successfully created card-on-file {card_id} for customer {customer_id}")
            
            # Get subscription_id for this user
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    "SELECT square_subscription_id FROM users.customer_membership_plans WHERE customer_id = %s AND status = 'active'::membership.membership_status ORDER BY created_at DESC LIMIT 1",
                    (user_id,)
                )
                result = cur.fetchone()
                
                if not result:
                    logger.error(f"[WEBHOOK] CRITICAL: No active subscription found for user {user_id} - cannot set up auto-charging")
                    await self._store_auto_charging_failure(user_id, "NO_SUBSCRIPTION", "No active subscription found")
                    return False
                
                subscription_id = result[0]
            
            # Update Square subscription with card_id for auto-charging
            update_data = {
                "subscription": {
                    "card_id": card_id
                }
            }
            
            update_response = await self.square_client.update_subscription(subscription_id, update_data)
            
            if "subscription" in update_response and update_response["subscription"].get("card_id") == card_id:
                logger.info(f"[WEBHOOK] ✅ SUCCESS: Auto-charging enabled for subscription {subscription_id} with card {card_id}")
                
                # Store success status in database
                await self._store_auto_charging_success(user_id, subscription_id, card_id)
                return True
            else:
                error_details = update_response.get("errors", [])
                logger.error(f"[WEBHOOK] CRITICAL: Failed to update subscription with card_id: {update_response}")
                await self._store_auto_charging_failure(user_id, "SUBSCRIPTION_UPDATE_FAILED", f"Subscription update failed: {error_details}")
                return False
            
        except Exception as e:
            logger.error(f"[WEBHOOK] CRITICAL: Exception during auto-charging setup: {e}")
            logger.exception(f"[WEBHOOK] Full exception traceback for card_id setup:")
            await self._store_auto_charging_failure(user_id, "EXCEPTION", str(e))
            return False
    
    async def _store_auto_charging_success(self, user_id: str, subscription_id: str, card_id: str) -> None:
        """Store successful auto-charging setup in database."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                # Get customer_membership_plan_id
                cur.execute(
                    "SELECT id FROM users.customer_membership_plans WHERE square_subscription_id = %s",
                    (subscription_id,)
                )
                result = cur.fetchone()
                
                if not result:
                    logger.error(f"[WEBHOOK] No membership plan found for subscription {subscription_id}")
                    return
                
                membership_plan_id = result[0]
                
                # Insert or update auto-charging record
                cur.execute(
                    """
                    INSERT INTO users.subscription_auto_charging 
                    (customer_membership_plan_id, square_subscription_id, enabled, card_id, updated_at)
                    VALUES (%s, %s, true, %s, CURRENT_TIMESTAMP)
                    ON CONFLICT (customer_membership_plan_id)
                    DO UPDATE SET 
                        enabled = true,
                        card_id = EXCLUDED.card_id,
                        error_type = NULL,
                        error_details = NULL,
                        updated_at = CURRENT_TIMESTAMP
                    """,
                    (membership_plan_id, subscription_id, card_id)
                )
                conn.commit()
                logger.info(f"[WEBHOOK] Stored auto-charging success for user {user_id}")
        except Exception as e:
            logger.error(f"[WEBHOOK] Error storing auto-charging success: {e}")
    
    async def _store_auto_charging_failure(self, user_id: str, error_type: str, error_details: str) -> None:
        """Store auto-charging setup failure in database for later retry."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                # Get customer_membership_plan_id for active subscription
                cur.execute(
                    """
                    SELECT id, square_subscription_id 
                    FROM users.customer_membership_plans 
                    WHERE customer_id = %s AND status = 'active'::membership.membership_status
                    ORDER BY created_at DESC LIMIT 1
                    """,
                    (user_id,)
                )
                result = cur.fetchone()
                
                if not result:
                    logger.error(f"[WEBHOOK] No active membership plan found for user {user_id}")
                    return
                
                membership_plan_id, subscription_id = result
                
                # Insert or update auto-charging record
                cur.execute(
                    """
                    INSERT INTO users.subscription_auto_charging 
                    (customer_membership_plan_id, square_subscription_id, enabled, error_type, error_details, updated_at)
                    VALUES (%s, %s, false, %s, %s, CURRENT_TIMESTAMP)
                    ON CONFLICT (customer_membership_plan_id)
                    DO UPDATE SET 
                        enabled = false,
                        error_type = EXCLUDED.error_type,
                        error_details = EXCLUDED.error_details,
                        retry_count = COALESCE(users.subscription_auto_charging.retry_count, 0),
                        updated_at = CURRENT_TIMESTAMP
                    """,
                    (membership_plan_id, subscription_id, error_type, error_details)
                )
                conn.commit()
                logger.info(f"[WEBHOOK] Stored auto-charging failure for user {user_id}: {error_type}")
        except Exception as e:
            logger.error(f"[WEBHOOK] Error storing auto-charging failure: {e}")
    
    async def _process_failed_payment(self, order_id: str, payment_id: str, status: str) -> Dict[str, str]:
        """Handle failed or canceled payment by deactivating subscription."""
        try:
            logger.info(f"[WEBHOOK] Processing failed payment {payment_id} status={status} for order {order_id}")
            
            # Get order details to find customer
            order_response = await self.square_client.make_request("GET", f"/v2/orders/{order_id}")
            order_data = order_response.json().get("order", {})
            customer_id = order_data.get("customer_id")
            
            if not customer_id:
                logger.warning(f"[WEBHOOK] No customer_id found in failed payment order {order_id}")
                return {"status": "no_customer"}
            
            # Find user by customer_id (with email fallback)
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    "SELECT id FROM users.users WHERE square_customer_id = %s",
                    (customer_id,)
                )
                user_result = cur.fetchone()
                
                if not user_result:
                    # Try email fallback like in successful payments
                    try:
                        customer_response = await self.square_client.make_request("GET", f"/v2/customers/{customer_id}")
                        customer_data = customer_response.json().get("customer", {})
                        email = customer_data.get("email_address")
                        
                        if email:
                            cur.execute("SELECT id FROM users.users WHERE email = %s", (email,))
                            user_result = cur.fetchone()
                        
                        if not user_result:
                            logger.warning(f"[WEBHOOK] No user found for failed payment customer {customer_id}")
                            return {"status": "no_user"}
                    except Exception as e:
                        logger.error(f"[WEBHOOK] Error looking up customer for failed payment: {e}")
                        return {"status": "lookup_error"}
                
                user_id = user_result[0]
                
                # Deactivate active subscriptions for this user
                cur.execute(
                    """
                    UPDATE users.customer_membership_plans 
                    SET status = 'inactive'::membership.membership_status
                    WHERE customer_id = %s 
                    AND subscription_source = 'subscription'
                    AND status = 'active'::membership.membership_status
                    """,
                    (user_id,)
                )
                
                updated_rows = cur.rowcount
                conn.commit()
                
                if updated_rows > 0:
                    logger.info(f"[WEBHOOK] Deactivated {updated_rows} subscription(s) for user {user_id} due to failed payment {payment_id}")
                    return {"status": "deactivated", "count": updated_rows}
                else:
                    logger.info(f"[WEBHOOK] No active subscriptions found to deactivate for user {user_id}")
                    return {"status": "no_active_subscriptions"}
                    
        except Exception as e:
            logger.exception(f"[WEBHOOK] Error processing failed payment: {e}")
            return {"status": "error"}
    
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
                        
                        if not user_result:
                            # Fallback: get customer email from Square and find user by email
                            logger.info(f"[WEBHOOK] Customer ID {customer_id} not found, trying email lookup for subscription")
                            try:
                                customer_response = await self.square_client.make_request("GET", f"/v2/customers/{customer_id}")
                                customer_data = customer_response.json().get("customer", {})
                                email = customer_data.get("email_address")
                                
                                if email:
                                    cur.execute(
                                        "SELECT id FROM users.users WHERE email = %s",
                                        (email,)
                                    )
                                    user_result = cur.fetchone()
                                    
                                    if user_result:
                                        # Update the user's square_customer_id to the new one
                                        logger.info(f"[WEBHOOK] Found user by email {email}, updating customer ID to {customer_id}")
                                        cur.execute(
                                            "UPDATE users.users SET square_customer_id = %s WHERE email = %s",
                                            (customer_id, email)
                                        )
                                        conn.commit()
                                    else:
                                        logger.warning(f"[WEBHOOK] No user found for email {email}")
                                else:
                                    logger.warning(f"[WEBHOOK] No email found for Square customer {customer_id}")
                            except Exception as e:
                                logger.error(f"[WEBHOOK] Error looking up customer by email for subscription: {e}")
                        
                        if user_result:
                            user_id = user_result[0]
                            
                            # Update the subscription record for this user (either existing or create new)
                            cur.execute(
                                """
                                UPDATE users.customer_membership_plans 
                                SET square_subscription_id = %s,
                                    subscription_status = %s,
                                    subscription_source = 'subscription',
                                    next_billing_date = %s,
                                    subscription_created_at = %s,
                                    status = 'inactive'::membership.membership_status
                                WHERE customer_id = %s 
                                AND subscription_source = 'subscription'
                                AND (square_subscription_id IS NULL OR square_subscription_id = %s)
                                """,
                                (subscription_id, status, next_billing_date, subscription_created_at, user_id, subscription_id)
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