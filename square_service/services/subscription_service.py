import os
import uuid
import logging
from datetime import datetime, timedelta
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
    
    async def get_subscription_plan_pricing_type(self, subscription_plan_variation_id: str) -> str:
        """Get the pricing type (STATIC or RELATIVE) for a subscription plan variation."""
        try:
            response = await self.square_client.make_request(
                "GET", 
                f"/v2/catalog/object/{subscription_plan_variation_id}"
            )
            
            if response.status_code != 200:
                raise HTTPException(status_code=404, detail="Subscription plan not found")
            
            data = response.json()
            pricing = data.get("object", {}).get("subscription_plan_variation_data", {}).get("phases", [{}])[0].get("pricing", {})
            return pricing.get("type", "RELATIVE")
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error getting pricing type: {e}")
            return "RELATIVE"  # Default to RELATIVE if we can't determine
    
    async def _calculate_subscription_start_date(self, subscription_plan_variation_id: str) -> str:
        """Calculate when the subscription should start based on billing cadence.
        
        The user pays an initial amount for immediate access, then the subscription
        starts billing at the next appropriate interval.
        """
        try:
            response = await self.square_client.make_request("GET", f"/v2/catalog/object/{subscription_plan_variation_id}")
            if response.status_code != 200:
                # Default to next day if we can't determine
                return (datetime.utcnow() + timedelta(days=1)).strftime("%Y-%m-%d")
            
            data = response.json()
            phases = data.get("object", {}).get("subscription_plan_variation_data", {}).get("phases", [])
            
            if not phases:
                return (datetime.utcnow() + timedelta(days=1)).strftime("%Y-%m-%d")
            
            # Get the cadence from the first phase
            phase = phases[0]
            cadence_data = phase.get("cadence", {})
            period = cadence_data.get("period", "")
            frequency = cadence_data.get("frequency", 1)
            
            now = datetime.utcnow()
            
            if period == "DAILY":
                # Daily: start tomorrow (they paid for today)
                start_date = now + timedelta(days=frequency)
            elif period == "WEEKLY":
                # Weekly: start next week
                start_date = now + timedelta(days=7 * frequency)
            elif period == "MONTHLY":
                # Monthly: start next month
                # Add 30 days per month (approximate)
                start_date = now + timedelta(days=30 * frequency)
            elif period == "ANNUAL":
                # Yearly: start next year
                start_date = now + timedelta(days=365 * frequency)
            else:
                # Default to daily if unknown
                start_date = now + timedelta(days=1)
            
            return start_date.strftime("%Y-%m-%d")
            
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error calculating start date: {e}")
            # Default to next day on error
            return (datetime.utcnow() + timedelta(days=1)).strftime("%Y-%m-%d")

    async def calculate_next_billing_date(self, subscription_plan_variation_id: str) -> str:
        """Calculate the next billing date based on subscription plan cadence."""
        try:
            
            response = await self.square_client.make_request("GET", f"/v2/catalog/object/{subscription_plan_variation_id}")
            if response.status_code != 200:
                # Default to monthly if we can't determine
                return (datetime.utcnow() + timedelta(days=30)).strftime("%Y-%m-%d")
            
            data = response.json()
            phases = data.get("object", {}).get("subscription_plan_variation_data", {}).get("phases", [])
            
            if not phases:
                # Default to monthly
                return (datetime.utcnow() + timedelta(days=30)).strftime("%Y-%m-%d")
            
            cadence = phases[0].get("cadence", "MONTHLY")
            
            # Calculate next billing based on cadence
            if cadence == "DAILY":
                next_billing = datetime.utcnow() + timedelta(days=1)
            elif cadence == "WEEKLY":
                next_billing = datetime.utcnow() + timedelta(days=7)
            elif cadence == "BIWEEKLY":
                next_billing = datetime.utcnow() + timedelta(days=14)
            elif cadence == "MONTHLY":
                next_billing = datetime.utcnow() + timedelta(days=30)
            elif cadence == "EVERY_TWO_MONTHS":
                next_billing = datetime.utcnow() + timedelta(days=60)
            elif cadence == "QUARTERLY":
                next_billing = datetime.utcnow() + timedelta(days=90)
            elif cadence == "EVERY_FOUR_MONTHS":
                next_billing = datetime.utcnow() + timedelta(days=120)
            elif cadence == "EVERY_SIX_MONTHS":
                next_billing = datetime.utcnow() + timedelta(days=180)
            elif cadence == "ANNUAL":
                next_billing = datetime.utcnow() + timedelta(days=365)
            else:
                # Default to monthly for unknown cadences
                next_billing = datetime.utcnow() + timedelta(days=30)
            
            return next_billing.strftime("%Y-%m-%d")
            
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error calculating billing date: {e}")
            # Default to monthly on error
            return (datetime.utcnow() + timedelta(days=30)).strftime("%Y-%m-%d")
    
    async def get_item_variation_from_subscription_plan(self, subscription_plan_variation_id: str) -> str:
        """Get the item variation ID associated with a subscription plan."""
        try:
            # Get subscription plan details from Square
            response = await self.square_client.make_request(
                "GET", 
                f"/v2/catalog/object/{subscription_plan_variation_id}"
            )
            
            if response.status_code != 200:
                raise HTTPException(status_code=404, detail="Subscription plan not found in Square")
            
            data = response.json()
            plan_variation = data.get("object", {})
            
            # Get the parent subscription plan ID
            parent_plan_id = plan_variation.get("subscription_plan_variation_data", {}).get("subscription_plan_id")
            if not parent_plan_id:
                raise HTTPException(status_code=400, detail="Invalid subscription plan structure")
            
            # Get the parent subscription plan to find eligible items
            parent_response = await self.square_client.make_request(
                "GET", 
                f"/v2/catalog/object/{parent_plan_id}"
            )
            
            if parent_response.status_code != 200:
                raise HTTPException(status_code=404, detail="Parent subscription plan not found")
            
            parent_data = parent_response.json()
            eligible_item_ids = parent_data.get("object", {}).get("subscription_plan_data", {}).get("eligible_item_ids", [])
            
            if not eligible_item_ids:
                raise HTTPException(status_code=400, detail="No items associated with subscription plan")
            
            # Get the first eligible item's variation
            item_id = eligible_item_ids[0]
            item_response = await self.square_client.make_request(
                "GET", 
                f"/v2/catalog/object/{item_id}"
            )
            
            if item_response.status_code != 200:
                raise HTTPException(status_code=404, detail="Item not found in Square")
            
            item_data = item_response.json()
            variations = item_data.get("object", {}).get("item_data", {}).get("variations", [])
            
            if not variations:
                raise HTTPException(status_code=400, detail="No variations found for item")
            
            # Return the first variation with both ID and price
            variation = variations[0]
            variation_id = variation["id"]
            price_money = variation.get("item_variation_data", {}).get("price_money", {})
            
            return variation_id, price_money
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error getting item variation: {e}")
            raise HTTPException(status_code=500, detail="Error retrieving item variation")
    
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
        
        # Check if subscription plan uses STATIC or RELATIVE pricing
        pricing_type = await self.get_subscription_plan_pricing_type(variation_id)
        item_variation_id = None
        item_price_money = None
        
        if pricing_type == "RELATIVE":
            # Only get item variation for RELATIVE pricing
            item_variation_id, item_price_money = await self.get_item_variation_from_subscription_plan(variation_id)
        
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
        
        # Build order differently based on pricing type
        order_data = {
            "location_id": os.getenv("SQUARE_LOCATION_ID"),
            "metadata": {
                "membership_plan_id": membership_plan_id,
                "user_id": user_id,
                "timezone": timezone
            }
        }
        
        # Include line items based on pricing type
        if pricing_type == "RELATIVE" and item_variation_id:
            # For RELATIVE pricing, don't include base_price_money since it should come from catalog
            order_data["line_items"] = [
                {
                    "catalog_object_id": item_variation_id,
                    "quantity": "1"
                }
            ]
        else:
            # For STATIC pricing, include a minimal line item (required by Square API)
            order_data["line_items"] = [
                {
                    "name": "Subscription",
                    "quantity": "1",
                    "base_price_money": {
                        "amount": 100,  # Minimal amount, will be overridden by subscription plan
                        "currency": "CAD"
                    }
                }
            ]
        
        checkout_payload = {
            "idempotency_key": idempotency_key,
            "order": order_data,
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
            # Handle RELATIVE pricing differently - use direct Subscriptions API
            if pricing_type == "RELATIVE":
                logger.info(f"[SUBSCRIPTION] Using direct Subscriptions API for RELATIVE pricing")
                return await self.create_relative_pricing_subscription(
                    user_id, membership_plan_id, variation_id, user_email, first_name, last_name
                )
            
            # For STATIC pricing, use checkout sessions as before
            logger.info(f"[DEBUG] STATIC pricing checkout payload: {checkout_payload}")
            
            # Create checkout session
            checkout_response = await self.square_client.create_checkout_session(checkout_payload)
            logger.info(f"[DEBUG] Square checkout response: {checkout_response}")
            
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
    
    async def _get_or_create_square_customer(self, user_id: str, user_email: str, first_name: str, last_name: str) -> str:
        """Get existing Square customer ID or create new one. Always returns customer_id."""
        try:
            # Check if user already has a Square customer ID
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    "SELECT square_customer_id FROM users.users WHERE id = %s",
                    (user_id,)
                )
                result = cur.fetchone()
                
                if result and result[0]:
                    existing_customer_id = result[0]
                    # Validate the customer still exists in Square
                    if await self._validate_square_customer(existing_customer_id):
                        logger.info(f"[SUBSCRIPTION] Reusing existing Square customer: {existing_customer_id}")
                        return existing_customer_id
                    else:
                        logger.warning(f"[SUBSCRIPTION] Square customer {existing_customer_id} no longer exists, creating new one")
            
            # No existing customer, create new one
            customer_data = await self.square_client.create_customer(
                user_id=user_id,
                email=user_email,
                first_name=first_name,
                last_name=last_name
            )
            
            if "customer" not in customer_data:
                logger.error(f"[SUBSCRIPTION] Failed to create customer: {customer_data}")
                raise HTTPException(status_code=502, detail="Failed to create customer")
            
            customer_id = customer_data["customer"]["id"]
            logger.info(f"[SUBSCRIPTION] Created new Square customer: {customer_id}")
            
            # Store customer ID in database
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    "UPDATE users.users SET square_customer_id = %s WHERE id = %s",
                    (customer_id, user_id)
                )
                conn.commit()
            
            return customer_id
            
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error getting/creating Square customer: {e}")
            raise HTTPException(status_code=500, detail="Failed to get or create customer")
    
    async def _validate_square_customer(self, customer_id: str) -> bool:
        """Validate that a Square customer ID still exists."""
        try:
            response = await self.square_client.make_request("GET", f"/v2/customers/{customer_id}")
            return response.status_code == 200
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error validating customer {customer_id}: {e}")
            return False
    
    async def create_relative_pricing_subscription(self, user_id: str, membership_plan_id: str, variation_id: str, user_email: str, first_name: str, last_name: str) -> dict:
        """Create subscription directly using Square's Subscriptions API for RELATIVE pricing."""
        try:
            from datetime import datetime, timedelta
            
            # Get or create Square customer (reuse existing if available)
            customer_id = await self._get_or_create_square_customer(user_id, user_email, first_name, last_name)
            
            # Get item variation details for the phases
            item_variation_id, item_price_money = await self.get_item_variation_from_subscription_plan(variation_id)
            
            # Create order template first
            order_template_response = await self.square_client.create_order_template(
                os.getenv("SQUARE_LOCATION_ID"), 
                item_variation_id
            )
            
            if "order" not in order_template_response:
                logger.error(f"[SUBSCRIPTION] Failed to create order template: {order_template_response}")
                raise HTTPException(status_code=502, detail="Failed to create order template")
            
            order_template_id = order_template_response["order"]["id"]
            logger.info(f"[SUBSCRIPTION] Created order template: {order_template_id}")
            
            # Create subscription using direct API with phases for RELATIVE pricing
            subscription_data = {
                "idempotency_key": f"sub-{user_id}-{membership_plan_id}-{os.urandom(8).hex()}",
                "location_id": os.getenv("SQUARE_LOCATION_ID"),
                "plan_variation_id": variation_id,
                "customer_id": customer_id,
                "start_date": await self._calculate_subscription_start_date(variation_id),
                "timezone": "UTC",
                "phases": [
                    {
                        "ordinal": 0,
                        "order_template_id": order_template_id
                    }
                ]
            }
            
            subscription_response = await self.square_client.create_subscription(subscription_data)
            logger.info(f"[SUBSCRIPTION] Direct subscription response: {subscription_response}")
            
            if "subscription" not in subscription_response:
                logger.error(f"[SUBSCRIPTION] Failed to create subscription: {subscription_response}")
                raise HTTPException(status_code=502, detail="Failed to create subscription")
            
            subscription = subscription_response["subscription"]
            subscription_id = subscription["id"]
            
            # Store subscription in database with PENDING_PAYMENT status until payment confirmed
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO users.customer_membership_plans 
                    (customer_id, membership_plan_id, subscription_source, status, square_subscription_id, subscription_status, next_billing_date)
                    VALUES (%s, %s, 'subscription', 'inactive'::membership.membership_status, %s, %s, %s)
                    ON CONFLICT (customer_id, membership_plan_id) 
                    DO UPDATE SET 
                        subscription_source = 'subscription', 
                        status = 'inactive'::membership.membership_status,
                        square_subscription_id = EXCLUDED.square_subscription_id,
                        subscription_status = EXCLUDED.subscription_status,
                        next_billing_date = EXCLUDED.next_billing_date
                    """,
                    (user_id, membership_plan_id, subscription_id, subscription.get("status"), await self.calculate_next_billing_date(variation_id))
                )
                conn.commit()
            
            logger.info(f"[SUBSCRIPTION] Created direct subscription: {subscription_id}")
            
            # Create a checkout session for payment method setup
            payment_checkout_payload = {
                "idempotency_key": f"payment-{user_id}-{subscription_id}-{os.urandom(4).hex()}",
                "order": {
                    "location_id": os.getenv("SQUARE_LOCATION_ID"),
                    "line_items": [
                        {
                            "name": "Subscription Setup",
                            "quantity": "1",
                            "base_price_money": {
                                "amount": item_price_money.get("amount", 1000),
                                "currency": "CAD"
                            }
                        }
                    ]
                },
                "payment_options": {
                    "autocomplete": True,
                    "accept_partial_authorization": False
                },
                "checkout_options": {
                    "redirect_url": f"{os.getenv('FRONTEND_URL', 'http://localhost:3000')}/subscription/success?subscriptionId={subscription_id}",
                    "merchant_support_email": user_email,
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
            
            checkout_response = await self.square_client.create_checkout_session(payment_checkout_payload)
            
            if "payment_link" not in checkout_response:
                logger.error(f"[SUBSCRIPTION] No payment_link in payment checkout: {checkout_response}")
                return {
                    "subscription_id": subscription_id,
                    "status": "PENDING_PAYMENT",
                    "message": "Subscription created but payment setup required",
                    "checkout_url": f"{os.getenv('FRONTEND_URL', 'http://localhost:3000')}/subscription/success?subscriptionId={subscription_id}"
                }
            
            payment_checkout_url = checkout_response["payment_link"]["url"]
            
            return {
                "subscription_id": subscription_id,
                "status": "PENDING_PAYMENT",
                "message": "Complete payment setup to activate subscription",
                "checkout_url": payment_checkout_url
            }
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error creating direct subscription: {e}")
            raise HTTPException(status_code=500, detail="Failed to create subscription")
    
    async def create_relative_pricing_with_payment_first(self, user_id: str, membership_plan_id: str, variation_id: str, user_email: str, first_name: str, last_name: str) -> dict:
        """Create RELATIVE pricing with payment first, subscription after payment success."""
        try:
            from datetime import datetime, timedelta
            
            # Get or create Square customer (reuse existing if available)
            customer_id = await self._get_or_create_square_customer(user_id, user_email, first_name, last_name)
            
            # Get item variation details for pricing
            item_variation_id, item_price_money = await self.get_item_variation_from_subscription_plan(variation_id)
            
            # Store subscription setup data in metadata for webhook processing
            metadata = {
                "user_id": user_id,
                "membership_plan_id": membership_plan_id, 
                "variation_id": variation_id,
                "customer_id": customer_id,
                "item_variation_id": item_variation_id,
                "subscription_setup": "true"
            }
            
            # Create checkout session FIRST - subscription created after payment via webhook
            payment_checkout_payload = {
                "idempotency_key": f"payment-first-{user_id}-{membership_plan_id}-{os.urandom(4).hex()}",
                "order": {
                    "location_id": os.getenv("SQUARE_LOCATION_ID"),
                    "metadata": metadata,
                    "line_items": [
                        {
                            "catalog_object_id": item_variation_id,
                            "quantity": "1"
                        }
                    ]
                },
                "payment_options": {
                    "autocomplete": True,
                    "accept_partial_authorization": False
                },
                "checkout_options": {
                    "redirect_url": f"{os.getenv('FRONTEND_URL', 'http://localhost:3000')}/subscription/success",
                    "merchant_support_email": user_email,
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
            
            checkout_response = await self.square_client.create_checkout_session(payment_checkout_payload)
            
            if "payment_link" not in checkout_response:
                logger.error(f"[SUBSCRIPTION] No payment_link in checkout: {checkout_response}")
                raise HTTPException(status_code=502, detail="Failed to create checkout session")
            
            payment_checkout_url = checkout_response["payment_link"]["url"]
            checkout_id = checkout_response["payment_link"]["id"]
            
            logger.info(f"[SUBSCRIPTION] Created payment-first checkout: {checkout_id}")
            
            return {
                "subscription_id": checkout_id,  # Using checkout_id for now, real subscription_id will be created after payment
                "status": "PAYMENT_PENDING", 
                "message": "Complete payment to activate subscription",
                "checkout_url": payment_checkout_url
            }
            
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error creating payment-first subscription: {e}")
            raise HTTPException(status_code=500, detail="Failed to create subscription checkout")
    
    async def complete_subscription_after_payment(self, order_metadata: dict) -> dict:
        """Complete subscription creation after successful payment using the exact same working API flow."""
        try:
            from datetime import datetime, timedelta
            
            user_id = order_metadata["user_id"]
            membership_plan_id = order_metadata["membership_plan_id"]
            variation_id = order_metadata["variation_id"]
            customer_id = order_metadata["customer_id"]
            item_variation_id = order_metadata["item_variation_id"]
            
            # Create order template using the same working code
            order_template_response = await self.square_client.create_order_template(
                os.getenv("SQUARE_LOCATION_ID"), 
                item_variation_id
            )
            
            if "order" not in order_template_response:
                logger.error(f"[SUBSCRIPTION] Failed to create order template: {order_template_response}")
                raise HTTPException(status_code=502, detail="Failed to create order template")
            
            order_template_id = order_template_response["order"]["id"]
            logger.info(f"[SUBSCRIPTION] Created order template: {order_template_id}")
            
            # Create subscription using the exact same working API structure
            subscription_data = {
                "idempotency_key": f"post-payment-sub-{user_id}-{membership_plan_id}-{os.urandom(8).hex()}",
                "location_id": os.getenv("SQUARE_LOCATION_ID"),
                "plan_variation_id": variation_id,
                "customer_id": customer_id,
                "start_date": await self._calculate_subscription_start_date(variation_id),
                "timezone": "UTC",
                "phases": [
                    {
                        "ordinal": 0,
                        "order_template_id": order_template_id
                    }
                ]
            }
            
            subscription_response = await self.square_client.create_subscription(subscription_data)
            logger.info(f"[SUBSCRIPTION] Post-payment subscription response: {subscription_response}")
            
            if "subscription" not in subscription_response:
                logger.error(f"[SUBSCRIPTION] Failed to create post-payment subscription: {subscription_response}")
                raise HTTPException(status_code=502, detail="Failed to create subscription")
            
            subscription = subscription_response["subscription"]
            subscription_id = subscription["id"]
            
            # Store in database using the same working structure
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO users.customer_membership_plans 
                    (customer_id, membership_plan_id, subscription_source, status, square_subscription_id, subscription_status, next_billing_date)
                    VALUES (%s, %s, 'subscription', 'active'::membership.membership_status, %s, %s, %s)
                    ON CONFLICT (customer_id, membership_plan_id) 
                    DO UPDATE SET 
                        subscription_source = 'subscription', 
                        status = 'active'::membership.membership_status,
                        square_subscription_id = EXCLUDED.square_subscription_id,
                        subscription_status = EXCLUDED.subscription_status,
                        next_billing_date = EXCLUDED.next_billing_date
                    """,
                    (user_id, membership_plan_id, subscription_id, subscription.get("status"), await self.calculate_next_billing_date(variation_id))
                )
                conn.commit()
            
            logger.info(f"[SUBSCRIPTION] Completed post-payment subscription setup: {subscription_id}")
            return {"subscription_id": subscription_id, "status": "ACTIVE"}
            
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error completing post-payment subscription: {e}")
            raise
    
    async def activate_membership_after_payment(self, square_subscription_id: str) -> bool:
        """Activate membership in database after payment confirmation."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    UPDATE users.customer_membership_plans 
                    SET status = 'active'::membership.membership_status
                    WHERE square_subscription_id = %s AND status = 'inactive'::membership.membership_status
                    """,
                    (square_subscription_id,)
                )
                rows_updated = cur.rowcount
                conn.commit()
                
                if rows_updated > 0:
                    logger.info(f"[SUBSCRIPTION] Activated membership for subscription: {square_subscription_id}")
                    return True
                else:
                    logger.warning(f"[SUBSCRIPTION] No pending membership found for subscription: {square_subscription_id}")
                    return False
                    
        except Exception as e:
            logger.error(f"[SUBSCRIPTION] Error activating membership: {e}")
            return False
    
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