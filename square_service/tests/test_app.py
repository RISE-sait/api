import pytest, os
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock, AsyncMock


# Ensure required env vars exist before importing the app
os.environ.setdefault("SQUARE_ACCESS_TOKEN", "token")
os.environ.setdefault("SQUARE_LOCATION_ID", "loc")
os.environ.setdefault("DATABASE_URL", "postgresql://user:pass@localhost/db")
os.environ.setdefault("SQUARE_ENV", "sandbox")
os.environ.setdefault("SQUARE_VERSION", "2025-07-16")
os.environ.setdefault("SQUARE_WEBHOOK_SIGNATURE_KEY", "secret")
os.environ.setdefault("JWT_SECRET", "secret")
os.environ.setdefault("FRONTEND_URL", "https://example.com")
os.environ.setdefault("WEBHOOK_URL", "https://example.com/webhook")

@pytest.fixture(autouse=True)
def set_env_vars(monkeypatch):
    monkeypatch.setenv("SQUARE_ACCESS_TOKEN", "token")
    monkeypatch.setenv("SQUARE_LOCATION_ID", "loc")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
    monkeypatch.setenv("SQUARE_ENV", "sandbox")
    monkeypatch.setenv("SQUARE_VERSION", "2025-07-16")
    monkeypatch.setenv("SQUARE_WEBHOOK_SIGNATURE_KEY", "secret")
    monkeypatch.setenv("JWT_SECRET", "secret")
    monkeypatch.setenv("FRONTEND_URL", "http://localhost")
    monkeypatch.setenv("WEBHOOK_URL", "http://testserver/webhook")
    yield


@pytest.fixture
def client():
    from square_service.app import app
    return TestClient(app)


def make_db_conn(rows=None):
    cursor = MagicMock()
    if rows is None:
        rows = []
    cursor.fetchone.side_effect = rows
    cursor.__enter__.return_value = cursor
    conn = MagicMock()
    conn.cursor.return_value = cursor
    conn.commit = MagicMock()
    conn.close = MagicMock()
    return conn


def test_checkout_membership_success(client):
    conn = make_db_conn(rows=[("price", None, 3)])
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = {"payment_link": {"url": "https://pay"}}

    with patch("square_service.app.get_db_conn", return_value=conn), \
         patch("square_service.app.httpx.AsyncClient") as ac, \
         patch("square_service.app.jwt.decode", return_value={"sub": "c1"}):
        ac.return_value.__aenter__.return_value.post = AsyncMock(return_value=mock_resp)
        resp = client.post(
            "/checkout/membership",
            json={"membership_plan_id": "m1"},
            headers={"Authorization": "Bearer t"},
        )

    assert resp.status_code == 200
    assert resp.json() == {"checkout_url": "https://pay"}


def test_checkout_membership_not_found(client):
    conn = make_db_conn(rows=[None])

    with patch("square_service.app.get_db_conn", return_value=conn), \
         patch("square_service.app.jwt.decode", return_value={"sub": "c1"}):
        resp = client.post(
            "/checkout/membership",
            json={"membership_plan_id": "bad"},
            headers={"Authorization": "Bearer t"},
        )
    assert resp.status_code == 404


def test_checkout_event_success(client):
    # Event exists and requires payment -> returns checkout link
    conn = make_db_conn(rows=[("price", None)])
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = {"payment_link": {"url": "https://pay"}}

    with patch("square_service.app.get_db_conn", return_value=conn), \
         patch("square_service.app.httpx.AsyncClient") as ac, \
         patch("square_service.app.jwt.decode", return_value={"sub": "c1"}):
        ac.return_value.__aenter__.return_value.post = AsyncMock(return_value=mock_resp)
        resp = client.post(
            "/checkout/event",
            json={"event_id": "e1"},
            headers={"Authorization": "Bearer t"},
        )

    assert resp.status_code == 200
    assert resp.json() == {"checkout_url": "https://pay"}


def test_checkout_event_requires_membership(client):
    # Event requires a membership plan; user lacks it -> 403
    conn = make_db_conn(rows=[("price", "plan"), None])

    with patch("square_service.app.get_db_conn", return_value=conn), \
         patch("square_service.app.jwt.decode", return_value={"sub": "c1"}):
        resp = client.post(
            "/checkout/event",
            json={"event_id": "e1"},
            headers={"Authorization": "Bearer t"},
        )
    assert resp.status_code == 403


def test_checkout_event_not_found(client):
    # Event not found -> 404
    conn = make_db_conn(rows=[None])

    with patch("square_service.app.get_db_conn", return_value=conn), \
         patch("square_service.app.jwt.decode", return_value={"sub": "c1"}):
        resp = client.post(
            "/checkout/event",
            json={"event_id": "e1"},
            headers={"Authorization": "Bearer t"},
        )
    assert resp.status_code == 404


def test_webhook_membership_created(client):
    # Successful webhook flow: signature valid, fetch order, do DB work -> 200
    conn = make_db_conn(rows=[(1,)])
    mock_order = MagicMock()
    mock_order.status_code = 200
    mock_order.json.return_value = {
        "order": {"metadata": {"user_id": "c1", "membership_plan_id": "m1", "amt_periods": "2"}}
    }
    mock_order.raise_for_status = MagicMock()

    with patch("square_service.app.get_db_conn", return_value=conn), \
         patch("square_service.app.httpx.AsyncClient") as ac:
        ac.return_value.__aenter__.return_value.get = AsyncMock(return_value=mock_order)

        # Build a valid HMAC signature over WEBHOOK_URL + raw body
        import os, json, hmac, hashlib, base64
        event = {
            "type": "payment.updated",
            "data": {"object": {"payment": {"status": "COMPLETED", "order_id": "o1"}}},
        }
        body = json.dumps(event).encode()
        webhook_url = os.environ["WEBHOOK_URL"]  # "http://testserver/webhook"
        key = os.environ["SQUARE_WEBHOOK_SIGNATURE_KEY"].encode()
        msg = webhook_url.encode() + body
        sig = base64.b64encode(hmac.new(key, msg, hashlib.sha256).digest()).decode()

        resp = client.post(
            "/webhook",
            content=body,  # send exact bytes
            headers={
                "x-square-hmacsha256-signature": sig,
                "content-type": "application/json",
            },
        )

    assert resp.status_code == 200
    conn.cursor.assert_called()
    assert resp.json() == {"status": "ok"}
