package custom_types

import (
	"database/sql/driver"
	"fmt"
	"github.com/shopspring/decimal"
)

// NullDecimal represents a nullable decimal
type NullDecimal struct {
	Decimal decimal.Decimal
	Valid   bool
}

// Scan converts the database value to decimal.Decimal
func (nd *NullDecimal) Scan(value interface{}) error {
	if value == nil {
		nd.Decimal, nd.Valid = decimal.Decimal{}, false
		return nil
	}
	nd.Valid = true
	switch v := value.(type) {
	case float64:
		nd.Decimal = decimal.NewFromFloat(v)
	case string:
		d, err := decimal.NewFromString(v)
		if err != nil {
			return err
		}
		nd.Decimal = d
	case []byte:
		d, err := decimal.NewFromString(string(v))
		if err != nil {
			return err
		}
		nd.Decimal = d
	default:
		return fmt.Errorf("cannot scan type %T into NullDecimal", value)
	}
	return nil
}

// Value converts the decimal.Decimal value to database value
func (nd NullDecimal) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	}
	return nd.Decimal.String(), nil
}
