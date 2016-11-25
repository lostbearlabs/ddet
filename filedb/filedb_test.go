package filedb

import (
	"testing"
)

func TestReadAllFileEntries(t *testing.T) {
	db := NewTempDB()
	defer db.Close()

	items := []*FileEntry{
		NewTestFileEntry().SetPath("/foo1.txt"),
		NewTestFileEntry().SetPath("/foo3.txt"),
	}
	db.StoreFileEntries(items)

	allFileEntries := db.ReadAllFileEntries()
	t.Log(allFileEntries)

	items2 := []*FileEntry{
		NewTestFileEntry().SetPath("/foo2.txt"),
		NewTestFileEntry().SetPath("/foo4.txt"),
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
	db := NewTempDB()
	defer db.Close()

	items := []*FileEntry{
		NewTestFileEntry().SetPath("/foo1.txt"),
		NewTestFileEntry().SetPath("/foo2.txt").SetLastMod(1).SetLength(2).SetScanTime(3).SetMd5("PQR1"),
		NewTestFileEntry().SetPath("/foo3.txt"),
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
	db := NewTempDB()
	defer db.Close()

	items := []*FileEntry{
		NewTestFileEntry().SetPath("/foo1.txt"),
		NewTestFileEntry().SetPath("/foo2.txt"),
		NewTestFileEntry().SetPath("/foo3.txt"),
	}
	db.StoreFileEntries(items)

	fileEntry := db.ReadFileEntry("/foo4.txt")

	if fileEntry != nil {
		t.Error("bad value, expected=nil, got=", fileEntry)
	}
}

func TestDeleteOldEntries(t *testing.T) {
	db := NewTempDB()
	defer db.Close()

	items := []*FileEntry{
		NewTestFileEntry().SetPath("/foo1.txt").SetScanTime(100),
		NewTestFileEntry().SetPath("/foo2.txt").SetScanTime(200),
		NewTestFileEntry().SetPath("/foo3.txt").SetScanTime(300),
	}
	db.StoreFileEntries(items)

	db.DeleteOldEntries("/", 200)
	allEntries := db.ReadAllFileEntries()
	if len(allEntries) != 2 {
		t.Error("should have got 2, got", len(allEntries))
	}

	db.DeleteOldEntries("/", 301)
	allEntries = db.ReadAllFileEntries()
	if len(allEntries) != 0 {
		t.Error("should have got 0, got", len(allEntries))
	}
}

func TestDeletePathPrefix(t *testing.T) {

	db := NewTempDB()
	defer db.Close()

	items := []*FileEntry{
		NewTestFileEntry().SetPath("/a/foo1.txt").SetScanTime(100),
		NewTestFileEntry().SetPath("/b/foo2.txt").SetScanTime(200),
		NewTestFileEntry().SetPath("/a/foo3.txt").SetScanTime(300),
	}
	db.StoreFileEntries(items)

	db.DeleteOldEntries("/a", 500)
	allEntries := db.ReadAllFileEntries()
	if len(allEntries) != 1 {
		t.Error("should have got 1, got", len(allEntries))
	}
}

func TestReadEntriesByMd5(t *testing.T) {
	db := NewTempDB()
	defer db.Close()

	items := []*FileEntry{
		NewTestFileEntry().SetPath("/foo1.txt"),
		NewTestFileEntry().SetPath("/foo2.txt").SetMd5("39879ddb5f9936cee72ff46ece623183"),
		NewTestFileEntry().SetPath("/foo3.txt"),
	}
	db.StoreFileEntries(items)

	items1 := db.ReadFileEntriesByKnownFileKey(items[0].Md5, items[0].Length)
	if len(items1) != 2 {
		t.Error("wrong number of items, got", len(items1))
	}

	items2 := db.ReadFileEntriesByKnownFileKey(items[1].Md5, items[1].Length)
	if len(items2) != 1 {
		t.Error("wrong number of items, got", len(items2))
	}
}
