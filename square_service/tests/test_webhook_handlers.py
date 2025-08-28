import pytest
import os
from unittest.mock import patch, MagicMock, AsyncMock

# Set up environment
os.environ.setdefault("SQUARE_ACCESS_TOKEN", "test_token")
os.environ.setdefault("SQUARE_LOCATION_ID", "test_location")
os.environ.setdefault("DATABASE_URL", "postgresql://user:pass@localhost/test_db")
os.environ.setdefault("SQUARE_ENV", "sandbox")
os.environ.setdefault("SQUARE_WEBHOOK_SIGNATURE_KEY", "test_secret")
os.environ.setdefault("JWT_SECRET", "test_jwt_secret")
os.environ.setdefault("FRONTEND_URL", "https://test.example.com")
os.environ.setdefault("WEBHOOK_URL", "https://test.example.com/webhook")

from webhooks.handlers import WebhookHandler
from services.square_client import SquareAPIClient


@pytest.fixture
def mock_square_client():
    """Create a mock Square API client."""
    client = MagicMock(spec=SquareAPIClient)
    client.get_subscription = AsyncMock()
    client.make_request = AsyncMock()
    return client


@pytest.fixture
def webhook_handler(mock_square_client):
    """Create webhook handler with mocked dependencies."""
    return WebhookHandler(mock_square_client)


@pytest.fixture
def mock_db_connection():
    """Mock database connection."""
    cursor = MagicMock()
    conn = MagicMock()
    conn.cursor.return_value.__enter__.return_value = cursor
    conn.cursor.return_value.__exit__.return_value = None
    conn.commit = MagicMock()
    return conn, cursor


class TestWebhookHandler:
    
    @pytest.mark.asyncio
    async def test_handle_payment_webhook_completed(self, webhook_handler, mock_square_client, mock_db_connection):
        """Test handling completed payment webhook."""
        conn, cursor = mock_db_connection
        
        # Mock order API response
        mock_response = MagicMock()
        mock_response.json.return_value = {
            "order": {
                "metadata": {
                    "user_id": "user_123",
                    "membership_plan_id": "plan_123"
                }
            }
        }
        mock_square_client.make_request.return_value = mock_response
        
        event = {
            "data": {
                "object": {
                    "payment": {
                        "id": "PAYMENT_123",
                        "status": "COMPLETED",
                        "order_id": "ORDER_123",
                        "amount_money": {"amount": 5000, "currency": "CAD"}
                    }
                }
            }
        }
        
        with patch("webhooks.handlers.get_db_conn", return_value=conn), \
             patch("webhooks.handlers.is_duplicate_payment_webhook", return_value=False), \
             patch.object(webhook_handler, '_activate_membership_and_create_subscription', return_value={"status": "ok"}):
            
            result = await webhook_handler.handle_payment_webhook(event)
        
        assert result["status"] == "ok"
        mock_square_client.make_request.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_handle_payment_webhook_duplicate(self, webhook_handler):
        """Test duplicate payment webhook prevention."""
        event = {
            "data": {
                "object": {
                    "payment": {
                        "id": "PAYMENT_123",
                        "status": "COMPLETED",
                        "order_id": "ORDER_123"
                    }
                }
            }
        }
        
        with patch("webhooks.handlers.is_duplicate_payment_webhook", return_value=True):
            result = await webhook_handler.handle_payment_webhook(event)
        
        assert result["status"] == "duplicate_ignored"
    
    @pytest.mark.asyncio
    async def test_handle_subscription_webhook_created(self, webhook_handler, mock_square_client, mock_db_connection):
        """Test handling subscription created webhook."""
        conn, cursor = mock_db_connection
        cursor.fetchone.return_value = ("user_123",)  # User found
        cursor.rowcount = 1
        
        # Mock subscription details
        mock_square_client.get_subscription.return_value = {
            "subscription": {
                "id": "SUB_123",
                "status": "ACTIVE",
                "created_at": "2025-01-01T00:00:00Z",
                "charged_through_date": "2025-02-01"
            }
        }
        
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
        
        with patch("webhooks.handlers.get_db_conn", return_value=conn):
            result = await webhook_handler.handle_subscription_webhook(event)
        
        assert result["status"] == "ok"
        mock_square_client.get_subscription.assert_called_once_with("SUB_123")
    
    @pytest.mark.asyncio
    async def test_handle_subscription_webhook_updated(self, webhook_handler, mock_db_connection):
        """Test handling subscription updated webhook."""
        conn, cursor = mock_db_connection
        cursor.rowcount = 1
        
        event = {
            "type": "subscription.updated",
            "data": {
                "object": {
                    "subscription": {
                        "id": "SUB_123",
                        "status": "PAUSED",
                        "customer_id": "CUSTOMER_123"
                    }
                }
            }
        }
        
        with patch("webhooks.handlers.get_db_conn", return_value=conn):
            result = await webhook_handler.handle_subscription_webhook(event)
        
        assert result["status"] == "ok"
        # Verify database update was called
        cursor.execute.assert_called()
        conn.commit.assert_called()
    
    @pytest.mark.asyncio
    async def test_handle_invoice_webhook_payment_made(self, webhook_handler, mock_square_client, mock_db_connection):
        """Test handling invoice payment made webhook."""
        conn, cursor = mock_db_connection
        
        # Mock subscription details
        mock_square_client.get_subscription.return_value = {
            "subscription": {
                "plan_variation_id": "VARIATION_123",
                "monthly_billing_anchor_date": 15
            }
        }
        
        # Mock plan details
        mock_square_client.get_subscription_plan.return_value = {
            "object": {
                "subscription_plan_variation_data": {
                    "phases": [{"cadence": "MONTHLY"}]
                }
            }
        }
        
        event = {
            "type": "invoice.payment_made",
            "data": {
                "object": {
                    "invoice": {
                        "id": "INVOICE_123",
                        "subscription_id": "SUB_123",
                        "invoice_status": "PAID"
                    }
                }
            }
        }
        
        with patch("webhooks.handlers.get_db_conn", return_value=conn):
            result = await webhook_handler.handle_invoice_webhook(event)
        
        assert result["status"] == "ok"
        cursor.execute.assert_called()
        conn.commit.assert_called()
    
    @pytest.mark.asyncio
    async def test_handle_invoice_webhook_failed(self, webhook_handler, mock_db_connection):
        """Test handling invoice failed webhook."""
        conn, cursor = mock_db_connection
        
        event = {
            "type": "invoice.failed",
            "data": {
                "object": {
                    "invoice": {
                        "id": "INVOICE_123",
                        "subscription_id": "SUB_123",
                        "invoice_status": "FAILED"
                    }
                }
            }
        }
        
        with patch("webhooks.handlers.get_db_conn", return_value=conn):
            result = await webhook_handler.handle_invoice_webhook(event)
        
        assert result["status"] == "ok"
        # Should update subscription to DELINQUENT
        cursor.execute.assert_called()
        conn.commit.assert_called()
    
    @pytest.mark.asyncio
    async def test_handle_webhook_missing_subscription_data(self, webhook_handler):
        """Test handling webhook with missing subscription data."""
        event = {
            "type": "subscription.created",
            "data": {"object": {}}  # Missing subscription data
        }
        
        result = await webhook_handler.handle_subscription_webhook(event)
        assert result["status"] == "ignored"
    
    @pytest.mark.asyncio
    async def test_activate_membership_and_create_subscription(self, webhook_handler, mock_db_connection):
        """Test membership activation flow."""
        conn, cursor = mock_db_connection
        cursor.rowcount = 1  # Simulate successful update
        cursor.fetchone.return_value = (None, "subscription")  # No existing subscription
        
        with patch("webhooks.handlers.get_db_conn", return_value=conn):
            result = await webhook_handler._activate_membership_and_create_subscription(
                user_id="user_123",
                membership_plan_id="plan_123"
            )
        
        assert result["status"] == "ok"
        assert result["did_anything"] is True
        cursor.execute.assert_called()
        conn.commit.assert_called()