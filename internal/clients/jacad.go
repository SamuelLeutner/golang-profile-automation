package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
)

type AuthResponse struct {
	Token string `json:"token"`
}

type City struct {
	IdCidade  int    `json:"idCidade"`
	Descricao string `json:"descricao"`
	Uf        string `json:"uf"`
	Estado    string `json:"estado"`
}

type CityIdResponse struct {
	Elements []City `json:"elements"`
}

func AuthenticateJacad(token string) (*AuthResponse, error) {
	jacadUrl := os.Getenv("JACAD_URL")

	if jacadUrl == "" {
		return nil, fmt.Errorf("JACAD_URL are not set")
	}

	url := jacadUrl + "/auth/token"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)

	resp, err := client.Do(req)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Error in `AuthenticateJacad`. Status: %s", resp.Status)
	}

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := json.Unmarshal(bodyBytes, &authResp); err != nil {
		return nil, err
	}

	return &authResp, nil
}

func CreatePerfil(bearerToken string, requestBody *m.Profile) (*http.Response, error) {
	jacadUrl := os.Getenv("JACAD_URL")
	if jacadUrl == "" {
		return nil, fmt.Errorf("JACAD_URL are not set")
	}

	url := jacadUrl + "/basicos/perfis"
	client := &http.Client{}

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Error creating profile, status: %s", resp.Status)
	}

	return resp, nil
}

func GetCityId(bearerToken string, uf string, search string) (*City, error) {
	jacadUrl := os.Getenv("JACAD_URL")
	if jacadUrl == "" {
		return nil, fmt.Errorf("JACAD_URL are not set")
	}

	url := jacadUrl + "/basicos/locais/cidades?uf=" + uf + "&search=" + search + "&currentPage=1&pageSize=10"
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Error to search city, status: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response CityIdResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		fmt.Println("Error in Json Unmarshal City:", err)
		return nil, err
	}

	if len(response.Elements) == 0 {
		return nil, fmt.Errorf("City not found for this search: '%s'", search)
	}

	city := response.Elements[0]

	if strings.ToLower(city.Uf) != strings.ToLower(uf) || strings.ToLower(city.Descricao) != strings.ToLower(search) {
		return nil, fmt.Errorf("City not found")
	}

	return &city, nil
}
