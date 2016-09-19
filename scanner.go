package ddet

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Scanner struct {
	Db      *sql.DB
	wg      *sync.WaitGroup
	results chan TestItem
}

func (scanner Scanner) saveResults() {
	for {
		// store file data in DB
		item := <-scanner.results
		fmt.Printf("stored: %v\n", item)
		StoreItem(scanner.Db, []TestItem{item})
		scanner.wg.Done()
	}
}

// TODO: check size and mod;  if they have not changed, then skip the MD5 and store
func (scanner Scanner) processFile(path string) {
	// process the file
	md5, _ := ComputeMd5(path)
	size, mod, _ := GetFileStats(path)
	item := TestItem{path, size, mod, hex.EncodeToString(md5)}
	scanner.results <- item
}

func (scanner Scanner) visit(path string, f os.FileInfo, err error) error {
	if !f.IsDir() {
		fmt.Printf("visited: %s\n", path)

		scanner.wg.Add(1)
		go scanner.processFile(path)
	}
	return nil
}

func (scanner Scanner) ScanFiles(dir string) {
	// goroutine to save results
	go scanner.saveResults()

	// parallel scanning
	filepath.Walk(dir, scanner.visit)
	fmt.Printf("all visited\n")

	// wait until all scanned files are stored
	scanner.wg.Wait()
	fmt.Printf("all stored\n")
}

func MakeScanner(db *sql.DB) Scanner {
	wg := new(sync.WaitGroup)
	return Scanner{db, wg, make(chan TestItem)}
}
