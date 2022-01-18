package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"order-confirmation-service/internal/model"
	"time"
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
	if paymentData.Payment_status == "confirmed" {
		orderStatus.PaymentConfirmationProcessed = true
	} else {
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
	if fraudCheckData.RiskPoints <= 60 {
		orderStatus.FraudCheckProcessed = true
	} else {
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
	if vendorConfirmationData.Status == "confirmed" {
		orderStatus.VendorConfirmationProcessed = true
	} else {
		orderStatus.ConfirmationsFailed = append(orderStatus.ConfirmationsFailed, "vendor")
	}

	//call delivery endpoint
	s.deliveryConfirmation(orderStatus)

	response(w, http.StatusCreated, "success")
}

// deliveryConfirmation is to call delivery confirmation endpoint.
func (s *Server) deliveryConfirmation(order *model.OrderStatus) {

	if order.PaymentConfirmationProcessed && order.FraudCheckProcessed && order.VendorConfirmationProcessed {
		order.ProcessingTimeMS = time.Since(order.StartTime).String()
		orderByte, err := json.Marshal(order)
		if err != nil {
			log.Println("json marshal failed: ", err)
			return
		}
		req, err := http.NewRequest(
			"PUT",
			s.DeliveryConfirmationEndpoint,
			bytes.NewBuffer(orderByte),
		)
		if err != nil {
			log.Println("new request failed: ", err)
			return
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("request failed: ", err)
			return
		}
		if res.StatusCode != http.StatusOK {
			log.Println("response not matching, code: ", res.StatusCode)
			return
		}
		fmt.Println("deliveryConfirmation completed")
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
