import os
import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock, AsyncMock

from square_service.app import app

client = TestClient(app)


@pytest.fixture(autouse=True)
def set_env_vars(monkeypatch):
    monkeypatch.setenv("SQUARE_ACCESS_TOKEN", "token")
    monkeypatch.setenv("SQUARE_LOCATION_ID", "loc")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
    monkeypatch.setenv("SQUARE_ENV", "sandbox")
    monkeypatch.setenv("SQUARE_VERSION", "2025-07-16")
    monkeypatch.setenv("SQUARE_WEBHOOK_SIGNATURE_KEY", "secret")
    monkeypatch.setenv("JWT_SECRET", "secret")
    yield


def make_db_conn(row=None):
    cursor = MagicMock()
    cursor.fetchone.return_value = row
    conn = MagicMock()
    conn.cursor.return_value = cursor
    conn.commit = MagicMock()
    conn.close = MagicMock()
    return conn


def test_checkout_membership_success():
    conn = make_db_conn(("price", None, 3))
    mock_resp = AsyncMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = {"payment_link": {"url": "https://pay"}}

    with patch("square_service.db.get_db_conn", return_value=conn), \
         patch("square_service.app.httpx.AsyncClient") as ac, \
         patch("square_service.app.jwt.decode", return_value={"sub": "c1"}):
        ac.return_value.__aenter__.return_value.post.return_value = mock_resp
        resp = client.post(
            "/checkout/membership",
            json={"membership_plan_id": "m1"},
            headers={"Authorization": "Bearer t"},
        )

    assert resp.status_code == 200
    assert resp.json() == {"checkout_url": "https://pay"}


def test_checkout_membership_not_found():
    conn = make_db_conn(None)
    with patch("square_service.db.get_db_conn", return_value=conn), \
         patch("square_service.app.jwt.decode", return_value={"sub": "c1"}):
        resp = client.post(
            "/checkout/membership",
            json={"membership_plan_id": "bad"},
            headers={"Authorization": "Bearer t"},
        )
    assert resp.status_code == 404





def test_webhook_membership_created():
    conn = make_db_conn()
    mock_order = AsyncMock()
    mock_order.status_code = 200
    mock_order.json.return_value = {
        "order": {"metadata": {"user_id": "c1", "membership_plan_id": "m1", "amt_periods": "2"}}
    }

    with patch("square_service.db.get_db_conn", return_value=conn), \
         patch("square_service.app.httpx.AsyncClient") as ac:
        ac.return_value.__aenter__.return_value.get.return_value = mock_order
        event = {
            "type": "payment.updated",
            "data": {"object": {"payment": {"status": "COMPLETED", "order_id": "o1"}}},
        }
        import hmac, hashlib, base64, json
        body = json.dumps(event).encode()
        sig = base64.b64encode(hmac.new(b"secret", body, hashlib.sha1).digest()).decode()
        resp = client.post("/webhook", data=body, headers={"x-square-signature": sig})

    assert resp.status_code == 200
    conn.cursor.assert_called()
    assert resp.json() == {"status": "ok"}