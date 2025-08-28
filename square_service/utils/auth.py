import os
import jwt
import hmac
import hashlib
import base64
from fastapi import HTTPException
from fastapi.security import HTTPAuthorizationCredentials
import logging

logger = logging.getLogger(__name__)

# JWT Configuration
JWT_SECRET = os.getenv("JWT_SECRET")
JWT_ALGORITHM = os.getenv("JWT_ALGORITHM")
WEBHOOK_SIGNATURE_KEY = os.getenv("SQUARE_WEBHOOK_SIGNATURE_KEY")
WEBHOOK_URL = os.getenv("WEBHOOK_URL")

def get_user_id(creds: HTTPAuthorizationCredentials) -> str:
    """Extract user_id from JWT token."""
    token = creds.credentials
    try:
        payload = jwt.decode(
            token,
            JWT_SECRET,
            algorithms=[JWT_ALGORITHM],
            options={"require": ["exp"]},  # require expiration
        )
        return payload.get("user_id") or payload.get("sub")
    except jwt.ExpiredSignatureError:
        logger.warning("JWT expired")
        raise HTTPException(status_code=401, detail="Token expired")
    except jwt.InvalidTokenError:
        raise HTTPException(status_code=401, detail="Invalid token")

def verify_signature(signature_header: str, body: bytes) -> bool:
    """Verify Square webhook signature."""
    if not WEBHOOK_SIGNATURE_KEY:
        logger.error("[WEBHOOK] No signature key configured - webhook verification required")
        return False
    
    try:
        # Square signs: HMAC_SHA256(URL + body)
        url = (WEBHOOK_URL or "").encode("utf-8")
        computed = base64.b64encode(
            hmac.new(WEBHOOK_SIGNATURE_KEY.encode("utf-8"), url + body, hashlib.sha256).digest()
        ).decode("utf-8")
        
        # Only log sensitive details at DEBUG
        logger.debug("[WEBHOOK] env_url=%r body_len=%s", WEBHOOK_URL, len(body))
        logger.debug("[WEBHOOK] header_sig=%r", signature_header)
        logger.debug("[WEBHOOK] computed_sig=%r", computed)
        
        return hmac.compare_digest(signature_header, computed)
    except Exception as e:
        logger.error(f"[WEBHOOK] Signature verification error: {e}")
        return False