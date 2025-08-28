from prometheus_client import Counter, Histogram, Gauge

# Metrics
subscription_requests = Counter('square_subscription_requests_total', 'Total subscription requests', ['operation', 'status'])
subscription_duration = Histogram('square_subscription_request_duration_seconds', 'Subscription request duration', ['operation'])
square_api_errors = Counter('square_api_errors_total', 'Total Square API errors', ['endpoint', 'status_code'])
active_subscriptions = Gauge('active_subscriptions_total', 'Total active subscriptions')
webhook_events = Counter('webhook_events_total', 'Total webhook events received', ['event_type'])