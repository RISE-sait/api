import os
import httpx
import logging
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type

logger = logging.getLogger(__name__)

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
                
                # Import here to avoid circular imports
                from metrics import square_api_errors
                
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
    
    async def create_order(self, order_data: dict) -> dict:
        """Create Square order (used for order templates)."""
        response = await self.make_request("POST", "/v2/orders", json=order_data)
        return response.json()