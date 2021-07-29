package server

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/protocol"
)

// Protocol ID
const (
	pid protocol.ID = "/reqresp/0.0.1"
)

// Options is a structure containing all the options that can be used when constructing the smart records env
type serverConfig struct {
	anOption interface{}
}

// Option type for smart records
type ServerOption func(*serverConfig) error

// defaults are the default smart record env options. This option will be automatically
// prepended to any options you pass to the constructor.
var serverDefaults = func(o *serverConfig) error {
	o.anOption = []int{1}

	return nil
}

// apply applies the given options to this Option
func (c *serverConfig) apply(opts ...ServerOption) error {
	for i, opt := range opts {
		if err := opt(c); err != nil {
			return fmt.Errorf("smart record server option %d failed: %s", i, err)
		}
	}
	return nil
}

// AnOption
func AnOption(o interface{}) ServerOption {
	return func(c *serverConfig) error {
		c.anOption = o
		return nil
	}
}
