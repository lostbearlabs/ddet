package ddet

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFnord(t *testing.T) {
	expected := 2
	x := Fnord()
	if x != expected {
		t.Error("Expected", expected, "got", x)
	}
}

func TestAll(t *testing.T) {
	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dbpath := dbdir + "/foo.db"

	db := InitDB(dbpath)
	defer db.Close()
	CreateTable(db)

	items := []FileEntry{
		FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		FileEntry{"/foo3.txt", 3, 4, "XYZ3", 200},
	}
	StoreFileEntry(db, items)

	ReadAllFileEntriess := ReadAllFileEntries(db)
	t.Log(ReadAllFileEntriess)

	items2 := []FileEntry{
		FileEntry{"/foo2.txt", 2, 3, "AXB2", 300},
		FileEntry{"/foo4.txt", 4, 5, "XYZ4", 400},
	}
	StoreFileEntry(db, items2)

	ReadAllFileEntriess2 := ReadAllFileEntries(db)
	t.Log(ReadAllFileEntriess2)

	if len(ReadAllFileEntriess2) != 4 {
		t.Error("wrong number of items, got ", len(ReadAllFileEntriess2))
	}

	expected := items[0]
	probe := ReadAllFileEntriess2[0]
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
	CreateTable(db)

	items := []FileEntry{
		FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		FileEntry{"/foo2.txt", 5, 6, "PQR1", 200},
		FileEntry{"/foo3.txt", 3, 4, "XYZ3", 300},
	}
	target := items[1]
	StoreFileEntry(db, items)

	ReadAllFileEntries := ReadFileEntry(db, target.Path)
	t.Log(ReadAllFileEntries)

	if ReadAllFileEntries == nil {
		t.Error("bad value, expected=", target, ", got=NIL")
		return
	}

	if *ReadAllFileEntries != target {
		t.Error("bad value, expected=", target, ", got=", ReadAllFileEntries)
	}
}

func TestReadAllFileEntriesForUnknownPath(t *testing.T) {

	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dbpath := dbdir + "/foo.db"

	db := InitDB(dbpath)
	defer db.Close()
	CreateTable(db)

	items := []FileEntry{
		FileEntry{"/foo1.txt", 1, 2, "AXB1", 100},
		FileEntry{"/foo2.txt", 5, 6, "PQR1", 200},
		FileEntry{"/foo3.txt", 3, 4, "XYZ3", 300},
	}
	StoreFileEntry(db, items)

	ReadAllFileEntries := ReadFileEntry(db, "/foo4.txt")

	if ReadAllFileEntries != nil {
		t.Error("bad value, expected=nil, got=", ReadAllFileEntries)
	}
}
