package config_merger

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"github.com/radovskyb/watcher"
	"time"
	"sync"
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

func (j *JsonSource) Watch(done chan bool, group *sync.WaitGroup) {

	if j.WatchHandler != nil {
		w := watcher.New()
		w.SetMaxEvents(1)

		w.FilterOps(watcher.Write)

		handler := func() {
			group.Add(1)
			j.WatchHandler()
			group.Done()
		}

		go func() {
			for {
				select {
				case <-w.Event:
					err := j.Load()
					if err == nil {
						handler()
					} else {
						fmt.Println(err)
					}
				case err := <-w.Error:
					fmt.Println(err)
				case <-done:
					w.Close()
					return
				}
			}
		}()

		if err := w.Add(j.Path); err != nil {
			fmt.Println(err)
		}

		if err := w.Start(time.Second); err != nil {
			fmt.Println(err)
		}
	}
}