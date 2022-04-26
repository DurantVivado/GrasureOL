package grasure

//diskInfo contains the disk-level information
//but not mechanical and electrical parameters
type diskInfo struct {
	//the disk path
	diskPath string

	//it's flag and when disk fails, it renders false.
	available bool

	//it tells how many blocks a disk holds
	numBlocks int

	//it's a disk with meta file?
	ifMetaExist bool

	//the capacity of a disk
	capacity int64

}


//ReadDiskPath reads the disk paths from diskFilePath.
//There should be exactly ONE disk path at each line.
//
//This func can NOT be called concurrently.
//func (n *Node) ReadDiskPath() error {
//	e.mu.RLock()
//	defer e.mu.RUnlock()
//	f, err := os.Open(e.DiskFilePath)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	buf := bufio.NewReader(f)
//	e.diskInfos = make([]*diskInfo, 0)
//	for {
//		line, _, err := buf.ReadLine()
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			return err
//		}
//		path := string(line)
//		if ok, err := pathExist(path); !ok && err == nil {
//			return &diskError{path, "disk path not exist"}
//		} else if err != nil {
//			return err
//		}
//		metaPath := filepath.Join(path, "META")
//		flag := false
//		if ok, err := pathExist(metaPath); ok && err == nil {
//			flag = true
//		} else if err != nil {
//			return err
//		}
//		diskInfo := &diskInfo{diskPath: string(line), available: true, ifMetaExist: flag}
//		e.diskInfos = append(e.diskInfos, diskInfo)
//	}
//	return nil
//}
