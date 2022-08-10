### MONOSERVICE - modular microservice using Golang plugins

The idea is to have a very simple service that uses Golang plugins compiled as modules (.so) to implement different HTTP handlers.
Each module can be worked on at different times, compiled and deployed, the service will detect a new module and load/use it.

#### Step 1

Each plugin needs to implement the following interface, defined in service.go:

```go
type HandlerInterface interface {
	Init()                                                 // this method will initialize the module
	Path() string                                          // the path handled
	Process(r *http.Request) utils.JSONResponse 		   // the logic for the handler
	Methods() []string                                     // HTTP methods used
}
```

so for example:

```go
type handlerPlugin struct{}
var HandlerPlugin handlerPlugin

func (p handlerPlugin) Init() {
	log.Println("hello plugin initialized")
}

func (p handlerPlugin) Path() string {
	return "/hello"
}

func (p handlerPlugin) Handler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			respondWithJSON(w, http.StatusOK, "hello from the plugin")
			return
		}
	}
}

func (p handlerPlugin) Methods() []string {
	return []string{http.MethodGet}
}
```

note that the HandlerPlugin is expected, so do not change this part, only the methods implementation.
Init() could be used to do specific task at start (connect to a database, for example). The rest is self explainatory.

#### Step 2

Build the plugin:

```sh
go build -buildmode=plugin -o ./modules/helloplugin.so ./plugin/hello_plugin.go
```

This will generate an ```helloplugin.so``` file inside the modules folder, that is an ELF shared object.

### Step 3

Use the module with the service. For example running it with:no 

```sh
go run service.go -MODULE_FOLDER=modules/
```

(that's by default the modules folder name)

This will scan for modules inside the folder and add/expose them, like defined in the plugin code.

A simple test will reveal that all is working:

```sh
curl http://localhost:8880/hello 

hello from the plugin
```


### Using Docker

The Dockerfile defined here will automagically compile the module defined inside /plugin folder and run the service:

```sh
docker build . -t modular1
docker run --publish 8880:8880 modular1
```
