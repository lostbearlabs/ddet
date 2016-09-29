package ddet

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Scanner struct {
	Db           *sql.DB
	wg           *sync.WaitGroup
	results      chan FileEntry
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

func (scanner *Scanner) saveResults() {
	for {
		// store file data in DB
		item := <-scanner.results
		//fmt.Printf("stored: %v\n", item)
		StoreFileEntry(scanner.Db, []FileEntry{item})
		scanner.wg.Done()
	}
}

func (scanner *Scanner) processFile(path string) {
	defer scanner.incFilesScanned()

	//fmt.Printf("scanning: %v\n", path)
	if scanner.isFileUnchanged(path) {
		scanner.wg.Done()
	} else {
		// process the file
		length, lastMod, _ := GetFileStats(path)
		md5, _ := ComputeMd5(path)
		item := FileEntry{path, length, lastMod, hex.EncodeToString(md5), time.Now().Unix()}
		scanner.results <- item
		scanner.incFilesUpdated()
	}
}

func (scanner *Scanner) isFileUnchanged(path string) bool {

	prev := ReadFileEntry(scanner.Db, path)

	if prev == nil {
		return false
	}

	length, lastMod, _ := GetFileStats(path)

	rc := prev.Length == length && prev.LastMod == lastMod
	//	if !rc {
	//		fmt.Printf("changed: %s\n", path)
	//	}
	return rc
}

func (scanner *Scanner) visit(path string, f os.FileInfo, err error) error {
	if f == nil {
		return nil
	}

	if !f.IsDir() {
		//fmt.Printf("visited: %s\n", path)

		scanner.wg.Add(1)
		scanner.incFilesFound()
		go scanner.processFile(path)
	}
	return nil
}

func (scanner *Scanner) ScanFiles(dir string) {
	// goroutine to save results
	go scanner.saveResults()

	// parallel scanning
	filepath.Walk(dir, scanner.visit)
	//fmt.Printf("all visited\n")

	// wait until all scanned files are stored
	scanner.wg.Wait()
	//fmt.Printf("all stored\n")
}

func (scanner *Scanner) PrintSummary(final bool) {
	if final {
		fmt.Printf("found %v files, %v have changed\n", scanner.getFilesFound(), scanner.getFilesUpdated())
	} else {
		elapsed := time.Now().Sub(scanner.startTime)
		fmt.Printf("... t=%v, found %v files, processed %v\n", int64(elapsed/1000000000), scanner.getFilesFound(), scanner.getFilesScanned())
	}
}

func MakeScanner(db *sql.DB) Scanner {
	wg := new(sync.WaitGroup)
	return Scanner{db, wg, make(chan FileEntry), 0, 0, 0, time.Now()}
}
