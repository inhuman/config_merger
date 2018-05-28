package configMerger

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"github.com/radovskyb/watcher"
	"log"
	"time"
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
		w := watcher.New()
		w.SetMaxEvents(1)

		w.FilterOps(watcher.Write)

		go func() {
			for {
				select {
				case event := <-w.Event:
					fmt.Println(event)
				case err := <-w.Error:
					log.Fatalln(err)
				case <-w.Closed:
					return
				}
			}
		}()

		if err := w.Add(j.Path); err != nil {
			fmt.Println(err)
		}

		go w.Wait()


		// Start the watching process - it'll check for changes every 100ms.
		if err := w.Start(time.Second); err != nil {
			fmt.Println(err)
		}
	}

}