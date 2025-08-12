import os
import psycopg2

def get_db_conn():
    """Return a new database connection using DATABASE_URL env variable."""
    return psycopg2.connect(os.getenv("DATABASE_URL"))