package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const weatherAPIKey = "c6e34b41fac04d51ad5115119250603"

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
}

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func getCityFromCEP(cep string) (string, error) {
	cep = strings.ReplaceAll(cep, "-", "")
	if len(cep) != 8 {
		return "", fmt.Errorf("invalid zipcode")
	}

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data ViaCEPResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	if data.Localidade == "" {
		return "", fmt.Errorf("can not find zipcode")
	}
	return data.Localidade, nil
}

func getWeather(city string) (float64, error) {
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", weatherAPIKey, city)
	fmt.Println("Fazendo requisição para:", url) // Debug

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Erro ao fazer requisição para WeatherAPI:", err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler resposta da WeatherAPI:", err)
		return 0, err
	}

	fmt.Println("Resposta da WeatherAPI:", string(body)) // Debug

	var data WeatherResponse
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Erro ao decodificar JSON da WeatherAPI:", err)
		return 0, err
	}

	return data.Current.TempC, nil
}

func convertTemperatures(tempC float64) TemperatureResponse {
	return TemperatureResponse{
		TempC: tempC,
		TempF: tempC*1.8 + 32,
		TempK: tempC + 273,
	}
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	// Extrai o CEP da URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "Invalid request: missing zipcode", http.StatusBadRequest)
		return
	}

	cep := parts[2]

	// Obtém a cidade a partir do CEP
	city, err := getCityFromCEP(cep)
	if err != nil {
		if err.Error() == "invalid zipcode" {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		} else {
			http.Error(w, "Can not find zipcode", http.StatusNotFound)
		}
		return
	}

	// Obtém a temperatura da cidade
	tempC, err := getWeather(city)
	if err != nil {
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}

	// Retorna os dados formatados
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(convertTemperatures(tempC))
}

func main() {
	http.HandleFunc("/weather/", weatherHandler)

	fmt.Println("Servidor rodando na porta 8080...")
	http.ListenAndServe(":8080", nil)
}
