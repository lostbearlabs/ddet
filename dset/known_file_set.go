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

// A set of files, from which we can extract any duplicate files,
// i.e. sets of multiple files with the same key.
type KnownFileSet struct {
	// total number of files processed, whether duplicated or not
	numFiles int64

	// weak filter of all MD5s
	bf1 bloom.BloomFilter
	// stronger filter of MD5s that occur more than onces in bf1
	mp2 map[string]bool
	// strongest filter, counting occurences of (MD5,Length) pairs for the
	// MD5 values present in mp2
	knownKeys map[KnownFileKey]uint64
}

func New() *KnownFileSet {
	knownKeys := make(map[KnownFileKey]uint64)
	mp2 := make(map[string]bool)
	bf, _ := bloom.New()

	return &KnownFileSet{0, bf, mp2, knownKeys}
}

func (k *KnownFileSet) GetNumFiles() int64 {
	return k.numFiles
}

// Adds all files from the database (prefixed by the specified path)
// to the KnownFileSet.
func (k *KnownFileSet) AddAll(db *filedb.FileDB, path string) {

	// Weakly identify all the MD5 values that occur more than once
	db.ProcessAllFileEntries(k.populateFilters, path)
	logger.Infof("first pass identified %d potential groups of duplicates", len(k.mp2))

	// Process the candidate MD5s and group into (MD5,Length) pairs.
	for md5, _ := range k.mp2 {
		items := db.ReadFileEntriesByMd5(md5)
		if len(items) > 1 {
			for _, item := range items {
				k.addToKnownKeys(item)
			}
		}
	}
}

func considerPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// Used to weakly identify candidates for MD5 duplication.
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

	k.numFiles++
}

// For files whose MD5 values are already suspected of being duplicated, this
// populates our main map of knownKeys with the count for each (MD5,Length) pair.
func (k *KnownFileSet) addToKnownKeys(e filedb.FileEntry) {
	if !k.mp2[e.Md5] {
		return
	}

	key := KnownFileKey{e.Md5, e.Length}
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
func (k *KnownFileSet) GetFileEntries(db *filedb.FileDB, key KnownFileKey) []filedb.FileEntry {

	ar := make([]filedb.FileEntry, 0)
	items := db.ReadFileEntriesByMd5(key.md5)
	for _, item := range items {
		if item.Length == key.length {
			ar = append(ar, item)
		}
	}

	return ar
}
