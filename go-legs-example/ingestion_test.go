package ingestion

import (
	"context"
	"testing"
	"time"

	ds "github.com/ipfs/go-datastore"

	// dagjson codec registered for encoding
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/willscott/go-legs"
)

func mkTestHost() host.Host {
	h, _ := libp2p.New(context.Background(), libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	return h
}

func TestRoundTrip(t *testing.T) {
	srcHost := mkTestHost()
	srcStore := ds.NewMapDatastore()
	lp, err := legs.Publish(context.Background(), srcStore, srcHost, "legs/testtopic")
	if err != nil {
		t.Fatal(err)
	}

	dstHost := mkTestHost()
	srcHost.Peerstore().AddAddrs(dstHost.ID(), dstHost.Addrs(), time.Hour)
	dstHost.Peerstore().AddAddrs(srcHost.ID(), srcHost.Addrs(), time.Hour)
	if err := srcHost.Connect(context.Background(), dstHost.Peerstore().PeerInfo(dstHost.ID())); err != nil {
		t.Fatal(err)
	}
	dstStore := ds.NewMapDatastore()
	ls, err := legs.Subscribe(context.Background(), dstStore, dstHost, "legs/testtopic")
	if err != nil {
		t.Fatal(err)
	}

	// per https://github.com/libp2p/go-libp2p-pubsub/blob/e6ad80cf4782fca31f46e3a8ba8d1a450d562f49/gossipsub_test.go#L103
	// we don't seem to have a way to manually trigger needed gossip-sub heartbeats for mesh establishment.
	time.Sleep(time.Second)

	watcher, cncl := ls.OnChange()
	_, ilnk1, err := GenesisIndex(srcStore, 3)
	if err != nil {
		t.Fatal(err)
	}
	_, ilnk2, err := RandomIndex(srcStore, 3, ilnk1)
	if err != nil {
		t.Fatal(err)
	}

	lnk2, err := ilnk2.AsLink()
	if err != nil {
		t.Fatal(err)
	}
	lnk1, err := ilnk1.AsLink()
	if err != nil {
		t.Fatal(err)
	}

	// Check if both indices stored in source store before fetching
	if _, err := srcStore.Get(ds.NewKey(lnk2.(cidlink.Link).Cid.String())); err != nil {
		t.Fatalf("data for lnk2 in source store: %v", err)
	}
	if _, err := srcStore.Get(ds.NewKey(lnk1.(cidlink.Link).Cid.String())); err != nil {
		t.Fatalf("data for lnk1 in source store: %v", err)
	}
	defer func() {
		cncl()
		lp.Close(context.Background())
		ls.Close(context.Background())
	}()

	// Share current root of the chain in pubsub channel (i.e. lnk2)
	if err := lp.UpdateRoot(context.Background(), lnk2.(cidlink.Link).Cid); err != nil {
		t.Fatal(err)
	}

	// Wait for DAG to be synced
	select {
	case <-time.After(time.Second * 5):
		t.Fatal("timed out waiting for sync to propogate")
	case downstream := <-watcher:
		if !downstream.Equals(lnk2.(cidlink.Link).Cid) {
			t.Fatalf("sync'd sid unexpected %s vs %s", downstream, lnk2)
		}
		// Check if all indices have been received.
		if _, err := dstStore.Get(ds.NewKey(downstream.String())); err != nil {
			t.Fatalf("data for lnk2 not in receiver store: %v", err)
		}
		if _, err := dstStore.Get(ds.NewKey(lnk1.(cidlink.Link).Cid.String())); err != nil {
			t.Fatalf("data for lnk1 in receiver store: %v", err)
		}
	}
}
