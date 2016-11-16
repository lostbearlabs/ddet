package dset

import (
	"com.lostbearlabs/ddet/filedb"
	"testing"
)

func TestEmptyReturnsNone(t *testing.T) {
	ks := New()
	dk := ks.GetDuplicateKeys()
	if dk == nil {
		t.Error("should not be nil")
	}
	if len(dk) > 0 {
		t.Error("should not have entries")
	}
}

func TestSingleDupReturnsIt(t *testing.T) {
	db := filedb.NewTempDB()
	defer db.Close()

	items := []*filedb.FileEntry{
		filedb.NewTestFileEntry().SetPath("/foo1.txt"),
		filedb.NewTestFileEntry().SetPath("/foo3.txt"),
	}
	db.StoreFileEntries(items)

	ks := New()
	ks.AddAll(db, "")

	dupKeys := ks.GetDuplicateKeys()
	if dupKeys == nil {
		t.Error("should not be nil")
	}
	if len(dupKeys) != 1 {
		t.Error("length should be 1, was", len(dupKeys))
	}

	entries := ks.GetFileEntries(db, dupKeys[0])
	if len(entries) != 2 {
		t.Error("length should be 2, was", len(entries))
	}
	for i, entry := range entries {
		if entry != *items[i] {
			t.Error("bad value at index ", i, " got ", entry, " expected ", *items[i])
		}
	}
}

func NonDupNotReturned(t *testing.T) {
	db := filedb.NewTempDB()
	defer db.Close()

	items := []*filedb.FileEntry{
		filedb.NewTestFileEntry().SetPath("/foo1.txt").SetMd5("ABX1"),
		filedb.NewTestFileEntry().SetPath("/foo3.txt").SetMd5("ABX2"),
	}
	db.StoreFileEntries(items)

	ks := New()
	ks.AddAll(db, "")

	dupKeys := ks.GetDuplicateKeys()
	if dupKeys == nil {
		t.Error("should not be nil")
	}
	if len(dupKeys) != 0 {
		t.Error("length should be 0, was", len(dupKeys))
	}
}
