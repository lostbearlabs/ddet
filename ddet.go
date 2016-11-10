package main

import (
	"com.lostbearlabs/ddet/dset"
	"com.lostbearlabs/ddet/filedb"
	"com.lostbearlabs/ddet/scanner"
	"com.lostbearlabs/ddet/util"
	"fmt"
	"github.com/juju/loggo"
	"os"
	"time"
)

var logger loggo.Logger = loggo.GetLogger("ddet.main")

func main() {
	path := ""
	numPaths := 0
	verbose := false

	for n, arg := range os.Args {
		if n > 0 {
			switch arg {
			case "-v":
				verbose = true
			default:
				path = arg
				numPaths++
			}
		}
	}

	if verbose {
		util.SetLogTrace()
	} else {
		util.SetLogInfo()
	}

	if numPaths == 1 {
		doScan(path)
	} else {
		fmt.Printf("Usage:\n")
		fmt.Printf("   ddet <folder>\n")
	}
}

func doScan(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if !fi.IsDir() {
		fmt.Printf("Not a directory: %s\n", path)
		return
	}

	dbpath := "ddet.db"

	db := filedb.InitDB(dbpath)
	defer db.Close()

	scanFiles(path, db)
	analyzeDuplicates(db, path)
}

func scanFiles(path string, db *filedb.FileDB) {

	logger.Tracef("BEGIN SCAN: %s", path)
	scanner := scanner.MakeScanner(db)

	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for range ticker.C {
			scanner.PrintSummary(false)
		}
	}()

	scanner.ScanFiles(path)

	ticker.Stop()

	scanner.PrintSummary(true)
	logger.Infof("COMPLETED SCAN: %s\n", path)

}

func analyzeDuplicates(db *filedb.FileDB, path string) {

	logger.Tracef("BEGIN ANALYSIS")
	start := time.Now()
	ks := dset.New()
	ks.AddAll(db, path)

	dupKeys := ks.GetDuplicateKeys()
	logger.Infof("COMPLETED ANALYSIS, elapsed=%v\n", time.Since(start))

	if dupKeys == nil || len(dupKeys) == 0 {
		logger.Infof("NO DUPLICATES FOUND, %d files total\n", ks.GetNumFiles())
		return
	}
	logger.Infof("found %d groups of duplicate files, %d files total", len(dupKeys), ks.GetNumFiles())

	for _, key := range dupKeys {
		entries := ks.GetFileEntries(key)
		fmt.Printf("Files with MD5 %s and length %d:\n", entries[0].Md5, entries[0].Length)
		for i, entry := range entries {
			if i > 5 {
				fmt.Printf("   ... and %d more\n", len(entries)-i)
				break
			}
			fmt.Printf("   %s\n", entry.Path)
		}
	}

}
