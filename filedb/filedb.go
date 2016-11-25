package filedb

import (
	"database/sql"
	"errors"
	"github.com/juju/loggo"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"sync"
)

var logger = loggo.GetLogger("scanner")

// This is the database we use to store FileEntry values during
// and between runs.  We use a persistent database in order to avoid
// re-hashing files on every run and also to allow for more files than
// we could handle in memory alone.
//
// A FileDB should be acquired via either InitDB() or NewTempDB() and
// *must* be closed by calling Close().
type FileDB struct {
	db       *sql.DB
	mx       *sync.Mutex
	tempDir  string
	tempFile string
}

func considerPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func InitDB(filepath string) (*FileDB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, errors.New("DB nil")
	}

	err = createTableIfNotExists(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	mx := new(sync.Mutex)

	filedb := FileDB{db, mx, "", ""}
	return &filedb, nil
}

func NewTempDB() (*FileDB, error) {
	dbdir, _ := ioutil.TempDir(os.TempDir(), "db")
	dbpath := dbdir + "/foo.db"

	f, err := InitDB(dbpath)
	if err != nil {
		return nil, err
	}
	f.tempDir = dbdir
	f.tempFile = dbpath
	return f, nil
}

func (filedb *FileDB) Close() {
	filedb.db.Close()

	if filedb.tempFile != "" {
		err := os.Remove(filedb.tempFile)
		if err != nil {
			logger.Errorf("Unable to clean up temp file %s, error is %v", filedb.tempFile, err)
		}
	}
	if filedb.tempDir != "" {
		err := os.Remove(filedb.tempDir)
		if err != nil {
			logger.Errorf("Unable to clean up temp folder %s, error is %v", filedb.tempDir, err)
		}
	}
}

func createTableIfNotExists(db *sql.DB) error {
	sql_table := `
	CREATE TABLE IF NOT EXISTS files(
		Path TEXT NOT NULL PRIMARY KEY,
		Length INT NOT NULL,
		LastMod INT NOT NULL,
		Md5 TEXT NOT NULL,
		ScanTime INT NOT NULL 
	);
	CREATE INDEX IF NOT EXISTS idx_md5
		ON files (Md5);
	`

	_, err := db.Exec(sql_table)
	return err
}

func (filedb *FileDB) StoreFileEntry(item FileEntry) error {
	return filedb.StoreFileEntries([]*FileEntry{&item})
}

func (filedb *FileDB) StoreFileEntries(items []*FileEntry) error {
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
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err := stmt.Exec(item.Path, item.Length, item.LastMod, item.Md5, item.ScanTime)
		if err != nil {
			return err
		}
	}

	return nil
}

func (filedb *FileDB) ProcessAllFileEntries(fn func(FileEntry), path string) error {
	filedb.mx.Lock()
	defer filedb.mx.Unlock()

	sql_readall := `
	SELECT Path, Length, LastMod, Md5, ScanTime 
	FROM files 
	WHERE Path LIKE ?
	ORDER BY Path
	`

	stmt, err := filedb.db.Prepare(sql_readall)
	if err != nil {
		return err
	}
	defer stmt.Close()

	//fmt.Printf("path is: %s\n", path)
	rows, err := stmt.Query(path + "%")
	considerPanic(err)
	defer rows.Close()

	for rows.Next() {
		item := NewBlankFileEntry()
		err := rows.Scan(&item.Path, &item.Length, &item.LastMod, &item.Md5, &item.ScanTime)
		if err != nil {
			return err
		}
		fn(*item)
	}
	return nil
}

func (filedb *FileDB) ReadAllFileEntries() ([]FileEntry, error) {
	var result []FileEntry
	appendFn := func(e FileEntry) {
		result = append(result, e)
	}
	err := filedb.ProcessAllFileEntries(appendFn, "/")
	return result, err
}

func (filedb *FileDB) ReadFileEntriesByKnownFileKey(md5 string, length int64) []FileEntry {
	var result []FileEntry
	filedb.mx.Lock()
	defer filedb.mx.Unlock()

	sql_readall := `
	SELECT Path, Length, LastMod, Md5, ScanTime 
	FROM files 
	WHERE MD5=? and Length=?
	ORDER BY Path
	`

	stmt, err := filedb.db.Prepare(sql_readall)
	considerPanic(err)
	defer stmt.Close()

	rows, err := stmt.Query(md5, length)
	considerPanic(err)
	defer rows.Close()

	for rows.Next() {
		item := NewBlankFileEntry()
		err2 := rows.Scan(&item.Path, &item.Length, &item.LastMod, &item.Md5, &item.ScanTime)
		//fmt.Printf("row: %v\n", item)
		considerPanic(err2)
		result = append(result, *item)
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

func (filedb *FileDB) DeleteOldEntries(path string, cutoff int64) (uint64, error) {
	filedb.mx.Lock()
	defer filedb.mx.Unlock()

	sql_delete := `
	DELETE
	FROM files
	WHERE ScanTime < ?
	AND Path LIKE ?
	`

	stmt, err := filedb.db.Prepare(sql_delete)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(cutoff, path+"%")
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return uint64(rows), nil
}
