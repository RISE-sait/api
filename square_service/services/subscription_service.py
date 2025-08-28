import os
import uuid
import logging
from typing import Tuple, Optional
from fastapi import HTTPException
from database.connection import get_db_conn
from services.square_client import SquareAPIClient
from utils.duplicate_prevention import (
    is_duplicate_checkout_request, 
    check_existing_active_subscription,
    check_recent_payment,
    generate_idempotency_key,
    record_payment_attempt
)

logger = logging.getLogger(__name__)

class SubscriptionService:
    def __init__(self, square_client: SquareAPIClient):
        self.square_client = square_client
    
    async def get_subscription_plan_details(self, plan_id: str) -> Tuple[str, str, int, str]:
        """Get subscription plan details from database."""
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """
                SELECT stripe_price_id, name, amt_periods
                FROM membership.membership_plans 
                WHERE id = %s
                """,
                (plan_id,)
            )
            result = cur.fetchone()
            
            if not result:
                raise HTTPException(status_code=404, detail="Membership plan not found")
            
            variation_id, name, amt_periods = result
            joining_fee_id = None  # You can add this logic if needed
            
            if not variation_id:
                raise HTTPException(status_code=404, detail="Square subscription plan not configured")
        
        return variation_id, joining_fee_id, amt_periods, name
    
    async def create_subscription_checkout(self, user_id: str, membership_plan_id: str, timezone: str = "UTC") -> dict:
        """Create Square checkout link for subscription signup."""
        logger.info(f"[SUBSCRIPTION] Creating checkout link user={user_id} plan={membership_plan_id}")
        
        # Duplicate payment prevention checks
        if is_duplicate_checkout_request(user_id, membership_plan_id):
            logger.warning(f"[DUPLICATE_PREVENTION] Duplicate checkout request blocked user={user_id} plan={membership_plan_id}")
            raise HTTPException(status_code=429, detail="Please wait before making another purchase")
        
        if await check_existing_active_subscription(user_id, membership_plan_id):
            logger.warning(f"[DUPLICATE_PREVENTION] User already has active subscription user={user_id} plan={membership_plan_id}")
            raise HTTPException(status_code=409, detail="You already have an active subscription for this plan")
        
        # Get plan details
        variation_id, joining_fee_id, amt_periods, plan_name = await self.get_subscription_plan_details(membership_plan_id)
        
        # Get user details and plan price
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute("SELECT email, first_name, last_name FROM users.users WHERE id = %s", (user_id,))
            user_result = cur.fetchone()
            
            if not user_result:
                raise HTTPException(status_code=404, detail="User not found")
            
            user_email, first_name, last_name = user_result
            
            cur.execute("SELECT unit_amount FROM membership.membership_plans WHERE id = %s", (membership_plan_id,))
            price_result = cur.fetchone()
            
            if not price_result:
                raise HTTPException(status_code=404, detail="Membership plan price not found")
            plan_price_cents = price_result[0] or 0
        
        # Check for recent identical payments
        if await check_recent_payment(user_id, membership_plan_id, plan_price_cents):
            logger.warning(f"[DUPLICATE_PREVENTION] Recent payment detected user={user_id} plan={membership_plan_id}")
            raise HTTPException(status_code=409, detail="You have already made this payment recently")
        
        # Create checkout with consistent idempotency
        idempotency_key = generate_idempotency_key(user_id, membership_plan_id)
        checkout_payload = {
            "idempotency_key": idempotency_key,
            "order": {
                "location_id": os.getenv("SQUARE_LOCATION_ID"),
                "line_items": [
                    {
                        "name": plan_name,
                        "quantity": "1",
                        "base_price_money": {
                            "amount": plan_price_cents,
                            "currency": "CAD"
                        }
                    }
                ],
                "metadata": {
                    "membership_plan_id": membership_plan_id,
                    "user_id": user_id,
                    "timezone": timezone
                }
            },
            "payment_options": {
                "autocomplete": True,
                "accept_partial_authorization": False
            },
            "checkout_options": {
                "subscription_plan_id": variation_id,
                "redirect_url": f"{os.getenv('FRONTEND_URL', 'http://localhost:3000')}/subscription/success",
                "merchant_support_email": os.getenv("MERCHANT_EMAIL", user_email),
                "ask_for_shipping_address": False
            },
            "pre_populated_data": {
                "buyer_email": user_email,
                "buyer_address": {
                    "first_name": first_name or "",
                    "last_name": last_name or ""
                }
            }
        }
        
        try:
            # Create checkout session
            checkout_response = await self.square_client.create_checkout_session(checkout_payload)
            
            if "payment_link" not in checkout_response:
                logger.error(f"[SUBSCRIPTION] No payment_link in response: {checkout_response}")
                raise HTTPException(status_code=502, detail="Failed to create checkout session")
            
            checkout_url = checkout_response["payment_link"]["url"]
            checkout_id = checkout_response["payment_link"]["id"]
            
            # Store checkout in database for tracking
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO users.customer_membership_plans 
                    (customer_id, membership_plan_id, subscription_source, status)
                    VALUES (%s, %s, 'subscription', 'inactive'::membership.membership_status)
                    ON CONFLICT (customer_id, membership_plan_id) 
                    DO UPDATE SET subscription_source = 'subscription', status = 'inactive'::membership.membership_status
                    """,
                    (user_id, membership_plan_id)
                )
                conn.commit()
            
            # Record payment attempt for duplicate prevention
            await record_payment_attempt(user_id, membership_plan_id, plan_price_cents, checkout_id)
            
            logger.info(f"[SUBSCRIPTION] Created checkout link user={user_id} checkout_id={checkout_id}")
            
            return {
                "subscription_id": checkout_id,
                "status": "CHECKOUT_PENDING",
                "message": f"Checkout link created for {plan_name}. Complete payment to activate subscription.",
                "checkout_url": checkout_url
            }
            
        except Exception as e:
            error_msg = str(e)
            logger.error(f"[SUBSCRIPTION] Error creating checkout: {e}")
            
            # Handle specific Square API errors
            if "IDEMPOTENCY_KEY_REUSED" in error_msg:
                raise HTTPException(status_code=409, detail="A checkout for this plan is already pending. Please use the existing checkout link or wait before creating a new one.")
            elif "400 Bad Request" in error_msg:
                raise HTTPException(status_code=400, detail="Invalid checkout request. Please check your plan configuration.")
            else:
                raise HTTPException(status_code=502, detail="Payment provider error")
    
    async def manage_subscription(self, subscription_id: str, action: str, **kwargs) -> dict:
        """Manage subscription actions (pause, resume, cancel, etc.)."""
        logger.info(f"[SUBSCRIPTION] Managing subscription {subscription_id} action={action}")
        
        # Get subscription from database
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute(
                """
                SELECT square_subscription_id, customer_id, subscription_status 
                FROM users.customer_membership_plans 
                WHERE square_subscription_id = %s OR checkout_id = %s
                """,
                (subscription_id, subscription_id)
            )
            result = cur.fetchone()
            
            if not result:
                raise HTTPException(status_code=404, detail="Subscription not found")
            
            square_subscription_id, user_id, current_status = result
            
            if not square_subscription_id:
                raise HTTPException(status_code=400, detail="Subscription not yet created in Square")
        
        try:
            response_data = None
            new_status = None
            
            if action == "pause":
                pause_data = {
                    "pause_effective_date": kwargs.get("pause_effective_date"),
                    "pause_cycle_duration": kwargs.get("pause_cycle_duration", 1)
                }
                response_data = await self.square_client.pause_subscription(square_subscription_id, pause_data)
                new_status = "PAUSED"
            
            elif action == "resume":
                resume_data = {
                    "resume_effective_date": kwargs.get("resume_effective_date")
                }
                response_data = await self.square_client.resume_subscription(square_subscription_id, resume_data)
                new_status = "ACTIVE"
            
            elif action == "cancel":
                response_data = await self.square_client.cancel_subscription(square_subscription_id)
                new_status = "CANCELED"
            
            elif action == "update_card":
                # This would require additional Square API calls for card management
                raise HTTPException(status_code=501, detail="Card update not yet implemented")
            
            else:
                raise HTTPException(status_code=400, detail="Invalid action")
            
            # Update database if Square operation succeeded
            if response_data and "subscription" in response_data:
                with get_db_conn() as conn, conn.cursor() as cur:
                    db_status = "active" if new_status == "ACTIVE" else "inactive"
                    cur.execute(
                        """
                        UPDATE users.customer_membership_plans 
                        SET subscription_status = %s, status = %s::membership.membership_status
                        WHERE square_subscription_id = %s
                        """,
                        (new_status, db_status, square_subscription_id)
                    )
                    conn.commit()
            
            logger.info(f"[SUBSCRIPTION] Action {action} completed for {subscription_id}")
            
            return {
                "subscription_id": subscription_id,
                "status": new_status or "UPDATED", 
                "message": f"Subscription {action} completed successfully"
            }
            
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error managing subscription {subscription_id}: {e}")
            raise HTTPException(status_code=502, detail="Subscription management error")
    
    async def get_subscription_details(self, subscription_id: str, user_id: str) -> dict:
        """Get subscription details from Square API."""
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
                    raise HTTPException(status_code=404, detail="Subscription not found or not owned by user")
                
                plan_name, subscription_status, next_billing_date = result
            
            # Fetch detailed subscription from Square
            detailed_subscription = await self.square_client.get_subscription(subscription_id)
            sub_details = detailed_subscription.get("subscription", {})
            
            logger.info(f"[MANUAL] Detailed subscription response: {sub_details}")
            
            # Extract billing date
            square_next_billing = (sub_details.get("charged_through_date") or 
                               sub_details.get("next_billing_date") or
                               sub_details.get("billing_anchor_date"))
            subscription_created_at = sub_details.get("created_at")
            square_status = sub_details.get("status")
            
            # Update database with latest info
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """UPDATE users.customer_membership_plans 
                       SET next_billing_date = %s,
                           subscription_created_at = %s,
                           subscription_status = %s
                       WHERE square_subscription_id = %s AND customer_id = %s""",
                    (square_next_billing, subscription_created_at, square_status, subscription_id, user_id)
                )
                conn.commit()
            
            return {
                "subscription_id": subscription_id,
                "plan_name": plan_name,
                "next_billing_date": square_next_billing,
                "subscription_created_at": subscription_created_at,
                "status": square_status,
                "full_response": sub_details
            }
            
        except HTTPException:
            raise
        except Exception as e:
            logger.exception(f"[SUBSCRIPTION] Error getting subscription details: {e}")
            raise HTTPException(status_code=500, detail="Error retrieving subscription details")
    
    async def create_event_checkout(self, user_id: str, event_id: str) -> dict:
        """Create checkout for event enrollment."""
        logger.info(f"[EVENT CHECKOUT] user={user_id} event={event_id}")

        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                # Fetch price and required membership for the event
                cur.execute(
                    "SELECT price_id, required_membership_plan_id FROM events.events WHERE id = %s",
                    (event_id,),
                )
                row = cur.fetchone()
                if not row:
                    raise HTTPException(status_code=404, detail="Event not found")
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
                    (user_id, event_id),
                )
                conn.commit()

            # Generate checkout payload
            metadata = {"customer_id": user_id, "event_id": event_id}
            payload = {
                "idempotency_key": f"event-{user_id}-{event_id}",
                "order": {
                    "location_id": os.getenv("SQUARE_LOCATION_ID"),
                    "line_items": [
                        {
                            "catalog_object_id": price_id,
                            "quantity": "1"
                        }
                    ],
                    "metadata": metadata
                },
                "payment_options": {"autocomplete": True},
                "checkout_options": {
                    "redirect_url": f"{os.getenv('FRONTEND_URL', 'http://localhost:3000')}/events/{event_id}/success"
                }
            }

            checkout_response = await self.square_client.create_checkout_session(payload)
            checkout_url = checkout_response["payment_link"]["url"]
            
            logger.info(f"[EVENT CHECKOUT] payment_link created user={user_id} event={event_id} price_id={price_id} url={checkout_url}")
            return {"checkout_url": checkout_url}
            
        except HTTPException:
            raise
        except Exception as e:
            logger.exception(f"[EVENT CHECKOUT] Error: {e}")
            raise HTTPException(status_code=502, detail="Payment provider error")