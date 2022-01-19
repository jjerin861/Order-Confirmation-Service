package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"order-confirmation-service/internal/model"
	"reflect"
	"testing"
	"time"

	"github.com/go-chi/chi"
)

func TestServer_deliveryConfirmation(t *testing.T) {

	type args struct {
		paymentConfirmationBody string
		fraudCheckBody          string
		vendorConfirmation      string
	}
	tests := []struct {
		name                string
		args                args
		expectedOrderStatus model.OrderStatus
	}{
		{
			name: "success",
			args: args{
				paymentConfirmationBody: `
				{
					"order_id": "de-ber-76898",
					"amount": 2000,
					"payment_reference": "h87j87y9q34rjo8rweqjo9fdckjhdslkmsdf",
					"payment_status": "confirmed"
				}
				`,
				fraudCheckBody: `
				{
					"reference_id": "de-ber-76898",
					"risk_points": 10 
				}
				`,
				vendorConfirmation: `
				{
					"order": "de-ber-76898",
					"status": "confirmed"
				}
				`,
			},
			expectedOrderStatus: model.OrderStatus{
				Amount: 2000,
				Status: "confirmed",
			},
		},
		{
			name: "payment confirmation failed",
			args: args{
				paymentConfirmationBody: `
				{
					"order_id": "de-ber-76898",
					"amount": 2000,
					"payment_reference": "h87j87y9q34rjo8rweqjo9fdckjhdslkmsdf",
					"payment_status": "failed"
				}
				`,
				fraudCheckBody: `
				{
					"reference_id": "de-ber-76898",
					"risk_points": 10 
				}
				`,
				vendorConfirmation: `
				{
					"order": "de-ber-76898",
					"status": "confirmed"
				}
				`,
			},
			expectedOrderStatus: model.OrderStatus{
				Amount:              2000,
				Status:              "errored",
				ConfirmationsFailed: []string{"payment"},
			},
		},
		{
			name: "fraud check failed",
			args: args{
				paymentConfirmationBody: `
				{
					"order_id": "de-ber-76898",
					"amount": 2000,
					"payment_reference": "h87j87y9q34rjo8rweqjo9fdckjhdslkmsdf",
					"payment_status": "confirmed"
				}
				`,
				fraudCheckBody: `
				{
					"reference_id": "de-ber-76898",
					"risk_points": 70 
				}
				`,
				vendorConfirmation: `
				{
					"order": "de-ber-76898",
					"status": "confirmed"
				}
				`,
			},
			expectedOrderStatus: model.OrderStatus{
				Amount:              2000,
				Status:              "errored",
				ConfirmationsFailed: []string{"fraud"},
			},
		},
		{
			name: "vendor confirmation failed",
			args: args{
				paymentConfirmationBody: `
				{
					"order_id": "de-ber-76898",
					"amount": 2000,
					"payment_reference": "h87j87y9q34rjo8rweqjo9fdckjhdslkmsdf",
					"payment_status": "confirmed"
				}
				`,
				fraudCheckBody: `
				{
					"reference_id": "de-ber-76898",
					"risk_points": 10 
				}
				`,
				vendorConfirmation: `
				{
					"order": "de-ber-76898",
					"status": "rejected"
				}
				`,
			},
			expectedOrderStatus: model.OrderStatus{
				Amount:              2000,
				Status:              "errored",
				ConfirmationsFailed: []string{"vendor"},
			},
		},
		{
			name: "payment, fraud, vendor failed",
			args: args{
				paymentConfirmationBody: `
				{
					"order_id": "de-ber-76898",
					"amount": 2000,
					"payment_reference": "h87j87y9q34rjo8rweqjo9fdckjhdslkmsdf",
					"payment_status": "failed"
				}
				`,
				fraudCheckBody: `
				{
					"reference_id": "de-ber-76898",
					"risk_points": 70 
				}
				`,
				vendorConfirmation: `
				{
					"order": "de-ber-76898",
					"status": "rejected"
				}
				`,
			},
			expectedOrderStatus: model.OrderStatus{
				Amount:              2000,
				Status:              "errored",
				ConfirmationsFailed: []string{"payment", "fraud", "vendor"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Port:                         ":8080",
				OrderData:                    make(map[string]*model.OrderStatus),
				DeliveryConfirmationEndpoint: "http://localhost:8080/delivery/de-ber-76898",
			}
			server := &http.Server{
				Addr:         s.Port,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
			}

			done := mockDeliveryConfirmationServer(t, server, tt.expectedOrderStatus)

			request(
				t,
				http.HandlerFunc(s.paymentConfirmation),
				"/order_status/webhook/payment_confirmation",
				[]byte(tt.args.paymentConfirmationBody),
				http.StatusOK,
			)
			request(
				t,
				http.HandlerFunc(s.fraudCheck),
				"/order_status/webhook/fraud_check",
				[]byte(tt.args.fraudCheckBody),
				http.StatusOK,
			)
			request(
				t,
				http.HandlerFunc(s.vendorConfirmation),
				"/order_status/webhook/vendor_confirmation",
				[]byte(tt.args.vendorConfirmation),
				http.StatusCreated,
			)
			<-done
			err := server.Shutdown(context.Background())
			if err != nil {
				t.Fatal("mock server shutdown failed:", err)
			}
		})
	}
}
func mockDeliveryConfirmationServer(t *testing.T, server *http.Server, expected model.OrderStatus) chan struct{} {
	done := make(chan struct{})
	r := chi.NewRouter()
	r.Put("/delivery/{order_id}", func(rw http.ResponseWriter, r *http.Request) {
		defer close(done)

		orderStatus := new(model.OrderStatus)
		if err := json.NewDecoder(r.Body).Decode(orderStatus); err != nil {
			t.Error("mockDeliveryConfirmationServer: json decoding failed:", err)
			return
		}

		expected.ProcessingTimeMS = orderStatus.ProcessingTimeMS
		if !reflect.DeepEqual(expected, *orderStatus) {
			t.Errorf("mockDeliveryConfirmationServer: order status not matching \n got: %v \n expected: %v", orderStatus, expected)
		}
	})

	server.Handler = r

	go func() {
		log.Printf("listening on port %s...\n", server.Addr)
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal("failed to start server: ", err)
		}
	}()
	return done
}
func request(t *testing.T, handler http.HandlerFunc, endpoint string, body []byte, expectedStatusCode int) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		t.Error("Error while creating payment_confirmation request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != expectedStatusCode {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
