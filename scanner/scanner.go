package scanner

import (
	"com.lostbearlabs/ddet/filedb"
	"encoding/hex"
	"github.com/juju/loggo"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

var logger = loggo.GetLogger("scanner")

type Scanner struct {
	Db           *filedb.FileDB
	wg           *sync.WaitGroup
	filesFound   uint64
	filesScanned uint64
	filesUpdated uint64
	filesDeleted uint64
	filesAdded   uint64
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
func (scanner *Scanner) incFilesDeleted(num uint64) {
	atomic.AddUint64(&scanner.filesDeleted, num)
}
func (scanner *Scanner) incFilesAdded(num uint64) {
	atomic.AddUint64(&scanner.filesAdded, num)
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
func (scanner *Scanner) getFilesDeleted() uint64 {
	return atomic.LoadUint64(&scanner.filesDeleted)
}
func (scanner *Scanner) getFilesAdded() uint64 {
	return atomic.LoadUint64(&scanner.filesAdded)
}

func (scanner *Scanner) processFile(path string) {
	defer scanner.incFilesScanned()
	defer scanner.wg.Done()

	//log.Trace("process: %s", path)
	length, lastMod, _ := GetFileStats(path)
	if length == 0 {
		return
	}
	changed, prev := scanner.isFileChanged(path, length, lastMod)

	if changed {
		// file has been added or updated ... recompute its MD5
		logger.Tracef(" ... changed since last scan: %s", path)
		md5, _ := ComputeMd5(path)
		if md5 == nil || len(md5) != 16 {
			logger.Warningf("unable to read file %s", path)
			return
		} else {
			item := filedb.NewBlankFileEntry().
				SetPath(path).
				SetLength(length).
				SetLastMod(lastMod).
				SetMd5(hex.EncodeToString(md5)).
				SetScanTime(time.Now().Unix())
			scanner.Db.StoreFileEntry(*item)
			if prev == nil {
				scanner.incFilesAdded(1)
			} else {
				scanner.incFilesUpdated()
			}
		}
	} else {
		// file has not been updated ... only need to get our current
		// scan time into the database
		prev.SetScanTime(time.Now().Unix())
		scanner.Db.StoreFileEntry(*prev)
	}
	//log.Trace("processed: %s", path)

}

func (scanner *Scanner) isFileChanged(path string, length int64, lastMod int64) (bool, *filedb.FileEntry) {

	prev := scanner.Db.ReadFileEntry(path)

	if prev == nil {
		return true, nil
	}

	rc := prev.Length != length || prev.LastMod != lastMod
	return rc, prev
}

func isRegularFile(f os.FileInfo) bool {
	if f == nil {
		return false
	}

	return (f.Mode() & os.ModeType) == 0
}

func (scanner *Scanner) visit(path string, f os.FileInfo, err error) error {
	if isRegularFile(f) {
		//log.Trace("visited: %s", path)

		scanner.wg.Add(1)
		scanner.incFilesFound()
		go scanner.processFile(path)
	}
	return nil
}

func (scanner *Scanner) ScanFiles(dir string) {
	scanTime := time.Now().Unix()
	logger.Infof("Scanning folder %v", dir)

	// parallel scanning
	filepath.Walk(dir, scanner.visit)
	logger.Tracef("all visited")

	// wait until all visited files are processed
	scanner.wg.Wait()
	logger.Tracef("all processed")

	deleted := scanner.Db.DeleteOldEntries(dir, scanTime)
	scanner.incFilesDeleted(deleted)
}

func (scanner *Scanner) PrintSummary(final bool) {
	if final {
		logger.Infof("found %v files, %v added, %v changed, %v deleted\n", scanner.getFilesFound(), scanner.getFilesAdded(),
			scanner.getFilesUpdated(), scanner.getFilesDeleted())
	} else {
		logger.Infof("... processed %v/%v files\n", scanner.getFilesScanned(), scanner.getFilesFound())
	}
}

func MakeScanner(db *filedb.FileDB) Scanner {
	wg := new(sync.WaitGroup)
	return Scanner{db, wg, 0, 0, 0, 0, 0}
}
