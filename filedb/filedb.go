package filedb

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

type FileDB struct {
	db *sql.DB
	mx *sync.Mutex
}

func considerPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func InitDB(filepath string) *FileDB {
	db, err := sql.Open("sqlite3", filepath)
	considerPanic(err)

	if db == nil {
		panic("db nil")
	}
	createTableIfNotExists(db)

	mx := new(sync.Mutex)

	filedb := FileDB{db, mx}
	return &filedb
}

func (filedb *FileDB) Close() {
	filedb.db.Close()
}

func createTableIfNotExists(db *sql.DB) {
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
	considerPanic(err)
}

func (filedb *FileDB) StoreFileEntry(item FileEntry) {
	filedb.StoreFileEntries([]FileEntry{item})
}

func (filedb *FileDB) StoreFileEntries(items []FileEntry) {
	filedb.mx.Lock()
	defer filedb.mx.Unlock()

	sql_additem := `
	INSERT OR REPLACE INTO files(
		Path,
		Length,
		LastMod,
		Md5,
		ScanTime
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := filedb.db.Prepare(sql_additem)
	considerPanic(err)
	defer stmt.Close()

	for _, item := range items {
		_, err2 := stmt.Exec(item.Path, item.Length, item.LastMod, item.Md5, item.ScanTime)
		considerPanic(err2)
	}
}

func (filedb *FileDB) ReadAllFileEntries() []FileEntry {
	filedb.mx.Lock()
	defer filedb.mx.Unlock()

	sql_readall := `
	SELECT Path, Length, LastMod, Md5, ScanTime 
	FROM files 
	ORDER BY Path
	`

	rows, err := filedb.db.Query(sql_readall)
	considerPanic(err)
	defer rows.Close()

	var result []FileEntry
	for rows.Next() {
		item := FileEntry{}
		err2 := rows.Scan(&item.Path, &item.Length, &item.LastMod, &item.Md5, &item.ScanTime)
		considerPanic(err2)
		result = append(result, item)
	}
	return result
}

func (filedb *FileDB) ReadFileEntry(path string) *FileEntry {
	filedb.mx.Lock()
	defer filedb.mx.Unlock()

	sql_read := `
	SELECT Path, Length, LastMod, Md5, ScanTime 
	FROM files 
	WHERE Path=?
	`

	item := new(FileEntry)
	err := filedb.db.QueryRow(sql_read, path).Scan(&item.Path, &item.Length, &item.LastMod, &item.Md5, &item.ScanTime)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		panic(err)
	default:
		return item
	}
}
