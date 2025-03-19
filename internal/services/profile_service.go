package services

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	j "github.com/SamuelLeutner/golang-profile-automation/internal/clients"
	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
	pp "github.com/SamuelLeutner/golang-profile-automation/pkg/profile"
	"github.com/gin-gonic/gin"
)

const (
	ID_NATIONALITY          = 0
	REQUEST_REFILL_RATE     = 1
	MAX_REQUESTS_PER_SECOND = 10
	MAX_REQUESTS_PER_HOUR   = 1000
	PROFILE_TYPE            = "FISICA"
)

var (
	BEARER_TOKEN string
	tokenMutex   sync.Mutex
	semaphore    = make(chan struct{}, 10)

	tokenRequests  = 10
	businessTokens = 1000

	lastRefill         = time.Now()
	lastBusinessRefill = time.Now()
)

func RefilTokens() {
	for {
		time.Sleep(1 * time.Second)
		tokenMutex.Lock()

		if tokenRequests < MAX_REQUESTS_PER_SECOND {
			tokenRequests++
		}

		if time.Since(lastBusinessRefill) >= time.Hour {
			businessTokens = MAX_REQUESTS_PER_HOUR
			lastBusinessRefill = time.Now()
		}

		tokenMutex.Unlock()
	}
}

func getAuthToken() (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	for tokenRequests <= 0 {
		tokenMutex.Unlock()
		time.Sleep(100 * time.Microsecond)
		tokenMutex.Lock()
	}

	tokenRequests--
	businessTokens--

	if businessTokens <= 0 {
		for businessTokens <= 0 {
			tokenMutex.Unlock()
			time.Sleep(10 * time.Second)
			tokenMutex.Lock()
		}
	}

	if BEARER_TOKEN == "" {
		respAuth, err := j.AuthenticateJacad(os.Getenv("JACAD_AUTH_TOKEN"))
		if err != nil {
			return "", err
		}

		BEARER_TOKEN = respAuth.Token
	}

	return BEARER_TOKEN, nil
}

func handleFileValues(c *gin.Context) ([]*m.Profile, string, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(2)

	semaphore <- struct{}{}
	defer func() {
		time.Sleep(1 * time.Second)
		<-semaphore
	}()

	var token string
	var authErr error
	var errProfile error
	var profiles []*m.Profile

	go func() {
		defer wg.Done()
		t, err := getAuthToken()
		mu.Lock()
		token, authErr = t, err
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		pe, err := pp.HandleFileRow(c)
		mu.Lock()
		errProfile = err
		profiles = pe
		mu.Unlock()
	}()
	wg.Wait()

	if authErr != nil {
		return nil, "", authErr
	}

	if errProfile != nil {
		return nil, "", errProfile
	}

	return profiles, token, nil
}

func CreateProfile(c *gin.Context) error {
	profiles, token, err := handleFileValues(c)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []*m.ErrProfile

	for i, p := range profiles {
		wg.Add(1)

		go func(p *m.Profile) {
			defer wg.Done()

			fmt.Printf("Processing user %d/%d: %s\n", i+1, len(profiles), p.Name)

			city, cityErr := j.GetCityId(token, p.Estado, p.Cidade)

			if cityErr != nil {
				mu.Lock()
				errors = append(errors, &m.ErrProfile{Line: i, Cpf: p.Cpf, Err: cityErr.Error()})
				mu.Unlock()
				return
			}

			orgId, err := strconv.Atoi(os.Getenv("ORG_ID"))
			if err != nil {
				mu.Lock()
				errors = append(errors, &m.ErrProfile{Line: i, Cpf: p.Cpf, Err: err.Error()})
				mu.Unlock()
				return
			}

			clientId, err := strconv.Atoi(os.Getenv("CLIENT_ID"))
			if err != nil {
				mu.Lock()
				errors = append(errors, &m.ErrProfile{Line: i, Cpf: p.Cpf, Err: err.Error()})
				mu.Unlock()
				return
			}

			p.OrgID = orgId
			p.ClientID = clientId
			p.ProfileType = PROFILE_TYPE
			p.IdCidadeEndereco = city.IdCidade
			p.IdNacionalidade = ID_NATIONALITY

			respProfile, err := j.CreateProfile(token, p)
			if err != nil {
				mu.Lock()
				errors = append(errors, &m.ErrProfile{Line: i, Cpf: p.Cpf, Err: err.Error()})
				mu.Unlock()
				return
			}

			mu.Lock()
			_ = respProfile
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	if len(errors) > 0 {
		if err := pp.SaveErrors(errors); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error saving errors file: %v", err)})
			return fmt.Errorf("error saving errors file: %v", err)
		}
		fmt.Println("📂 Errors saved in ProfileErrors.xlsx")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile creation process completed.", "errors": len(errors)})
	return nil
}
