package client

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/protocol"
)

// Protocol ID
const (
	pid protocol.ID = "/reqresp/0.0.1"
)

type clientConfig struct {
	anOption interface{}
}

// Option type for smart records
type ClientOption func(*clientConfig) error

// apply applies the given options to this Option
func (c *clientConfig) apply(opts ...ClientOption) error {
	for i, opt := range opts {
		if err := opt(c); err != nil {
			return fmt.Errorf("smart record client option %d failed: %s", i, err)
		}
	}
	return nil
}

var clientDefaults = func(o *clientConfig) error {
	o.anOption = []int{1}
	return nil
}

// AnOption
func AnOption(o interface{}) ClientOption {
	return func(c *clientConfig) error {
		c.anOption = o
		return nil
	}
}
