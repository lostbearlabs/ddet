package scanner 

import (
	"com.lostbearlabs/ddet/filedb"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("scanner")

type Scanner struct {
	Db           *filedb.FileDB
	wg           *sync.WaitGroup
	filesFound   uint64
	filesScanned uint64
	filesUpdated uint64
	startTime    time.Time
}

func (scanner *Scanner) incFilesFound() {
	atomic.AddUint64(&scanner.filesFound, 1)
}
func (scanner *Scanner) incFilesScanned() {
	atomic.AddUint64(&scanner.filesScanned, 1)
}
func (scanner *Scanner) incFilesUpdated() {
	atomic.AddUint64(&scanner.filesUpdated, 1)
}
func (scanner *Scanner) getFilesScanned() uint64 {
	return atomic.LoadUint64(&scanner.filesScanned)
}
func (scanner *Scanner) getFilesUpdated() uint64 {
	return atomic.LoadUint64(&scanner.filesUpdated)
}
func (scanner *Scanner) getFilesFound() uint64 {
	return atomic.LoadUint64(&scanner.filesFound)
}

func (scanner *Scanner) processFile(path string) {
	defer scanner.incFilesScanned()
	defer scanner.wg.Done()

	//log.Trace("process: %s", path)
	if !scanner.isFileUnchanged(path) {
		//log.Trace("unchanged: %s", path)
		// process the file
		length, lastMod, _ := GetFileStats(path)
		md5, _ := ComputeMd5(path)
		item := filedb.FileEntry{path, length, lastMod, hex.EncodeToString(md5), time.Now().Unix()}
		scanner.Db.StoreFileEntry(item)
		scanner.incFilesUpdated()
	}
	//log.Trace("processed: %s", path)

}

func (scanner *Scanner) isFileUnchanged(path string) bool {

	prev := scanner.Db.ReadFileEntry(path)

	if prev == nil {
		return false
	}

	length, lastMod, _ := GetFileStats(path)

	rc := prev.Length == length && prev.LastMod == lastMod
	return rc
}

func (scanner *Scanner) visit(path string, f os.FileInfo, err error) error {
	if f != nil && !f.IsDir() {
		//log.Trace("visited: %s", path)

		scanner.wg.Add(1)
		scanner.incFilesFound()
		go scanner.processFile(path)
	}
	return nil
}

func (scanner *Scanner) ScanFiles(dir string) {
	scanTime := time.Now().Unix()

	// parallel scanning
	filepath.Walk(dir, scanner.visit)
	//log.Trace("all visited")

	// wait until all visited files are processed
	scanner.wg.Wait()
	//log.Trace("all processed")

	scanner.Db.DeleteOldEntries(dir, scanTime)
}

func (scanner *Scanner) PrintSummary(final bool) {
	elapsed := time.Since(scanner.startTime)
	if final {
		logger.Infof("found %v files, %v have changed, elapsed=%v\n", scanner.getFilesFound(), scanner.getFilesUpdated(), elapsed)
	} else {
		logger.Infof("... found %v files, processed %v, elapsed=%v\n", scanner.getFilesFound(), scanner.getFilesScanned(), elapsed)
	}
}

func MakeScanner(db *filedb.FileDB) Scanner {
	wg := new(sync.WaitGroup)
	return Scanner{db, wg, 0, 0, 0, time.Now()}
}
