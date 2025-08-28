from pydantic import BaseModel, Field
from typing import Optional

class SubscriptionResponse(BaseModel):
    subscription_id: str
    status: str
    message: str
    subscription_url: Optional[str] = None
    checkout_url: Optional[str] = None

class CheckoutResponse(BaseModel):
    checkout_url: Optional[str] = Field(None, description="Square checkout URL or null if free")
    status: str = Field(default="success", description="Transaction status")

class ErrorResponse(BaseModel):
    error: str
    detail: Optional[str] = None
    code: Optional[str] = None