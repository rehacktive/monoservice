package monoservice

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/radovskyb/watcher"
)

type ModulesManager struct {
	modulesFolder string
	modules       map[string]*HandlerInterface
	chanEvents    chan Module
}

type Module struct {
	Name    string
	Handler *HandlerInterface
}

func NewModulesManager(folder string, chanEvents chan Module) ModulesManager {
	return ModulesManager{
		modulesFolder: folder,
		chanEvents:    chanEvents,
		modules:       map[string]*HandlerInterface{},
	}
}

// scan a specific directory
// return a modules when  detected
func (m ModulesManager) WatchFolder() {
	w := watcher.New()

	// Only notify rename and move events.
	w.FilterOps(watcher.Create)
	// Only files that match the regular expression during file listings
	// will be watched.
	r := regexp.MustCompile("[a-z0-9]+.so")
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		for {
			select {
			case event := <-w.Event:
				if !event.IsDir() {
					m.setupModule(event.Name())
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch this folder for changes.
	if err := w.Add(m.modulesFolder); err != nil {
		log.Fatalln(err)
	}

	// Print a list of all of the files and folders currently
	// being watched and their paths.
	for _, f := range w.WatchedFiles() {
		if !f.IsDir() {
			m.setupModule(f.Name())
		}
	}

	fmt.Println()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}

func (m ModulesManager) setupModule(name string) {
	if _, ok := m.modules[name]; ok {
		// module was already loaded, skip it
		return
	}

	module := Module{
		Name: name,
	}

	handler, err := LoadPlugin(m.modulesFolder, module.Name)
	if err != nil {
		fmt.Println("error add a plugin ", err)
		return
	}
	m.modules[module.Name] = &handler
	module.Handler = &handler
	m.chanEvents <- module
}
