"""
API Key management system for Square Payment Service.
"""
import secrets
import hashlib
import time
import logging
from typing import Optional, Dict, List
from dataclasses import dataclass
from enum import Enum
from fastapi import Request, HTTPException
from database.connection import get_db_conn

logger = logging.getLogger(__name__)

class ApiKeyScope(Enum):
    """API key permission scopes."""
    READ_ONLY = "read_only"          # Health, status endpoints
    WEBHOOKS = "webhooks"            # Webhook processing only
    SUBSCRIPTIONS = "subscriptions"   # Subscription operations
    ADMIN = "admin"                  # Full admin access
    SYNC = "sync"                    # Sync operations

@dataclass
class ApiKey:
    """API Key data structure."""
    key_id: str
    name: str
    scopes: List[ApiKeyScope]
    created_at: int
    last_used: Optional[int] = None
    is_active: bool = True
    description: Optional[str] = None

class ApiKeyManager:
    """Manages API keys for service-to-service authentication."""
    
    def __init__(self):
        self.ensure_api_keys_table()
    
    def ensure_api_keys_table(self):
        """Create API keys table if it doesn't exist."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute("""
                    CREATE TABLE IF NOT EXISTS api_keys (
                        key_id VARCHAR(64) PRIMARY KEY,
                        key_hash VARCHAR(128) NOT NULL,
                        name VARCHAR(255) NOT NULL,
                        description TEXT,
                        scopes TEXT[] NOT NULL,
                        created_at TIMESTAMP DEFAULT NOW(),
                        last_used TIMESTAMP,
                        is_active BOOLEAN DEFAULT TRUE,
                        created_by VARCHAR(255),
                        UNIQUE(name)
                    );
                    CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active, key_hash);
                """)
                conn.commit()
        except Exception as e:
            logger.error(f"Failed to create api_keys table: {e}")
    
    def generate_api_key(self, name: str, scopes: List[ApiKeyScope], 
                        description: str = None, created_by: str = None) -> Dict[str, str]:
        """Generate a new API key."""
        # Generate key components
        key_id = self._generate_key_id()
        secret = self._generate_secret()
        api_key = f"rsp_{key_id}_{secret}"  # rsp = Rise Square Payment
        key_hash = self._hash_key(api_key)
        
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute("""
                    INSERT INTO api_keys (key_id, key_hash, name, description, scopes, created_by)
                    VALUES (%s, %s, %s, %s, %s, %s)
                """, (
                    key_id,
                    key_hash,
                    name,
                    description,
                    [scope.value for scope in scopes],
                    created_by
                ))
                conn.commit()
                
                logger.info(f"Created API key '{name}' with scopes: {[s.value for s in scopes]}")
                
                return {
                    "key_id": key_id,
                    "api_key": api_key,
                    "name": name,
                    "scopes": [scope.value for scope in scopes],
                    "created_at": int(time.time())
                }
                
        except Exception as e:
            logger.error(f"Failed to create API key: {e}")
            raise HTTPException(status_code=500, detail="Failed to create API key")
    
    def validate_api_key(self, api_key: str, required_scope: ApiKeyScope = None) -> Optional[ApiKey]:
        """Validate an API key and return key info if valid."""
        if not api_key or not api_key.startswith("rsp_"):
            return None
        
        try:
            parts = api_key.split("_", 2)
            if len(parts) != 3:
                return None
                
            key_id = parts[1]
            key_hash = self._hash_key(api_key)
            
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute("""
                    SELECT key_id, name, scopes, created_at, last_used, is_active, description
                    FROM api_keys 
                    WHERE key_id = %s AND key_hash = %s AND is_active = TRUE
                """, (key_id, key_hash))
                
                row = cur.fetchone()
                if not row:
                    return None
                
                # Update last_used timestamp
                cur.execute("""
                    UPDATE api_keys SET last_used = NOW() WHERE key_id = %s
                """, (key_id,))
                conn.commit()
                
                scopes = [ApiKeyScope(scope) for scope in row[2]]
                
                # Check if required scope is present
                if required_scope and required_scope not in scopes:
                    logger.warning(f"API key '{row[1]}' missing required scope: {required_scope.value}")
                    return None
                
                return ApiKey(
                    key_id=row[0],
                    name=row[1],
                    scopes=scopes,
                    created_at=int(row[3].timestamp()) if row[3] else 0,
                    last_used=int(row[4].timestamp()) if row[4] else None,
                    is_active=row[5],
                    description=row[6]
                )
                
        except Exception as e:
            logger.error(f"API key validation error: {e}")
            return None
    
    def list_api_keys(self) -> List[Dict]:
        """List all API keys (without exposing the actual keys)."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute("""
                    SELECT key_id, name, description, scopes, created_at, last_used, is_active
                    FROM api_keys 
                    ORDER BY created_at DESC
                """)
                
                keys = []
                for row in cur.fetchall():
                    keys.append({
                        "key_id": row[0],
                        "name": row[1],
                        "description": row[2],
                        "scopes": row[3],
                        "created_at": row[4].isoformat() if row[4] else None,
                        "last_used": row[5].isoformat() if row[5] else None,
                        "is_active": row[6]
                    })
                
                return keys
                
        except Exception as e:
            logger.error(f"Failed to list API keys: {e}")
            return []
    
    def revoke_api_key(self, key_id: str) -> bool:
        """Revoke an API key."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute("""
                    UPDATE api_keys SET is_active = FALSE WHERE key_id = %s
                """, (key_id,))
                
                if cur.rowcount > 0:
                    conn.commit()
                    logger.info(f"Revoked API key: {key_id}")
                    return True
                
                return False
                
        except Exception as e:
            logger.error(f"Failed to revoke API key: {e}")
            return False
    
    def _generate_key_id(self) -> str:
        """Generate a unique key ID."""
        return secrets.token_urlsafe(16)[:16]
    
    def _generate_secret(self) -> str:
        """Generate a secure secret."""
        return secrets.token_urlsafe(32)
    
    def _hash_key(self, api_key: str) -> str:
        """Hash an API key for secure storage."""
        return hashlib.sha256(api_key.encode()).hexdigest()

# Global API key manager instance
api_key_manager = ApiKeyManager()

def get_api_key_from_request(request: Request) -> Optional[str]:
    """Extract API key from request headers."""
    # Check X-API-Key header first
    api_key = request.headers.get("X-API-Key")
    if api_key:
        return api_key
    
    # Check Authorization header for API key format
    auth_header = request.headers.get("Authorization", "")
    if auth_header.startswith("ApiKey "):
        return auth_header[7:]  # Remove "ApiKey " prefix
    
    return None

def require_api_key(scope: ApiKeyScope = None):
    """Dependency to require valid API key with optional scope."""
    def dependency(request: Request) -> ApiKey:
        api_key = get_api_key_from_request(request)
        if not api_key:
            raise HTTPException(
                status_code=401,
                detail="API key required",
                headers={"WWW-Authenticate": "ApiKey"}
            )
        
        key_info = api_key_manager.validate_api_key(api_key, scope)
        if not key_info:
            raise HTTPException(
                status_code=401,
                detail="Invalid or insufficient API key permissions"
            )
        
        return key_info
    
    return dependency

# Endpoint-specific scope requirements
ENDPOINT_SCOPES = {
    "/webhook": ApiKeyScope.WEBHOOKS,
    "/admin/sync": ApiKeyScope.SYNC,
    "/admin/sync/force": ApiKeyScope.ADMIN,
    "/admin/sync/status": ApiKeyScope.READ_ONLY,
    "/health": None,  # No scope required
    "/metrics": ApiKeyScope.READ_ONLY,
    "/subscriptions": ApiKeyScope.SUBSCRIPTIONS,
}

def get_required_scope_for_endpoint(path: str) -> Optional[ApiKeyScope]:
    """Get required scope for an endpoint."""
    # Exact match first
    if path in ENDPOINT_SCOPES:
        return ENDPOINT_SCOPES[path]
    
    # Pattern matching for parameterized endpoints
    if path.startswith("/subscriptions/") and path.endswith("/checkout"):
        return ApiKeyScope.SUBSCRIPTIONS
    elif path.startswith("/subscriptions/") and any(action in path for action in ["/pause", "/resume", "/cancel", "/update-card"]):
        return ApiKeyScope.SUBSCRIPTIONS
    elif path.startswith("/admin/"):
        return ApiKeyScope.ADMIN
    elif path.startswith("/payments/"):
        return ApiKeyScope.READ_ONLY
    
    return None