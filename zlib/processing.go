package zlib

import (
	"encoding/json"

	"github.com/zmap/zgrab/ztools/processing"
)

// GrabWorker implements ztools.processing.Worker
type GrabWorker struct {
	success uint
	failure uint

	statuses chan status

	config *Config
}

type status uint

const (
	status_success status = iota
	status_failure status = iota
)

func (g *GrabWorker) Success() uint {
	return g.success
}

func (g *GrabWorker) Failure() uint {
	return g.failure
}

func (g *GrabWorker) Total() uint {
	return g.success + g.failure
}

func (g *GrabWorker) RunCount() uint {
	return g.config.ConnectionsPerHost
}

func (g *GrabWorker) Done() {
	close(g.statuses)
}

func (g *GrabWorker) MakeHandler(id uint) processing.Handler {
	return func(v interface{}) interface{} {
		target, ok := v.(GrabTarget)
		if !ok {
			return nil
		}
		grab := GrabBanner(g.config, &target)
		s := grab.status()
		g.statuses <- s
		return grab
	}
}

func NewGrabWorker(config *Config) processing.Worker {
	w := new(GrabWorker)
	w.statuses = make(chan status, config.Senders*4)
	w.config = config
	go func() {
		for s := range w.statuses {
			switch s {
			case status_success:
				w.success++
			case status_failure:
				w.failure++
			default:
				continue
			}
		}
	}()
	return w
}

type grabMarshaler struct{}

func (gm *grabMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func NewGrabMarshaler() processing.Marshaler {
	return new(grabMarshaler)
}
