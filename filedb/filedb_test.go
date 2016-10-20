package filedb

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestReadAllFileEntries(t *testing.T) {
	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dbpath := dbdir + "/foo.db"

	db := InitDB(dbpath)
	defer db.Close()

	items := []*FileEntry{
		&FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		&FileEntry{"/foo3.txt", 3, 4, "XYZ3", 200},
	}
	db.StoreFileEntries(items)

	allFileEntries := db.ReadAllFileEntries()
	t.Log(allFileEntries)

	items2 := []*FileEntry{
		&FileEntry{"/foo2.txt", 2, 3, "AXB2", 300},
		&FileEntry{"/foo4.txt", 4, 5, "XYZ4", 400},
	}
	db.StoreFileEntries(items2)

	allFileEntriess2 := db.ReadAllFileEntries()
	t.Log(allFileEntriess2)

	if len(allFileEntriess2) != 4 {
		t.Error("wrong number of items, got ", len(allFileEntriess2))
	}

	expected := *items[0]
	probe := allFileEntriess2[0]
	if expected != probe {
		t.Error("bad value, expected=", expected, ", got=", probe)
	}
}

func TestReadFileEntry(t *testing.T) {

	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dbpath := dbdir + "/foo.db"

	db := InitDB(dbpath)
	defer db.Close()

	items := []*FileEntry{
		&FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		&FileEntry{"/foo2.txt", 5, 6, "PQR1", 200},
		&FileEntry{"/foo3.txt", 3, 4, "XYZ3", 300},
	}
	target := *items[1]
	db.StoreFileEntries(items)

	fileEntry := db.ReadFileEntry(target.Path)
	t.Log(fileEntry)

	if fileEntry == nil {
		t.Error("bad value, expected=", target, ", got=NIL")
		return
	}

	if *fileEntry != target {
		t.Error("bad value, expected=", target, ", got=", fileEntry)
	}
}

func TestReadFileEntryForUnknownPath(t *testing.T) {

	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dbpath := dbdir + "/foo.db"

	db := InitDB(dbpath)
	defer db.Close()

	items := []*FileEntry{
		&FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		&FileEntry{"/foo2.txt", 5, 6, "PQR1", 200},
		&FileEntry{"/foo3.txt", 3, 4, "XYZ3", 300},
	}
	db.StoreFileEntries(items)

	fileEntry := db.ReadFileEntry("/foo4.txt")

	if fileEntry != nil {
		t.Error("bad value, expected=nil, got=", fileEntry)
	}
}

func TestDeleteOldEntries(t *testing.T) {

	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dbpath := dbdir + "/foo.db"

	db := InitDB(dbpath)
	defer db.Close()

	items := []*FileEntry{
		&FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		&FileEntry{"/foo2.txt", 5, 6, "PQR1", 200},
		&FileEntry{"/foo3.txt", 3, 4, "XYZ3", 300},
	}
	db.StoreFileEntries(items)

	db.DeleteOldEntries(200)
	allEntries := db.ReadAllFileEntries()
	if len(allEntries) != 2 {
		t.Error("should have got 2, got", len(allEntries))
	}

	db.DeleteOldEntries(301)
	allEntries = db.ReadAllFileEntries()
	if len(allEntries) != 0 {
		t.Error("should have got 0, got", len(allEntries))
	}
}
