package api

import (
	"log"
	"net/http"
	"order-confirmation-service/internal/model"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Server is the top level struct
type Server struct {
	Port                         string
	OrderData                    map[string]*model.OrderStatus
	DeliveryConfirmationEndpoint string
}

// NewServer returns new Server instance.
func NewServer(port string) *Server {
	return &Server{
		Port:                         port,
		OrderData:                    make(map[string]*model.OrderStatus),
		DeliveryConfirmationEndpoint: "https://en8ml4ab05csyji.m.pipedream.net",
	}
}

// Serve starts a service backed by an http.Server using default options.
func (s *Server) Serve() {
	if !strings.HasPrefix(s.Port, ":") {
		s.Port = ":" + s.Port
	}
	server := http.Server{
		Addr:         s.Port,
		Handler:      s.InitRouter(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on port %s...\n", server.Addr)
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal("failed to start server: ", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
