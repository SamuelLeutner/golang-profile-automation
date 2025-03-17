package services

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	jacad "github.com/SamuelLeutner/golang-profile-automation/internal/clients"
	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
	"github.com/gin-gonic/gin"
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
		respAuth, err := jacad.AuthenticateJacad(os.Getenv("JACAD_AUTH_TOKEN"))
		if err != nil {
			return "", err
		}

		BEARER_TOKEN = respAuth.Token
		tokenRequestCount = 0
	}

	return BEARER_TOKEN, nil
}

func CreatePerfil(c *gin.Context) (*http.Response, error) {
	var profile m.Profile

	if err := c.ShouldBindJSON(&profile); err != nil {
		return nil, err
	}

	semaphore <- struct{}{}
	defer func() {
		time.Sleep(1 * time.Second)
		<-semaphore
	}()

	token, err := getAuthToken()
	if err != nil {
		return nil, err
	}

	respCityId, err := jacad.GetCityId(token, "pr", "guarapuava")
	if err != nil {
		return nil, err
	}

	orgId, err := strconv.Atoi(os.Getenv("ORG_ID"))
	if err != nil {
		return nil, fmt.Errorf("Error to convert ORG_ID: %v", err)
	}

	clientId, err := strconv.Atoi(os.Getenv("CLIENT_ID"))
	if err != nil {
		return nil, fmt.Errorf("Error to convert CLIENT_ID: %v", err)
	}

	profile.OrgID = orgId
	profile.ClientID = clientId
	profile.IdCidadeEndereço = respCityId.IdCidade

	fmt.Println(&profile)
	os.Exit(1)

	resp, err := jacad.CreatePerfil(token, &profile)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
