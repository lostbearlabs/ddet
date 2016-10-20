package filedb

// TODO: we should use a builder to create this.

type FileEntry struct {
	Path     string
	Length   int64
	LastMod  int64
	Md5      string
	ScanTime int64
}
