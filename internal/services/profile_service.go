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
	MAX_CONCURRENT_REQUESTS = 3
	PROFILE_TYPE            = "FISICA"
)

var (
	BEARER_TOKEN     string
	lastMinuteRefill time.Time
	tokenMutex       sync.Mutex
)

func getAuthToken() (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

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

	var mu sync.Mutex
	var errors []*m.ErrProfile

	semaphore := make(chan struct{}, MAX_CONCURRENT_REQUESTS)

	var wg sync.WaitGroup

	for i, p := range profiles {
		wg.Add(1)

		semaphore <- struct{}{}

		go func(i int, p *m.Profile) {
			defer wg.Done()
			defer func() { <-semaphore }()

			fmt.Printf("✅ Processing user %d/%d: %s\n", i+1, len(profiles), p.Name)

			time.Sleep(100 * time.Millisecond)

			city, cityErr := j.GetCityId(token, p.Estado, p.Cidade)
			if cityErr != nil {
				mu.Lock()
				errors = append(errors, &m.ErrProfile{Line: i, Cpf: p.Cpf, Err: cityErr.Error()})
				mu.Unlock()
				return
			}

			time.Sleep(100 * time.Millisecond)

			orgId, err := strconv.Atoi(os.Getenv("ORG_ID"))
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
		}(i, p)
	}

	wg.Wait()

	if len(errors) > 0 {
		if err := pp.SaveErrors(errors); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao salvar logs: %v", err)})
			return fmt.Errorf("erro ao salvar logs: %v", err)
		}
		fmt.Println("📂 Logs salvos em ProfileErrors.xlsx")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Processo de criação concluído.", "errors": len(errors)})
	return nil
}
