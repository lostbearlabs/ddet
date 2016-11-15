package dset

import (
	"com.lostbearlabs/ddet/bloom"
	"com.lostbearlabs/ddet/filedb"
	"encoding/hex"
	"github.com/juju/loggo"
	"sort"
)

var logger = loggo.GetLogger("dset")

// The key used to compare files.  Files with the same key are
// treated as identical.
type KnownFileKey struct {
	md5    string
	length int64
}

// ByLength implements sort.Interface for []KnownFileKey based on
// the length field first, then the md5
type ByLength []KnownFileKey

func (a ByLength) Len() int      { return len(a) }
func (a ByLength) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByLength) Less(i, j int) bool {
	return (a[i].length < a[j].length) || ((a[i].length == a[j].length) && (a[i].md5 < a[j].md5))
}

// A set of files with the same key.
type FilesWithSameKey struct {
	key KnownFileKey

	// The individual entries are keyed by path.
	// (We could just use a list, but this strictly prevents
	// duplicates and avoids the mis-apprehension that the
	// set might be sorted)
	entries map[string]*filedb.FileEntry
}

// A set of files, from which we can extract any duplicate files,
// i.e. sets of multiple files with the same key.
type KnownFileSet struct {
	mp       map[KnownFileKey]*FilesWithSameKey
	numFiles int64
	bf1      bloom.BloomFilter
	mp2      map[string]bool
}

func New() *KnownFileSet {
	mp := make(map[KnownFileKey]*FilesWithSameKey)
	mp2 := make(map[string]bool)
	bf, _ := bloom.New()

	return &KnownFileSet{mp, 0, bf, mp2}
}

func (k *KnownFileSet) GetNumFiles() int64 {
	return k.numFiles
}

func (k *KnownFileSet) AddAll(db *filedb.FileDB, path string) {
	// Identify all the MD5 values that occur more than once
	db.ProcessAllFileEntries(k.populateFilters, path)
	logger.Infof("first pass identified %d potential groups of duplicates", len(k.mp2))

	// Group into file sets.

	// Approach #1: Re-scan database:
	// db.ProcessAllFileEntries(k.Add, path)

	// Approach #2: Query by MD5
	for md5, _ := range k.mp2 {
		items := db.ReadFileEntriesByMd5(md5)
		for _, item := range items {
			k.Add(item)
		}
	}
}

func considerPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func (k *KnownFileSet) populateFilters(e filedb.FileEntry) {
	md5, err := hex.DecodeString(e.Md5)
	considerPanic(err)

	known, err := k.bf1.Contains(md5)
	if err != nil {
		logger.Infof("ERROR: [%v] from [%v] with md5 decoded as [%v]", err, e, md5)
		return
	}

	if known && !k.mp2[e.Md5] {
		logger.Tracef("found interesting MD5: %v", e.Md5)
		k.mp2[e.Md5] = true
	} else {
		k.bf1.Add(md5)
	}
}

func (k *KnownFileSet) Add(e filedb.FileEntry) {
	if !k.mp2[e.Md5] {
		return
	}

	key := KnownFileKey{e.Md5, e.Length}
	l, prs := k.mp[key]
	if !prs {
		entries := make(map[string]*filedb.FileEntry)
		l = &FilesWithSameKey{key, entries}
		k.mp[key] = l
		logger.Tracef("created bucket for %v", key)
	}
	if l.entries[e.Path] == nil {
		tag := ""
		if len(l.entries) > 0 {
			tag = "  !!DUPLICATE!!  "
		}
		logger.Tracef("key=%v, adding path %v%s\n", key, e.Path, tag)
		l.entries[e.Path] = &e
		k.numFiles++
	}
}

func (k *KnownFileSet) GetDuplicateKeys() []KnownFileKey {
	keys := make([]KnownFileKey, 0)

	for x, val := range k.mp {
		if len(val.entries) > 1 {
			keys = append(keys, x)
		}
	}

	sort.Sort(ByLength(keys))

	logger.Tracef("duplicate keys: %v", keys)
	return keys
}

func (k *KnownFileSet) GetFileEntries(key KnownFileKey) []*filedb.FileEntry {
	list := k.mp[key]

	paths := make([]string, 0)
	for path, _ := range list.entries {
		paths = append(paths, path)
	}
	logger.Tracef("paths: %v\n", paths)
	sort.Strings(paths)
	logger.Tracef("sorted paths: %s\n", paths)

	ar := make([]*filedb.FileEntry, 0)
	for _, path := range paths {
		ar = append(ar, list.entries[path])
	}
	logger.Tracef("ar: %v\n", ar)
	return ar
}
