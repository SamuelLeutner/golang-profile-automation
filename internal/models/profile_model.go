package interfaces

import (
	"encoding/json"
	"time"
)

type Profile struct {
	Name             string     `json:"name"`
	Email            string     `json:"email"`
	Sexo             string     `json:"sexo"`
	DateOfBirth      CustomTime `json:"date_of_birth"`
	ClientID         int        `json:"client_id,omitempty"`
	OrgID            int        `json:"org_id,omitempty"`
	PerfilType       string     `json:"perfil_type"`
	Cpf              string     `json:"cpf"`
	EstadoCivil      string     `json:"estado_civil"`
	RG               string     `json:"rg"`
	RGOrgaoExpedidor string     `json:"rg_orgao_expedidor"`
	IdCidadeEndereço int        `json:"id_cidade_endereco"`
	Bairro           string     `json:"bairro"`
	Logradouro       string     `json:"logradouro"`
	Numero           string     `json:"numero"`
}

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	t, err := time.Parse("02/01/2006", s)
	if err != nil {
		return err
	}

	ct.Time = t
	return nil
}
