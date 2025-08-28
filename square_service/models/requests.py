from pydantic import BaseModel, validator, Field
from typing import Optional
from enum import Enum
import uuid

class SubscriptionAction(str, Enum):
    PAUSE = "pause"
    RESUME = "resume" 
    CANCEL = "cancel"
    UPDATE_CARD = "update_card"

class SubscriptionRequest(BaseModel):
    membership_plan_id: str = Field(..., description="UUID of the membership plan")
    card_id: Optional[str] = Field(None, description="Card ID for automatic billing")
    start_date: Optional[str] = Field(None, description="ISO date string for subscription start")
    timezone: str = Field(default="UTC", description="Timezone for the subscription")
    
    @validator('membership_plan_id')
    def validate_membership_plan_id(cls, v):
        try:
            uuid.UUID(v)
            return v
        except ValueError:
            raise ValueError('membership_plan_id must be a valid UUID')

class SubscriptionManagementRequest(BaseModel):
    action: SubscriptionAction = Field(..., description="Action to perform on subscription")
    card_id: Optional[str] = Field(None, description="New card ID for update_card action")
    new_plan_variation_id: Optional[str] = Field(None, description="For upgrade/downgrade actions")
    pause_cycle_duration: Optional[int] = Field(None, description="Number of billing cycles to pause")
    cancel_reason: Optional[str] = Field(None, max_length=500, description="Reason for cancellation")

class CheckoutRequest(BaseModel):
    membership_plan_id: str = Field(..., min_length=1, max_length=100, description="Valid membership plan UUID")
    discount_code: Optional[str] = Field(None, max_length=50, description="Optional discount code")
    
    @validator('membership_plan_id')
    def validate_membership_plan_id(cls, v):
        try:
            uuid.UUID(v)
            return v
        except ValueError:
            raise ValueError('membership_plan_id must be a valid UUID')

class ProgramCheckoutRequest(BaseModel):
    program_id: str = Field(..., min_length=1, max_length=100, description="Valid program UUID")
    
    @validator('program_id')
    def validate_program_id(cls, v):
        try:
            uuid.UUID(v)
            return v
        except ValueError:
            raise ValueError('program_id must be a valid UUID')

class EventCheckoutRequest(BaseModel):
    event_id: str = Field(..., min_length=1, max_length=100, description="Valid event UUID")
    
    @validator('event_id')
    def validate_event_id(cls, v):
        try:
            uuid.UUID(v)
            return v
        except ValueError:
            raise ValueError('event_id must be a valid UUID')