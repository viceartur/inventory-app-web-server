package handlers

type SuccessResponseJSON struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponseJSON struct {
	Message string `json:"message"`
}
