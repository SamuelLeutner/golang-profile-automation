package interfaces

type City struct {
	IdCidade  int    `json:"idCidade"`
	Descricao string `json:"descricao"`
	Uf        string `json:"uf"`
	Estado    string `json:"estado"`
}

type CityIdResponse struct {
	Elements []City `json:"elements"`
}
