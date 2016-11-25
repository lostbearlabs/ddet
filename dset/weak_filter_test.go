package dset

import (
	"crypto/md5"
	"testing"
)

func TestCreate(t *testing.T) {
	_, err := newWeakFilter()
	if err != nil {
		t.Error(err)
	}
}

func TestHappyPath(t *testing.T) {
	filter, _ := newWeakFilter()

	hash := md5.Sum([]byte("test string"))
	len := int64(17)

	rc, err := filter.contains(hash[:], len)
	if err != nil {
		t.Error(err)
	}
	if rc {
		t.Error("should not contain entry ")
	}

	err = filter.add(hash[:], len)
	if err != nil {
		t.Error(err)
	}

	rc, err = filter.contains(hash[:], len)
	if err != nil {
		t.Error(err)
	}
	if !rc {
		t.Error("should contain entry ")
	}
}
