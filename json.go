package config_merger

import (
	"encoding/json"
	"fmt"
	"github.com/radovskyb/watcher"
	"io/ioutil"
	"sync"
	"time"
)

type JsonSource struct {
	SourceModel
	Path string
}

//TODO: implement tag ids in json source

func (s *JsonSource) Load() error {

	file, err := ioutil.ReadFile(s.Path)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(file), s.TargetStruct)
	if err != nil {
		return err
	}

	return nil
}

func (s *JsonSource) SetTargetStruct(i interface{}) {
	s.TargetStruct = i
}

func (s *JsonSource) Watch(done chan bool, group *sync.WaitGroup) {

	if s.WatchHandler != nil {
		w := watcher.New()
		w.SetMaxEvents(1)

		w.FilterOps(watcher.Write)

		handler := func() {
			group.Add(1)
			s.WatchHandler()
			group.Done()
		}

		go func() {
			for {
				select {
				case <-w.Event:
					err := s.Load()
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

		if err := w.Add(s.Path); err != nil {
			fmt.Println(err)
		}

		if err := w.Start(time.Second); err != nil {
			fmt.Println(err)
		}
	}
}

func (s *JsonSource) GetTagIds() map[string]string {
	return s.TagIds
}
