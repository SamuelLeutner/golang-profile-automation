package interfaces

type ErrProfile struct {
	Line int    `json:"line"`
	Cpf  string `json:"cpf"`
	Err  string  `json:"err,omitempty"`
}
