// +build cbg

package main

import (
	"os"
	"path"

	msg "github.com/adlrocha/libp2p-boilerplate/cbor-example/msg"
	cborgen "github.com/whyrusleeping/cbor-gen"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	obj_file := path.Clean(path.Join(wd, "..", "msg_cbor_gen.go"))
	err = cborgen.WriteMapEncodersToFile(
		obj_file,
		"msg",
		msg.Msg{},
	)
	if err != nil {
		panic(err)
	}
}
