package filedb

// This is the information we store for each file.
// The KnownFileSet relies on the Md5 and Length to identify
// duplicates;  the Scanner relies on the LastMod, Length, and
// ScanTime to identify which files need to be re-hashed.
type FileEntry struct {
	Path     string
	Length   int64
	LastMod  int64
	Md5      string
	ScanTime int64
}

func NewBlankFileEntry() *FileEntry {
	return &FileEntry{}
}

func NewTestFileEntry() *FileEntry {
	return &FileEntry{"a.txt", 128, 0, "8d9ace9df01c0c0876a95c3f810e7e9a", 100000}
}

func (f *FileEntry) SetPath(path string) *FileEntry {
	f.Path = path
	return f
}

func (f *FileEntry) SetLength(length int64) *FileEntry {
	f.Length = length
	return f
}

func (f *FileEntry) SetLastMod(lastMod int64) *FileEntry {
	f.LastMod = lastMod
	return f
}

func (f *FileEntry) SetMd5(md5 string) *FileEntry {
	f.Md5 = md5
	return f
}

func (f *FileEntry) SetScanTime(scanTime int64) *FileEntry {
	f.ScanTime = scanTime
	return f
}
