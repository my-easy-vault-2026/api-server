package entities

// SystemParameterVO represents the system parameter data for response.
type ListSystemParametersVO struct {
	Records []*SystemParameterVO `json:"records"`
}

// SystemParameter represents a system parameter.
type SystemParameterVO struct {
	ID          int64  `json:"id,string"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
