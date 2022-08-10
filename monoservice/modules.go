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

type ActionType int64

const (
	NEW ActionType = iota
	UPDATE
	REMOVE
	NONE
)

type Module struct {
	Name    string
	Action  ActionType
	Handler *HandlerInterface
}

func NewModulesManager(folder string, chanEvents chan Module) ModulesManager {
	return ModulesManager{
		modulesFolder: folder,
		modules:       make(map[string]*HandlerInterface, 0),
		chanEvents:    chanEvents,
	}
}

// scan a specific directory
// return a modules when  detected
func (m ModulesManager) WatchFolder() {
	w := watcher.New()

	// Only files that match the regular expression during file listings
	// will be watched.
	r := regexp.MustCompile("[a-z0-9]+.so")
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		for {
			select {
			case event := <-w.Event:
				if !event.IsDir() {
					m.processEvent(event)
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
	for path, f := range w.WatchedFiles() {
		fmt.Printf("%s: %s\n", path, f.Name())
	}

	fmt.Println()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}

func (m *ModulesManager) processEvent(event watcher.Event) {
	var err error

	module := Module{
		Name: event.Name(),
	}

	switch event.Op {
	case watcher.Create:
		{
			module.Action = NEW
			module.Handler, err = LoadPlugin(m.modulesFolder, event.Name())
			if err != nil {
				log.Println("error on loading plugin: ", err)
				return
			}

		}
	case watcher.Write:
		{
			module.Action = UPDATE
			module.Handler, err = LoadPlugin(m.modulesFolder, event.Name())
			if err != nil {
				log.Println("error on loading plugin: ", err)
				return
			}
		}
	case watcher.Remove:
		{
			module.Action = REMOVE
			module.Handler = m.modules[module.Name]
		}
	default:
		module.Action = NONE
	}

	if module.Action != NONE {
		m.chanEvents <- module
	}
}
