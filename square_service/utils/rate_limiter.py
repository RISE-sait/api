"""
Per-user rate limiting for Square Payment Service endpoints.
"""
import time
import logging
from typing import Dict, Optional, Tuple
from fastapi import Request, HTTPException
from fastapi.responses import JSONResponse
from collections import defaultdict
import asyncio
import jwt
from config.settings import JWT_SECRET

logger = logging.getLogger(__name__)

class RateLimitEntry:
    """Rate limit tracking entry."""
    def __init__(self, limit: int, window: int):
        self.limit = limit
        self.window = window
        self.requests = []
        self.lock = asyncio.Lock()
    
    async def is_allowed(self) -> bool:
        """Check if request is allowed within rate limit."""
        async with self.lock:
            now = time.time()
            # Remove old requests outside the window
            self.requests = [req_time for req_time in self.requests if now - req_time < self.window]
            
            if len(self.requests) < self.limit:
                self.requests.append(now)
                return True
            return False
    
    def get_reset_time(self) -> int:
        """Get timestamp when rate limit resets."""
        if not self.requests:
            return int(time.time())
        return int(self.requests[0] + self.window)

class UserRateLimiter:
    """Per-user rate limiter with different limits for different endpoints."""
    
    def __init__(self):
        self.user_limits: Dict[str, Dict[str, RateLimitEntry]] = defaultdict(dict)
        self.cleanup_task = None
        self._cleanup_started = False
    
    def start_cleanup_task(self):
        """Start background cleanup task."""
        if not self._cleanup_started:
            try:
                if self.cleanup_task is None or self.cleanup_task.done():
                    self.cleanup_task = asyncio.create_task(self._cleanup_expired_entries())
                    self._cleanup_started = True
            except RuntimeError:
                # No event loop running, will start later
                pass
    
    async def _cleanup_expired_entries(self):
        """Clean up expired rate limit entries periodically."""
        while True:
            try:
                await asyncio.sleep(300)  # Clean up every 5 minutes
                now = time.time()
                users_to_remove = []
                
                for user_id, endpoints in self.user_limits.items():
                    endpoints_to_remove = []
                    for endpoint, entry in endpoints.items():
                        async with entry.lock:
                            entry.requests = [req_time for req_time in entry.requests if now - req_time < entry.window]
                            if not entry.requests:
                                endpoints_to_remove.append(endpoint)
                    
                    for endpoint in endpoints_to_remove:
                        del endpoints[endpoint]
                    
                    if not endpoints:
                        users_to_remove.append(user_id)
                
                for user_id in users_to_remove:
                    del self.user_limits[user_id]
                    
            except Exception as e:
                logger.error(f"Error in rate limiter cleanup: {e}")
    
    def get_rate_limit_config(self, endpoint: str) -> Tuple[int, int]:
        """Get rate limit configuration for endpoint."""
        # Rate limits: (requests_per_window, window_seconds)
        endpoint_limits = {
            # Critical endpoints
            "/subscriptions/checkout": (5, 60),      # 5 requests per minute
            "/subscriptions/manage": (10, 60),       # 10 requests per minute
            "/webhook": (100, 60),                   # 100 webhooks per minute
            
            # Admin endpoints
            "/admin/sync": (2, 300),                 # 2 syncs per 5 minutes
            "/admin/sync/status": (20, 60),          # 20 status checks per minute
            "/admin/sync/force": (1, 300),           # 1 force sync per 5 minutes
            
            # Subscription management
            "/subscriptions/{subscription_id}": (20, 60),  # 20 requests per minute
            "/subscriptions/{subscription_id}/pause": (5, 300),   # 5 pauses per 5 minutes
            "/subscriptions/{subscription_id}/resume": (5, 300),  # 5 resumes per 5 minutes
            "/subscriptions/{subscription_id}/cancel": (3, 300),  # 3 cancels per 5 minutes
            "/subscriptions/{subscription_id}/update-card": (5, 300),  # 5 updates per 5 minutes
            
            # General endpoints
            "/health": (100, 60),                    # 100 health checks per minute
            "/metrics": (60, 60),                    # 60 metrics requests per minute
            "/payments/{payment_id}": (30, 60),      # 30 payment queries per minute
            "/subscriptions": (30, 60),              # 30 subscription queries per minute
        }
        
        # Default rate limit
        return endpoint_limits.get(endpoint, (60, 60))  # 60 requests per minute default
    
    def extract_user_id(self, request: Request) -> Optional[str]:
        """Extract user ID from request for rate limiting."""
        # Try to get user ID from JWT token
        auth_header = request.headers.get("Authorization")
        if auth_header and auth_header.startswith("Bearer "):
            try:
                token = auth_header[7:]
                payload = jwt.decode(token, JWT_SECRET, algorithms=["HS256"])
                return payload.get("user_id") or payload.get("sub")
            except jwt.InvalidTokenError:
                pass
        
        # Fall back to IP address for unauthenticated requests
        client_ip = request.client.host if request.client else "unknown"
        forwarded_for = request.headers.get("X-Forwarded-For")
        if forwarded_for:
            client_ip = forwarded_for.split(",")[0].strip()
        
        return f"ip:{client_ip}"
    
    def normalize_endpoint(self, path: str) -> str:
        """Normalize endpoint path for rate limiting."""
        # Replace path parameters with placeholders
        import re
        
        # UUID pattern
        uuid_pattern = r'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
        path = re.sub(uuid_pattern, '{subscription_id}', path, flags=re.IGNORECASE)
        path = re.sub(uuid_pattern, '{payment_id}', path, flags=re.IGNORECASE)
        
        # Other ID patterns
        path = re.sub(r'/\d+', '/{id}', path)
        
        return path
    
    async def is_allowed(self, request: Request) -> Tuple[bool, Optional[RateLimitEntry]]:
        """Check if request is allowed under rate limits."""
        # Start cleanup task if not already started
        if not self._cleanup_started:
            self.start_cleanup_task()
            
        user_id = self.extract_user_id(request)
        if not user_id:
            return True, None  # Allow if we can't identify user
        
        endpoint = self.normalize_endpoint(request.url.path)
        limit, window = self.get_rate_limit_config(endpoint)
        
        # Get or create rate limit entry for this user and endpoint
        if endpoint not in self.user_limits[user_id]:
            self.user_limits[user_id][endpoint] = RateLimitEntry(limit, window)
        
        entry = self.user_limits[user_id][endpoint]
        allowed = await entry.is_allowed()
        
        return allowed, entry

# Global rate limiter instance
rate_limiter = UserRateLimiter()

async def rate_limit_middleware(request: Request, call_next):
    """Rate limiting middleware."""
    try:
        # Skip rate limiting for health check from internal sources
        if request.url.path == "/health" and request.client and request.client.host in ["127.0.0.1", "localhost"]:
            return await call_next(request)
        
        allowed, entry = await rate_limiter.is_allowed(request)
        
        if not allowed:
            headers = {}
            if entry:
                headers.update({
                    "X-RateLimit-Limit": str(entry.limit),
                    "X-RateLimit-Remaining": "0",
                    "X-RateLimit-Reset": str(entry.get_reset_time()),
                    "Retry-After": str(max(1, entry.get_reset_time() - int(time.time())))
                })
            
            logger.warning(f"Rate limit exceeded for user: {rate_limiter.extract_user_id(request)}, endpoint: {request.url.path}")
            
            return JSONResponse(
                status_code=429,
                content={
                    "detail": "Too many requests. Please try again later.",
                    "error_code": "RATE_LIMIT_EXCEEDED"
                },
                headers=headers
            )
        
        # Add rate limit headers to successful responses
        response = await call_next(request)
        
        if entry:
            remaining = max(0, entry.limit - len(entry.requests))
            response.headers["X-RateLimit-Limit"] = str(entry.limit)
            response.headers["X-RateLimit-Remaining"] = str(remaining)
            response.headers["X-RateLimit-Reset"] = str(entry.get_reset_time())
        
        return response
        
    except Exception as e:
        logger.error(f"Rate limit middleware error: {e}")
        # Continue processing request if rate limiter fails
        return await call_next(request)

def get_rate_limiter_status() -> Dict:
    """Get current rate limiter status for monitoring."""
    active_users = len(rate_limiter.user_limits)
    total_endpoints = sum(len(endpoints) for endpoints in rate_limiter.user_limits.values())
    
    return {
        "active_users": active_users,
        "total_tracked_endpoints": total_endpoints,
        "cleanup_task_running": rate_limiter.cleanup_task and not rate_limiter.cleanup_task.done()
    }