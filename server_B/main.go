package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type CEPRequest struct {
	CEP string `json:"cep"`
}

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func main() {
	log.Println("Starting server-b...")
	
	_, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log.Println("Starting HTTP server...")
	http.HandleFunc("/weather", weatherHandler)
	http.ListenAndServe(":8080", http.DefaultServeMux)
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Inicio de Servico weatherHandler B")
	var req CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("r.Body: %v", r.Body)
		log.Printf("req: %v", req)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.CEP) != 8 {
		log.Printf("Cep invalido: %v", req.CEP)
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()

	location, err := getLocation(ctx, req.CEP)
	if err != nil {
		log.Printf("náo encotrou o cep: %v", err)
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	tempC, err := getTemperature(ctx, location)
	if err != nil {
		log.Printf("error fetching temperature: %v", err)
		http.Error(w, "error fetching temperature", http.StatusInternalServerError)
		return
	}

	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	response := map[string]interface{}{
		"city":   location,
		"temp_C": tempC,
		"temp_F": tempF,
		"temp_K": tempK,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getLocation(ctx context.Context, cep string) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Get(fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var viaCEP ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viaCEP); err != nil {
		return "", err
	}

	return viaCEP.Localidade, nil
}

func getTemperature(ctx context.Context, location string) (float64, error) {
	log.Printf("Inicio getTemperature")

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return 0, fmt.Errorf("WEATHER_API_KEY is not set")
	}
	log.Printf("Using WEATHER_API_KEY: %s", apiKey)

	resp, err := http.Get(fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, location))
	if err != nil {
		log.Printf("error fetching temperature: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch temperature: %s", resp.Status)
	}

	var weatherAPI WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherAPI); err != nil {
		return 0, err
	}

	return weatherAPI.Current.TempC, nil
}
