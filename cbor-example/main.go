package main

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/adlrocha/libp2p-msg/cbor-example/msg"
	cborutil "github.com/filecoin-project/go-cbor-util"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type h struct {
	ch chan bool
}

var pid protocol.ID = "/test/v0"

func main() {
	ctx := context.Background()
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

	fmt.Println("[*] Connecting peers")
	// Connect h1-h2
	err = DialOtherPeer(ctx, h1, *host.InfoFromHost(h2))
	if err != nil {
		panic(err)
	}

	h := h{ch: make(chan bool)}
	h1.SetStreamHandler(pid, h.handleNewStream)
	time.Sleep(time.Second)
	fmt.Println("Starting to send messages")
	send(ctx, h2, h1.ID(), msg.Msg{[]byte("test"), 1})
	<-h.ch
}

func (h *h) handleNewStream(s network.Stream) {
	obj := msg.Msg{}
	err := cborutil.ReadCborRPC(bufio.NewReader(s), &obj)
	fmt.Println("Received", err, obj)
	close(h.ch)

}

func send(ctx context.Context, h host.Host, p peer.ID, obj msg.Msg) error {

	s, err := h.NewStream(ctx, p, []protocol.ID{pid}...)
	//defer s.Close()
	if err != nil {
		return err
	}
	buffered := bufio.NewWriter(s)
	err = cborutil.WriteCborRPC(buffered, &obj)
	if err != nil {
		panic(err)
	}
	err = buffered.Flush()
	fmt.Println("Message sent", err)
	return nil
}

// DialOtherPeers connects to a set of peers in the experiment.
func DialOtherPeer(ctx context.Context, self host.Host, ai peer.AddrInfo) error {
	if err := self.Connect(ctx, ai); err != nil {
		return fmt.Errorf("Error while dialing peer %v: %w", ai.Addrs, err)
	}
	return nil
}
