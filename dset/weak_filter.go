package dset

import (
	"com.lostbearlabs/ddet/bloom"
	"crypto/md5"
	"encoding/binary"
)

// This is the weak filter used internally to the KnownFileSet in order
// to improve the space efficiency of duplicate detection.  Our KnownFileKey
// contains both the file MD5 and the file length, but our Bloom filter
// implementation can only handle MD5 hashes, so this class just re-hashes the
// full key and then keeps the result in a bloom filter.
type weakFilter struct {
	bloom bloom.BloomFilter
}

func newWeakFilter() (*weakFilter, error) {
	bloomFilter, err := bloom.New()
	if err != nil {
		return nil, err
	}
	return &weakFilter{bloomFilter}, nil
}

func rehash(fileHash []byte, length int64) []byte {
	tmp := make([]byte, len(fileHash)+8)
	copy(tmp[0:len(fileHash)], fileHash)
	binary.PutVarint(tmp[len(fileHash):], length)

	newHash := md5.Sum(tmp)
	return newHash[:]
}

func (f *weakFilter) add(hash []byte, length int64) error {
	return f.bloom.Add(rehash(hash, length))
}

func (f *weakFilter) contains(hash []byte, length int64) (bool, error) {
	return f.bloom.Contains(rehash(hash, length))
}
