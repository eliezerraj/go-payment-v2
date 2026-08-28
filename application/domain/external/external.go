package external

import "time"

type OrderRequest struct {
	ID			int			`json:"id,omitempty"`
	OrderNumber string		`json:"order_number,omitempty"`
}

type PaymentDetailRequest struct {
	DetailDate	time.Time 	`json:"payment_detail_date,omitempty"`
	Status		string 		`json:"status,omitempty"`
	Currency	string 		`json:"currency,omitempty"`
	Amount		float64 	`json:"amount,omitempty"`
	CreditCard	*CreditCardRequest	`json:"credit_card,omitempty"`
}

type CreditCardRequest struct {
	Pan		string	`json:"pan,omitempty"`
	Holder	string	`json:"holder,omitempty"`
	CVV		string	`json:"cvv,omitempty"`
	Password string	`json:"password,omitempty"`
}

type PaymentRequest struct {
	PaymentNumber 	string	`json:"payment_number,omitempty"`
	TransactionID	string	`json:"transaction_id,omitempty"`
	Type			string	`json:"type,omitempty"`
	Order			OrderRequest	`json:"order,omitempty"`
	PaymentDetail	[]*PaymentDetailRequest	`json:"payment_detail,omitempty"`
}

type PaymentResponse struct {
	Response    string	`json:"response"`
	Payment		any	`json:"payment,omitempty"`
}
