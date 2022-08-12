package monoservice

import (
	"io"
	"log"
	"net/http"
)

type HTTPRequest struct {
	Method      string
	Host        string
	Body        []byte
	Header      http.Header
	QueryString map[string][]string
}

type JSONResponse struct {
	JSONContent string
	Code        int
}

func MapToHTTPRequest(r *http.Request) (*HTTPRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return &HTTPRequest{
		Method:      r.Method,
		Host:        r.Host,
		Body:        body,
		Header:      r.Header,
		QueryString: r.URL.Query(),
	}, nil
}

func RespondWithJSON(w http.ResponseWriter, code int, jsonContent string) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(code)
	_, err := w.Write([]byte(jsonContent))
	if err != nil {
		log.Println("error sending response: ", err)
	}
}
