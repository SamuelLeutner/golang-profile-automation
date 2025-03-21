package profile

import (
	"fmt"
	"strconv"
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
		if len(row) < len(requiredColumns) {
			errProfile = append(errProfile, interfaces.ErrProfile{
				Line: i + 2,
				Cpf:  "Unknown",
				Err:  fmt.Sprintf("row %d has insufficient columns (expected %d)", i+2, len(requiredColumns))})
			continue
		}

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

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no valid profiles found in the uploaded file")
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

	p.Name = strings.TrimSpace(strings.ToUpper(row[columnIndex["Name"]]))
	p.Email = row[columnIndex["Email"]]
	p.Sexo = sexoMap[row[columnIndex["Sexo"]]]
	p.Cpf = strings.TrimSpace(strings.ToUpper(row[columnIndex["Cpf"]]))
	p.EstadoCivil = maritalStatus[row[columnIndex["EstadoCivil"]]]
	p.RGOrgaoExpedidor = strings.TrimSpace(strings.ToUpper(row[columnIndex["RGOrgaoExpedidor"]]))
	p.Bairro = strings.TrimSpace(strings.ToUpper(row[columnIndex["Bairro"]]))
	p.Logradouro = strings.TrimSpace(strings.ToUpper(row[columnIndex["Logradouro"]]))
	p.Estado = strings.TrimSpace(strings.ToUpper(row[columnIndex["Estado"]]))
	p.Cidade = strings.TrimSpace(strings.ToUpper(row[columnIndex["Cidade"]]))

	houseNumber := strings.TrimSpace(strings.ToUpper(row[columnIndex["Numero"]]))
	if houseNumber == "" {
		p.Numero = "0"
	} else {
		_, err := strconv.Atoi(houseNumber)
		if err != nil {
			p.Numero = "0"
		} else {
			p.Numero = houseNumber
		}
	}

	rgTrim := strings.TrimSpace(strings.ToUpper(row[columnIndex["RG"]]))
	if rgTrim == "" {
		return fmt.Errorf("Missing RG value.")
	}

	rgFormatted := strings.Replace(rgTrim, ".", "", -1)
	rgFormatted = strings.Replace(rgFormatted, "-", "", -1)
	if len(rgFormatted) <= 5 {
		return fmt.Errorf("Formatting the RG is wrong.")
	}

	p.RG = rgFormatted

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
