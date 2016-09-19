package main 

import( 
    "fmt"
    "com.lostbearlabs/ddet"
    "os"
)

func main() {
    argsWithProg := os.Args
    if len(argsWithProg)==3 && argsWithProg[1]=="scan" {
        doScan(argsWithProg[2])
    } else {
        fmt.Printf("Usage:\n")
        fmt.Printf("   ddet scan <folder>\n")
        fmt.Printf("   ddet analyze\n")
    }
}

// TODO ... display summary stats after scan?
// TODO ... display progress during scan?

func doScan(path string) {
    fi, err := os.Stat(path)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return;
    }
    if !fi.IsDir() {
        fmt.Printf("Not a directory: %s\n", path)
        return;
    }
    
   	dbpath := "ddet.db"

	db := ddet.InitDB(dbpath)
	defer db.Close()
	ddet.CreateTable(db)
	
	scanner := ddet.MakeScanner(db)
	scanner.ScanFiles(path)
	fmt.Printf("COMPLETED SCAN: %s\n", path)
}
