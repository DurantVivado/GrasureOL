package grasure

//fileInfo defines the file-level information,
//it's concurrently safe
type fileInfo struct {
	//file name
	FileName string `json:"fileName"`

	//file size
	FileSize int64 `json:"fileSize"`

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

	//loadBalancedScheme is the most load-balanced scheme derived by SGA algo
	loadBalancedScheme [][]int
}

type blockStat uint8
type blockInfo struct {
	stat blockStat
	filename string
	actualOffset int64
}

//constant variables
const (
	blkOK         blockStat = 0
	blkFail       blockStat = 1
)