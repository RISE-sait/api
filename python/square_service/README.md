# Square Payment Service

This directory contains a small Python service for interacting with the Square API. It exposes simple HTTP endpoints to create payment links, create subscriptions and handle webhooks.

## Setup

```bash
cd python/square_service
pip install -r requirements.txt
python main.py
```

The service expects `SQUARE_ACCESS_TOKEN` to be set in the environment. Optional variables like `SQUARE_START_DATE` can be used to override the default subscription start date.