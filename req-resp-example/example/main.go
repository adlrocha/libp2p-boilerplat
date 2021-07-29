package main

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {
	ctx := context.Background()

	fmt.Println("[*] Starting hosts")

	// Instantiating hosts
	h1, err := libp2p.New(ctx)
	if err != nil {
		panic(err)
	}
	h2, err := libp2p.New(ctx)
	if err != nil {
		panic(err)
	}
	defer h1.Close()
	defer h2.Close()

	// Wait until hosts are ready
	time.Sleep(1 * time.Second)

	fmt.Println("[*] Connecting peers")
	// Connect h1-h2
	err = DialOtherPeer(ctx, h1, *host.InfoFromHost(h2))
	if err != nil {
		panic(err)
	}

}

// DialOtherPeers connects to a set of peers in the experiment.
func DialOtherPeer(ctx context.Context, self host.Host, ai peer.AddrInfo) error {
	if err := self.Connect(ctx, ai); err != nil {
		return fmt.Errorf("Error while dialing peer %v: %w", ai.Addrs, err)
	}
	return nil
}
