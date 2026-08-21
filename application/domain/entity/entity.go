package entity

import (
	"time"
)

type Payment struct {
	ID			int			`json:"id,omitempty"`
	PaymentNumber string	`json:"payment_number,omitempty"`
	TransactionID string	`json:"transaction_id,omitempty"`
	OrderID     int         `json:"order_id,omitempty"`
	OrderNumber	string		`json:"order_number,omitempty"`
	Type		string 		`json:"type,omitempty"`
	Status		string 		`json:"status,omitempty"`
	CustomerID	string 		`json:"customer_id,omitempty"`
	PaymentDate	time.Time 	`json:"payment_date,omitempty"`
	Currency	string 		`json:"currency,omitempty"`
	Amount		float64 	`json:"amount,omitempty"`
	CreditCard	*CreditCard	`json:"credit_card,omitempty"`
	CreatedAt	time.Time 	`json:"created_at,omitempty"`
	UpdatedAt	*time.Time 	`json:"updated_at,omitempty"`
	StepProcess	*[]StepProcess `json:"step_process,omitempty"`			
}

type CreditCard struct {
	Pan				string	`json:"pan,omitempty"`
	Holder			string	`json:"holder,omitempty"`
	ExpirationDate	string	`json:"expiration_date,omitempty"`
	Password		string	`json:"password,omitempty"`
	CVV				string	`json:"cvv,omitempty"`
}

type Order struct {
	ID			int			`json:"id,omitempty"`
	OrderNumber string		`json:"order_number,omitempty"`
}

type StepProcess struct {
	Name		string  	`json:"step_process,omitempty"`
	ProcessedAt	time.Time 	`json:"processed_at,omitempty"`
}