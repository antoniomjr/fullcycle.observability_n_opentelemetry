package main

import (
	"context"
	"errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type CEPRequest struct {
	CEP string `json:"cep"`
}

func main() {
	log.Println("Starting the application...")
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() (err error) {
	log.Println("Setting up signal handling...")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log.Println("Setting up OpenTelemetry...")
	otelShutdown, err := SetupOTelSDK(ctx)
	if err != nil {
		log.Printf("Error setting up OpenTelemetry: %v", err)
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	log.Println("Starting HTTP server...")
	srv := &http.Server{
		Addr:         ":8081",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	log.Println("Waiting for interruption...")
	select {
	case err = <-srvErr:
		log.Printf("HTTP server error: %v", err)
		return
	case <-ctx.Done():
		log.Println("Received interrupt signal, shutting down...")
		stop()
	}

	log.Println("Shutting down HTTP server...")
	err = srv.Shutdown(context.Background())
	return
}

func newHTTPHandler() http.Handler {
	log.Printf("Setting up HTTP handler...")
	mux := http.NewServeMux()

	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		log.Printf("pattern: %v", pattern)
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}
	log.Printf("handleFunc: %v", handleFunc)
	handleFunc("/input", InputHandler)

	handler := otelhttp.NewHandler(mux, "/")
	log.Printf("Fim HTTP handler...")
	return handler
}
