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

	items := []TestItem{
		TestItem{"/foo1.txt", 1, 2, "AXB1"},
		TestItem{"/foo3.txt", 3, 4, "XYZ3"},
	}
	StoreItem(db, items)

	readItems := ReadItem(db)
	t.Log(readItems)

	items2 := []TestItem{
		TestItem{"/foo2.txt", 2, 3, "AXB2"},
		TestItem{"/foo4.txt", 4, 5, "XYZ4"},
	}
	StoreItem(db, items2)

	readItems2 := ReadItem(db)
	t.Log(readItems2)

	if len(readItems2) != 4 {
		t.Error("wrong number of items, got ", len(readItems2))
	}

	expected := items[0]
	probe := readItems2[0]
	if expected != probe {
		t.Error("bad value, expected=", expected, ", got=", probe)
	}
}
