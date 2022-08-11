package monoservice

import (
	"errors"
	"net/http"
	"path/filepath"
	"plugin"
)

type HandlerInterface interface {
	Init()                                // this method will initialize the module
	Path() string                         // the path handled
	Process(r *http.Request) JSONResponse // the logic for the handler
	Methods() []string                    // HTTP methods used
}

func LoadPlugin(moduleFolder string, moduleName string) (HandlerInterface, error) {
	plug, err := plugin.Open(filepath.Join(moduleFolder, moduleName))
	if err != nil {
		return nil, err
	}

	symHandler, err := plug.Lookup("HandlerPlugin")
	if err != nil {
		return nil, err
	}

	var handler HandlerInterface
	handler, ok := symHandler.(HandlerInterface)
	if !ok {
		return nil, errors.New("unexpected type from module symbol")
	}

	return handler, nil
}
