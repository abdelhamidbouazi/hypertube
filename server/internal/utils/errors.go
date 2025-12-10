package utils

type HTTPError struct {
	Message string `json:"message"`
}

type HTTPErrorUnauthorized struct {
	Message string `json:"message" example:"unauthorized"`
}
