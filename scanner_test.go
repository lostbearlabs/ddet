package ddet

import (
	"io/ioutil"
	"os"
	"testing"
)

func confirmItem(t *testing.T, it FileEntry, path string) {
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

	ReadAllFileEntriess := ReadAllFileEntries(db)
	if len(ReadAllFileEntriess) != 3 {
		t.Error("wrong length, expected=3, got=", len(ReadAllFileEntriess))
	}
	confirmItem(t, ReadAllFileEntriess[0], name1)
	confirmItem(t, ReadAllFileEntriess[1], name2)
	confirmItem(t, ReadAllFileEntriess[2], name3)

}

func TestScanUnchangedFile(t *testing.T) {

	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dir, _ := ioutil.TempDir(os.TempDir(), "data")
	defer os.Remove(dir)

	name1 := dir + "/file1"
	ioutil.WriteFile(name1, []byte("constant text string 1"), 0644)

	dbpath := dbdir + "/foo.db"
	db := InitDB(dbpath)
	defer db.Close()
	CreateTable(db)

	scanner := MakeScanner(db)
	scanner.ScanFiles(dir)
	read1 := ReadFileEntry(db, name1)

	scanner2 := MakeScanner(db)
	scanner2.ScanFiles(dir)
	read2 := ReadFileEntry(db, name1)

	if *read2 != *read1 {
		t.Error("File should not have been scanned with no change, read1=", read1, ", read2=", read2)
	}

}

func TestScanChangedFile(t *testing.T) {

	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	defer os.Remove(dbdir)

	dir, _ := ioutil.TempDir(os.TempDir(), "data")
	defer os.Remove(dir)

	name1 := dir + "/file1"
	ioutil.WriteFile(name1, []byte("constant text string 1"), 0644)

	dbpath := dbdir + "/foo.db"
	db := InitDB(dbpath)
	defer db.Close()
	CreateTable(db)

	scanner := MakeScanner(db)
	scanner.ScanFiles(dir)

	read1 := ReadFileEntry(db, name1)

	// change length of file.  (We can't reliably test mod time to within 1-second
	// unless we stick a sleep in here?)
	ioutil.WriteFile(name1, []byte("constant text string 22"), 0644)

	scanner2 := MakeScanner(db)
	scanner2.ScanFiles(dir)
	read2 := ReadFileEntry(db, name1)

	if *read2 == *read1 {
		t.Error("File should have been scanned after change, read1=", read1, ", read2=", read2)
	}

}
