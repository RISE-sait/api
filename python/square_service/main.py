import os
import uuid
from flask import Flask, request, jsonify
from square.client import Client
import requests

# Initialize Square client using environment variables
client = Client(
    access_token=os.getenv("SQUARE_ACCESS_TOKEN"),
    environment=os.getenv("SQUARE_ENVIRONMENT", "sandbox")
)

app = Flask(__name__)

BACKEND_CALLBACK_URL = os.getenv("GO_BACKEND_URL")
LOCATION_ID = os.getenv("SQUARE_LOCATION_ID")


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
            "price_money": {"amount": amount, "currency": "USD"},
            "location_id": LOCATION_ID,
        },
    }

    result = client.checkout.create_payment_link(body)
    if result.is_success():
        notify_backend(result.body)
        return jsonify(result.body)
    return jsonify(result.errors), 400


@app.route("/checkout/subscription", methods=["POST"])
def create_subscription():
    data = request.json or {}
    plan_id = data.get("plan_id")
    customer_id = data.get("customer_id")
    if not plan_id or not customer_id:
        return jsonify({"error": "missing plan_id or customer_id"}), 400

    body = {
        "idempotency_key": str(uuid.uuid4()),
        "location_id": LOCATION_ID,
        "plan_id": plan_id,
        "customer_id": customer_id,
    }

    result = client.subscriptions.create_subscription(body)
    if result.is_success():
        notify_backend(result.body)
        return jsonify(result.body)
    return jsonify(result.errors), 400


@app.route("/webhook", methods=["POST"])
def handle_webhook():
    event = request.json or {}
    notify_backend(event)
    return "", 200


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.getenv("PORT", 8000)))