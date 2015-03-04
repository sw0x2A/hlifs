package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/codegangsta/cli"
)

type FileData struct {
	name  string // Full path
	hash  []byte // MD5 checksum
	dev   uint64 // Device ID
	inode uint64 // Inode number
	nlink uint64 // Number of hard links
	mode  uint32 // File permissions
	uid   uint32 // User ID
	gid   uint32 // Group ID
	size  int64  // File size in bytes
}

var fdb []*FileData

func HashFile(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

func CompareFileData(a, b FileData) bool {
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

func main() {
	app := cli.NewApp()
	app.Name = "hlifs"
	app.Usage = "Hard link identical files"

	app.Action = func(c *cli.Context) {
		if len(c.Args()) > 0 {
			dir := c.Args()[0]
			src, err := os.Stat(dir)
			if err != nil {
				panic(err)
			}

			if !src.IsDir() {
				fmt.Println("Source is not a directory")
				os.Exit(1)
			}

			err = filepath.Walk(dir, walker)
			if err != nil {
				panic(err)
			}

			// Group by file size
			smap := make(map[int64][]*FileData)
			for _, file := range fdb {
				smap[file.size] = append(smap[file.size], file)
			}

			// If more than one file with same size:
			// create md5sum of all files with same size
			// compare and hardlink if possible
			for _, sfile := range smap {
				if len(sfile) > 1 {
					hmap := make(map[string][]*FileData)
					for _, hfile := range sfile {
						h, err := HashFile(hfile.name)
						hstring := hex.EncodeToString(h)
						if err == nil {
							hmap[hstring] = append(hmap[hstring], hfile)
						}
					}

					for _, cfile := range hmap {
						if len(cfile) > 1 {
							fmt.Printf("Files with hash: %s\n", cfile)
						}
					}
				}
			}

		}
	}

	app.Run(os.Args)
}
