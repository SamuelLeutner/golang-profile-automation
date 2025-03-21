package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
)

type AuthResponse struct {
	Token string `json:"token"`
}

const (
	JACAD_URL_AUTHENTICATE = "/auth/token"
	JACAD_URL_STORE        = "/basicos/perfis"
	JACAD_URL_GET_CITY_ID  = "/basicos/locais/cidades"

	MAX_REQUESTS_PER_SECOND = 5
	MAX_REQUESTS_PER_MINUTE = 250
)

var (
	rateLimiterOnce  sync.Once
	secondTicker     *time.Ticker
	secondLimiter    chan struct{}
	minuteMutex      sync.Mutex
	minuteTokens     int
	lastMinuteRefill time.Time
)

func initRateLimiter() {
	rateLimiterOnce.Do(func() {
		log.Println("Initializing rate limiter...")
		secondLimiter = make(chan struct{}, MAX_REQUESTS_PER_SECOND)
		for i := 0; i < MAX_REQUESTS_PER_SECOND; i++ {
			secondLimiter <- struct{}{}
		}

		secondTicker = time.NewTicker(time.Second / MAX_REQUESTS_PER_SECOND)
		go func() {
			for range secondTicker.C {
				select {
				case secondLimiter <- struct{}{}:
				default:
				}
			}
		}()

		minuteTokens = MAX_REQUESTS_PER_MINUTE
		lastMinuteRefill = time.Now()

		go func() {
			ticker := time.NewTicker(10 * time.Second)
			for range ticker.C {
				minuteMutex.Lock()
				now := time.Now()
				elapsed := now.Sub(lastMinuteRefill)
				if elapsed >= time.Minute {
					minuteTokens = MAX_REQUESTS_PER_MINUTE
					lastMinuteRefill = now
				}
				minuteMutex.Unlock()
			}
		}()
	})
}

func waitForRateLimit() {
	initRateLimiter()
	<-secondLimiter
	minuteMutex.Lock()
	if minuteTokens <= 0 {
		log.Println("Waiting for rate limit...")
	}
	minuteTokens--
	minuteMutex.Unlock()
}

func handleErrorResponse(resp *http.Response) error {
	if resp.StatusCode == 429 {
		log.Println("Rate limit exceeded - 429")
		return fmt.Errorf("rate limit exceeded - 429")
	}
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Error: status: %s, body: %s\n", resp.Status, string(bodyBytes))
		return fmt.Errorf("error: status: %s, body: %s", resp.Status, string(bodyBytes))
	}
	return nil
}

func AuthenticateJacad(token string) (*AuthResponse, error) {
	log.Println("Authenticating with Jacad...")
	waitForRateLimit()
	jacadUrl := os.Getenv("JACAD_URL")
	if jacadUrl == "" {
		log.Println("JACAD_URL is not set")
		return nil, fmt.Errorf("JACAD_URL is not set")
	}
	url := jacadUrl + JACAD_URL_AUTHENTICATE
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := json.Unmarshal(bodyBytes, &authResp); err != nil {
		log.Printf("Error unmarshalling response: %v\n", err)
		return nil, err
	}

	log.Println("Authentication successful")
	return &authResp, nil
}

func CreateProfile(bt string, requestBody *m.Profile) (*http.Response, error) {
	log.Println("Creating profile...")
	waitForRateLimit()
	jacadUrl := os.Getenv("JACAD_URL")
	if jacadUrl == "" {
		log.Println("JACAD_URL is not set")
		return nil, fmt.Errorf("JACAD_URL is not set")
	}
	url := jacadUrl + JACAD_URL_STORE
	client := &http.Client{Timeout: 30 * time.Second}
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(requestBody)
	if err != nil {
		log.Printf("Error encoding request body: %v\n", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+bt)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request error: %v\n", err)
		return nil, err
	}

	if resp.StatusCode == 422 {
		log.Println("Profile already exists.")
		resp.Body.Close()
		return nil, fmt.Errorf("this profile already exists")
	}

	if resp.StatusCode == 400 {
		b, _ := io.ReadAll(resp.Body)
		log.Println("Error in request body for profile.", string(b))
		resp.Body.Close()
		return nil, fmt.Errorf("Error in request body for profile.")
	}

	if err := handleErrorResponse(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}

	log.Println("Profile created successfully")
	return resp, nil
}

func GetCityId(bt string, uf string, search string) (*m.City, error) {
	log.Println("Getting city...")

	waitForRateLimit()

	jacadUrl := os.Getenv("JACAD_URL")
	if jacadUrl == "" {
		return nil, fmt.Errorf("JACAD_URL are not set")
	}

	ufFormat := strings.TrimSpace(strings.ToLower(uf))
	searchFormat := strings.TrimSpace(strings.ToLower(search))
	ufEncoded := url.QueryEscape(ufFormat)
	cityNameEncoded := url.QueryEscape(searchFormat)

	baseUrl := jacadUrl + JACAD_URL_GET_CITY_ID + fmt.Sprintf("?uf=%s&search=%s&currentPage=0&pageSize=10", ufEncoded, cityNameEncoded)

	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		log.Println("error creating request.")
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+bt)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := handleErrorResponse(resp); err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response m.CityIdResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		log.Println("error parsing response.")
		return nil, fmt.Errorf("error parsing response: %w, body: %s", err, string(bodyBytes))
	}

	if len(response.Elements) == 0 {
		log.Println("city not found for search.")
		return nil, fmt.Errorf("City not found for this search: '%s'", search)
	}

	city := response.Elements[0]

	log.Println("Get city successfully")

	return &city, nil
}
