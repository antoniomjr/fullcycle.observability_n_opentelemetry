package main

import (
	"bytes"
	"context"
	"encoding/json"
	//"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	//"go.opentelemetry.io/otel/exporters/zipkin"
	//"go.opentelemetry.io/otel/semconv/v1.7.0"
	//"go.opentelemetry.io/otel/sdk/resource"
	//sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"io/ioutil"
	"net/http"
)

type CEPRequest struct {
	CEP string `json:"cep"`
}

func main() {
	//initTracer()

	http.HandleFunc("/input", inputHandler)
	http.ListenAndServe(":8081", otelhttp.NewHandler(http.DefaultServeMux, "input-server"))
}

func inputHandler(w http.ResponseWriter, r *http.Request) {
	var req CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.CEP) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	tracer := otel.Tracer("input-service")
	ctx, span := tracer.Start(ctx, "forward-to-service-b")
	defer span.End()

	resp, err := forwardToServiceB(ctx, req.CEP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response body", http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func forwardToServiceB(ctx context.Context, cep string) (*http.Response, error) {
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/weather", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	reqBody, _ := json.Marshal(map[string]string{"cep": cep})
	req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))

	return client.Do(req)
}

//func initTracer() {
//	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
//	if err != nil {
//		fmt.Printf("failed to initialize zipkin exporter: %v\n", err)
//		return
//	}
//
//	tp := sdktrace.NewTracerProvider(
//		sdktrace.WithBatcher(exporter),
//		sdktrace.WithResource(resource.NewWithAttributes(
//			semconv.ServiceNameKey.String("input-service"),
//		)),
//	)
//	otel.SetTracerProvider(tp)
//}
