package configMerger

import (
	"io/ioutil"
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"fmt"
)

type JsonSource struct {
	Path         string
	TargetStruct interface{}
	WatchHandler func()
}

func (j *JsonSource) Load() error {

	file, err := ioutil.ReadFile(j.Path)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(file), j.TargetStruct)
	if err != nil {
		return err
	}

	return nil
}

func (j *JsonSource) SetTargetStruct(i interface{}) {
	j.TargetStruct = i
}

func (j *JsonSource) Watch() {

	if j.WatchHandler != nil {
		// creates a new file watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Println("ERROR", err)
		}
		defer watcher.Close()

		//
		done := make(chan bool)

		//TODO: make it work

		//
		go func() {
			for {
				select {
				// watch for events
				case event := <-watcher.Events:
					fmt.Printf("EVENT! %#v\n", event)

					// watch for errors
				case err := <-watcher.Errors:
					fmt.Println("ERROR", err)
				}
			}
		}()

		// out of the box fsnotify can watch a single file, or a single directory
		if err := watcher.Add(j.Path); err != nil {
			fmt.Println("ERROR", err)
		}

		<-done
	}

}