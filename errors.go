package server

import (
	"net/http"
	"encoding/json"
)

type ErrorMessage struct{
	ErrorCode string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func (s CLinkServer) BadRequestError(w http.ResponseWriter, error_code string, error_message string){
	w.WriteHeader(400)

	response_body, err := json.Marshal(ErrorMessage{error_code, error_message})

	if err != nil {
		panic("Error writing error response:" + err.Error())
	}

	w.Write(response_body)
}

func (s CLinkServer) InternalError(w http.ResponseWriter){
	w.WriteHeader(500)

	body := ErrorMessage{"internal", "Interanl server error"}

	response_body, err := json.Marshal(body)

	if err != nil {
		panic("Error writing error response:" + err.Error())
	}

	w.Write(response_body)
}