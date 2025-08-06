package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"io/ioutil"
	"log"
	"net/http"
)

const name = "go.opentelemetry.io/otel/example/weather"

var (
	tracer     = otel.Tracer(name)
	meter      = otel.Meter(name)
	logger     = otelslog.NewLogger(name)
	requestCnt metric.Int64Counter
)

func init() {
	var err error
	requestCnt, err = meter.Int64Counter("weather.requests",
		metric.WithDescription("The number of weather measurement requests"),
		metric.WithUnit("{requests}"))
	if err != nil {
		panic(err)
	}
}

func InputHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Inicio de InputHandler")
	var req CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.CEP) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx, span := tracer.Start(r.Context(), "forward-to-service-b")
	defer span.End()

	resp, err := forwardToServiceB(ctx, req.CEP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	log.Println(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Falha ao ler o corpo da resposta")
		http.Error(w, "failed to read response body", http.StatusInternalServerError)
		return
	}

	weatherValue := attribute.String("weather", string(body))
	span.SetAttributes(weatherValue)
	requestCnt.Add(ctx, 1, metric.WithAttributes(weatherValue))

	w.Write(body)
}

func forwardToServiceB(ctx context.Context, cep string) (*http.Response, error) {
	log.Println("Inicio de forwardToServiceB")
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://server-b:8080/weather", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	reqBody, _ := json.Marshal(map[string]string{"cep": cep})
	req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))

	return client.Do(req)
}
