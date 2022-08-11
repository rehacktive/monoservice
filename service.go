package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"github.com/rehacktive/monoservice/monoservice"
)

const (
	envModule  = "MODULE_FOLDER"
	envAddress = "ADDRESS"
)

var modulesFolder string

type Service struct {
	address          string
	router           *mux.Router
	registeredRoutes []string
}

func main() {
	var address string

	flag.StringVar(&modulesFolder, envModule, "modules/", "modules folder")
	flag.StringVar(&address, envAddress, ":8880", "address:port to listen from")

	flag.Parse()

	routes := make([]string, 0)
	srv := Service{
		address:          address,
		router:           mux.NewRouter(),
		registeredRoutes: routes,
	}

	server := &http.Server{
		Addr:              srv.address,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		Handler:           srv.router,
	}

	moduleEvents := make(chan monoservice.Module, 1)
	modulesManager := monoservice.NewModulesManager(modulesFolder, moduleEvents)

	// wait for modules changes
	go modulesManager.WatchFolder()
	go func() {
		for m := range moduleEvents {
			srv.useModule(m)
		}
	}()

	log.Println("service up and running")
	go server.ListenAndServe()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("unable to stop gracefully server: %v", err)
	}

	log.Println("best regards.")
}

func (srv *Service) useModule(m monoservice.Module) {
	defer func() {
		if r := recover(); r != nil {
			// the plugin method may panic
			fmt.Println("Recovered. Error:\n", r)
		}
	}()

	log.Println("add plugin ", m.Name)
	log.Println(m)
	handler, err := monoservice.LoadPlugin(modulesFolder, m.Name)
	if err != nil {
		fmt.Println("[skipping] error on LoadPlugin ", err)
		return
	}
	// check if already registered, reload
	if contains(srv.registeredRoutes, handler.Path()) {
		srv.reloadHandler(handler)
	} else {
		// else add it
		srv.addHandler(handler)
	}
}

func (srv *Service) addHandler(handler monoservice.HandlerInterface) {
	srv.registeredRoutes = append(srv.registeredRoutes, handler.Path())

	handler.Init()
	srv.router.HandleFunc(handler.Path(), func(w http.ResponseWriter, r *http.Request) {
		response := handler.Process(r)
		monoservice.RespondWithJSON(w, response.Code, response.JSONContent)
	}).Methods(handler.Methods()...)
}

func (srv *Service) reloadHandler(handler monoservice.HandlerInterface) {
	srv.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}

		if t == handler.Path() {
			handler.Init()
			route.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := handler.Process(r)
				monoservice.RespondWithJSON(w, response.Code, response.JSONContent)
			}).Methods(handler.Methods()...)
		}
		return nil
	})
}

// utils

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
