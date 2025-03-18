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

const PROFILE_TYPE = "FISICA"

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

func handleCreateProfile(c *gin.Context) (*m.Profile, string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	p := &m.Profile{}
	var errProfile error
	var token string
	var authErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		err := pp.GetProfileContent(c, p)
		mu.Lock()
		errProfile = err
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		t, err := getAuthToken()
		mu.Lock()
		token, authErr = t, err
		mu.Unlock()
	}()

	wg.Wait()

	if errProfile != nil {
		return nil, "", errProfile
	}
	if authErr != nil {
		return nil, "", authErr
	}

	return p, token, nil
}

func CreateProfile(c *gin.Context) (*http.Response, error) {
	p, token, err := handleCreateProfile(c)
	if err != nil {
		return nil, err
	}

	var city *m.City
	var cityErr error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		city, cityErr = j.GetCityId(token, p.Estado, p.Cidade)
	}()

	wg.Wait()
	if cityErr != nil {
		return nil, cityErr
	}

	semaphore <- struct{}{}
	defer func() {
		time.Sleep(1 * time.Second)
		<-semaphore
	}()

	orgId, err := strconv.Atoi(os.Getenv("ORG_ID"))
	if err != nil {
		return nil, fmt.Errorf("Error to convert ORG_ID: %v", err)
	}

	clientId, err := strconv.Atoi(os.Getenv("CLIENT_ID"))
	if err != nil {
		return nil, fmt.Errorf("Error to convert CLIENT_ID: %v", err)
	}

	// TODO: Verify why ClientID not be set in `CreateProfile`
	p.OrgID = orgId
	p.ClientID = clientId
	p.ProfileType = PROFILE_TYPE
	p.IdCidadeEndereco = city.IdCidade

	resp, err := j.CreateProfile(token, p)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
