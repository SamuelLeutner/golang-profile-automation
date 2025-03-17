package interfaces

import (
	"encoding/json"
	"strings"
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

func (p *Profile) UnmarshalJSON(b []byte) error {
	type Alias Profile
	aux := &struct {
		DateOfBirth string `json:"date_of_birth"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	p.Name = strings.ToUpper(p.Name)
	p.Sexo = strings.ToUpper(p.Sexo)
	p.PerfilType = strings.ToUpper(p.PerfilType)
	p.Cpf = strings.ToUpper(p.Cpf)
	p.EstadoCivil = strings.ToUpper(p.EstadoCivil)
	p.RG = strings.ToUpper(p.RG)
	p.RGOrgaoExpedidor = strings.ToUpper(p.RGOrgaoExpedidor)
	p.Bairro = strings.ToUpper(p.Bairro)
	p.Logradouro = strings.ToUpper(p.Logradouro)
	p.Numero = strings.ToUpper(p.Numero)

	if aux.DateOfBirth != "" {
		t, err := time.Parse("02/01/2006", aux.DateOfBirth)
		if err == nil {
			p.DateOfBirth.Time = t
		} else {
			return err
		}
	}
	return nil
}

func (ct CustomTime) String() string {
	return ct.Time.Format("2006-01-02")
}
