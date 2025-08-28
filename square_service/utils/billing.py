from datetime import datetime, timedelta
import calendar

def calculate_next_billing_date(cadence: str, anchor_day: int = None, from_date: str = None) -> str:
    """Calculate next billing date for any Square subscription cadence."""
    if from_date:
        current_date = datetime.strptime(from_date, "%Y-%m-%d")
    else:
        current_date = datetime.utcnow()
    
    # Handle different billing cadences
    cadence_upper = cadence.upper() if cadence else "MONTHLY"
    
    if cadence_upper == "DAILY":
        next_billing = current_date + timedelta(days=1)
        
    elif cadence_upper == "WEEKLY":
        next_billing = current_date + timedelta(days=7)
        
    elif cadence_upper == "EVERY_TWO_WEEKS":
        next_billing = current_date + timedelta(days=14)
        
    elif cadence_upper == "EVERY_THREE_WEEKS":
        next_billing = current_date + timedelta(days=21)
        
    elif cadence_upper == "EVERY_FOUR_WEEKS":
        next_billing = current_date + timedelta(days=28)
        
    elif cadence_upper == "EVERY_TWO_MONTHS":
        # Add 2 months
        if current_date.month <= 10:
            next_billing = current_date.replace(month=current_date.month + 2)
        else:
            next_billing = current_date.replace(year=current_date.year + 1, 
                                              month=current_date.month - 10)
                                              
    elif cadence_upper == "QUARTERLY":
        # Add 3 months
        if current_date.month <= 9:
            next_billing = current_date.replace(month=current_date.month + 3)
        else:
            next_billing = current_date.replace(year=current_date.year + 1, 
                                              month=current_date.month - 9)
                                              
    elif cadence_upper == "EVERY_SIX_MONTHS":
        # Add 6 months
        if current_date.month <= 6:
            next_billing = current_date.replace(month=current_date.month + 6)
        else:
            next_billing = current_date.replace(year=current_date.year + 1, 
                                              month=current_date.month - 6)
                                              
    elif cadence_upper == "ANNUAL":
        # Add 1 year
        next_billing = current_date.replace(year=current_date.year + 1)
        
    elif cadence_upper == "EVERY_TWO_YEARS":
        # Add 2 years
        next_billing = current_date.replace(year=current_date.year + 2)
        
    else:  # MONTHLY (default)
        anchor_day = anchor_day or 1  # Default to 1st of month
        
        # If current day is before anchor day, next billing is this month
        if current_date.day < anchor_day:
            try:
                next_billing = current_date.replace(day=anchor_day)
            except ValueError:
                # Handle months with fewer days (e.g., Feb 30th doesn't exist)
                last_day = calendar.monthrange(current_date.year, current_date.month)[1]
                next_billing = current_date.replace(day=min(anchor_day, last_day))
        else:
            # Next billing is next month
            if current_date.month == 12:
                next_year = current_date.year + 1
                next_month = 1
            else:
                next_year = current_date.year
                next_month = current_date.month + 1
                
            try:
                next_billing = current_date.replace(year=next_year, month=next_month, day=anchor_day)
            except ValueError:
                # Handle months with fewer days
                last_day = calendar.monthrange(next_year, next_month)[1]
                next_billing = current_date.replace(year=next_year, month=next_month, day=min(anchor_day, last_day))
    
    return next_billing.strftime("%Y-%m-%d")