package ingestion

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/multiformats/go-multicodec"
)

var prefix = cid.Prefix{
	Version:  1,
	Codec:    uint64(multicodec.DagJson),
	MhType:   uint64(multicodec.Sha2_256),
	MhLength: 16,
}

/* Requires latest go-ipld-prime
var linkproto = cidlink.LinkPrototype{
	Prefix: prefix,
}
*/

var linkproto = cidlink.LinkBuilder{
	Prefix: prefix,
}

func RandomCids(n int) ([]cid.Cid, []_String, error) {
	var prng = rand.New(rand.NewSource(time.Now().UnixNano()))

	res := make([]cid.Cid, n)
	resString := make([]_String, n)
	for i := 0; i < n; i++ {
		b := make([]byte, 10*n)
		prng.Read(b)
		c, err := prefix.Sum(b)
		if err != nil {
			return nil, nil, err
		}
		res[i] = c
		resString[i] = _String{x: c.String()}
	}
	return res, resString, nil
}

func RandomIndex(srcStore ds.Datastore, numEnt int, previous Link_Index) (Index, Link_Index, error) {
	entries := make([]_Entry, numEnt)
	for i := 0; i < numEnt; i++ {
		_, cids, _ := RandomCids(10)
		entries[i] = _Entry{
			Cids: _List_String__Maybe{m: schema.Maybe_Value, v: &_List_String{cids}},
		}
	}
	lentries := _List_Entry{x: entries}

	index := _Index{
		Entries:  lentries,
		Previous: _Link_Index__Maybe{m: schema.Maybe_Value, v: previous},
	}

	storer := func(_ ipld.LinkContext) (io.Writer, ipld.StoreCommitter, error) {
		buf := bytes.NewBuffer(nil)
		return buf, func(lnk ipld.Link) error {
			c := lnk.(cidlink.Link).Cid
			return srcStore.Put(datastore.NewKey(c.String()), buf.Bytes())
		}, nil
	}

	lnk, err := linkproto.Build(context.Background(), ipld.LinkContext{}, &lentries, storer)
	if err != nil {
		return nil, nil, err
	}
	return &index, &_Link_Index{lnk}, nil

}

/* Using new linksystem for latests version of go-ipld-prime. Require changes in go-legs.
func RandomIndex(srcStore ds.Datastore, numEnt int, previous Link_Index) (Index, Link_Index, error) {
	entries := make([]_Entry, numEnt)
	for i := 0; i < numEnt; i++ {
		_, cids, _ := RandomCids(10)
		entries[i] = _Entry{
			Cids: _List_String__Maybe{m: schema.Maybe_Value, v: &_List_String{cids}},
		}
	}
	lentries := _List_Entry{x: entries}

	index := _Index{
		Entries:  lentries,
		Previous: _Link_Index__Maybe{m: schema.Maybe_Value, v: previous},
	}

	lsys := cidlink.DefaultLinkSystem()
	lsys.StorageWriteOpener = func(_ ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		buf := bytes.NewBuffer(nil)
		return buf, func(lnk ipld.Link) error {
			c := lnk.(cidlink.Link).Cid
			return srcStore.Put(datastore.NewKey(c.String()), buf.Bytes())
		}, nil
	}

	lnk, err := lsys.Store(ipld.LinkContext{}, linkproto, &index)
	if err != nil {
		return nil, nil, err
	}
	return &index, &_Link_Index{lnk}, nil

}
*/
