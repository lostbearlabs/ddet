package ddet

import (
	"io/ioutil"
	"os"
	"testing"
)

func confirmItem(t *testing.T, it TestItem, path string) {
	if it.Path != path {
		t.Error("wrong path, expected=", path, ", got=", it.Path)
	}
	// TODO: other fields
}

func TestScan(t *testing.T) {
	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dir, _ := ioutil.TempDir(os.TempDir(), "data")
	defer os.Remove(dir)

	name1 := dir + "/file1"
	name2 := dir + "/file2"
	name3 := dir + "/x/file3"
	os.Mkdir(dir+"/x", 0755)
	ioutil.WriteFile(name1, []byte("constant text string 1"), 0644)
	ioutil.WriteFile(name2, []byte("constant text string 22"), 0644)
	ioutil.WriteFile(name3, []byte("constant text string 333"), 0644)

	dbpath := dbdir + "/foo.db"
	db := InitDB(dbpath)
	defer db.Close()
	CreateTable(db)

	scanner := MakeScanner(db)
	scanner.ScanFiles(dir)

	readItems := ReadItem(db)
	if len(readItems) != 3 {
		t.Error("wrong length, expected=3, got=", len(readItems))
	}
	confirmItem(t, readItems[0], name1)
	confirmItem(t, readItems[1], name2)
	confirmItem(t, readItems[2], name3)

}
