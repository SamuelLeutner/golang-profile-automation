package profile

import (
	"fmt"

	m "github.com/SamuelLeutner/golang-profile-automation/internal/models"
	e "github.com/xuri/excelize/v2"
)

func SaveErrors(errors []*m.ErrProfile) error {
	if len(errors) == 0 {
		return nil
	}

	f := e.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)
	f.SetCellValue(sheet, "A1", "Linha")
	f.SetCellValue(sheet, "B1", "CPF")
	f.SetCellValue(sheet, "C1", "Erro")

	for i, err := range errors {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), err.Line)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), err.Cpf)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+2), err.Err)
	}

	if err := f.SaveAs("ProfileErrors.xlsx"); err != nil {
		return err
	}

	return nil
}
