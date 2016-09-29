package ddet

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func Fnord() int {
	return 2
}

type FileEntry struct {
	Path     string
	Length   int64
	LastMod  int64
	Md5      string
	ScanTime int64
}

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	return db
}

func CreateTable(db *sql.DB) {
	// create table if not exists
	sql_table := `
	CREATE TABLE IF NOT EXISTS files(
		Path TEXT NOT NULL PRIMARY KEY,
		Length INT NOT NULL,
		LastMod INT NOT NULL,
		Md5 TEXT NOT NULL,
		ScanTime INT NOT NULL 
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func StoreFileEntry(db *sql.DB, items []FileEntry) {
	sql_additem := `
	INSERT OR REPLACE INTO files(
		Path,
		Length,
		LastMod,
		Md5,
		ScanTime
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_additem)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, item := range items {
		_, err2 := stmt.Exec(item.Path, item.Length, item.LastMod, item.Md5, item.ScanTime)
		if err2 != nil {
			panic(err2)
		}
	}
}

func ReadAllFileEntries(db *sql.DB) []FileEntry {
	sql_readall := `
	SELECT Path, Length, LastMod, Md5, ScanTime 
	FROM files 
	ORDER BY Path
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []FileEntry
	for rows.Next() {
		item := FileEntry{}
		err2 := rows.Scan(&item.Path, &item.Length, &item.LastMod, &item.Md5, &item.ScanTime)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result
}

func ReadFileEntry(db *sql.DB, path string) *FileEntry {

	sql_read := `
	SELECT Path, Length, LastMod, Md5, ScanTime 
	FROM files 
	WHERE Path=?
	ORDER BY Path
	`

	item := new(FileEntry)
	err := db.QueryRow(sql_read, path).Scan(&item.Path, &item.Length, &item.LastMod, &item.Md5, &item.ScanTime)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	default:
		return item
	}
}
