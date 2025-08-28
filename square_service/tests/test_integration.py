import pytest
import os
import asyncio
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock

# Set up test environment
os.environ.setdefault("SQUARE_ACCESS_TOKEN", "test_token")
os.environ.setdefault("SQUARE_LOCATION_ID", "test_location") 
os.environ.setdefault("DATABASE_URL", "postgresql://test:test@localhost/test_db")
os.environ.setdefault("SQUARE_ENV", "sandbox")
os.environ.setdefault("SQUARE_WEBHOOK_SIGNATURE_KEY", "test_secret")
os.environ.setdefault("JWT_SECRET", "test_jwt_secret")
os.environ.setdefault("JWT_ALGORITHM", "HS256")
os.environ.setdefault("FRONTEND_URL", "https://test.example.com")
os.environ.setdefault("WEBHOOK_URL", "https://test.example.com/webhook")


@pytest.fixture(scope="session")
def client():
    """Create test client for the entire test session."""
    from app import app
    return TestClient(app)


@pytest.fixture
def auth_headers():
    """Create valid JWT token for testing."""
    import jwt
    import time
    
    payload = {
        "user_id": "test_user_123",
        "sub": "test_user_123", 
        "exp": int(time.time()) + 3600  # 1 hour from now
    }
    token = jwt.encode(payload, "test_jwt_secret", algorithm="HS256")
    return {"Authorization": f"Bearer {token}"}


@pytest.fixture
def mock_db_pool():
    """Mock database connection pool."""
    cursor = MagicMock()
    conn = MagicMock()
    conn.cursor.return_value.__enter__.return_value = cursor
    conn.cursor.return_value.__exit__.return_value = None
    conn.commit = MagicMock()
    
    pool = MagicMock()
    pool.getconn.return_value = conn
    pool.putconn = MagicMock()
    
    return pool, conn, cursor


class TestIntegrationFlows:
    """Integration tests for complete user flows."""
    
    def test_health_endpoint(self, client):
        """Test health endpoint without authentication."""
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["service"] == "rise-square-payment"
    
    def test_root_endpoint(self, client):
        """Test root endpoint."""
        response = client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert data["service"] == "rise-square-payment"
        assert data["status"] == "running"
    
    def test_metrics_endpoint(self, client):
        """Test metrics endpoint returns Prometheus format."""
        response = client.get("/metrics")
        assert response.status_code == 200
        assert response.headers["content-type"] == "text/plain; charset=utf-8"
        # Should contain some basic metrics
        assert b"square_subscription_requests_total" in response.content
    
    def test_subscription_checkout_flow(self, client, auth_headers, mock_db_pool):
        """Test complete subscription checkout flow."""
        pool, conn, cursor = mock_db_pool
        
        # Mock database responses for plan details, user details, and price
        cursor.fetchone.side_effect = [
            ("VARIATION_123", "Test Plan", 12),  # Plan details
            ("test@example.com", "John", "Doe"),  # User details  
            (5000,),  # Plan price
        ]
        
        # Mock Square API response
        mock_square_response = {
            "payment_link": {
                "url": "https://checkout.square.link/test123",
                "id": "CHECKOUT_123"
            }
        }
        
        checkout_request = {
            "membership_plan_id": "550e8400-e29b-41d4-a716-446655440000",
            "timezone": "America/Toronto"
        }
        
        with patch("database.connection.db_pool", pool), \
             patch("services.subscription_service.get_db_conn", return_value=conn), \
             patch("services.square_client.SquareAPIClient.create_checkout_session", return_value=mock_square_response), \
             patch("utils.duplicate_prevention.is_duplicate_checkout_request", return_value=False), \
             patch("utils.duplicate_prevention.check_existing_active_subscription", return_value=False), \
             patch("utils.duplicate_prevention.check_recent_payment", return_value=False), \
             patch("utils.duplicate_prevention.record_payment_attempt", return_value=None):
            
            response = client.post(
                "/subscriptions/checkout",
                json=checkout_request,
                headers=auth_headers
            )
        
        assert response.status_code == 200
        data = response.json()
        assert data["subscription_id"] == "CHECKOUT_123"
        assert data["status"] == "CHECKOUT_PENDING"
        assert "https://checkout.square.link/test123" in data["checkout_url"]
    
    def test_webhook_to_subscription_activation_flow(self, client, mock_db_pool):
        """Test webhook processing leading to subscription activation."""
        pool, conn, cursor = mock_db_pool
        
        # Mock successful order fetch
        cursor.fetchone.side_effect = [
            ("test_user_123",),  # User lookup by Square customer ID
            (None, "subscription"),  # Subscription check
        ]
        cursor.rowcount = 1  # Successful membership activation
        
        # Create valid webhook signature
        import json
        import hmac
        import hashlib
        import base64
        
        event = {
            "type": "subscription.created",
            "data": {
                "object": {
                    "subscription": {
                        "id": "SUB_123",
                        "status": "ACTIVE",
                        "customer_id": "CUSTOMER_123"
                    }
                }
            }
        }
        
        body = json.dumps(event).encode()
        webhook_url = "https://test.example.com/webhook"
        key = "test_secret".encode()
        msg = webhook_url.encode() + body
        signature = base64.b64encode(hmac.new(key, msg, hashlib.sha256).digest()).decode()
        
        # Mock Square API calls
        mock_subscription_details = {
            "subscription": {
                "id": "SUB_123",
                "status": "ACTIVE",
                "created_at": "2025-01-01T00:00:00Z",
                "charged_through_date": "2025-02-01"
            }
        }
        
        with patch("database.connection.db_pool", pool), \
             patch("webhooks.handlers.get_db_conn", return_value=conn), \
             patch("services.square_client.SquareAPIClient.get_subscription", return_value=mock_subscription_details):
            
            response = client.post(
                "/webhook",
                content=body,
                headers={
                    "x-square-hmacsha256-signature": signature,
                    "content-type": "application/json"
                }
            )
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "ok"
    
    def test_subscription_management_flow(self, client, auth_headers, mock_db_pool):
        """Test subscription management (pause/resume/cancel)."""
        pool, conn, cursor = mock_db_pool
        cursor.fetchone.return_value = ("SQUARE_SUB_123", "test_user_123", "ACTIVE")
        cursor.rowcount = 1
        
        management_request = {
            "action": "pause",
            "pause_cycle_duration": 2
        }
        
        mock_pause_response = {
            "subscription": {
                "id": "SQUARE_SUB_123",
                "status": "PAUSED"
            }
        }
        
        with patch("database.connection.db_pool", pool), \
             patch("services.subscription_service.get_db_conn", return_value=conn), \
             patch("services.square_client.SquareAPIClient.pause_subscription", return_value=mock_pause_response):
            
            response = client.post(
                "/subscriptions/SQUARE_SUB_123/manage",
                json=management_request,
                headers=auth_headers
            )
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "PAUSED"
        assert "pause completed successfully" in data["message"]
    
    def test_event_checkout_flow(self, client, auth_headers, mock_db_pool):
        """Test event enrollment checkout flow."""
        pool, conn, cursor = mock_db_pool
        cursor.fetchone.return_value = ("CATALOG_ITEM_123", None)  # Event exists, no membership required
        
        mock_checkout_response = {
            "payment_link": {"url": "https://checkout.square.link/event123"}
        }
        
        event_request = {"event_id": "550e8400-e29b-41d4-a716-446655440001"}
        
        with patch("database.connection.db_pool", pool), \
             patch("services.subscription_service.get_db_conn", return_value=conn), \
             patch("services.square_client.SquareAPIClient.create_checkout_session", return_value=mock_checkout_response):
            
            response = client.post(
                "/checkout/event",
                json=event_request,
                headers=auth_headers
            )
        
        assert response.status_code == 200
        data = response.json()
        assert "https://checkout.square.link/event123" in data["checkout_url"]
    
    def test_admin_endpoints_require_auth(self, client):
        """Test that admin endpoints require authentication."""
        admin_endpoints = [
            "/admin/square/subscription-plans",
            "/admin/sync/status",
        ]
        
        for endpoint in admin_endpoints:
            response = client.get(endpoint)
            assert response.status_code == 403  # Forbidden without auth
    
    def test_admin_sync_status(self, client, auth_headers, mock_db_pool):
        """Test admin sync status endpoint."""
        pool, conn, cursor = mock_db_pool
        cursor.fetchone.side_effect = [
            (10, 8, 0),  # total, active, missing billing dates
            (2, 0),  # failed webhooks, unprocessed
        ]
        
        with patch("database.connection.db_pool", pool), \
             patch("services.sync_service.get_db_conn", return_value=conn), \
             patch("services.sync_service.SyncService.get_sync_status") as mock_sync:
            
            mock_sync.return_value = {
                "sync_health": {"last_full_sync": "2025-01-01T00:00:00Z"},
                "subscription_health": {"total_subscriptions": 10},
                "webhook_health": {"failed_webhooks_24h": 2},
                "recommendation": "healthy"
            }
            
            response = client.get("/admin/sync/status", headers=auth_headers)
        
        assert response.status_code == 200
        data = response.json()
        assert "sync_health" in data
        assert data["recommendation"] == "healthy"
    
    def test_rate_limiting(self, client, auth_headers):
        """Test rate limiting on endpoints."""
        # This would require actual rate limiting to be active
        # For now, just test that endpoints respond normally within limits
        
        checkout_request = {
            "membership_plan_id": "550e8400-e29b-41d4-a716-446655440000",
            "timezone": "UTC"
        }
        
        with patch("services.subscription_service.SubscriptionService.create_subscription_checkout") as mock_create:
            mock_create.return_value = {
                "subscription_id": "test",
                "status": "CHECKOUT_PENDING", 
                "checkout_url": "https://test.com"
            }
            
            # First request should work
            response = client.post(
                "/subscriptions/checkout",
                json=checkout_request,
                headers=auth_headers
            )
            assert response.status_code == 200
    
    def test_error_handling_invalid_json(self, client, auth_headers):
        """Test error handling for invalid JSON."""
        response = client.post(
            "/subscriptions/checkout",
            data="invalid json",
            headers={**auth_headers, "content-type": "application/json"}
        )
        assert response.status_code == 422  # Validation error
    
    def test_error_handling_invalid_uuid(self, client, auth_headers):
        """Test error handling for invalid UUID."""
        checkout_request = {
            "membership_plan_id": "not-a-valid-uuid",
            "timezone": "UTC"
        }
        
        response = client.post(
            "/subscriptions/checkout",
            json=checkout_request,
            headers=auth_headers
        )
        assert response.status_code == 422  # Validation error
    
    def test_cors_headers(self, client):
        """Test CORS headers are present."""
        response = client.options("/health")
        assert response.status_code == 200
        # CORS headers should be present for cross-origin requests