package interfaces

type Profile struct {
	Name             string `json:"nome"`
	Email            string `json:"email"`
	Sexo             string `json:"sexo"`
	DateOfBirth      string `json:"dataNascimento"`
	ClientID         int    `json:"idCliente,omitempty"`
	OrgID            int    `json:"idOrg"`
	Estado           string `json:"estado,omitempty"`
	Cidade           string `json:"cidade"`
	ProfileType      string `json:"tipoPerfil"`
	Cpf              string `json:"cpf"`
	EstadoCivil      string `json:"estadoCivil"`
	RG               string `json:"rg"`
	RGOrgaoExpedidor string `json:"rgOrgaoExpedidor"`
	IdCidadeEndereco int    `json:"idCidadeEndereco"`
	Bairro           string `json:"bairro"`
	Logradouro       string `json:"logradouro"`
	Numero           string `json:"numero"`
}
