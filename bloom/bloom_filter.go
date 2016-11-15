package bloom

import (
	"errors"
	//	"fmt"
)

const md5_len = 16
const num_slots = 5192
const slots_per_entry = 2

type BloomFilter struct {
	ar []byte
}

func New() (BloomFilter, error) {
	return BloomFilter{make([]byte, num_slots)}, nil
}

func toSlot(i int, ar []byte) int {
	a := int(ar[i*2])
	b := int(ar[i*2+1])
	slot := (a<<8 + b) % num_slots
	// fmt.Printf("a=%d, b=%d, slot=%d\n", a, b, slot)
	return slot
}

func (filter *BloomFilter) Add(ar []byte) error {
	if len(ar) != md5_len {
		return errors.New("not an MD5")
	}
	for i := 0; i < slots_per_entry; i++ {
		k := toSlot(i, ar)
		filter.ar[k] = 1
	}
	return nil
}

func (filter *BloomFilter) Contains(ar []byte) (bool, error) {
	if len(ar) != md5_len {
		return false, errors.New("not an MD5")
	}
	for i := 0; i < slots_per_entry; i++ {
		k := toSlot(i, ar)
		if filter.ar[k] == 0 {
			return false, nil
		}
	}
	return true, nil
}
