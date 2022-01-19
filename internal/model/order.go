package model

import "time"

// Location model to hold latitude an longitude.
type PaymentConfirmation struct {
	OrderID        string `json:"order_id"`
	Amount         int    `json:"amount"`
	Payment_status string `json:"payment_status"`
}

type OrderStatus struct {
	OrderID                      string    `json:"-"`
	Status                       string    `json:"status"`
	Amount                       int       `json:"amount"`
	ConfirmationsFailed          []string  `json:"confirmations_failed"`
	ProcessingTimeMS             string    `json:"processing_time_ms"`
	PaymentConfirmationProcessed bool      `json:"-"`
	FraudCheckProcessed          bool      `json:"-"`
	VendorConfirmationProcessed  bool      `json:"-"`
	StartTime                    time.Time `json:"-"`
}

type FraudCheck struct {
	ReferenceID string `json:"reference_id"`
	RiskPoints  int    `json:"risk_points"`
}
type VendorConfirmation struct {
	OrderID string `json:"order"`
	Status  string `json:"status"`
}
