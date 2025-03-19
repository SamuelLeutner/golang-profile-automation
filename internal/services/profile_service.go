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
	ID_NATIONALITY = 0
	PROFILE_TYPE   = "FISICA"
)

var (
	tokenRequestCount = 0
	BEARER_TOKEN      = ""
	tokenMutex        sync.Mutex
	semaphore         = make(chan struct{}, 10)
)

func getAuthToken() (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if BEARER_TOKEN == "" || tokenRequestCount >= 10 {
		respAuth, err := j.AuthenticateJacad(os.Getenv("JACAD_AUTH_TOKEN"))
		if err != nil {
			return "", err
		}

		BEARER_TOKEN = respAuth.Token
		tokenRequestCount = 0
	}

	return BEARER_TOKEN, nil
}

func handleFileValues(c *gin.Context) ([]*m.Profile, string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(2)

	semaphore <- struct{}{}
	defer func() {
		time.Sleep(1 * time.Second)
		<-semaphore
	}()

	var token string
	var authErr error

	go func() {
		defer wg.Done()
		tokenRequestCount += 1
		t, err := getAuthToken()
		mu.Lock()
		token, authErr = t, err
		mu.Unlock()
	}()
	wg.Wait()

	if authErr != nil {
		return nil, "", authErr
	}

	var profiles []*m.Profile
	var errProfile error

	go func() {
		defer wg.Done()
		tokenRequestCount += 1
		pe, err := pp.HandleFileRow(c)
		mu.Lock()
		errProfile = err
		profiles = pe
		mu.Unlock()
	}()
	wg.Wait()

	if errProfile != nil {
		return nil, "", errProfile
	}

	return profiles, token, nil
}

func CreateProfile(c *gin.Context) (*http.Response, error) {
	profiles, token, err := handleFileValues(c)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var resp *http.Response
	var respErr error

	for _, p := range profiles {
		wg.Add(1)

		go func(p *m.Profile) {
			defer wg.Done()
			tokenRequestCount += 1
			city, cityErr := j.GetCityId(token, p.Estado, p.Cidade)

			if cityErr != nil {
				mu.Lock()
				respErr = cityErr
				mu.Unlock()
				return
			}

			orgId, err := strconv.Atoi(os.Getenv("ORG_ID"))
			if err != nil {
				mu.Lock()
				respErr = fmt.Errorf("Error to convert ORG_ID: %v", err)
				mu.Unlock()
				return
			}

			clientId, err := strconv.Atoi(os.Getenv("CLIENT_ID"))
			if err != nil {
				mu.Lock()
				respErr = fmt.Errorf("Error to convert CLIENT_ID: %v", err)
				mu.Unlock()
				return
			}

			p.OrgID = orgId
			p.ClientID = clientId
			p.ProfileType = PROFILE_TYPE
			p.IdCidadeEndereco = city.IdCidade
			p.IdNacionalidade = ID_NATIONALITY

			tokenRequestCount += 1

			// TODO: Remove the exit program after final test
			fmt.Println("User", p)
			os.Exit(1)

			respProfile, err := j.CreateProfile(token, p)
			if err != nil {
				mu.Lock()
				respErr = err
				mu.Unlock()
				return
			}

			mu.Lock()
			resp = respProfile
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	if respErr != nil {
		return nil, respErr
	}

	return resp, nil
}
