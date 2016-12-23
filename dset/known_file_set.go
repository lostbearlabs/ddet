package dset

import (
	"encoding/hex"
	"github.com/juju/loggo"
	"lostbearlabs.com/ddet/filedb"
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

// A set of files, from which we can extract any duplicate files,
// i.e. sets of multiple files with the same key.
type KnownFileSet struct {
	// total number of files processed, whether duplicated or not
	numFiles int64

	// weak filter of all keys
	wf *weakFilter
	// stronger filter of keys that occur more than once in wf
	mp2 map[KnownFileKey]bool
	// strongest filter, counting occurences of each key for each
	// key that really is duplicated
	knownKeys map[KnownFileKey]uint64
}

func New() *KnownFileSet {
	knownKeys := make(map[KnownFileKey]uint64)
	mp2 := make(map[KnownFileKey]bool)
	wf, _ := newWeakFilter()

	return &KnownFileSet{0, wf, mp2, knownKeys}
}

func (k *KnownFileSet) GetNumFiles() int64 {
	return k.numFiles
}

// Adds all files from the database (prefixed by the specified path)
// to the KnownFileSet.
func (k *KnownFileSet) AddAll(db *filedb.FileDB, path string) {

	// Weakly identify all the keys that occur more than once
	err := db.ProcessAllFileEntries(k.populateFilters, path)
	if err != nil {
		logger.Errorf("error processing file entries [%v]", err)
		return
	}
	logger.Infof("first pass identified %d potential groups of duplicates", len(k.mp2))

	// Re-process all the candidate keys to identify the ones that are really duplicated
	for key, _ := range k.mp2 {
		items, err := db.ReadFileEntriesByKnownFileKey(key.md5, key.length)
		if err != nil {
			logger.Errorf("error reading entries for key [%v]: [%v]", key, err)
			return
		}
		if len(items) > 1 {
			for _, item := range items {
				k.addToKnownKeys(item)
			}
		}
	}
}

// Used to weakly identify candidates for MD5 duplication.
func (k *KnownFileSet) populateFilters(e filedb.FileEntry) {
	md5, err := hex.DecodeString(e.Md5)
	if err != nil {
		logger.Errorf("got error [%v] decoding md5 for [%v]", err, e)
		return
	}

	known, err := k.wf.contains(md5, e.Length)
	if err != nil {
		logger.Errorf("got error [%v] from [%v] with md5 decoded as [%v]", err, e, md5)
		return
	}

	key := KnownFileKey{e.Md5, e.Length}
	if known && !k.mp2[key] {
		logger.Tracef("found interesting key: %v", key)
		k.mp2[key] = true
	} else {
		k.wf.add(md5, e.Length)
	}

	k.numFiles++
}

// For files whose keys are already suspected of being duplicated, this
// populates our main map of knownKeys with the count for each (MD5,Length) pair.
func (k *KnownFileSet) addToKnownKeys(e filedb.FileEntry) {
	key := KnownFileKey{e.Md5, e.Length}
	if !k.mp2[key] {
		return
	}

	count, _ := k.knownKeys[key]
	k.knownKeys[key] = count + 1
}

// Returns the (MD5,Length) pairs that really do correspond to duplicate files.
func (k *KnownFileSet) GetDuplicateKeys() []KnownFileKey {
	keys := make([]KnownFileKey, 0)

	for key, count := range k.knownKeys {
		if count > 1 {
			keys = append(keys, key)
		}
	}

	sort.Sort(ByLength(keys))

	logger.Tracef("duplicate keys: %v", keys)
	return keys
}

// Returns the file entries for a particular (MD5,Length) pair.
func (k *KnownFileSet) GetFileEntries(db *filedb.FileDB, key KnownFileKey) ([]filedb.FileEntry, error) {

	ar := make([]filedb.FileEntry, 0)
	items, err := db.ReadFileEntriesByKnownFileKey(key.md5, key.length)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Length == key.length {
			ar = append(ar, item)
		}
	}

	return ar, nil
}
