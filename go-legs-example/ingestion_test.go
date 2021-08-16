package ingestion

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"

	// dagjson codec registered for encoding
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/fluent"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	selectorbuilder "github.com/ipld/go-ipld-prime/traversal/selector/builder"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/multiformats/go-multicodec"
	"github.com/willscott/go-legs"
)

func mkTestHost() host.Host {
	h, _ := libp2p.New(context.Background(), libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	return h
}

func mkLinkSystem(ds datastore.Batching) ipld.LinkSystem {
	lsys := cidlink.DefaultLinkSystem()
	lsys.StorageReadOpener = func(lctx ipld.LinkContext, lnk ipld.Link) (io.Reader, error) {
		c := lnk.(cidlink.Link).Cid
		val, err := ds.Get(datastore.NewKey(c.String()))
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(val), nil
	}
	lsys.StorageWriteOpener = func(_ ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		buf := bytes.NewBuffer(nil)
		return buf, func(lnk ipld.Link) error {
			c := lnk.(cidlink.Link).Cid
			return ds.Put(datastore.NewKey(c.String()), buf.Bytes())
		}, nil
	}
	return lsys
}

func initPubSub(t *testing.T, srcStore, dstStore datastore.Batching) (legs.LegPublisher, legs.LegSubscriber) {
	srcHost := mkTestHost()
	srcLnkS := mkLinkSystem(srcStore)
	lp, err := legs.NewPublisher(context.Background(), srcStore, srcHost, "legs/testtopic", srcLnkS)
	if err != nil {
		t.Fatal(err)
	}

	dstHost := mkTestHost()
	srcHost.Peerstore().AddAddrs(dstHost.ID(), dstHost.Addrs(), time.Hour)
	dstHost.Peerstore().AddAddrs(srcHost.ID(), srcHost.Addrs(), time.Hour)
	if err := srcHost.Connect(context.Background(), dstHost.Peerstore().PeerInfo(dstHost.ID())); err != nil {
		t.Fatal(err)
	}
	dstLnkS := mkLinkSystem(dstStore)
	ls, err := legs.NewSubscriber(context.Background(), dstStore, dstHost, "legs/testtopic", dstLnkS)
	if err != nil {
		t.Fatal(err)
	}
	return lp, ls
}

func mkRoot(srcStore datastore.Batching, n ipld.Node) (ipld.Link, error) {
	linkproto := cidlink.LinkPrototype{
		Prefix: cid.Prefix{
			Version:  1,
			Codec:    uint64(multicodec.DagJson),
			MhType:   uint64(multicodec.Sha2_256),
			MhLength: 16,
		},
	}
	lsys := cidlink.DefaultLinkSystem()
	lsys.StorageWriteOpener = func(_ ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		buf := bytes.NewBuffer(nil)
		return buf, func(lnk ipld.Link) error {
			c := lnk.(cidlink.Link).Cid
			return srcStore.Put(datastore.NewKey(c.String()), buf.Bytes())
		}, nil
	}

	return lsys.Store(ipld.LinkContext{}, linkproto, n)
}

func ExploreRecursive(limit selector.RecursionLimit, sequence selectorbuilder.SelectorSpec, stopLnk ipld.Link) ipld.Node {
	np := basicnode.Prototype__Map{}
	return fluent.MustBuildMap(np, 1, func(na fluent.MapAssembler) {
		// RecursionLimit
		na.AssembleEntry(selector.SelectorKey_ExploreRecursive).CreateMap(3, func(na fluent.MapAssembler) {
			na.AssembleEntry(selector.SelectorKey_Limit).CreateMap(1, func(na fluent.MapAssembler) {
				switch limit.Mode() {
				case selector.RecursionLimit_Depth:
					na.AssembleEntry(selector.SelectorKey_LimitDepth).AssignInt(limit.Depth())
				case selector.RecursionLimit_None:
					na.AssembleEntry(selector.SelectorKey_LimitNone).CreateMap(0, func(na fluent.MapAssembler) {})
				default:
					panic("Unsupported recursion limit type")
				}
			})
			// Sequence
			na.AssembleEntry(selector.SelectorKey_Sequence).AssignNode(sequence.Node())

			// Stop condition
			if stopLnk != nil {
				cond := fluent.MustBuildMap(basicnode.Prototype__Map{}, 1, func(na fluent.MapAssembler) {
					na.AssembleEntry(string(selector.ConditionMode_Link)).AssignLink(stopLnk)
				})
				na.AssembleEntry(selector.SelectorKey_StopAt).AssignNode(cond)
			}
		})
	})

}
func TestRoundTrip(t *testing.T) {
	var latestAd cid.Cid

	// Init legs publisher and subscriber
	srcStore := ds.NewMapDatastore()
	dstStore := ds.NewMapDatastore()
	lp, ls := initPubSub(t, srcStore, dstStore)

	_, ilnk1, err := GenesisIndex(srcStore, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, ilnk2, err := RandomIndex(srcStore, 1, ilnk1)
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

	// Fetch-all recursively selector
	np := basicnode.Prototype__Any{}
	ssb := selectorbuilder.NewSelectorSpecBuilder(np)
	// sn := ssb.ExploreRecursive(selector.RecursionLimitNone(), ssb.ExploreAll(ssb.ExploreRecursiveEdge())).Node()
	sn := ExploreRecursive(selector.RecursionLimitNone(), ssb.ExploreAll(ssb.ExploreRecursiveEdge()), lnk1)

	// Subscription handler to perform a plain exchange
	handler, err := legs.PlainExchangeWithSelector(ls, sn)
	if err != nil {
		t.Fatal(err)
	}

	err = ls.Subscribe(context.Background(), handler)
	if err != nil {
		t.Fatal(err)
	}

	// per https://github.com/libp2p/go-libp2p-pubsub/blob/e6ad80cf4782fca31f46e3a8ba8d1a450d562f49/gossipsub_test.go#L103
	// we don't seem to have a way to manually trigger needed gossip-sub heartbeats for mesh establishment.
	time.Sleep(time.Second)

	watcher, cncl := ls.OnChange()

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
	case <-time.After(time.Second * 5000):
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
		// Update latestAd
		latestAd = lnk2.(cidlink.Link).Cid
	}
	fmt.Println(latestAd)
}
