package storage_node

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"regexp"
	"syscall"
	"time"
)

type FileInfo struct {
	FreeSpace            int64
	NumRequestsProcessed int32
	NewFiles             []string
	AllFiles             []string
}

func (s *StorageNode) potentiallyCorrupt(f string) bool {

	// Get file information
	fileInfo, err := os.Stat(s.dir + "/" + f)
	if err != nil {
		fmt.Println(err)

		//s.logger.Error("Error getting file information", zap.Error(err))
		return true
	}

	stat := fileInfo.Sys().(*syscall.Stat_t)
	createTime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))

	currentTime := time.Now()

	timeDiff := currentTime.Sub(createTime)

	if timeDiff > time.Minute && timeDiff < 2*time.Minute {
		s.logger.Info("File needs to be checked for corruption", zap.String("file", f))
		return true
	} else {
		//s.logger.Info("File does not need to be checked for corruption", zap.String("file", f))
		return false
	}

}

func (s *StorageNode) isNewFile(f string) bool {

	// Get file information
	fileInfo, err := os.Stat(s.dir + "/" + f)
	if err != nil {
		fmt.Println(err)

		//s.logger.Error("Error getting file information", zap.Error(err))
		return true
	}

	stat := fileInfo.Sys().(*syscall.Stat_t)
	createTime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))

	currentTime := time.Now()

	timeDiff := currentTime.Sub(createTime)

	if timeDiff < time.Minute {
		return true
	} else {
		return false
	}

}

func (s *StorageNode) GetNewFiles() (allFiles []string) {

	files, err := os.ReadDir(s.dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	var filesFound []string
	for _, file := range files {

		if !file.IsDir() {
			if s.isNewFile(file.Name()) {
				filesFound = append(filesFound, file.Name())
			}
		}
	}
	return filesFound

}

func (s *StorageNode) GetAllFiles() (allFiles []string) {

	files, err := os.ReadDir(s.dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	var filesFound []string
	for _, file := range files {

		if !file.IsDir() {
			filesFound = append(filesFound, file.Name())
		}
	}
	return filesFound

	//s.fileInfo.AllFiles = filesFound
}

func (s *StorageNode) isFileFragment(f string) bool {

	//check if the file name ends with _x, where x is a number
	//if it does, then it is a file fragment
	//if it doesn't, then it is a file

	pattern := regexp.MustCompile(`_\d+$`)

	if pattern.MatchString(f) {
		return true
	}

	return false

}

// check for free space
func (s *StorageNode) CheckFreeSpace() int64 {

	var stat syscall.Statfs_t
	err := syscall.Statfs(s.dir, &stat)
	if err != nil {
		fmt.Println("Error getting file system statistics:", err)
		return 0
	}

	freeSpace := stat.Bavail * uint64(stat.Bsize)

	//	fmt.Printf("Free space in %s: %s\n", DIR, s.byteCountSI(freeSpace))

	return int64(freeSpace)
}

func (s *StorageNode) byteCountSI(b uint64) string {

	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
