package profile

import (
	"fmt"

	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
	"github.com/gin-gonic/gin"
	e "github.com/xuri/excelize/v2"
)

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
	"NASCIMENTO":   "DateOfBirth",
}

func GetProfileContent(c *gin.Context, p *m.Profile) error {
	file, err := c.FormFile("file")
	if err != nil {
		return fmt.Errorf("failed to get uploaded file: %v", err)
	}

	f, err := e.OpenFile(file.Filename)
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
		for _, required := range requiredColumns {
			if colName == required {
				columnIndex[required] = i
			}
		}
	}

	for _, col := range requiredColumns {
		if _, exists := columnIndex[col]; !exists {
			return fmt.Errorf("missing required column: %s", col)
		}
	}

	dataRow := rows[1]

	// TODO: Add another struct type for send to jacad
	p.Name = dataRow[columnIndex["Name"]]
	p.Email = dataRow[columnIndex["Email"]]
	p.Sexo = dataRow[columnIndex["Sexo"]]
	p.Cpf = dataRow[columnIndex["Cpf"]]
	p.EstadoCivil = dataRow[columnIndex["EstadoCivil"]]
	p.RG = dataRow[columnIndex["RG"]]
	p.RGOrgaoExpedidor = dataRow[columnIndex["RGOrgaoExpedidor"]]
	p.Bairro = dataRow[columnIndex["Bairro"]]
	p.Logradouro = dataRow[columnIndex["Logradouro"]]
	p.Numero = dataRow[columnIndex["Numero"]]
	p.DateOfBirth = dataRow[columnIndex["DateOfBirth"]]

	fmt.Printf("Profile parsed: %+v\n", p)

	return nil
}
