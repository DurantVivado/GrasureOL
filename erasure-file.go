package grasure

type Version string

//fileInfo defines the file-level information,
//it's concurrently safe
type fileInfo struct {
	//file name
	BaseName string `json:"baseName"`

	//filepath is path in the dfs
	FilePath string `json:"fileName"`

	//file size
	FileSize int64 `json:"fileSize"`

	//version is file's version
	version Version

	//prev is the last version of file
	prev *fileInfo
	//next is the next version of file
	next *fileInfo

	//hash value (SHA256 by default)
	Hash string `json:"fileHash"`

	//distribution forms a block->node mapping
	Distribution [][]int `json:"fileDist"`

	//blockToOffset has the same size as Distribution but points to the block offset relative to a disk.
	blockToOffset [][]int

	//block state, default to blkOK otherwise blkFail in case of bit-rot.
	blockInfos [][]*blockInfo

	//system-level file info
	// metaInfo     *os.fileInfo

}

type blockStat uint8
type blockInfo struct {
	stat         blockStat
	filename     string
	actualOffset int64
}

//constant variables
const (
	blkOK   blockStat = 0
	blkFail blockStat = 1
)
