import os
import psycopg2
from psycopg2 import pool
from contextlib import closing, contextmanager
import logging

logger = logging.getLogger(__name__)

# Global database pool
db_pool = None

def init_db_pool():
    """Initialize database connection pool."""
    global db_pool
    
    database_url = os.getenv("DATABASE_URL")
    if not database_url:
        raise RuntimeError("Missing required environment variable: DATABASE_URL")
    
    try:
        db_pool = psycopg2.pool.ThreadedConnectionPool(
            1, 20,  # min/max connections
            dsn=database_url
        )
        logger.info("[DB] Connection pool initialized")
    except Exception as e:
        logger.error(f"[DB] Failed to initialize connection pool: {e}")
        raise

@contextmanager
def get_db_conn():
    """Get database connection from pool."""
    global db_pool
    
    if not db_pool:
        raise RuntimeError("Database pool not initialized. Call init_db_pool() first.")
    
    conn = None
    try:
        conn = db_pool.getconn()
        yield conn
    except Exception as e:
        if conn:
            conn.rollback()
        raise e
    finally:
        if conn:
            db_pool.putconn(conn)

def _warn_if_db_superuser():
    """Best-effort warning if connected as a superuser (least-privilege reminder)."""
    try:
        with get_db_conn() as conn, conn.cursor() as cur:
            cur.execute("SELECT current_user")
            user = cur.fetchone()[0]
            if user == "postgres":
                logger.warning("[SECURITY] Connected as DB superuser 'postgres'. "
                               "Use a least-privileged DB user in production.")
    except Exception as e:
        logger.error(f"[DB] Error checking database user: {e}")

def close_db_pool():
    """Close all database connections."""
    global db_pool
    if db_pool:
        db_pool.closeall()
        logger.info("[DB] Connection pool closed")