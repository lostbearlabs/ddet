package scanner 

import (
	"crypto/md5"
	"io"
	"os"
)

// from: http://dev.pawelsz.eu/2014/11/google-golang-compute-md5-of-file.html
func ComputeMd5(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

func GetFileStats(filePath string) (int64, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 0, 0, err
	}

	return stat.Size(), stat.ModTime().Unix(), nil
}
