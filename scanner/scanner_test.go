package scanner

import (
	"com.lostbearlabs/ddet/filedb"
	"io/ioutil"
	"os"
	"testing"
)

func confirmItem(t *testing.T, it filedb.FileEntry, path string, length int64) {
	if it.Path != path {
		t.Error("wrong path, expected=", path, ", got=", it.Path)
	}
	if it.Length != length {
		t.Error("wrong length, expected=", length, ", got=", it.Length)
	}
}

func TestScan(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "data")
	defer os.Remove(dir)

	name1 := dir + "/file1"
	name2 := dir + "/file2"
	name3 := dir + "/x/file3"
	os.Mkdir(dir+"/x", 0755)
	ioutil.WriteFile(name1, []byte("constant text string 1"), 0644)
	ioutil.WriteFile(name2, []byte("constant text string 22"), 0644)
	ioutil.WriteFile(name3, []byte("constant text string 333"), 0644)

	db, _ := filedb.NewTempDB()
	defer db.Close()

	scanner := MakeScanner(db)
	scanner.ScanFiles(dir)

	allFileEntriess, _ := db.ReadAllFileEntries()
	if len(allFileEntriess) != 3 {
		t.Error("wrong length, expected=3, got=", len(allFileEntriess))
	}
	confirmItem(t, allFileEntriess[0], name1, 22)
	confirmItem(t, allFileEntriess[1], name2, 23)
	confirmItem(t, allFileEntriess[2], name3, 24)

}

func TestScanUnchangedFile(t *testing.T) {

	dir, _ := ioutil.TempDir(os.TempDir(), "data")
	defer os.Remove(dir)

	name1 := dir + "/file1"
	ioutil.WriteFile(name1, []byte("constant text string 1"), 0644)

	db, _ := filedb.NewTempDB()
	defer db.Close()

	scanner := MakeScanner(db)
	scanner.ScanFiles(dir)
	read1 := db.ReadFileEntry(name1)

	scanner2 := MakeScanner(db)
	scanner2.ScanFiles(dir)
	read2 := db.ReadFileEntry(name1)

	if *read2 != *read1 {
		t.Error("File should not have been scanned with no change, read1=", read1, ", read2=", read2)
	}

}

func TestScanChangedFile(t *testing.T) {

	dir, _ := ioutil.TempDir(os.TempDir(), "data")
	defer os.Remove(dir)

	name1 := dir + "/file1"
	ioutil.WriteFile(name1, []byte("constant text string 1"), 0644)

	db, _ := filedb.NewTempDB()
	defer db.Close()

	scanner := MakeScanner(db)
	scanner.ScanFiles(dir)

	read1 := db.ReadFileEntry(name1)

	// change length of file.  (We can't reliably test mod time to within 1-second
	// unless we stick a sleep in here?)
	ioutil.WriteFile(name1, []byte("constant text string 22"), 0644)

	scanner2 := MakeScanner(db)
	scanner2.ScanFiles(dir)
	read2 := db.ReadFileEntry(name1)

	if *read2 == *read1 {
		t.Error("File should have been scanned after change, read1=", read1, ", read2=", read2)
	}

}
