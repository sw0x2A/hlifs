package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"syscall"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// FileData type stores information about path, stats and hashsum of a file
type FileData struct {
	name  string // Full path
	hash  []byte // SHA256 checksum
	dev   uint64 // Device ID
	inode uint64 // Inode number
	nlink uint64 // Number of hard links
	mode  uint32 // File permissions
	uid   uint32 // User ID
	gid   uint32 // Group ID
	size  int64  // File size in bytes
}

type FileGroup []*FileData

type OrderedFileGroup map[string]FileGroup // map[string][]*FileData

var fdb FileGroup

func (fg FileGroup) Len() int {
	return len(fg)
}

func (fg FileGroup) Less(i, j int) bool {
	if fg[i].size < fg[j].size {
		return true
	}
	return false
}

func (fg FileGroup) Swap(i, j int) {
	fg[i], fg[j] = fg[j], fg[i]
}

func (fg FileGroup) OrderBySize() map[int64][]*FileData {
	smap := make(map[int64][]*FileData)
	for _, file := range fg {
		smap[file.size] = append(smap[file.size], file)
	}
	return smap
}

// Returns hash sum of a file
func getFileHash(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

func compareFileData(a, b FileData) bool {
	if &a == &b {
		return true
	}
	if a.dev != b.dev || a.uid != b.uid || a.gid != b.gid || a.mode != b.mode {
		return false
	}
	return true
}

func walker(path string, f os.FileInfo, err error) error {
	if !f.IsDir() {
		fd := FileData{
			name:  path,
			dev:   f.Sys().(*syscall.Stat_t).Dev,
			inode: f.Sys().(*syscall.Stat_t).Ino,
			nlink: f.Sys().(*syscall.Stat_t).Nlink,
			mode:  f.Sys().(*syscall.Stat_t).Mode,
			uid:   f.Sys().(*syscall.Stat_t).Uid,
			gid:   f.Sys().(*syscall.Stat_t).Gid,
			size:  f.Size(),
		}
		fdb = append(fdb, &fd)
	}
	return err
}

// Returns random string with length between min and max
func getRandStringBytes(min, max int) (string, error) {
	if min <= 0 {
		return "", fmt.Errorf("min (%d) <= 0", min)
	}
	if max < min {
		return "", fmt.Errorf("max (%d) < min (%d)", max, min)
	}
	b := make([]byte, rand.Intn(max-min)+min)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b), nil
}

func main() {
	flag.Usage = func() {
		log.Printf("Usage: %s [options] DIR\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	log.SetFlags(0)
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	dir := flag.Arg(0)
	src, err := os.Stat(dir)
	if err != nil {
		log.Fatal(err)
	}

	if !src.IsDir() {
		log.Fatal("Source is not a directory")
	}

	err = filepath.Walk(dir, walker)
	if err != nil {
		log.Fatal(err)
	}

	// Group by file size
	smap := make(map[int64][]*FileData)
	for _, file := range fdb {
		smap[file.size] = append(smap[file.size], file)
	}

	// If more than one file with same size:
	// create sha256sum of all files with same size
	// compare and hardlink if possible
	for _, sfile := range smap {
		if len(sfile) > 1 {
			hmap := make(map[string][]*FileData)
			for _, hfile := range sfile {
				h, err := getFileHash(hfile.name)
				hstring := hex.EncodeToString(h)
				if err == nil {
					hmap[hstring] = append(hmap[hstring], hfile)
				}
			}

			for _, cfile := range hmap {
				if len(cfile) > 1 {
					firstFile := cfile[0]
					cfile := append(cfile[:0], cfile[1:]...)
					for _, file := range cfile {
						// If not already hardlink of first file...
						if firstFile.dev == file.dev && firstFile.inode != file.inode {
							// Make sure new file does not exist
							suffix, _ := getRandStringBytes(8, 16)
							for _, err = os.Stat(file.name + suffix); ; os.IsExist(err) {
								suffix, _ = getRandStringBytes(8, 16)
							}
							os.Rename(file.name, file.name+suffix)
							err = os.Link(firstFile.name, file.name)
							if err != nil {
								os.Rename(file.name+suffix, file.name)
								log.Fatal(err)
							}
							os.Remove(file.name + suffix)
						}
					}
				}
			}
		}
	}
}
