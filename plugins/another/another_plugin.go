package main

import (
	"log"
	"net/http"

	monoservice "github.com/rehacktive/monoservice/monoservice"
)

type handlerPlugin struct{}

var HandlerPlugin handlerPlugin

func (p handlerPlugin) Init() {
	log.Println("another plugin initialized 3")
}

func (p handlerPlugin) Path() string {
	return "/another"
}

func (p handlerPlugin) Process(r *http.Request) monoservice.JSONResponse {
	return monoservice.JSONResponse{
		JSONContent: `{"message":"hello from the another plugin 3"}`,
		Code:        http.StatusOK,
	}
}

func (p handlerPlugin) Methods() []string {
	return []string{http.MethodGet}
}
