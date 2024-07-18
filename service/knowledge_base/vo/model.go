package vo

type ModelList struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    []ModelListData `json:"data"`
}

type ModelListData struct {
	ID        string      `json:"id"`
	Provider  string      `json:"provider"`
	Name      string      `json:"name"`
	ModelType string      `json:"model_type"`
	ModelName string      `json:"model_name"`
	Status    string      `json:"status"`
	Meta      interface{} `json:"meta"`
}
