package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Address struct {
	CEP        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Source     string `json:"source"`
}

func fetchFromAPI(url string, source string, resultChan chan<- Address, errChan chan<- error) {
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		errChan <- err
		return
	}
	defer resp.Body.Close()

	var address Address
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		errChan <- err
		return
	}

	address.Source = source
	resultChan <- address
}

func main() {
	cep := "01153000"
	api1 := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	api2 := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	resultChan := make(chan Address, 2)
	errChan := make(chan error, 2)

	go fetchFromAPI(api1, "Brasil API", resultChan, errChan)

	go fetchFromAPI(api2, "ViaCEP", resultChan, errChan)

	select {
	case result := <-resultChan:
		fmt.Printf("API mais rápida: %s\n", result.Source)
		fmt.Printf("Endereço: %s, %s, %s - %s, %s\n", result.Logradouro, result.Bairro, result.Localidade, result.UF, result.CEP)
	case <-time.After(1 * time.Second):
		fmt.Println("Erro: Timeout. Nenhuma API respondeu a tempo.")
	}
}
