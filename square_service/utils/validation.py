import re
import logging
from typing import Any, Dict, List, Optional
from fastapi import Request, HTTPException
from fastapi.responses import JSONResponse

logger = logging.getLogger(__name__)

class ValidationError(Exception):
    """Custom validation error."""
    def __init__(self, message: str, field: str = None):
        self.message = message
        self.field = field
        super().__init__(self.message)

class InputValidator:
    """Input validation utilities."""
    
    @staticmethod
    def validate_uuid(value: str, field_name: str = "id") -> str:
        """Validate UUID format."""
        uuid_pattern = r'^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
        if not re.match(uuid_pattern, value, re.IGNORECASE):
            raise ValidationError(f"Invalid UUID format for {field_name}", field_name)
        return value
    
    @staticmethod
    def validate_email(email: str, field_name: str = "email") -> str:
        """Validate email format."""
        email_pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        if not re.match(email_pattern, email):
            raise ValidationError(f"Invalid email format for {field_name}", field_name)
        return email
    
    @staticmethod
    def validate_string_length(value: str, min_length: int = 1, max_length: int = 1000, field_name: str = "field") -> str:
        """Validate string length."""
        if len(value) < min_length:
            raise ValidationError(f"{field_name} must be at least {min_length} characters long", field_name)
        if len(value) > max_length:
            raise ValidationError(f"{field_name} must be no more than {max_length} characters long", field_name)
        return value
    
    @staticmethod
    def validate_amount(amount: int, min_amount: int = 1, max_amount: int = 1000000, field_name: str = "amount") -> int:
        """Validate monetary amount (in cents)."""
        if amount < min_amount:
            raise ValidationError(f"{field_name} must be at least {min_amount} cents", field_name)
        if amount > max_amount:
            raise ValidationError(f"{field_name} must be no more than {max_amount} cents", field_name)
        return amount
    
    @staticmethod
    def sanitize_string(value: str) -> str:
        """Sanitize string input to prevent injection attacks."""
        # Remove null bytes and control characters
        sanitized = re.sub(r'[\x00-\x1F\x7F]', '', value)
        # Trim whitespace
        sanitized = sanitized.strip()
        return sanitized
    
    @staticmethod
    def validate_timezone(timezone: str, field_name: str = "timezone") -> str:
        """Validate timezone format."""
        # Basic timezone validation - in production you might want to use pytz
        valid_timezones = [
            'UTC', 'America/New_York', 'America/Chicago', 'America/Denver', 
            'America/Los_Angeles', 'America/Toronto', 'America/Vancouver',
            'Europe/London', 'Europe/Paris', 'Asia/Tokyo'
        ]
        
        if timezone not in valid_timezones:
            # Allow standard timezone formats
            tz_pattern = r'^[A-Za-z_]+/[A-Za-z_]+$|^UTC$|^GMT[+-]\d{1,2}$'
            if not re.match(tz_pattern, timezone):
                raise ValidationError(f"Invalid timezone format for {field_name}", field_name)
        
        return timezone
    
    @staticmethod
    def validate_subscription_action(action: str, field_name: str = "action") -> str:
        """Validate subscription management action."""
        valid_actions = ['pause', 'resume', 'cancel', 'update_card']
        if action not in valid_actions:
            raise ValidationError(f"Invalid action '{action}'. Must be one of: {', '.join(valid_actions)}", field_name)
        return action

class SecurityValidator:
    """Security-focused validation."""
    
    SUSPICIOUS_PATTERNS = [
        r'<script[^>]*>.*?</script>',  # XSS
        r'javascript:',  # JavaScript protocol
        r'on\w+\s*=',  # Event handlers
        r'expression\s*\(',  # CSS expression
        r'import\s+\w+',  # Import statements
        r'__\w+__',  # Python dunder methods
        r'eval\s*\(',  # Eval function
        r'exec\s*\(',  # Exec function
    ]
    
    @staticmethod
    def check_for_malicious_patterns(value: str, field_name: str = "field") -> str:
        """Check for potentially malicious patterns."""
        for pattern in SecurityValidator.SUSPICIOUS_PATTERNS:
            if re.search(pattern, value, re.IGNORECASE):
                logger.warning(f"Suspicious pattern detected in {field_name}: {pattern}")
                raise ValidationError(f"Invalid content detected in {field_name}", field_name)
        return value
    
    @staticmethod
    def validate_content_length(content: bytes, max_size: int = 1024 * 1024) -> bytes:  # 1MB default
        """Validate content length to prevent DoS attacks."""
        if len(content) > max_size:
            raise ValidationError(f"Content too large. Maximum size is {max_size} bytes")
        return content

async def validation_middleware(request: Request, call_next):
    """Middleware to validate requests."""
    try:
        # Validate content length for POST/PUT requests
        if request.method in ['POST', 'PUT', 'PATCH']:
            content_length = request.headers.get('content-length')
            if content_length:
                content_length = int(content_length)
                if content_length > 10 * 1024 * 1024:  # 10MB limit
                    return JSONResponse(
                        status_code=413,
                        content={"detail": "Request entity too large"}
                    )
        
        # Validate common headers
        user_agent = request.headers.get('user-agent', '')
        if len(user_agent) > 1000:  # Reasonable user agent length
            return JSONResponse(
                status_code=400,
                content={"detail": "Invalid user agent header"}
            )
        
        # Check for suspicious patterns in headers (skip legitimate headers)
        skip_validation_headers = {
            'x-square-signature', 'x-square-hmacsha256-signature', 
            'authorization', 'x-api-key', 'user-agent', 'referer',
            'origin', 'x-forwarded-for', 'x-real-ip'
        }
        
        for header_name, header_value in request.headers.items():
            if header_name.lower() not in skip_validation_headers:
                try:
                    SecurityValidator.check_for_malicious_patterns(str(header_value), f"header:{header_name}")
                except ValidationError:
                    logger.warning(f"Suspicious pattern in header {header_name}")
                    return JSONResponse(
                        status_code=400,
                        content={"detail": "Invalid request headers"}
                    )
        
        response = await call_next(request)
        return response
        
    except Exception as e:
        logger.error(f"Validation middleware error: {e}")
        return JSONResponse(
            status_code=500,
            content={"detail": "Internal server error"}
        )

def create_field_validator(validator_func):
    """Decorator to create field validators for Pydantic models."""
    def decorator(field_value):
        return validator_func(field_value)
    return decorator

# Custom Pydantic validators
def validate_membership_plan_id(v: str) -> str:
    """Pydantic validator for membership plan IDs."""
    return InputValidator.validate_uuid(v, "membership_plan_id")

def validate_timezone_field(v: str) -> str:
    """Pydantic validator for timezone fields."""
    return InputValidator.validate_timezone(v, "timezone")

def validate_action_field(v: str) -> str:
    """Pydantic validator for action fields."""
    return InputValidator.validate_subscription_action(v, "action")

def sanitize_string_field(v: str) -> str:
    """Pydantic validator to sanitize string fields."""
    sanitized = InputValidator.sanitize_string(v)
    SecurityValidator.check_for_malicious_patterns(sanitized)
    return sanitized