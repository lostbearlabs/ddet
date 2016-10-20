package dset

import (
	"com.lostbearlabs/ddet/filedb"
	"io/ioutil"
	"os"
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

// TODO: need some kind of test database so we don't duplicate this dir/path/InitDB
// stuff all over the place.
func TestSingleDupReturnsIt(t *testing.T) {

	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dbpath := dbdir + "/foo.db"
	db := filedb.InitDB(dbpath)
	defer db.Close()

	items := []*filedb.FileEntry{
		&filedb.FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		&filedb.FileEntry{"/foo3.txt", 1, 4, "AXB1", 300},
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

	entries := ks.GetFileEntries(dupKeys[0])
	if len(entries) != 2 {
		t.Error("length should be 2, was", len(entries))
	}
	for i, entry := range entries {
		if *entry != *items[i] {
			t.Error("bad value at index ", i, " got ", *items[i], " expected ", *entry)
		}
	}
}

func NonDupNotReturned(t *testing.T) {

	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dbpath := dbdir + "/foo.db"
	db := filedb.InitDB(dbpath)
	defer db.Close()

	items := []*filedb.FileEntry{
		&filedb.FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		&filedb.FileEntry{"/foo3.txt", 1, 4, "AXB2", 300},
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
