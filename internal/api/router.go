package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// InitRouter initialize a new chi router instance.
func (s *Server) InitRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Route("/order_status/webhook", func(r chi.Router) {

		r.Post("/payment_confirmation", s.paymentConfirmation)
		r.Post("/fraud_check", s.fraudCheck)
		r.Post("/vendor_confirmation", s.vendorConfirmation)

	})

	r.NotFound(s.handleNotFound)

	return r
}
