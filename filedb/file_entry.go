package filedb

type FileEntry struct {
	Path     string
	Length   int64
	LastMod  int64
	Md5      string
	ScanTime int64
}
