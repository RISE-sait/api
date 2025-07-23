import os
import uuid
import requests
import hmac
import hashlib
import base64
from flask import Flask, request, jsonify
from dotenv import load_dotenv

load_dotenv()

ENVIRONMENT = os.getenv("SQUARE_ENVIRONMENT", "sandbox").lower()
BASE_URL = (
    "https://connect.squareupsandbox.com"
    if ENVIRONMENT == "sandbox"
    else "https://connect.squareup.com"
)
ACCESS_TOKEN = os.getenv("SQUARE_ACCESS_TOKEN")

HEADERS = {
    "Authorization": f"Bearer {ACCESS_TOKEN}",
    "Content-Type": "application/json",
    "Square-Version": os.getenv("SQUARE_VERSION", "2024-04-17"),
}

app = Flask(__name__)

BACKEND_CALLBACK_URL = os.getenv("GO_BACKEND_URL")
LOCATION_ID = os.getenv("SQUARE_LOCATION_ID")
WEBHOOK_SIGNATURE_KEY = os.getenv("SQUARE_WEBHOOK_SIGNATURE_KEY")
WEBHOOK_URL = os.getenv("SQUARE_WEBHOOK_URL")


def notify_backend(payload):
    """Send JSON payload to the Go backend if callback URL is set."""
    if not BACKEND_CALLBACK_URL:
        return
    try:
        requests.post(BACKEND_CALLBACK_URL, json=payload, timeout=5)
    except Exception as e:
        app.logger.error(f"failed to notify backend: {e}")


@app.route("/checkout/payment", methods=["POST"])
def create_payment():
    data = request.json or {}
    amount = int(data.get("amount", 0))
    if amount <= 0:
        return jsonify({"error": "invalid amount"}), 400

    body = {
        "idempotency_key": str(uuid.uuid4()),
        "quick_pay": {
            "name": data.get("description", "Payment"),
            "price_money": {"amount": amount, "currency": "CAD"},
            "location_id": LOCATION_ID,
        },
    }

    resp = requests.post(
        f"{BASE_URL}/v2/online-checkout/payment-links",
        headers=HEADERS,
        json=body,
    )
    if resp.ok:
        payload = resp.json()
        notify_backend(payload)
        return jsonify(payload)
    return jsonify({"error": resp.text}), resp.status_code


@app.route("/checkout/subscription", methods=["POST"])
def create_subscription():
    data = request.json or {}
    plan_variation_id = data.get("plan_variation_id") or data.get("plan_id")
    customer_id = data.get("customer_id")
    if not plan_variation_id or not customer_id:
        return jsonify({"error": "missing plan_variation_id or customer_id"}), 400

    body = {
        "idempotency_key": str(uuid.uuid4()),
        "location_id": LOCATION_ID,
        "plan_variation_id": plan_variation_id,
        "customer_id": customer_id,
    }

    start_date = data.get("start_date") or os.getenv("SQUARE_START_DATE")
    if start_date:
        body["start_date"] = start_date

    resp = requests.post(
        f"{BASE_URL}/v2/subscriptions",
        headers=HEADERS,
        json=body,
    )
    if resp.ok:
        payload = resp.json()
        notify_backend(payload)
        return jsonify(payload)
    return jsonify({"error": resp.text}), resp.status_code


@app.route("/webhook", methods=["POST"])
def handle_webhook():
    raw_body = request.get_data(as_text=True)

    if WEBHOOK_SIGNATURE_KEY:
        signature = request.headers.get("x-square-hmacsha256-signature", "")
        payload_to_sign = (WEBHOOK_URL or request.url) + raw_body
        digest = hmac.new(
            WEBHOOK_SIGNATURE_KEY.encode(),
            payload_to_sign.encode(),
            hashlib.sha256,
        ).digest()
        expected = base64.b64encode(digest).decode()
        if not hmac.compare_digest(expected, signature):
            return "invalid signature", 400

    event = request.get_json(silent=True) or {}
    notify_backend(event)
    return "", 200


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.getenv("PORT", 8000)))
