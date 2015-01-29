package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"
	"path/filepath"
)

type FileData struct {
		name 	string
		hash 	[]byte
        dev     uint64
        inode   uint64
        nlink   uint64
        mode    uint32
        uid     uint32
        gid     uint32
        size 	int64
}

var fdb []FileData

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

/*
   stat := fi.Sys().(*syscall.Stat_t)
   type Stat_t struct {
        Dev       uint64
        Ino       uint64
        Nlink     uint64
        Mode      uint32
        Uid       uint32
        Gid       uint32
        X__pad0   int32
        Rdev      uint64
        Size      int64
        Blksize   int64
        Blocks    int64
        Atim      Timespec
        Mtim      Timespec
        Ctim      Timespec
        X__unused [3]int64
}		 * Name() string       // base name of the file
         * Size() int64        // length in bytes for regular files; system-dependent for others
         * Mode() FileMode     // file mode bits
         * ModTime() time.Time // modification time
         * IsDir() bool        // abbreviation for Mode().IsDir()
         * Sys() interface{}   // underlying data source (can return nil)
*/


 func walker(path string, f os.FileInfo, err error) error {
   if !f.IsDir() {
	   fd := FileData{
			name:	path,
			dev:	f.Sys().(*syscall.Stat_t).Dev,
			inode:	f.Sys().(*syscall.Stat_t).Ino,
			nlink:	f.Sys().(*syscall.Stat_t).Nlink,
			mode:	f.Sys().(*syscall.Stat_t).Mode,
			uid:	f.Sys().(*syscall.Stat_t).Uid,
			gid:	f.Sys().(*syscall.Stat_t).Gid,
			size:	f.Size(),
			}
	   fdb = append(fdb, fd)
   }
   return err
 }
 
func main() {
	flag.Parse()
	dir := flag.Arg(0)
	
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

	for _, file := range fdb {
		fmt.Printf("%s with %d bytes\n", file.name, file.size)
	}

}