package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

// disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

type FileDir struct {
	File     string
	FileSize int64
	Dir      string
	DirSize  int64
}

func (f *FileDir) PercentUsage() (float64, error) {
	if f.FileSize == 0 || f.DirSize == 0 {
		return float64(-1), fmt.Errorf("Neither file or dir can be zero: Filesize: %d, Dirsize: %d", f.FileSize, f.DirSize)
	}
	return float64(f.FileSize) * 100 / float64(f.DirSize), nil
}

func FindLargeFiles(dirPath string, filterRegex *regexp.Regexp, threshold int64) ([]FileDir, error) {
	var files = []FileDir{}
	err := filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if filterRegex != nil {
				// match only on filename
				basePath := path.Base(filePath)
				if !filterRegex.MatchString(basePath) {
					return nil
				}
			}
			if info.Size() > threshold {
				dir := filepath.Dir(filePath)
				dirSize, _ := DirSize(dir)
				files = append(files, FileDir{
					File:     filePath,
					FileSize: info.Size(),
					Dir:      filepath.Dir(filePath),
					DirSize:  dirSize,
				})
			}
		}
		return err
	})
	return files, err
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func PrintDiskUsage(dsk string) {
	disk := DiskUsage(dsk)
	fmt.Printf("%s All: %.2f GB Used: %.2f GB Free: %.2f GB\n", dsk, float64(disk.All)/float64(GB), float64(disk.Used)/float64(GB), float64(disk.Free)/float64(GB))
}

func GetDisks() ([]string, error) {
	if file, err := os.Open("/proc/mounts"); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		disks := []string{}
		for scanner.Scan() {
			//log.Println(scanner.Text())
			line := scanner.Text()
			items := strings.Split(line, " ")
			if strings.HasPrefix(items[0], "/dev") {
				disks = append(disks, items[1])
			}
		}

		// check for errors
		if err = scanner.Err(); err != nil {
			return nil, err
		}
		return disks, nil
	} else {

		return nil, err
	}
}

func parseThreshold(s string) (int64, error) {
	l := strings.ToUpper(s)
	unit := string(l[len(l)-1])
	if unit == `B` {
		l = l[:len(l)-1]
	}
	val := l[:len(l)-1]
	var mult int64
	switch string(unit) {
	case "M":
		mult = MB
	case "K":
		mult = KB
	case "G":
		mult = GB
	default:
		return -1, fmt.Errorf("Unknown unit in %s", s)
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		return -1, fmt.Errorf("Failed to parse %s", val)
	}
	return int64(num) * mult, nil
}

func main() {
	thresholdStr := flag.String("t", "100M", "Threshold to look files for")
	filter := flag.String("f", "", "Filter regex for files")
	flag.Parse()

	var filterRegex *regexp.Regexp

	threshold, err := parseThreshold(*thresholdStr)
	if err != nil {
		log.Fatalf("Unable to parse the string %s, it needs to be <num>(Mm|Kk|Gg)", *thresholdStr)
	}

	if *filter != "" {
		filterRegex = regexp.MustCompilePOSIX(*filter)
	}

	args := flag.Args()
	if len(args) == 0 {
		disks, err := GetDisks()
		if err != nil {
			log.Fatalf("Failed to get the disks. %s", err)
		}
		for _, disk := range disks {
			PrintDiskUsage(disk)
		}
	} else {
		for _, path := range args {
			dirSize, err := DirSize(path)
			if err != nil {
				log.Fatalf("Failed to read size for %s. %s", path, err)
			}
			fmt.Printf("path: %s, Size: %.2f MB\n", path, float64(dirSize)/float64(MB))

			largeFiles, err := FindLargeFiles(path, filterRegex, threshold)
			if err != nil {
				log.Fatalf("Failed to search for large files for %s. %s", path, err)
			}

			for _, largeFile := range largeFiles {
				filePercent, err := largeFile.PercentUsage()
				if err != nil {
					log.Fatalf("Failed to get file usage %s", largeFile.File)
				}
				fmt.Printf("%s in %s, Size: %.2f MB => %.2f%%\n", largeFile.File, largeFile.Dir, float64(largeFile.FileSize)/float64(MB), filePercent)
			}
		}
	}
}
