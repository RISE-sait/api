from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel
import os
import psycopg2
import httpx
from typing import Optional
from dotenv import load_dotenv
from datetime import datetime
from dateutil.relativedelta import relativedelta

load_dotenv()

app = FastAPI()

SQUARE_BASE_URL = "https://connect.squareupsandbox.com" if os.getenv("SQUARE_ENV") != "production" else "https://connect.squareup.com"
HEADERS = {
    "Square-Version": "2025-07-16",
    "Authorization": f"Bearer {os.getenv('SQUARE_ACCESS_TOKEN')}",
    "Content-Type": "application/json"
}

def get_db_conn():
    url = os.getenv("DATABASE_URL")
    if not url:
        raise RuntimeError("DATABASE_URL not configured")
    return psycopg2.connect(url)

def generate_checkout_payload(location_id, catalog_ids, metadata):
    return {
        "idempotency_key": os.urandom(16).hex(),
        "order": {
            "location_id": location_id,
            "line_items": [{"catalog_object_id": cid, "quantity": "1"} for cid in catalog_ids],
            "metadata": metadata,
        },
        "ask_for_shipping_address": False,
    }

class CheckoutRequest(BaseModel):
    membership_plan_id: str
    customer_id: str
    discount_code: Optional[str] = None

class ProgramCheckoutRequest(BaseModel):
    program_id: str
    customer_id: str

class EventCheckoutRequest(BaseModel):
    event_id: str
    customer_id: str

@app.post("/checkout/membership")
async def checkout_membership(req: CheckoutRequest):
    conn = get_db_conn()
    cur = conn.cursor()
    cur.execute(
        "SELECT stripe_price_id, stripe_joining_fee_id, amt_periods "
        "FROM membership.membership_plans WHERE id = %s",
        (req.membership_plan_id,)
    )
    row = cur.fetchone()
    if not row:
        raise HTTPException(status_code=404, detail='membership plan not found')
    price_id, joining_fee_id, amt_periods = row

    catalog_ids = [price_id]
    if joining_fee_id:
        catalog_ids.append(joining_fee_id)

    payload = generate_checkout_payload(
        os.getenv("SQUARE_LOCATION_ID"),
        catalog_ids,
        {
            "user_id": req.customer_id,
            "membership_plan_id": req.membership_plan_id,
            "amt_periods": str(amt_periods)
        }
    )

    async with httpx.AsyncClient() as client:
        resp = await client.post(f"{SQUARE_BASE_URL}/v2/online-checkout/payment-links", headers=HEADERS, json=payload)
        if resp.status_code >= 400:
            raise HTTPException(status_code=500, detail=resp.text)
        return {"checkout_url": resp.json()["payment_link"]["url"]}

@app.post("/checkout/program")
async def checkout_program(req: ProgramCheckoutRequest):
    conn = get_db_conn()
    cur = conn.cursor()
    cur.execute(
        """INSERT INTO program.customer_enrollment (customer_id, program_id, payment_expired_at, payment_status)
           VALUES (%s, %s, NOW() + interval '10 minute', 'pending')
           ON CONFLICT (customer_id, program_id)
           DO UPDATE SET payment_expired_at = EXCLUDED.payment_expired_at,
                         payment_status = EXCLUDED.payment_status
           WHERE program.customer_enrollment.payment_status != 'paid'""",
        (req.customer_id, req.program_id),
    )
    cur.execute(
        """SELECT f.stripe_price_id, p.pay_per_event FROM program.fees f
           JOIN program.programs p ON p.id = f.program_id
           WHERE f.membership_id = (
               SELECT mp.membership_id FROM users.customer_membership_plans cmp
               JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
               WHERE customer_id = %s AND status = 'active'
               ORDER BY cmp.start_date DESC LIMIT 1)
           AND f.program_id = %s""",
        (req.customer_id, req.program_id),
    )
    row = cur.fetchone()
    if not row:
        raise HTTPException(status_code=403, detail='customer not eligible')
    price_id, pay_per_event = row
    if pay_per_event:
        raise HTTPException(status_code=400, detail='program is pay-per-event')

    payload = generate_checkout_payload(
        os.getenv("SQUARE_LOCATION_ID"),
        [price_id],
        {"user_id": req.customer_id, "program_id": req.program_id}
    )

    async with httpx.AsyncClient() as client:
        resp = await client.post(f"{SQUARE_BASE_URL}/v2/online-checkout/payment-links", headers=HEADERS, json=payload)
        if resp.status_code >= 400:
            raise HTTPException(status_code=500, detail=resp.text)
        return {"checkout_url": resp.json()["payment_link"]["url"]}

@app.post("/checkout/event")
async def checkout_event(req: EventCheckoutRequest):
    conn = get_db_conn()
    cur = conn.cursor()
    cur.execute("SELECT program_id FROM events.events WHERE id = %s", (req.event_id,))
    row = cur.fetchone()
    if not row:
        raise HTTPException(status_code=404, detail='event not found')
    program_id = row[0]

    cur.execute(
        """INSERT INTO events.customer_enrollment (customer_id, event_id, payment_expired_at, payment_status)
           VALUES (%s, %s, NOW() + interval '10 minute', 'pending')
           ON CONFLICT (customer_id, event_id)
           DO UPDATE SET payment_expired_at = EXCLUDED.payment_expired_at,
                         payment_status = EXCLUDED.payment_status
           WHERE events.customer_enrollment.payment_status != 'paid'""",
        (req.customer_id, req.event_id),
    )
    cur.execute(
        """SELECT f.stripe_price_id, p.pay_per_event FROM program.fees f
           JOIN program.programs p ON p.id = f.program_id
           WHERE f.membership_id = (
               SELECT mp.membership_id FROM users.customer_membership_plans cmp
               JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
               WHERE customer_id = %s AND status = 'active'
               ORDER BY cmp.start_date DESC LIMIT 1)
           AND f.program_id = %s""",
        (req.customer_id, program_id),
    )
    row = cur.fetchone()
    if not row:
        raise HTTPException(status_code=403, detail='customer not eligible')
    price_id, pay_per_event = row
    if not pay_per_event:
        raise HTTPException(status_code=400, detail='event is pay-per-program')

    payload = generate_checkout_payload(
        os.getenv("SQUARE_LOCATION_ID"),
        [price_id],
        {"user_id": req.customer_id, "event_id": req.event_id}
    )

    async with httpx.AsyncClient() as client:
        resp = await client.post(f"{SQUARE_BASE_URL}/v2/online-checkout/payment-links", headers=HEADERS, json=payload)
        if resp.status_code >= 400:
            raise HTTPException(status_code=500, detail=resp.text)
        return {"checkout_url": resp.json()["payment_link"]["url"]}

class WebhookEvent(BaseModel):
    type: str
    data: dict

@app.post("/webhook")
async def handle_webhook(event: WebhookEvent):
    if event.type == "payment.updated":
        payment = event.data.get("object", {}).get("payment")
        if payment and payment.get("status") == "COMPLETED":
            order_id = payment.get("order_id")
            async with httpx.AsyncClient() as client:
                resp = await client.get(f"{SQUARE_BASE_URL}/v2/orders/{order_id}", headers=HEADERS)
                if resp.status_code >= 400:
                    raise HTTPException(status_code=500, detail=resp.text)
                metadata = resp.json()["order"].get("metadata", {})

            customer_id = metadata.get("user_id")
            plan_id = metadata.get("membership_plan_id")
            program_id = metadata.get("program_id")
            event_id = metadata.get("event_id")

            conn = get_db_conn()
            cur = conn.cursor()

            if customer_id and plan_id:
                renewal_date = None
                if amt := metadata.get("amt_periods"):
                    try:
                        renewal_date = datetime.utcnow() + relativedelta(months=int(amt))
                    except Exception:
                        renewal_date = None
                cur.execute(
                    "INSERT INTO users.customer_membership_plans (customer_id, membership_plan_id, renewal_date) VALUES (%s, %s, %s)",
                    (customer_id, plan_id, renewal_date),
                )

            elif customer_id and program_id:
                cur.execute(
                    "UPDATE program.customer_enrollment SET payment_status='paid' WHERE customer_id=%s AND program_id=%s",
                    (customer_id, program_id),
                )

            elif customer_id and event_id:
                cur.execute(
                    "UPDATE events.customer_enrollment SET payment_status='paid' WHERE customer_id=%s AND event_id=%s",
                    (customer_id, event_id),
                )
            conn.commit()

    return {"status": "ok"}
