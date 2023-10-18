package cfilter

import (
	"bytes"
	"hash"
)

type fingerprint struct {
	key []byte
	metadata []byte
}

func fprint(item []byte, fpSize uint8, metaSize uint8, hashfn hash.Hash) fingerprint {
	hashfn.Reset()
	hashfn.Write(item)
	h := hashfn.Sum(nil)

	fp := make([]byte, fpSize)
	copy(fp, h)


	meta := make([]byte, metaSize)

	return fingerprint{key: fp, metadata: meta}

}

func hashfp(f []byte) uint {
	var h uint = 5381


	for i := range f {
		h = ((h << 5) + h) + uint(f[i])
	}

	return h
}

func match(a, b fingerprint) bool {
	return bytes.Equal(a.key, b.key)
}

func getKey(f fingerprint) []byte {
	return f.key
}

func getMeta(f fingerprint) []byte {
	return f.metadata
}
