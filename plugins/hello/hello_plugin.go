package main

import (
	"log"
	"net/http"

	monoservice "github.com/rehacktive/monoservice/monoservice"
)

type handlerPlugin struct{}

var HandlerPlugin handlerPlugin

func (p handlerPlugin) Init() {
	log.Println("hello plugin initialized")
}

func (p handlerPlugin) Path() string {
	return "/hello"
}

func (p handlerPlugin) Process(r *http.Request) monoservice.JSONResponse {
	return monoservice.JSONResponse{
		JSONContent: `{"message":"hello from the plugin"}`,
		Code:        http.StatusOK,
	}
}

func (p handlerPlugin) Methods() []string {
	return []string{http.MethodGet}
}
