package grasure

import (
	"bufio"
	"io"
	"os"

	"github.com/DurantVivado/GrasureOL/xlog"
)

type DiskInfo struct {
	path      string
	available bool

	//the capacity of a disk (in Bytes)
	total uint64
	used  uint64
	free  uint64
}

//NewDiskInfo news a disk with basic information
func NewDiskInfo(path string) *DiskInfo {
	return &DiskInfo{
		path:      path,
		available: true,
	}
}

type DiskState int

const (
	Normal DiskState = iota
	Fail
	BitRot
)

//DiskArray contains the low-level disk information
type DiskArray struct {
	//diskFilePath is a file containing all disk's path
	diskFilePath string

	//the disk info array
	diskInfos []*DiskInfo

	//it's flag and when disk fails, it renders false.
	state DiskState

	//the capacity of a disk (in Bytes)
	total uint64
	used  uint64
	free  uint64
}

func NewDiskArray(diskFilePath string) *DiskArray {

	diskArray := &DiskArray{
		diskFilePath: diskFilePath,
		state:        Normal,
		total:        0,
		used:         0,
		free:         0,
	}
	diskArray.ReadDiskPath()
	return diskArray
}

//ReadDiskPath reads the disk paths from diskFilePath.
//There should be exactly ONE disk path at each line.
//
//This func can NOT be called concurrently.
func (d *DiskArray) ReadDiskPath() {
	f, err := os.Open(d.diskFilePath)
	if err != nil {
		xlog.Fatal(err)
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			xlog.Fatal(err)
		}
		path := string(line)
		if ok, err := pathExist(path); !ok && err == nil {
			xlog.Fatal(errDiskNotFound)
		} else if err != nil {
			xlog.Fatal(err)
		}
		total, free := DiskUsage(path)
		diskInfo := NewDiskInfo(path)
		diskInfo.total = total
		diskInfo.free = free
		diskInfo.used = total - free
		d.diskInfos = append(d.diskInfos, diskInfo)
		d.total += diskInfo.total
		d.free += diskInfo.free
		d.used += diskInfo.used
	}
}
