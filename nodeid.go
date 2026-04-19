package main

import (
	"bytes"
	"encoding/gob"
	"hash/maphash"
	"sync/atomic"
)

var root atomic.Uint64
var seed maphash.Seed = maphash.MakeSeed()

type NodeID struct {
	id     uint64
	digest uint64
}

func NewNodeIDFrom(v any) NodeID {
	id := root.Add(1)
	return NodeID{id: id, digest: nodeDigest(id, v)}
}

func nodeDigest(id uint64, v any) uint64 {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		panic(err)
	}

	var h maphash.Hash
	h.SetSeed(seed)
	maphash.WriteComparable(&h, id)
	maphash.WriteComparable(&h, string(buf.Bytes()))

	return h.Sum64()
}
