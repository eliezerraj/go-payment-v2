package external

type PaymentRequest struct {
	PaymentNumber string	`json:"payment_number,omitempty"`
	TransactionID	string	`json:"transaction_id,omitempty"`
	Type			string	`json:"type,omitempty"`
	OrderID			int		`json:"order_id,omitempty"`
	OrderNumber		string	`json:"order_number,omitempty"`
	Currency		string	`json:"currency,omitempty"`
	Amount			float64	`json:"amount,omitempty"`
	CreditCard		*CreditCardRequest	`json:"credit_card,omitempty"`
}

type CreditCardRequest struct {
	Pan		string	`json:"pan,omitempty"`
	Holder	string	`json:"holder,omitempty"`
	CVV		string	`json:"cvv,omitempty"`
	Password string	`json:"password,omitempty"`
}

type PaymentResponse struct {
	Response    string	`json:"response"`
	Payment		any	`json:"payment,omitempty"`
}

type OrderResponse struct {
	Response    string	`json:"response"`
	Order		any	`json:"order,omitempty"`
}
