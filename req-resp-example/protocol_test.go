package protocol_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/adlrocha/libp2p-boilerplate/req-resp-example/client"
	"github.com/adlrocha/libp2p-boilerplate/req-resp-example/server"
	"github.com/libp2p/go-libp2p-core/host"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

func setupServer(ctx context.Context, t *testing.T) (server.Server, host.Host) {

	h := bhost.New(swarmt.GenSwarm(t, ctx, swarmt.OptDisableReuseport))
	s, err := server.NewServer(ctx, h)
	if err != nil {
		t.Fatal(err)
	}
	return s, h
}

func setupClient(ctx context.Context, t *testing.T) (client.Client, host.Host) {
	h := bhost.New(swarmt.GenSwarm(t, ctx, swarmt.OptDisableReuseport))
	c, err := client.NewClient(ctx, h)
	if err != nil {
		t.Fatal(err)
	}
	return c, h
}

func connect(ctx context.Context, t *testing.T, h1 host.Host, h2 host.Host) {
	if err := h1.Connect(ctx, *host.InfoFromHost(h2)); err != nil {
		t.Fatal(err)
	}
}

func TestE2E(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, ch := setupClient(ctx, t)
	_, sh := setupServer(ctx, t)
	connect(ctx, t, ch, sh)

	value := []byte("{\"hello\": \"world\"}")

	resp, err := c.SendGet(ctx, value, sh.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(resp.([]byte), value) {
		t.Fatal("Something in the protocol failed", resp, value)
	}

}
