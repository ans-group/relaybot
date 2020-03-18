package relay

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
)

// InitFunc represents a function which when called returns a slice of servers
type InitFunc func() []Server

var initFuncs []InitFunc

// Debug specifies debug logging should be enabled
var Debug bool

// RegisterServer registers provided initFunc with relay
func RegisterServer(initFunc InitFunc) {
	initFuncs = append(initFuncs, initFunc)
}

func getServerConfigs(t string, v interface{}) {
	viper.UnmarshalKey(fmt.Sprintf("servers.%s", t), v)
}

type Server interface {
	Connect() error
	SetTargets(targets []Target) error
	Read(context.Context, chan *TargetMessage) error
	Name() string
	Write(*TargetMessage) error
}

// ServerBase provides common methods for use with implementations of Server
type ServerBase struct {
	name    string
	Targets []Target
}

// NewServerBase returns a pointer to an initialised ServerBase
func NewServerBase(name string) *ServerBase {
	return &ServerBase{
		name: name,
	}
}

// Name returns the name of the server
func (b *ServerBase) Name() string {
	n := b.name
	return n
}

// GetTarget returns a target matching given name or an error if target isn't found
func (b *ServerBase) GetTarget(name string) (Target, error) {
	for _, target := range b.Targets {
		if target.Name == name {
			return target, nil
		}
	}

	return Target{}, fmt.Errorf("Cannot find target [%s]", name)
}
