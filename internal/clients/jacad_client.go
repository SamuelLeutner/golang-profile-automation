package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
)

const (
	JACAD_URL_AUTHENTICATE = "/auth/token"
	JACAD_URL_STORE        = "/basicos/perfis"
	JACAD_URL_GET_CITY_ID  = "/basicos/locais/cidades"
)

type AuthResponse struct {
	Token string `json:"token"`
}

func AuthenticateJacad(token string) (*AuthResponse, error) {
	jacadUrl := os.Getenv("JACAD_URL")

	if jacadUrl == "" {
		return nil, fmt.Errorf("JACAD_URL are not set")
	}

	url := jacadUrl + JACAD_URL_AUTHENTICATE

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

func CreateProfile(bt string, requestBody *m.Profile) (*http.Response, error) {
	jacadUrl := os.Getenv("JACAD_URL")
	if jacadUrl == "" {
		return nil, fmt.Errorf("JACAD_URL are not set")
	}

	url := jacadUrl + JACAD_URL_STORE
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
	req.Header.Add("Authorization", "Bearer "+bt)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Error creating profile, status: %s", resp.Status)
	}

	if resp.StatusCode == 422 {
		return nil, fmt.Errorf("This profile already exists.")
	}

	return resp, nil
}

func GetCityId(bt string, uf string, search string) (*m.City, error) {
	jacadUrl := os.Getenv("JACAD_URL")
	if jacadUrl == "" {
		return nil, fmt.Errorf("JACAD_URL are not set")
	}

	ufFormat := strings.TrimSpace(strings.ToLower(uf))
	searchFormat := strings.TrimSpace(strings.ToLower(search))

	ufEncoded := url.QueryEscape(ufFormat)
	cityNameEncoded := url.QueryEscape(searchFormat)

	baseUrl := jacadUrl + JACAD_URL_GET_CITY_ID + fmt.Sprintf("?uf=%s&search=%s&currentPage=1&pageSize=10", ufEncoded, cityNameEncoded)
	client := &http.Client{}

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+bt)

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

	var response m.CityIdResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, err
	}

	if len(response.Elements) == 0 {
		return nil, fmt.Errorf("City not found for this search: '%s'", search)
	}

	city := response.Elements[0]

	if strings.ToLower(city.Uf) != ufFormat || strings.ToLower(city.Descricao) != searchFormat {
		return nil, fmt.Errorf("City not found")
	}

	return &city, nil
}
