# ddet
Duplicate file detection (in go)

## Introduction

In this project I am using the problem of duplicate file detection to explore go programming.  My goals are:
* write a non-trivial program in go
* detect and report duplicate files in a folder tree
* run quickly
* scale to large numbers of files


## Compilation

ddet is a simple go program, so it builds with the standard toolchain:

    $> cd src/com.lostbearlabs/ddet
    $> go build

To work on this project, I have used a public workspace in the Cloud9 hosted IDE:  https://c9.io/lostbearlabs/lostbearlabs-dup-detection

## Usage

Usage:

    ddet {folder} [-v]

Examples:

    $> ddet /home/eric.johnson
    $> ddet /etc -v
    
Outputs are:
* groups of duplicate files are written to stdout
* logging is written to stderr


## Design


### Scanning

The go library method "filepath.Walk" runs quickly on large file hierarchies, so we can invoke this directly on the target path.  Individual files are then processed by a goroutine for optimal parallelization.

As files are processed, they are stored to a SQLite database.  This has two advantages:

* our working set is stored on disk, rather than in memory.  This improves scalability, letting us run on larger file sets.
* our working set is persistent, which means that on subsequent runs we don't need to re-examine a file's contents if its size and modification time are unchanged.  This improves performance over multiple runs.

The SQLite database is named "~/.ddetdb" -- this file can be deleted to force all files to be re-hashed.

Files with length zero are ignored.

To deal with deleted files, we update each scanned file with a timestamp.  At the end of a scan we delete any unmarked files.

Our main performance constraint is the database -- we query (by primary key) and insert (which also updates a secondary key used later during analysis).  Per-file goroutines contend for the database, which is currently locked with a mutex;  an active task queue might be more performant.

Our second performance constraint is file I/O and MD5 calculation.

### Analysis


We consider two files to be duplicates if they have the same key, where the key is composed of:
* the file length
* the MD5 hash of the entire file contents

Analysis proceeds in three stages:
* working file-by-file we use a weak hash (a bloom filter) to identify keys that might be duplicated
* working key-by-key from the possible duplicates we confirm keys that are definitely duplicated
* working key-by-key from the definite duplicates we report file details for all the files having that key

Our main performance constraint is, again, the database.  We perform a full scan and then we repeatedly query by the (MD5+length) secondary key.


### Libraries

* the library "github.com/juju/loggo" provides logging
* the library "github.com/mattn/go-sqlite3" provides SQLite

