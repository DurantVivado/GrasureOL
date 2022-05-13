package grasure

//ReadFile reads ONE file  on the system and save it to local `savePath`.
//
//In case of any failure within fault tolerance, the file will be decoded first.
//`degrade` indicates whether degraded read is enabled.
//func (la *LocalAccess) ReadFile(filename string, savepath string, options *Options) error {
//	baseFileName := filepath.Base(filename)
//	intFi, ok := lafileMap.Load(baseFileName)
//	if !ok {
//		return errFileNotFound
//	}
//	fi := intFi.(*fileInfo)
//
//	fileSize := fi.FileSize
//	stripeNum := int(ceilFracInt64(fileSize, ladataStripeSize))
//	dist := fi.Distribution
//	//first we check the number of alive disks
//	// to judge if any part need reconstruction
//	alive := int32(0)
//	ifs := make([]*os.File, laDiskNum)
//	erg := new(errgroup.Group)
//
//	for i, disk := range ladiskInfos[:laDiskNum] {
//		i := i
//		disk := disk
//		erg.Go(func() error {
//			folderPath := filepath.Join(disk.diskPath, baseFileName)
//			blobPath := filepath.Join(folderPath, "BLOB")
//			if !disk.available {
//				return &diskError{disk.diskPath, " available flag set false"}
//			}
//			ifs[i], err = os.Open(blobPath)
//			if err != nil {
//				disk.available = false
//				return err
//			}
//
//			disk.available = true
//			atomic.AddInt32(&alive, 1)
//			return nil
//		})
//	}
//	if err := erg.Wait(); err != nil {
//		if !laQuiet {
//			log.Printf("%s", err.Error())
//		}
//	}
//	defer func() {
//		for i := 0; i < laDiskNum; i++ {
//			if ifs[i] != nil {
//				ifs[i].Close()
//			}
//		}
//	}()
//	if int(alive) < laK {
//		//the disk renders inrecoverable
//		return errTooFewDisksAlive
//	}
//	if int(alive) == laDiskNum {
//		if !laQuiet {
//			log.Println("start reading blocks")
//		}
//	} else {
//		if !laQuiet {
//			log.Println("start reconstructing blocks")
//		}
//	}
//	//for local save path
//	sf, err := os.OpenFile(savepath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
//	if err != nil {
//		return err
//	}
//	defer sf.Close()
//
//	//Since the file is striped, we have to reconstruct each stripe
//	//for each stripe we rejoin the data
//	numBlob := ceilFracInt(stripeNum, laConStripes)
//	stripeCnt := 0
//	nextStripe := 0
//	for blob := 0; blob < numBlob; blob++ {
//		if stripeCnt+laConStripes > stripeNum {
//			nextStripe = stripeNum - stripeCnt
//		} else {
//			nextStripe = laConStripes
//		}
//		eg := laerrgroupPool.Get().(*errgroup.Group)
//		blobBuf := makeArr2DByte(laConStripes, int(laallStripeSize))
//		for s := 0; s < nextStripe; s++ {
//			s := s
//			stripeNo := stripeCnt + s
//			// offset := int64(subCnt) * laallStripeSize
//			eg.Go(func() error {
//				erg := laerrgroupPool.Get().(*errgroup.Group)
//				defer laerrgroupPool.Put(erg)
//				//read all blocks in parallel
//				//We only have to read k blocks to rec
//				failList := make(map[int]bool)
//				for i := 0; i < laK+laM; i++ {
//					i := i
//					diskId := dist[stripeNo][i]
//					disk := ladiskInfos[diskId]
//					blkStat := fi.blockInfos[stripeNo][i]
//					if !disk.available || blkStat.bstat != blkOK {
//						failList[diskId] = true
//						continue
//					}
//					erg.Go(func() error {
//
//						//we also need to know the block's accurate offset with respect to disk
//						offset := fi.blockToOffset[stripeNo][i]
//						_, err := ifs[diskId].ReadAt(blobBuf[s][int64(i)*laBlockSize:int64(i+1)*laBlockSize],
//							int64(offset)*laBlockSize)
//						// fmt.Println("Read ", n, " bytes at", i, ", block ", block)
//						if err != nil && err != io.EOF {
//							return err
//						}
//						return nil
//					})
//				}
//				if err := erg.Wait(); err != nil {
//					return err
//				}
//				//Split the blob into k+m parts
//				splitData, err := lasplitStripe(blobBuf[s])
//				if err != nil {
//					return err
//				}
//				//verify and reconstruct if broken
//				ok, err := laenc.Verify(splitData)
//				if err != nil {
//					return err
//				}
//				if !ok {
//					// fmt.Println("reconstruct data of stripe:", stripeNo)
//					err = laenc.ReconstructWithList(splitData,
//						&failList,
//						&(fi.Distribution[stripeNo]),
//						options.Degrade)
//
//					// err = laenc.ReconstructWithKBlocks(splitData,
//					// 	&failList,
//					// 	&loadBalancedScheme[stripeNo],
//					// 	&(fi.Distribution[stripeNo]),
//					// 	degrade)
//					if err != nil {
//						return err
//					}
//				}
//				//join and write to output file
//
//				for i := 0; i < laK; i++ {
//					i := i
//					writeOffset := int64(stripeNo)*ladataStripeSize + int64(i)*laBlockSize
//					if fileSize-writeOffset <= laBlockSize {
//						leftLen := fileSize - writeOffset
//						_, err := sf.WriteAt(splitData[i][:leftLen], writeOffset)
//						if err != nil {
//							return err
//						}
//						break
//					}
//					erg.Go(func() error {
//						// fmt.Println("i:", i, "writeOffset", writeOffset+laBlockSize, "at stripe", subCnt)
//						_, err := sf.WriteAt(splitData[i], writeOffset)
//						if err != nil {
//							return err
//						}
//						// sf.Sync()
//						return nil
//					})
//
//				}
//				if err := erg.Wait(); err != nil {
//					return err
//				}
//				return nil
//			})
//
//		}
//		if err := eg.Wait(); err != nil {
//			return err
//		}
//		laerrgroupPool.Put(eg)
//		stripeCnt += nextStripe
//
//	}
//	if !laQuiet {
//		log.Printf("reading %s...", filename)
//	}
//	return nil
//}
//
//func (e *Erasure) splitStripe(data []byte) ([][]byte, error) {
//	if len(data) == 0 {
//		return nil, reedsolomon.ErrShortData
//	}
//	// Calculate number of bytes per data shard.
//	perShard := ceilFracInt(len(data), laK+laM)
//
//	// Split into equal-length shards.
//	dst := make([][]byte, laK+laM)
//	i := 0
//	for ; i < len(dst) && len(data) >= perShard; i++ {
//		dst[i], data = data[:perShard:perShard], data[perShard:]
//	}
//
//	return dst, nil
//}

//Read Phase
// normal mode:
//First, figure out which nodes storage the blocks according to the meta;
//Second, send response to storage nodes when all is completed.
// degraded mode:
//First,  figure out which nodes storage the blocks according to the meta and algorithm;
//Second, send response to storage nodes when all is completed.
//Third, the computing node reconstruct the file.

func readFromNode(filename string, nodeId int) (data []byte, err error) {
	return nil, err
}
