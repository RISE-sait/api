import pytest
import os
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi import HTTPException

# Set up environment before imports
os.environ.setdefault("SQUARE_ACCESS_TOKEN", "test_token")
os.environ.setdefault("SQUARE_LOCATION_ID", "test_location")
os.environ.setdefault("DATABASE_URL", "postgresql://user:pass@localhost/test_db")
os.environ.setdefault("SQUARE_ENV", "sandbox")
os.environ.setdefault("SQUARE_VERSION", "2025-07-16")
os.environ.setdefault("SQUARE_WEBHOOK_SIGNATURE_KEY", "test_secret")
os.environ.setdefault("JWT_SECRET", "test_jwt_secret")
os.environ.setdefault("JWT_ALGORITHM", "HS256")
os.environ.setdefault("FRONTEND_URL", "https://test.example.com")
os.environ.setdefault("WEBHOOK_URL", "https://test.example.com/webhook")

from services.subscription_service import SubscriptionService
from services.square_client import SquareAPIClient


@pytest.fixture
def mock_square_client():
    """Create a mock Square API client."""
    return MagicMock(spec=SquareAPIClient)


@pytest.fixture
def subscription_service(mock_square_client):
    """Create a subscription service with mocked dependencies."""
    return SubscriptionService(mock_square_client)


@pytest.fixture
def mock_db_connection():
    """Mock database connection with cursor."""
    cursor = MagicMock()
    conn = MagicMock()
    conn.cursor.return_value.__enter__.return_value = cursor
    conn.cursor.return_value.__exit__.return_value = None
    return conn, cursor


class TestSubscriptionService:
    
    @pytest.mark.asyncio
    async def test_get_subscription_plan_details_success(self, subscription_service, mock_db_connection):
        """Test successful subscription plan details retrieval."""
        conn, cursor = mock_db_connection
        cursor.fetchone.return_value = ("VARIATION_123", "Test Plan", 12)
        
        with patch("services.subscription_service.get_db_conn", return_value=conn):
            result = await subscription_service.get_subscription_plan_details("plan_123")
            
        assert result == ("VARIATION_123", None, 12, "Test Plan")
        cursor.execute.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_get_subscription_plan_details_not_found(self, subscription_service, mock_db_connection):
        """Test subscription plan not found scenario."""
        conn, cursor = mock_db_connection
        cursor.fetchone.return_value = None
        
        with patch("services.subscription_service.get_db_conn", return_value=conn):
            with pytest.raises(HTTPException) as exc_info:
                await subscription_service.get_subscription_plan_details("invalid_plan")
            
        assert exc_info.value.status_code == 404
        assert "Membership plan not found" in str(exc_info.value.detail)
    
    @pytest.mark.asyncio
    async def test_create_subscription_checkout_success(self, subscription_service, mock_square_client, mock_db_connection):
        """Test successful subscription checkout creation."""
        conn, cursor = mock_db_connection
        
        # Mock database responses
        cursor.fetchone.side_effect = [
            ("VARIATION_123", "Test Plan", 12),  # Plan details
            ("test@example.com", "John", "Doe"),  # User details
            (5000,),  # Plan price
        ]
        
        # Mock Square API response
        mock_square_client.create_checkout_session.return_value = {
            "payment_link": {
                "url": "https://checkout.square.link/test",
                "id": "CHECKOUT_123"
            }
        }
        
        with patch("services.subscription_service.get_db_conn", return_value=conn), \
             patch("services.subscription_service.is_duplicate_checkout_request", return_value=False), \
             patch("services.subscription_service.check_existing_active_subscription", return_value=False), \
             patch("services.subscription_service.check_recent_payment", return_value=False), \
             patch("services.subscription_service.record_payment_attempt", return_value=None):
            
            result = await subscription_service.create_subscription_checkout(
                user_id="user_123",
                membership_plan_id="plan_123",
                timezone="UTC"
            )
        
        assert result["subscription_id"] == "CHECKOUT_123"
        assert result["status"] == "CHECKOUT_PENDING"
        assert "https://checkout.square.link/test" in result["checkout_url"]
        mock_square_client.create_checkout_session.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_create_subscription_checkout_duplicate_prevention(self, subscription_service, mock_db_connection):
        """Test duplicate checkout prevention."""
        with patch("services.subscription_service.is_duplicate_checkout_request", return_value=True):
            with pytest.raises(HTTPException) as exc_info:
                await subscription_service.create_subscription_checkout(
                    user_id="user_123",
                    membership_plan_id="plan_123"
                )
            
        assert exc_info.value.status_code == 429
        assert "wait before making another purchase" in str(exc_info.value.detail)
    
    @pytest.mark.asyncio
    async def test_manage_subscription_pause(self, subscription_service, mock_square_client, mock_db_connection):
        """Test subscription pause functionality."""
        conn, cursor = mock_db_connection
        cursor.fetchone.return_value = ("SQUARE_SUB_123", "user_123", "ACTIVE")
        
        mock_square_client.pause_subscription.return_value = {
            "subscription": {"id": "SQUARE_SUB_123", "status": "PAUSED"}
        }
        
        with patch("services.subscription_service.get_db_conn", return_value=conn):
            result = await subscription_service.manage_subscription(
                subscription_id="SQUARE_SUB_123",
                action="pause",
                pause_cycle_duration=1
            )
        
        assert result["status"] == "PAUSED"
        assert "pause completed successfully" in result["message"]
        mock_square_client.pause_subscription.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_manage_subscription_not_found(self, subscription_service, mock_db_connection):
        """Test managing non-existent subscription."""
        conn, cursor = mock_db_connection
        cursor.fetchone.return_value = None
        
        with patch("services.subscription_service.get_db_conn", return_value=conn):
            with pytest.raises(HTTPException) as exc_info:
                await subscription_service.manage_subscription(
                    subscription_id="INVALID_SUB",
                    action="pause"
                )
            
        assert exc_info.value.status_code == 404
        assert "Subscription not found" in str(exc_info.value.detail)
    
    @pytest.mark.asyncio
    async def test_create_event_checkout_success(self, subscription_service, mock_square_client, mock_db_connection):
        """Test successful event checkout creation."""
        conn, cursor = mock_db_connection
        cursor.fetchone.return_value = ("CATALOG_123", None)  # Event with no required membership
        
        mock_square_client.create_checkout_session.return_value = {
            "payment_link": {"url": "https://checkout.square.link/event"}
        }
        
        with patch("services.subscription_service.get_db_conn", return_value=conn):
            result = await subscription_service.create_event_checkout(
                user_id="user_123",
                event_id="event_123"
            )
        
        assert "https://checkout.square.link/event" in result["checkout_url"]
        mock_square_client.create_checkout_session.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_create_event_checkout_event_not_found(self, subscription_service, mock_db_connection):
        """Test event checkout with non-existent event."""
        conn, cursor = mock_db_connection
        cursor.fetchone.return_value = None
        
        with patch("services.subscription_service.get_db_conn", return_value=conn):
            with pytest.raises(HTTPException) as exc_info:
                await subscription_service.create_event_checkout(
                    user_id="user_123",
                    event_id="invalid_event"
                )
            
        assert exc_info.value.status_code == 404
        assert "Event not found" in str(exc_info.value.detail)