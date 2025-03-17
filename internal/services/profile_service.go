package services

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	jacad "github.com/SamuelLeutner/golang-profile-automation/internal/clients"
	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
	"github.com/gin-gonic/gin"
)

func CreatePerfil(c *gin.Context) (*http.Response, error) {
	var profile m.Profile

	if err := c.ShouldBindJSON(&profile); err != nil {
		return nil, err
	}

	respAuth, err := jacad.AuthenticateJacad(os.Getenv("JACAD_AUTH_TOKEN"))
	if err != nil {
		return nil, err
	}

	respCityId, err := jacad.GetCityId(respAuth.Token, "pr", "guarapuava")
	if err != nil {
		return nil, err
	}

	orgId, err := strconv.Atoi(os.Getenv("ORG_ID"))
	if err != nil {
		return nil, fmt.Errorf("Erro ao converter ORG_ID: %v", err)
	}

	clientId, err := strconv.Atoi(os.Getenv("CLIENT_ID"))
	if err != nil {
		return nil, fmt.Errorf("Erro ao converter CLIENT_ID: %v", err)
	}

	profile.OrgID = orgId
	profile.ClientID = clientId
	profile.IdCidadeEndereço = respCityId.IdCidade

	parsedDate, err := time.Parse("02/01/2006", profile.DateOfBirth.Format("02/01/2006"))
	if err != nil {
		return nil, fmt.Errorf("Erro ao formatar a data de nascimento: %v", err)
	}

	profile.DateOfBirth = m.CustomTime{Time: parsedDate}

	fmt.Println("Criando perfil com os dados:", profile)
	fmt.Println("orgId:", profile.OrgID)
	fmt.Println("clientId:", profile.ClientID)

	os.Exit(1)

	resp, err := jacad.CreatePerfil(respAuth.Token, &profile)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
