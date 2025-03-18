package profile

import (
	"fmt"
	"regexp"
	"strings"
	"time"

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

func GetProfileContent(c *gin.Context, p *m.Profile) error {
	file, err := c.FormFile("file")
	if err != nil {
		return fmt.Errorf("failed to get uploaded file: %v", err)
	}

	tempPath := fmt.Sprintf("/tmp/%s", file.Filename)

	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		return fmt.Errorf("failed to save uploaded file: %v", err)
	}

	f, err := e.OpenFile(tempPath)
	if err != nil {
		return err
	}
	defer f.Close()

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}

	if len(rows) < 2 {
		return fmt.Errorf("invalid file format")
	}

	header := rows[0]
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

	dataRow := rows[1]
	p.Name = strings.ToUpper(dataRow[columnIndex["Name"]])
	p.Email = dataRow[columnIndex["Email"]]
	p.Sexo = sexoMap[dataRow[columnIndex["Sexo"]]]
	p.Cpf = strings.ToUpper(dataRow[columnIndex["Cpf"]])
	p.EstadoCivil = maritalStatus[dataRow[columnIndex["EstadoCivil"]]]
	p.RGOrgaoExpedidor = strings.ToUpper(dataRow[columnIndex["RGOrgaoExpedidor"]])
	p.Bairro = strings.ToUpper(dataRow[columnIndex["Bairro"]])
	p.Logradouro = strings.ToUpper(dataRow[columnIndex["Logradouro"]])
	p.Numero = strings.ToUpper(dataRow[columnIndex["Numero"]])
	p.Estado = strings.ToUpper(dataRow[columnIndex["Estado"]])
	p.Cidade = strings.ToUpper(dataRow[columnIndex["Cidade"]])

	rgTrim := strings.TrimSpace(strings.ToUpper(dataRow[columnIndex["RG"]]))
	re := regexp.MustCompile(`[^A-Z0-9]`)
	p.RG = re.ReplaceAllString(rgTrim, "")

	d, err := formatDate(dataRow[columnIndex["DateOfBirth"]])
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
