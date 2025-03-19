package profile

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	interfaces "github.com/SamuelLeutner/golang-profile-automation/internal/models"
	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
	"github.com/gin-gonic/gin"
	e "github.com/xuri/excelize/v2"
)

var sexoMap = map[string]string{
	"M": "MASCULINO",
	"F": "FEMININO",
}

var maritalStatus = map[string]string{
	"Solteiro(a)":   "SOLTEIRO",
	"Casado(a)":     "CASADO",
	"Divorciado(a)": "DIVORCIADO",
	"Viuvo(a)":      "VIUVO",
	"União Estável": "UNIAO_ESTAVEL",
	"Desquitado(a)": "DESQUITADO",
	"":              "NAO_INFORMADO",
}

var requiredColumns = map[string]string{
	"ALUNO":        "Name",
	"E-MAIL":       "Email",
	"SEXO":         "Sexo",
	"CPF":          "Cpf",
	"ESTADO CIVIL": "EstadoCivil",
	"RG":           "RG",
	"EMISSOR":      "RGOrgaoExpedidor",
	"BAIRRO":       "Bairro",
	"LOGRADOURO":   "Logradouro",
	"NUMERO":       "Numero",
	"ESTADO":       "Estado",
	"CIDADE":       "Cidade",
	"NASCIMENTO":   "DateOfBirth",
}

func HandleFileRow(c *gin.Context) ([]*m.Profile, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get uploaded file: %v", err)
	}

	tempPath := fmt.Sprintf("/tmp/%s", file.Filename)

	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %v", err)
	}

	f, err := e.OpenFile(tempPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("invalid file format")
	}

	var profiles []*m.Profile
	var errProfile []m.ErrProfile
	for i, row := range rows[1:] {
		p := &m.Profile{}
		err := getProfileContent(row, rows[0], p)
		if err != nil {

			errProfile = append(errProfile, interfaces.ErrProfile{
				Line: i + 2, // Excel line start at
				Cpf:  p.Cpf,
				Err:  err.Error(),
			})
			continue
		}

		profiles = append(profiles, p)
	}

	return profiles, nil
}

func getProfileContent(row []string, header []string, p *m.Profile) error {
	columnIndex := make(map[string]int)

	for i, colName := range header {
		if mappedName, exists := requiredColumns[colName]; exists {
			columnIndex[mappedName] = i
		}
	}

	for _, alias := range requiredColumns {
		if _, exists := columnIndex[alias]; !exists {
			return fmt.Errorf("missing required column: %s", alias)
		}
	}

	p.Name = strings.ToUpper(row[columnIndex["Name"]])
	p.Email = row[columnIndex["Email"]]
	p.Sexo = sexoMap[row[columnIndex["Sexo"]]]
	p.Cpf = strings.ToUpper(row[columnIndex["Cpf"]])
	p.EstadoCivil = maritalStatus[row[columnIndex["EstadoCivil"]]]
	p.RGOrgaoExpedidor = strings.ToUpper(row[columnIndex["RGOrgaoExpedidor"]])
	p.Bairro = strings.ToUpper(row[columnIndex["Bairro"]])
	p.Logradouro = strings.ToUpper(row[columnIndex["Logradouro"]])
	p.Numero = strings.ToUpper(row[columnIndex["Numero"]])
	p.Estado = strings.ToUpper(row[columnIndex["Estado"]])
	p.Cidade = strings.ToUpper(row[columnIndex["Cidade"]])

	rgTrim := strings.TrimSpace(strings.ToUpper(row[columnIndex["RG"]]))
	re := regexp.MustCompile(`[^A-Z0-9]`)
	p.RG = re.ReplaceAllString(rgTrim, "")

	d, err := formatDate(row[columnIndex["DateOfBirth"]])
	if err != nil {
		return err
	}

	p.DateOfBirth = d

	return nil
}

func formatDate(dateStr string) (string, error) {
	if dateStr == "" {
		return "", nil
	}

	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return "", err
	}

	return t.Format("2006-01-02"), nil
}
