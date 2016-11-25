package scanner

import (
	"sync/atomic"
)

// The scannerStats type holds counts of files processed as the scanner runs.
// All the accessor methods are thread-safe.
type scannerStats struct {
	filesFound   uint64
	filesScanned uint64
	filesUpdated uint64
	filesDeleted uint64
	filesAdded   uint64
}

func newScannerStats() *scannerStats {
	return &scannerStats{}
}

func (stats *scannerStats) incFilesFound() {
	atomic.AddUint64(&stats.filesFound, 1)
}
func (stats *scannerStats) incFilesScanned() {
	atomic.AddUint64(&stats.filesScanned, 1)
}
func (stats *scannerStats) incFilesUpdated() {
	atomic.AddUint64(&stats.filesUpdated, 1)
}
func (stats *scannerStats) incFilesDeleted(num uint64) {
	atomic.AddUint64(&stats.filesDeleted, num)
}
func (stats *scannerStats) incFilesAdded(num uint64) {
	atomic.AddUint64(&stats.filesAdded, num)
}

func (stats *scannerStats) getFilesScanned() uint64 {
	return atomic.LoadUint64(&stats.filesScanned)
}
func (stats *scannerStats) getFilesUpdated() uint64 {
	return atomic.LoadUint64(&stats.filesUpdated)
}
func (stats *scannerStats) getFilesFound() uint64 {
	return atomic.LoadUint64(&stats.filesFound)
}
func (stats *scannerStats) getFilesDeleted() uint64 {
	return atomic.LoadUint64(&stats.filesDeleted)
}
func (stats *scannerStats) getFilesAdded() uint64 {
	return atomic.LoadUint64(&stats.filesAdded)
}
