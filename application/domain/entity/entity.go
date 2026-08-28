package entity

import (
	"time"
)

type Payment struct {
	ID				int		`json:"id,omitempty"`
	PaymentNumber 	string	`json:"payment_number,omitempty"`
	TransactionID 	string	`json:"transaction_id,omitempty"`
	Type			string 	`json:"type,omitempty"`
	Order			Order	`json:"order,omitempty"`	
	PaymentDetail 	[]*PaymentDetail	`json:"payment_detail,omitempty"`
	CreatedAt		time.Time 	`json:"created_at,omitempty"`
	UpdatedAt		*time.Time 	`json:"updated_at,omitempty"`		
}

type PaymentDetail struct {
	ID			int			`json:"id,omitempty"`
	DetailDate	time.Time 	`json:"payment_detail_date,omitempty"`
	Status		string 		`json:"status,omitempty"`
	Currency	string 		`json:"currency,omitempty"`
	Amount		float64 	`json:"amount,omitempty"`
	CreditCard	*CreditCard	`json:"credit_card,omitempty"`
	CreatedAt	time.Time 	`json:"created_at,omitempty"`
	UpdatedAt	*time.Time 	`json:"updated_at,omitempty"`
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

type Event struct {
	ID			string			`json:"event_id,omitempty"`
	Date		time.Time	`json:"event_date,omitempty"`
	Type		string		`json:"event_type,omitempty"`
	Metadata	map[string]interface{}	`json:"metadata,omitempty"`
	Data		interface{}	`json:"data,omitempty"`
}
