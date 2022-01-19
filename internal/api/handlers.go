package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"order-confirmation-service/internal/model"
	"time"

	"github.com/cenkalti/backoff"
)

// paymentConfirmation handler.
func (s *Server) paymentConfirmation(w http.ResponseWriter, r *http.Request) {
	paymentData := new(model.PaymentConfirmation)
	if err := json.NewDecoder(r.Body).Decode(paymentData); err != nil {
		log.Println("paymentConfirmation: json decoding failed:", err)
		response(w, http.StatusBadRequest, err.Error())
		return
	}
	orderStatus := &model.OrderStatus{}
	val, ok := s.OrderData[paymentData.OrderID]
	if ok {
		orderStatus = val
	} else {
		s.OrderData[paymentData.OrderID] = orderStatus
		orderStatus.OrderID = paymentData.OrderID
		orderStatus.StartTime = time.Now()
	}
	orderStatus.Amount = paymentData.Amount
	orderStatus.PaymentConfirmationProcessed = true
	if paymentData.Payment_status != "confirmed" {
		orderStatus.ConfirmationsFailed = append(orderStatus.ConfirmationsFailed, "payment")
	}

	//call delivery endpoint
	s.deliveryConfirmation(orderStatus)

	response(w, http.StatusOK, "success")
}

// fraudCheck handler.
func (s *Server) fraudCheck(w http.ResponseWriter, r *http.Request) {
	fraudCheckData := new(model.FraudCheck)
	if err := json.NewDecoder(r.Body).Decode(fraudCheckData); err != nil {
		log.Println("fraudCheck: json decoding failed:", err)
		response(w, http.StatusBadRequest, err.Error())
		return
	}
	orderStatus := &model.OrderStatus{}
	val, ok := s.OrderData[fraudCheckData.ReferenceID]
	if ok {
		orderStatus = val
	} else {
		s.OrderData[fraudCheckData.ReferenceID] = orderStatus
		orderStatus.OrderID = fraudCheckData.ReferenceID
		orderStatus.StartTime = time.Now()
	}
	orderStatus.FraudCheckProcessed = true
	if fraudCheckData.RiskPoints > 60 {
		orderStatus.ConfirmationsFailed = append(orderStatus.ConfirmationsFailed, "fraud")
	}

	//call delivery endpoint
	s.deliveryConfirmation(orderStatus)

	response(w, http.StatusOK, "success")
}

// vendorConfirmation handler.
func (s *Server) vendorConfirmation(w http.ResponseWriter, r *http.Request) {
	vendorConfirmationData := new(model.VendorConfirmation)
	if err := json.NewDecoder(r.Body).Decode(vendorConfirmationData); err != nil {
		log.Println("fraudCheck: json decoding failed:", err)
		response(w, http.StatusBadRequest, err.Error())
		return
	}
	orderStatus := &model.OrderStatus{}
	val, ok := s.OrderData[vendorConfirmationData.OrderID]
	if ok {
		orderStatus = val
	} else {
		s.OrderData[vendorConfirmationData.OrderID] = orderStatus
		orderStatus.OrderID = vendorConfirmationData.OrderID
		orderStatus.StartTime = time.Now()
	}
	orderStatus.VendorConfirmationProcessed = true
	if vendorConfirmationData.Status != "confirmed" {
		orderStatus.ConfirmationsFailed = append(orderStatus.ConfirmationsFailed, "vendor")
	}

	//call delivery endpoint
	s.deliveryConfirmation(orderStatus)

	response(w, http.StatusCreated, "success")
}

// deliveryConfirmation is to call delivery confirmation endpoint.
func (s *Server) deliveryConfirmation(order *model.OrderStatus) {

	if order.PaymentConfirmationProcessed && order.FraudCheckProcessed && order.VendorConfirmationProcessed {
		request := func() error {

			if len(order.ConfirmationsFailed) == 0 {
				order.Status = "confirmed"
			} else {
				order.Status = "errored"

			}
			order.ProcessingTimeMS = fmt.Sprintf("%dms", time.Since(order.StartTime).Milliseconds())
			orderByte, err := json.Marshal(order)
			if err != nil {
				log.Println("json marshal failed: ", err)
				return err
			}
			req, err := http.NewRequest(
				"PUT",
				s.DeliveryConfirmationEndpoint,
				bytes.NewBuffer(orderByte),
			)
			if err != nil {
				log.Println("new request failed: ", err)
				return err
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("request failed: ", err)
				return err
			}
			if res.StatusCode != http.StatusOK {
				log.Println("response not matching, code: ", res.StatusCode)
				return nil
			}
			fmt.Println("deliveryConfirmation completed")
			return nil
		}

		b := backoff.NewExponentialBackOff()
		err := backoff.Retry(request, backoff.WithMaxRetries(b, 3))
		if err != nil {
			log.Println("retry failed: ", err)
		}
	}
}

// handleNotFound sets up custom not found response.
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	response(
		w,
		http.StatusNotFound,
		"OOPS. We tried 404 times, but couldn't find that resource.",
	)
}
