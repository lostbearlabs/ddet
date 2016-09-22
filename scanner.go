package ddet

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type Scanner struct {
	Db              *sql.DB
	wg              *sync.WaitGroup
	results         chan TestItem
	filesScanned	uint64
	filesUpdated	uint64
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

func (scanner *Scanner) saveResults() {
	for {
		// store file data in DB
		item := <-scanner.results
		fmt.Printf("stored: %v\n", item)
		StoreItem(scanner.Db, []TestItem{item})
		scanner.wg.Done()
	}
}

// TODO: check size and mod;  if they have not changed, then skip the MD5 and store
func (scanner *Scanner) processFile(path string) {
	atomic.AddUint64(&scanner.filesScanned, 1)

	prev_size := int64(0)
	prev_mod := int64(0)
	
	size, mod, _ := GetFileStats(path)
	if prev_size!=size || prev_mod!=mod {
		// process the file
		md5, _ := ComputeMd5(path)
		item := TestItem{path, size, mod, hex.EncodeToString(md5)}
		scanner.results <- item
		atomic.AddUint64(&scanner.filesUpdated,1)
	}
}
func (scanner *Scanner) visit(path string, f os.FileInfo, err error) error {
	if f==nil {
		return nil
	}
	
	if !f.IsDir() {
		fmt.Printf("visited: %s\n", path)

		scanner.wg.Add(1)
		go scanner.processFile(path)
	}
	return nil
}

func (scanner *Scanner) ScanFiles(dir string) {
	// goroutine to save results
	go scanner.saveResults()

	// parallel scanning
	filepath.Walk(dir, scanner.visit)
	fmt.Printf("all visited\n")

	// wait until all scanned files are stored
	scanner.wg.Wait()
	fmt.Printf("all stored\n")
}

func (scanner *Scanner) PrintSummary() {
	fmt.Printf("found %v files, %v modified\n", scanner.getFilesScanned(), scanner.getFilesUpdated())
}

func MakeScanner(db *sql.DB) Scanner {
	wg := new(sync.WaitGroup)
	return Scanner{db, wg, make(chan TestItem), 0, 0}
}
