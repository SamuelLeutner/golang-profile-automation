package interfaces

type Profile struct {
	Name             string `json:"nome"`
	Email            string `json:"email"`
	Sexo             string `json:"sexo"`
	DateOfBirth      string `json:"dataNascimento"`
	ClientID         int    `json:"idCliente"`
	OrgID            int    `json:"idOrg"`
	Estado           string `json:"estado"`
	Cidade           string `json:"cidade"`
	ProfileType      string `json:"tipoPerfil"`
	Cpf              string `json:"cpf"`
	EstadoCivil      string `json:"estadoCivil"`
	RG               string `json:"rg"`
	RGOrgaoExpedidor string `json:"rgOrgaoExpedidor"`
	IdCidadeEndereco int    `json:"idCidadeEndereco"`
	IdNacionalidade  int    `json:"idNacionalidade"`
	Bairro           string `json:"bairro"`
	Logradouro       string `json:"logradouro"`
	Numero           string `json:"numero"`
}
