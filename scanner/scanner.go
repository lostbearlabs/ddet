package scanner

import (
	"com.lostbearlabs/ddet/filedb"
	"encoding/hex"
	"github.com/juju/loggo"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var logger = loggo.GetLogger("scanner")

// Scanner walks a file tree, updating the FileDB with current information
// for each file found and collecting some statistics along the way.
type Scanner struct {
	Db    *filedb.FileDB
	wg    *sync.WaitGroup
	stats *scannerStats
}

func (scanner *Scanner) processFile(path string) {
	defer scanner.stats.incFilesScanned()
	defer scanner.wg.Done()

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
			err := scanner.Db.StoreFileEntry(*item)
			if err != nil {
				logger.Errorf("Error [%v] storing [%v]", err, item)
			}
			if prev == nil {
				scanner.stats.incFilesAdded(1)
			} else {
				scanner.stats.incFilesUpdated()
			}
		}
	} else {
		// file has not been updated ... only need to get our current
		// scan time into the database
		prev.SetScanTime(time.Now().Unix())
		err := scanner.Db.StoreFileEntry(*prev)
		if err != nil {
			logger.Errorf("Error [%v] storing [%v]", err, *prev)
		}
	}

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
		scanner.stats.incFilesFound()
		go scanner.processFile(path)
	}
	return nil
}

func (scanner *Scanner) ScanFiles(dir string) error {
	scanTime := time.Now().Unix()
	logger.Infof("Scanning folder %v", dir)

	// Walk the file tree, and kick off a separate parallel goroutine
	// to process each file that's visited.
	filepath.Walk(dir, scanner.visit)
	logger.Tracef("all visited")

	// Wait until all visited files are processed
	scanner.wg.Wait()
	logger.Tracef("all processed")

	// Clean up any old database entries that were not refreshed
	// during this scan.
	deleted, err := scanner.Db.DeleteOldEntries(dir, scanTime)
	if err != nil {
		return err
	}
	scanner.stats.incFilesDeleted(deleted)

	return nil
}

func (scanner *Scanner) PrintSummary(final bool) {
	if final {
		logger.Infof("found %v files, %v added, %v changed, %v deleted\n", scanner.stats.getFilesFound(), scanner.stats.getFilesAdded(),
			scanner.stats.getFilesUpdated(), scanner.stats.getFilesDeleted())
	} else {
		logger.Infof("... processed %v/%v files\n", scanner.stats.getFilesScanned(), scanner.stats.getFilesFound())
	}
}

func MakeScanner(db *filedb.FileDB) Scanner {
	wg := new(sync.WaitGroup)
	stats := newScannerStats()
	return Scanner{db, wg, stats}
}
