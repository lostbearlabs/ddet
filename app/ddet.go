package main

import (
	"com.lostbearlabs/ddet"
	"com.lostbearlabs/ddet/dset"
	"com.lostbearlabs/ddet/filedb"
	"fmt"
	"os"
	"time"
)

func main() {
	argsWithProg := os.Args
	if len(argsWithProg) == 2 {
		doScan(argsWithProg[1])
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

	scanner := ddet.MakeScanner(db)

	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for range ticker.C {
			scanner.PrintSummary(false)
		}
	}()

	scanner.ScanFiles(path)

	ticker.Stop()

	scanner.PrintSummary(true)
	fmt.Printf("COMPLETED SCAN: %s\n", path)

}

func analyzeDuplicates(db *filedb.FileDB, path string) {
	
	start := time.Now()
	ks := dset.New()
	ks.AddAll(db, path)

	dupKeys := ks.GetDuplicateKeys()
	fmt.Printf("COMPLETED ANALYSIS, elapsed=%v\n", time.Since(start))
	
	if dupKeys == nil || len(dupKeys) == 0 {
		fmt.Printf("NO DUPLICATES FOUND\n")
		return
	}

	for _, key := range dupKeys {
		entries := ks.GetFileEntries(key)
		fmt.Printf("Files with MD5 %s and length %d:\n", entries[0].Md5, entries[0].Length)
		for _, entry := range entries {
			fmt.Printf("   %s\n", entry.Path)
		}
	}
	
}
