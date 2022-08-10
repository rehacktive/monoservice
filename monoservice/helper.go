package monoservice

import (
	"log"
	"net/http"
)

type JSONResponse struct {
	JSONContent string
	Code        int
}

func RespondWithJSON(w http.ResponseWriter, code int, jsonContent string) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(code)
	_, err := w.Write([]byte(jsonContent))
	if err != nil {
		log.Println("error sending response: ", err)
	}
}
