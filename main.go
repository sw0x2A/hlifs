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

//var ofg = make(map[string][]*FileData)
var ofg = make(map[string]FileGroup)

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

func (fg FileGroup) CalcAllHashes() {
	for _, f := range fg {
		h, err := getFileHash(f.name)
		if err != nil {
			log.Fatal(err)
		}
		f.hash = h
	}
}

func (fg FileGroup) IndexesPerSameHash() map[string][]int {
	imap := make(map[string][]int)
	for i, f := range fg {
		fh := hex.EncodeToString(f.hash)
		imap[fh] = append(imap[fh], i)
	}
	return imap
}

func (fg FileGroup) HardLink(i, j int) error {
	// Make sure new file does not exist
	suffix, _ := getRandStringBytes(8, 16)
	// FIXME: Endless loop
	// for _, err := os.Stat(fg[j].name + suffix); ; os.IsExist(err) {
	// 	fmt.Println(fg[j].name + suffix)
	// 	suffix, _ = getRandStringBytes(8, 16)
	// }
	os.Rename(fg[j].name, fg[j].name+suffix)
	err := os.Link(fg[i].name, fg[j].name)
	if err != nil {
		os.Rename(fg[j].name+suffix, fg[j].name)
		return err
	}
	os.Remove(fg[j].name + suffix)
	return nil
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
	knownInodes := make(map[string]bool)
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
		// Only add new files if we haven't seen their dev/inode combination yet.
		// They are already hard-links
		ikey := fmt.Sprintf("%x_%x", fd.inode, fd.dev)
		if !knownInodes[ikey] {
			knownInodes[ikey] = true
			key := fmt.Sprintf("%x_%x_%x_%x_%x", fd.dev, fd.mode, fd.uid, fd.gid, fd.size)
			ofg[key] = append(ofg[key], &fd)
		}
	}
	// Remove all keys that have only one file (nothing to hard-link)
	for k, fg := range ofg {
		if fg.Len() < 2 {
			delete(ofg, k)
		}
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

	// ofg has been grouped by dev, mode, uid, gid and size by walker func
	for key, fg := range ofg {
		fmt.Println(key, fg)
		if fg.Len() > 1 {
			fg.CalcAllHashes()
			for _, hlable := range fg.IndexesPerSameHash() {
				fi := hlable[0]
				test := append(hlable[:0], hlable[1:]...)
				for _, i := range test {
					fg.HardLink(fi, i)
				}
			}
		}
	}
}
