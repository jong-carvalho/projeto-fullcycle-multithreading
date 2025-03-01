package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Estrutura para Brasil API
type BrasilAPIResponse struct {
	CEP          string `json:"cep"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

// Estrutura para ViaCEP
type ViaCEPResponse struct {
	CEP        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
}

// Estrutura de endereço padronizada
type Address struct {
	CEP        string
	Logradouro string
	Bairro     string
	Localidade string
	UF         string
	Source     string
}

// Função para fazer requisições e tratar os dados corretamente
func fetchFromAPI(url, source string, resultChan chan<- Address, errChan chan<- error) {
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		errChan <- fmt.Errorf("Erro ao acessar %s: %v", source, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errChan <- fmt.Errorf("Erro ao ler resposta de %s: %v", source, err)
		return
	}

	var address Address
	if source == "Brasil API" {
		var data BrasilAPIResponse
		if err := json.Unmarshal(body, &data); err != nil {
			errChan <- fmt.Errorf("Erro ao decodificar resposta de %s: %v", source, err)
			return
		}
		address = Address{
			CEP:        data.CEP,
			Logradouro: data.Street,
			Bairro:     data.Neighborhood,
			Localidade: data.City,
			UF:         data.State,
			Source:     source,
		}
	} else {
		var data ViaCEPResponse
		if err := json.Unmarshal(body, &data); err != nil {
			errChan <- fmt.Errorf("Erro ao decodificar resposta de %s: %v", source, err)
			return
		}
		address = Address{
			CEP:        data.CEP,
			Logradouro: data.Logradouro,
			Bairro:     data.Bairro,
			Localidade: data.Localidade,
			UF:         data.UF,
			Source:     source,
		}
	}

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
	case err := <-errChan:
		fmt.Printf("Erro: %v\n", err)
	case <-time.After(1 * time.Second):
		fmt.Println("Erro: Timeout. Nenhuma API respondeu a tempo.")
	}
}
