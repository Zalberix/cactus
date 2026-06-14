package watcher

import (
	"strings"
	"time"

	"github.com/pterm/pterm"

	"github.com/rjeczalik/notify"
	"github.com/samber/lo"
)

type Watcher struct {
	chann chan notify.EventInfo
}

func New() *Watcher {
	return &Watcher{
		chann: make(chan notify.EventInfo, 1),
	}
}

func (c *Watcher) Start(path string) (chan struct{}, error) {
	updateChannel := make(chan struct{}, 1)

	events, err := c.StartEvents(path)
	if err != nil {
		return nil, err
	}

	reload, _ := lo.NewDebounce(
		1*time.Second,
		func() {
			updateChannel <- struct{}{}
		},
	)

	go func() {
		for path := range events {
			pterm.Warning.Printfln("File changed, triggering restart: %s", path)
			reload()
		}
	}()

	return updateChannel, nil
}

func (c *Watcher) StartEvents(path string) (chan string, error) {
	watchPath := path + "..."
	updateChannel := make(chan string, 1)

	if err := notify.Watch(watchPath, c.chann, notify.All); err != nil {
		return nil, err
	}

	go func() {
		for event := range c.chann {
			if strings.HasSuffix(event.Path(), "~") || strings.Contains(event.Path(), ".out") {
				continue
			}

			select {
			case updateChannel <- event.Path():
			default:
			}
		}
	}()

	return updateChannel, nil
}

func (c *Watcher) Stop() {
	notify.Stop(c.chann)
}
