# Rise Stripe Payment Service

A comprehensive, secure Stripe integration for Rise API providing subscription management, one-time payments, and webhook processing.

## 🚀 Features

### ✅ Complete Stripe Integration
- **One-time payments** with checkout sessions
- **Recurring subscriptions** with automatic billing  
- **Subscription management** (pause, resume, cancel)
- **Customer portal** for self-service management
- **Discount/coupon** support
- **Comprehensive webhook** handling

### 🔒 Security Features
- **JWT authentication** with user ID validation
- **Webhook signature verification** for all Stripe events
- **Rate limiting** by user and endpoint
- **Input validation** and sanitization
- **Security headers** (CSP, HSTS, etc.)
- **SQL injection protection**
- **CORS configuration**

### 📊 Monitoring & Logging
- **Structured logging** with request tracing
- **Performance monitoring** with request duration
- **Error tracking** with detailed error messages
- **Security event logging**

## 🏗️ Architecture

```
internal/domains/payment/
├── handler/
│   ├── checkout.go          # Checkout endpoints
│   ├── webhook.go           # Stripe webhook handler
│   └── subscription.go      # Subscription management
├── services/
│   ├── checkout.go          # Business logic for checkout
│   ├── webhooks.go          # Webhook event processing
│   └── stripe/
│       ├── stripe.go        # Core Stripe operations
│       └── integration_test.go
├── middleware/
│   └── rate_limit.go        # Security & rate limiting
└── README.md               # This file
```

## 🔧 Configuration

### Environment Variables

```bash
# Required
STRIPE_SECRET_KEY=sk_test_...           # Stripe secret API key
STRIPE_WEBHOOK_SECRET=whsec_...         # Webhook endpoint secret
JWT_SECRET=your-jwt-secret              # JWT signing secret
DATABASE_URL=postgres://...             # Database connection

# Optional
LOG_LEVEL=INFO                          # Logging level
```

### Stripe Configuration

1. **Create Products and Prices** in Stripe Dashboard
2. **Set up Webhook Endpoint** pointing to `/webhooks/stripe`
3. **Configure Webhook Events**:
   - `checkout.session.completed`
   - `customer.subscription.created`
   - `customer.subscription.updated` 
   - `customer.subscription.deleted`
   - `invoice.payment_succeeded`
   - `invoice.payment_failed`

## 📡 API Endpoints

### Authentication Required
All endpoints require `Authorization: Bearer <jwt_token>` header.

### Checkout Endpoints

#### Create Membership Checkout
```http
POST /checkout/membership_plans/{id}?discount_code={code}
```
Creates Stripe checkout session for membership plan subscription.

**Response:**
```json
{
  "payment_url": "https://checkout.stripe.com/pay/..."
}
```

#### Create Program Checkout  
```http
POST /checkout/programs/{id}
```
Creates checkout session for program enrollment.

#### Create Event Checkout
```http
POST /checkout/events/{id} 
```
Creates checkout session for event registration.

### Subscription Management

#### Get Subscription Details
```http
GET /subscriptions/{id}
```
Retrieves subscription with ownership verification.

**Response:**
```json
{
  "id": "sub_...",
  "status": "active",
  "current_period_start": 1640995200,
  "current_period_end": 1643673600,
  "cancel_at_period_end": false,
  "items": [...],
  "latest_invoice": {...}
}
```

#### Cancel Subscription
```http
POST /subscriptions/{id}/cancel?immediate=false
```
Cancels subscription immediately or at period end.

#### Pause Subscription
```http
POST /subscriptions/{id}/pause?resume_at=2024-03-01T00:00:00Z
```
Pauses subscription with optional auto-resume date.

#### Resume Subscription  
```http
POST /subscriptions/{id}/resume
```
Resumes a paused subscription.

#### Get Customer Subscriptions
```http
GET /subscriptions
```
Lists all subscriptions for authenticated user.

#### Create Portal Session
```http
POST /subscriptions/portal?return_url=https://app.rise.com/dashboard
```
Creates Stripe Customer Portal session for self-service management.

**Response:**
```json
{
  "portal_url": "https://billing.stripe.com/session/...",
  "message": "Portal session created successfully"
}
```

### Webhook Endpoint

#### Stripe Webhooks
```http
POST /webhooks/stripe
```
Processes Stripe webhook events with signature verification.

## 🔒 Security Implementation

### Rate Limiting
- **Checkout endpoints**: 5 requests/minute per user
- **Subscription management**: 10 requests/minute per user  
- **Customer portal**: 3 requests/minute per user
- **Webhooks**: 100 requests/minute (Stripe can send bursts)

### Input Validation
- **UUID validation** for all ID parameters
- **URL validation** for return URLs
- **Date format validation** (RFC3339)
- **String sanitization** to prevent injection

### Webhook Security
- **Signature verification** using Stripe webhook secret
- **Replay attack protection** via Stripe's timestamp validation
- **Event idempotency** handling

### Security Headers
```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY  
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
Content-Security-Policy: default-src 'self'; connect-src 'self' https://api.stripe.com
Cache-Control: no-cache, no-store, must-revalidate  # For sensitive endpoints
```

## 🧪 Testing

### Running Tests
```bash
# Unit tests
go test ./internal/domains/payment/...

# Integration tests (requires STRIPE_API_KEY)
STRIPE_API_KEY=sk_test_... go test ./internal/domains/payment/services/stripe/...

# Benchmarks
go test -bench=. ./internal/domains/payment/services/stripe/...
```

### Test Coverage
- ✅ **Input validation** for all endpoints
- ✅ **Authentication** requirement testing
- ✅ **Rate limiting** behavior
- ✅ **Error handling** scenarios
- ✅ **Webhook signature** validation
- ✅ **Stripe API integration**

## 🚨 Error Handling

### Standard Error Format
```json
{
  "error": {
    "message": "Subscription not found",
    "code": 404,
    "details": "..."
  }
}
```

### Common Error Codes
- `400` - Invalid input parameters
- `401` - Missing or invalid authentication  
- `403` - Access denied (resource not owned by user)
- `404` - Resource not found
- `409` - Conflict (e.g., already cancelled)
- `429` - Rate limit exceeded
- `500` - Internal server error

## 📊 Monitoring

### Logging Examples
```
[REQUEST] POST /checkout/membership_plans/123 - User: abc-def - Status: 200 - Duration: 145ms
[STRIPE] Subscription created: sub_1234567890
[WEBHOOK] Processing checkout.session.completed for user abc-def
[RATE_LIMIT] Rate limit exceeded for user:abc-def:endpoint:checkout
[SECURITY] Missing Stripe signature for webhook
```

### Key Metrics to Monitor
- **Request latency** by endpoint
- **Error rates** by status code
- **Rate limit** violations
- **Webhook processing** success/failure
- **Subscription** lifecycle events

## 🔄 Webhook Events Processing

### checkout.session.completed
- Validates session and line items
- Updates reservation status for programs/events  
- Activates membership plans for subscriptions
- Sends confirmation emails

### customer.subscription.created
- Logs subscription creation
- Updates membership status to active

### customer.subscription.updated  
- Handles status changes (active, past_due, canceled)
- Updates membership records accordingly

### customer.subscription.deleted
- Marks membership as expired
- Logs cancellation event

### invoice.payment_succeeded
- Confirms membership is active
- Updates payment history

### invoice.payment_failed
- Handles dunning process
- Sends payment failure notifications

## 🚀 Deployment

### Prerequisites
1. **Stripe Account** with API keys
2. **Database** with membership/customer tables
3. **JWT authentication** middleware
4. **HTTPS endpoint** for webhooks

### Configuration Steps
1. Set environment variables
2. Configure Stripe webhook endpoint
3. Test webhook delivery
4. Monitor logs for errors
5. Set up alerting for failures

## 🛠️ Development

### Adding New Payment Features

1. **Create service method** in `stripe/stripe.go`
2. **Add handler** in appropriate handler file  
3. **Apply middleware** for security/rate limiting
4. **Write tests** for validation and integration
5. **Update documentation**

### Code Style
- **Error handling**: Use `errLib.CommonError` for consistent responses
- **Logging**: Use structured logging with `[COMPONENT]` prefixes  
- **Validation**: Validate all inputs before processing
- **Security**: Apply authentication and rate limiting to all endpoints

## 📞 Support

For issues or questions:
1. Check logs for detailed error messages
2. Verify Stripe configuration and webhook delivery
3. Review authentication token validity  
4. Test with Stripe's test cards and webhooks
5. Monitor rate limiting and adjust as needed

## 🔐 Security Considerations

### Production Checklist
- [ ] Use production Stripe API keys
- [ ] Enable webhook signature verification
- [ ] Configure rate limiting appropriately  
- [ ] Set up HTTPS for all endpoints
- [ ] Enable security headers
- [ ] Monitor for suspicious activity
- [ ] Set up log aggregation and alerting
- [ ] Regular security audits

---

