package validator

import (
    "strconv"
	"context"
	"errors"

	"github.com/go-payment-v2/application/domain/external"
)

// Schema struct defines a validation schema for product requests.
type Schema struct {
    Validate func(context.Context, any) error
}

func (s *Schema) PaymentAddSchema() Schema {
    return Schema{
        Validate: func(ctx context.Context, data any) error {
			
			req, ok := data.(external.PaymentRequest)
			if !ok {
                return errors.New("schema validation failed ! Please check the request body and try again.")
            }

			if req.PaymentDetail == nil {
                 return errors.New("schema validation failed ! field payment_detail is mandatory")
            }
   
            if req.Order == nil {
                return errors.New("schema validation failed ! field order.id is mandatory")
            }

            for i, item := range req.PaymentDetail {
                if item.Amount <= 0 {
                    return errors.New("schema validation failed ! field amount must be greater than 0 for payment detail at index " + strconv.Itoa(i))
                }
                if item.Currency == "" {
                    return errors.New("schema validation failed ! field currency is mandatory for payment detail at index " + strconv.Itoa(i))
                }
                if item.CreditCard == nil {
                    return errors.New("schema validation failed ! field credit_card is mandatory for payment detail at index " + strconv.Itoa(i))
                }
            }

            return nil
        },
    }
}
