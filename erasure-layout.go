package grasure

type Pattern int

const (
	Random Pattern = iota
	// see http://www.accs.com/p_and_p/RAID/LinuxRAID.html for more infos
	LeftSymmetric
	LeftAsymmetric
	RightSymmetric
	RightAsymmetric
)

//Layout determines the block location relative to the nodes, usually stored on name node
type Layout struct {
	//Pattern is how blocks are mapped to nodes
	pattern Pattern

	//distribution maps the shards of each stripe to node
	distribution [][]int

	//blockToOffset determines the block's offset with repsect to the node
	blockToOffset [][]int
}

//NewLayout news a layout for certain patterns
//stripeNum = totalPoolSize // blockSize // stripeLen
//stripeLen = dataShards + parityShards
func NewLayout(pattern Pattern, stripeNum, dataShards, parityShards, nodeNum int) *Layout {
	stripeLen := dataShards + parityShards
	layout := &Layout{
		pattern:       pattern,
		distribution:  make([][]int, stripeNum),
		blockToOffset: makeArr2DInt(stripeNum, stripeLen),
	}
	if pattern == Random {
		countSum := make([]int, nodeNum)
		for i := 0; i < stripeNum; i++ {
			layout.distribution[i] = genRandomArr(nodeNum, 0)[:stripeLen]
			for j := 0; j < stripeLen; j++ {
				diskId := layout.distribution[i][j]
				layout.blockToOffset[i][j] = countSum[diskId]
				countSum[diskId]++
			}

		}
	} else if pattern == LeftSymmetric {
		countSum := make([]int, nodeNum)
		for i := 0; i < stripeNum; i++ {
			layout.distribution[i] = make([]int, stripeLen)
			for c := 0; c < stripeLen; c++ {
				layout.distribution[i][c] = (c + i) % stripeLen
				diskId := layout.distribution[i][c]
				layout.blockToOffset[i][c] = countSum[diskId]
				countSum[diskId]++
			}
		}
	} else if pattern == LeftAsymmetric {
		countSum := make([]int, nodeNum)
		for i := 0; i < stripeNum; i++ {
			layout.distribution[i] = make([]int, stripeLen)
			parity := dataShards
			for c := 0; c < parityShards; c++ {
				j := (dataShards - i%stripeLen + c + stripeLen) % stripeLen
				layout.distribution[i][j] = parity
				parity++
				diskId := layout.distribution[i][j]
				layout.blockToOffset[i][j] = countSum[diskId]
				countSum[diskId]++
			}
			data := 0
			for c := 0; c < stripeLen; c++ {
				if layout.distribution[i][c] == 0 {
					layout.distribution[i][c] = data
					data++
					diskId := layout.distribution[i][c]
					layout.blockToOffset[i][c] = countSum[diskId]
					countSum[diskId]++
				}
			}
		}
	} else if pattern == RightSymmetric {
		countSum := make([]int, nodeNum)
		for i := 0; i < stripeNum; i++ {
			layout.distribution[i] = make([]int, stripeLen)
			parity := dataShards
			for c := 0; c < parityShards; c++ {
				j := (i%stripeLen + c) % stripeLen
				layout.distribution[i][j] = parity
				parity++
				diskId := layout.distribution[i][j]
				layout.blockToOffset[i][j] = countSum[diskId]
				countSum[diskId]++
			}
			data := i
			for c := 0; c < stripeLen; c++ {
				if layout.distribution[i][c] == 0 {
					layout.distribution[i][c] = (stripeLen - data) % stripeLen
					data++
					diskId := layout.distribution[i][c]
					layout.blockToOffset[i][c] = countSum[diskId]
					countSum[diskId]++
				}
			}
		}
	} else if pattern == RightAsymmetric {
		countSum := make([]int, nodeNum)
		for i := 0; i < stripeNum; i++ {
			layout.distribution[i] = make([]int, stripeLen)
			parity := dataShards
			for c := 0; c < parityShards; c++ {
				j := (i%stripeLen + c) % stripeLen
				layout.distribution[i][j] = parity
				parity++
				diskId := layout.distribution[i][j]
				layout.blockToOffset[i][j] = countSum[diskId]
				countSum[diskId]++
			}
			data := 0
			for c := 0; c < stripeLen; c++ {
				if layout.distribution[i][c] == 0 {
					layout.distribution[i][c] = data
					data++
					diskId := layout.distribution[i][c]
					layout.blockToOffset[i][c] = countSum[diskId]
					countSum[diskId]++
				}
			}
		}
	}
	return layout
}
