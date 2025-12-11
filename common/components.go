package common

import (
	"fmt"
	"reflect"
	"sync"

	"slices"

	"github.com/rs/zerolog/log"
)

type Components struct {
	lock       sync.Mutex
	components []interface{}
}

func NewComponents() *Components {
	return &Components{
		components: make([]interface{}, 0),
	}
}

func (c *Components) AllComponents() []interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.components
}

func (in *Components) AddComponent(feature interface{}) {
	in.lock.Lock()
	defer in.lock.Unlock()
	in.components = append(in.components, feature)
}

func (in *Components) GetComponent(t reflect.Type) interface{} {
	in.lock.Lock()
	defer in.lock.Unlock()
	for _, component := range in.components {
		if reflect.TypeOf(component).AssignableTo(t) {
			return component
		}
	}
	return nil
}

func (in *Components) RemoveComponent(feature interface{}) {
	in.lock.Lock()
	defer in.lock.Unlock()
	for i, component := range in.components {
		if component == feature {
			in.components = slices.Delete(in.components, i, i+1)
			return
		}
	}
}

func (s *Components) Start() error {
	err := StartAll(s.components...)
	if err != nil {
		return fmt.Errorf("failed to start components: %w", err)
	}
	log.Debug().Msgf("started")
	return nil
}

func (s *Components) Close() error {
	log.Debug().Msgf("closing")
	err := CloseAll(s.components...)
	if err != nil {
		return fmt.Errorf("failed to close components: %w", err)
	}
	log.Debug().Msg("closed")
	return nil
}
