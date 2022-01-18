package model

import "time"

// Location model to hold latitude an longitude.
type PaymentConfirmation struct {
	OrderID        string `json:"order_id,omitempty"`
	Amount         int    `json:"amount,omitempty"`
	Payment_status string `json:"payment_status,omitempty"`
}

type OrderStatus struct {
	OrderID                      string
	Status                       string   `json:"status,omitempty"`
	Amount                       int      `json:"amount,omitempty"`
	ConfirmationsFailed          []string `json:"confirmations_failed,omitempty"`
	ProcessingTimeMS             string   `json:"processing_time_ms,omitempty"`
	PaymentConfirmationProcessed bool
	FraudCheckProcessed          bool
	VendorConfirmationProcessed  bool
	StartTime                    time.Time
}

type FraudCheck struct {
	ReferenceID string `json:"reference_id,omitempty"`
	RiskPoints  int    `json:"risk_points,omitempty"`
}
type VendorConfirmation struct {
	OrderID string `json:"order,omitempty"`
	Status  string `json:"status,omitempty"`
}
