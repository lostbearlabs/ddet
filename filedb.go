package ddet

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func Fnord() int {
	return 2
}

type TestItem struct {
	Path    string
	Length  int64
	LastMod int64
	Md5     string
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
		Md5 TEXT NOT NULL
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func StoreItem(db *sql.DB, items []TestItem) {
	sql_additem := `
	INSERT OR REPLACE INTO files(
		Path,
		Length,
		LastMod,
		Md5	
	) values(?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_additem)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, item := range items {
		_, err2 := stmt.Exec(item.Path, item.Length, item.LastMod, item.Md5)
		if err2 != nil {
			panic(err2)
		}
	}
}

func ReadItem(db *sql.DB) []TestItem {
	sql_readall := `
	SELECT Path, Length, LastMod, Md5 FROM files 
	ORDER BY Path
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []TestItem
	for rows.Next() {
		item := TestItem{}
		err2 := rows.Scan(&item.Path, &item.Length, &item.LastMod, &item.Md5)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result
}
