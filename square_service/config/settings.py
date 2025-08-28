import os
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Database Configuration  
DATABASE_URL = os.getenv("DATABASE_URL")

# Square API Configuration
SQUARE_ACCESS_TOKEN = os.getenv("SQUARE_ACCESS_TOKEN")
SQUARE_LOCATION_ID = os.getenv("SQUARE_LOCATION_ID")
SQUARE_ENV = os.getenv("SQUARE_ENV", "sandbox")
SQUARE_VERSION = os.getenv("SQUARE_VERSION", "2025-07-16")

# JWT Configuration
JWT_SECRET = os.getenv("JWT_SECRET")
JWT_ALGORITHM = os.getenv("JWT_ALGORITHM", "HS256")

# Webhook Configuration
SQUARE_WEBHOOK_SIGNATURE_KEY = os.getenv("SQUARE_WEBHOOK_SIGNATURE_KEY")
WEBHOOK_URL = os.getenv("WEBHOOK_URL")
FRONTEND_URL = os.getenv("FRONTEND_URL")

# Logging Configuration
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
DEFAULT_LEVEL = "INFO" if SQUARE_ENV == "production" else "DEBUG"

# Required environment variables
REQUIRED_ENV_VARS = [
    "DATABASE_URL",
    "JWT_SECRET", 
    "SQUARE_ACCESS_TOKEN",
    "SQUARE_LOCATION_ID",
    "SQUARE_WEBHOOK_SIGNATURE_KEY",
    "FRONTEND_URL",
    "WEBHOOK_URL"
]

def validate_environment():
    """Validate that all required environment variables are set."""
    missing = [var for var in REQUIRED_ENV_VARS if not os.getenv(var)]
    if missing:
        raise RuntimeError(f"Missing required environment variables: {missing}")
    
    # Additional validation for production security
    if SQUARE_ENV == "production" and not SQUARE_WEBHOOK_SIGNATURE_KEY:
        raise RuntimeError("SQUARE_WEBHOOK_SIGNATURE_KEY is required in production environment")

# Square API URLs
SQUARE_BASE_URL = (
    "https://connect.squareupsandbox.com" 
    if SQUARE_ENV != "production" 
    else "https://connect.squareup.com"
)

# Headers for Square API
SQUARE_HEADERS = {
    "Square-Version": SQUARE_VERSION,
    "Authorization": f"Bearer {SQUARE_ACCESS_TOKEN}",
    "Content-Type": "application/json",
}