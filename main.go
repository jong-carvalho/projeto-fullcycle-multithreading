package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"regexp"
)

const weatherAPIKey = "c6e34b41fac04d51ad5115119250603" // Substitua pela sua chave da WeatherAPI

// Estrutura para ViaCEP
type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
}

// Estrutura para WeatherAPI
type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

// Estrutura de resposta do sistema
type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

// Valida se o CEP tem exatamente 8 dígitos numéricos
func isValidCEP(cep string) bool {
	return regexp.MustCompile(`^\d{8}$`).MatchString(cep)
}

// Busca cidade e estado pelo CEP via ViaCEP
func getCityFromCEP(cep string) (string, string, error) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var data ViaCEPResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", "", err
	}

	if data.Localidade == "" || data.UF == "" {
		return "", "", fmt.Errorf("CEP não encontrado")
	}
	return data.Localidade, data.UF, nil
}

// Obtém a temperatura da cidade pela WeatherAPI
func getTemperature(city string) (float64, error) {
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", weatherAPIKey, city)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data WeatherAPIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}

	return data.Current.TempC, nil
}

func getCEPHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cep := vars["cep"]

	if !isValidCEP(cep) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	city, _, err := getCityFromCEP(cep)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "can not find zipcode"})
		return
	}

	tempC, err := getTemperature(city)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch temperature"})
		return
	}

	response := TemperatureResponse{
		TempC: tempC,
		TempF: tempC*1.8 + 32,
		TempK: tempC + 273,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/cep/{cep}", getCEPHandler).Methods("GET")

	fmt.Println("Servidor rodando na porta 8080...")
	http.ListenAndServe(":8080", r)
}
