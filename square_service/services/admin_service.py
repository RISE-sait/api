import logging
from database.connection import get_db_conn
from services.square_client import SquareAPIClient

logger = logging.getLogger(__name__)

class AdminService:
    def __init__(self, square_client: SquareAPIClient):
        self.square_client = square_client
    
    async def expire_memberships(self) -> dict:
        """Expire memberships that have passed their renewal date."""
        try:
            with get_db_conn() as conn, conn.cursor() as cur:
                cur.execute(
                    """
                    UPDATE users.customer_membership_plans
                       SET status = 'inactive'::membership.membership_status
                     WHERE status = 'active'
                       AND renewal_date IS NOT NULL
                       AND renewal_date < NOW()
                    """
                )
                changed = cur.rowcount
                conn.commit()
            
            logger.info(f"[ADMIN] expired {changed} memberships")
            return {"expired": changed}
            
        except Exception as e:
            logger.error(f"[ADMIN] Error expiring memberships: {e}")
            raise
    
    async def list_square_subscription_plans(self) -> dict:
        """List available Square subscription plans."""
        try:
            # This would need to be implemented in the SquareAPIClient
            # For now, return a placeholder response
            return {
                "subscription_plans": [],
                "total_count": 0
            }
        except Exception as e:
            logger.error(f"[ADMIN] Error listing subscription plans: {e}")
            raise