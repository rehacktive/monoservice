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

type Service struct {
	address string
	router  *mux.Router
}

func main() {
	var modulesFolder, address string

	flag.StringVar(&modulesFolder, envModule, "modules/", "modules folder")
	flag.StringVar(&address, envAddress, ":8880", "address:port to listen from")

	flag.Parse()

	srv := Service{
		address: address,
		router:  mux.NewRouter(),
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
			switch m.Action {
			case monoservice.NEW:
				{
					log.Println("add plugin ", m.Name)
					srv.addHandler(*m.Handler)
				}
			case monoservice.UPDATE:
				{
					log.Println("replace plugin ", m.Name)
					srv.reloadHandler(*m.Handler)

				}
			case monoservice.REMOVE:
				{
					log.Println("remove plugin ", m.Name)
				}
			}

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

func (srv *Service) addHandler(handler monoservice.HandlerInterface) {
	handler.Init()
	srv.router.HandleFunc(handler.Path(), func(w http.ResponseWriter, r *http.Request) {
		response := handler.Process(r)
		monoservice.RespondWithJSON(w, response.Code, response.JSONContent)
	}).Methods(handler.Methods()...)
}

func (srv *Service) reloadHandler(handler monoservice.HandlerInterface) {
	srv.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		fmt.Println("1")
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}

		if t == handler.Path() {
			fmt.Println("replacing handler for path ", t)
			handler.Init()
			route.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := handler.Process(r)
				monoservice.RespondWithJSON(w, response.Code, response.JSONContent)
			}).Methods(handler.Methods()...)
		}
		return nil
	})
}
