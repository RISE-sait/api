import pytest
import httpx
from unittest.mock import AsyncMock, MagicMock, patch
from fastapi.testclient import TestClient
from square_service.app import app, get_user_id
import json
import uuid

# Test client
client = TestClient(app)

# Mock authentication
@pytest.fixture
def mock_auth():
    with patch('square_service.app.get_user_id') as mock_get_user_id:
        mock_get_user_id.return_value = str(uuid.uuid4())
        yield mock_get_user_id

@pytest.fixture
def mock_db():
    with patch('square_service.app.get_db_conn') as mock_get_conn:
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__.return_value = mock_cursor
        mock_get_conn.return_value.__enter__.return_value = mock_conn
        yield mock_cursor

@pytest.fixture
def mock_square_client():
    with patch('square_service.app.square_client') as mock_client:
        yield mock_client

class TestSubscriptionCreation:
    """Test subscription creation functionality."""
    
    def test_create_subscription_success(self, mock_auth, mock_db, mock_square_client):
        """Test successful subscription creation."""
        user_id = str(uuid.uuid4())
        plan_id = str(uuid.uuid4())
        subscription_id = "sub_123456789"
        
        mock_auth.return_value = user_id
        
        # Mock database responses
        mock_db.fetchone.side_effect = [
            # Plan details query
            ("plan_var_123", None, 12, "Premium Plan"),
            # User email query
            ("user@example.com",),
            # Square customer ID check
            (None,),  # No existing customer
        ]
        
        # Mock Square API responses
        mock_square_client.create_customer.return_value = {
            "customer": {"id": "cust_123456789"}
        }
        
        mock_square_client.create_subscription.return_value = {
            "subscription": {
                "id": subscription_id,
                "status": "ACTIVE",
                "charged_through_date": "2025-09-26T00:00:00Z"
            }
        }
        
        # Test request
        response = client.post(
            "/subscriptions/create",
            json={
                "membership_plan_id": plan_id,
                "card_id": "card_123",
                "timezone": "America/New_York"
            },
            headers={"Authorization": "Bearer valid_token"}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["subscription_id"] == subscription_id
        assert data["status"] == "ACTIVE"
        assert "Premium Plan" in data["message"]
        
        # Verify Square API calls
        mock_square_client.create_customer.assert_called_once()
        mock_square_client.create_subscription.assert_called_once()
    
    def test_create_subscription_missing_plan(self, mock_auth, mock_db):
        """Test subscription creation with missing plan."""
        user_id = str(uuid.uuid4())
        mock_auth.return_value = user_id
        mock_db.fetchone.return_value = None  # No plan found
        
        response = client.post(
            "/subscriptions/create",
            json={"membership_plan_id": str(uuid.uuid4())},
            headers={"Authorization": "Bearer valid_token"}
        )
        
        assert response.status_code == 404
        assert "not found" in response.json()["detail"]
    
    def test_create_subscription_invalid_uuid(self, mock_auth):
        """Test subscription creation with invalid UUID."""
        user_id = str(uuid.uuid4())
        mock_auth.return_value = user_id
        
        response = client.post(
            "/subscriptions/create",
            json={"membership_plan_id": "invalid-uuid"},
            headers={"Authorization": "Bearer valid_token"}
        )
        
        assert response.status_code == 422  # Validation error

class TestSubscriptionManagement:
    """Test subscription management functionality."""
    
    def test_pause_subscription_success(self, mock_auth, mock_db, mock_square_client):
        """Test successful subscription pause."""
        user_id = str(uuid.uuid4())
        subscription_id = "sub_123456789"
        
        mock_auth.return_value = user_id
        mock_db.fetchone.return_value = (1,)  # User owns subscription
        
        mock_square_client.pause_subscription.return_value = {
            "subscription": {"id": subscription_id, "status": "PAUSED"}
        }
        
        response = client.post(
            f"/subscriptions/{subscription_id}/manage",
            json={"action": "pause", "pause_cycle_duration": 2},
            headers={"Authorization": "Bearer valid_token"}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["subscription_id"] == subscription_id
        assert data["status"] == "PAUSED"
        assert "pause completed" in data["message"]
        
        mock_square_client.pause_subscription.assert_called_once()
    
    def test_resume_subscription_success(self, mock_auth, mock_db, mock_square_client):
        """Test successful subscription resume."""
        user_id = str(uuid.uuid4())
        subscription_id = "sub_123456789"
        
        mock_auth.return_value = user_id
        mock_db.fetchone.return_value = (1,)  # User owns subscription
        
        mock_square_client.resume_subscription.return_value = {
            "subscription": {"id": subscription_id, "status": "ACTIVE"}
        }
        
        response = client.post(
            f"/subscriptions/{subscription_id}/manage",
            json={"action": "resume"},
            headers={"Authorization": "Bearer valid_token"}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["subscription_id"] == subscription_id
        assert data["status"] == "ACTIVE"
        
        mock_square_client.resume_subscription.assert_called_once()
    
    def test_cancel_subscription_success(self, mock_auth, mock_db, mock_square_client):
        """Test successful subscription cancellation."""
        user_id = str(uuid.uuid4())
        subscription_id = "sub_123456789"
        
        mock_auth.return_value = user_id
        mock_db.fetchone.return_value = (1,)  # User owns subscription
        
        mock_square_client.cancel_subscription.return_value = {
            "subscription": {"id": subscription_id, "status": "CANCELED"}
        }
        
        response = client.post(
            f"/subscriptions/{subscription_id}/manage",
            json={"action": "cancel", "cancel_reason": "User request"},
            headers={"Authorization": "Bearer valid_token"}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["subscription_id"] == subscription_id
        assert data["status"] == "CANCELED"
        
        mock_square_client.cancel_subscription.assert_called_once()
    
    def test_manage_subscription_not_owned(self, mock_auth, mock_db):
        """Test managing subscription not owned by user."""
        user_id = str(uuid.uuid4())
        subscription_id = "sub_123456789"
        
        mock_auth.return_value = user_id
        mock_db.fetchone.return_value = None  # User doesn't own subscription
        
        response = client.post(
            f"/subscriptions/{subscription_id}/manage",
            json={"action": "pause"},
            headers={"Authorization": "Bearer valid_token"}
        )
        
        assert response.status_code == 404
        assert "not found" in response.json()["detail"]

class TestSubscriptionDetails:
    """Test subscription details retrieval."""
    
    def test_get_subscription_details_success(self, mock_auth, mock_db, mock_square_client):
        """Test successful subscription details retrieval."""
        user_id = str(uuid.uuid4())
        subscription_id = "sub_123456789"
        
        mock_auth.return_value = user_id
        mock_db.fetchone.return_value = (
            "Premium Plan", "ACTIVE", "2025-09-26T00:00:00Z"
        )
        
        mock_square_client.get_subscription.return_value = {
            "subscription": {
                "id": subscription_id,
                "status": "ACTIVE",
                "charged_through_date": "2025-09-26T00:00:00Z",
                "created_at": "2025-08-26T00:00:00Z"
            }
        }
        
        response = client.get(
            f"/subscriptions/{subscription_id}",
            headers={"Authorization": "Bearer valid_token"}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["subscription_id"] == subscription_id
        assert data["plan_name"] == "Premium Plan"
        assert data["status"] == "ACTIVE"
        assert "square_details" in data
        
        mock_square_client.get_subscription.assert_called_once_with(subscription_id)

class TestWebhookHandlers:
    """Test webhook event handling."""
    
    def test_subscription_webhook_handler(self, mock_db):
        """Test subscription webhook event handling."""
        webhook_payload = {
            "type": "subscription.updated",
            "data": {
                "object": {
                    "subscription": {
                        "id": "sub_123456789",
                        "status": "ACTIVE",
                        "customer_id": "cust_123456789"
                    }
                }
            }
        }
        
        with patch('square_service.app.verify_signature', return_value=True):
            response = client.post(
                "/webhook",
                json=webhook_payload,
                headers={"x-square-hmacsha256-signature": "valid_signature"}
            )
        
        assert response.status_code == 200
        assert response.json()["status"] == "ok"
        
        # Verify database update was called
        mock_db.execute.assert_called()
    
    def test_invoice_webhook_handler_payment_made(self, mock_db):
        """Test invoice payment made webhook."""
        webhook_payload = {
            "type": "invoice.payment_made",
            "data": {
                "object": {
                    "invoice": {
                        "id": "inv_123456789",
                        "subscription_id": "sub_123456789",
                        "invoice_status": "PAID"
                    }
                }
            }
        }
        
        with patch('square_service.app.verify_signature', return_value=True):
            response = client.post(
                "/webhook",
                json=webhook_payload,
                headers={"x-square-hmacsha256-signature": "valid_signature"}
            )
        
        assert response.status_code == 200
        assert response.json()["status"] == "ok"
        
        # Verify activation query was called
        mock_db.execute.assert_called()
    
    def test_webhook_invalid_signature(self):
        """Test webhook with invalid signature."""
        webhook_payload = {"type": "subscription.updated"}
        
        with patch('square_service.app.verify_signature', return_value=False):
            response = client.post(
                "/webhook",
                json=webhook_payload,
                headers={"x-square-hmacsha256-signature": "invalid_signature"}
            )
        
        assert response.status_code == 400
        assert "Invalid signature" in response.json()["detail"]

class TestSquareAPIClient:
    """Test Square API client functionality."""
    
    @pytest.mark.asyncio
    async def test_square_client_retry_logic(self):
        """Test Square client retry logic on failures."""
        from square_service.app import SquareAPIClient
        
        client = SquareAPIClient()
        
        # Mock httpx client to raise request error twice, then succeed
        with patch('httpx.AsyncClient') as mock_httpx:
            mock_response = AsyncMock()
            mock_response.status_code = 200
            mock_response.raise_for_status.return_value = None
            
            mock_client_instance = AsyncMock()
            mock_client_instance.request.side_effect = [
                httpx.RequestError("Network error"),
                httpx.RequestError("Network error"),
                mock_response
            ]
            mock_httpx.return_value.__aenter__.return_value = mock_client_instance
            
            response = await client.make_request("GET", "/v2/test")
            
            # Should have been called 3 times due to retries
            assert mock_client_instance.request.call_count == 3
            assert response == mock_response

class TestHealthAndMetrics:
    """Test health check and metrics endpoints."""
    
    def test_health_check(self, mock_db):
        """Test health check endpoint."""
        mock_db.fetchone.side_effect = [(1,), (5,)]  # DB check, active subscriptions count
        
        response = client.get("/health")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["service"] == "rise-square-payment"
        assert "active_subscriptions" in data
        assert "square_api_version" in data
    
    def test_metrics_endpoint(self):
        """Test Prometheus metrics endpoint."""
        response = client.get("/metrics")
        
        assert response.status_code == 200
        assert response.headers["content-type"] == "text/plain; charset=utf-8"
        # Should contain Prometheus metrics format
        assert "# HELP" in response.text or "# TYPE" in response.text

if __name__ == "__main__":
    pytest.main([__file__, "-v"])