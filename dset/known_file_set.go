package dset

import (
	"com.lostbearlabs/ddet/filedb"
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
	mp map[KnownFileKey]*FilesWithSameKey
	numFiles int64
}

func New() *KnownFileSet {
	mp := make(map[KnownFileKey]*FilesWithSameKey)
	return &KnownFileSet{mp, 0}
}

func (k *KnownFileSet) GetNumFiles() int64 {
	return k.numFiles
}

// TODO: if we process a zillion files, the KnownFileSet will get full of single-file entries.
// Use a bloom filter to provide some pre-filtering?
func (k *KnownFileSet) AddAll(db *filedb.FileDB, path string) {
	db.ProcessAllFileEntries(k.Add, path)
}

func (k *KnownFileSet) Add(e filedb.FileEntry) {
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
		k.numFiles ++
	}
}

func (k *KnownFileSet) GetDuplicateKeys() []KnownFileKey {
	keys := make([]KnownFileKey, 0)

	for x, val := range k.mp {
		if len(val.entries) > 1 {
			keys = append(keys, x)
		}
	}
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
