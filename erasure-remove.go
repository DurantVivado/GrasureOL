package grasure

////RemoveFile deletes specific file `filename`in the system.
////
////Both the file blobs and meta data are deleted. It's currently irreversible.
//func (e *Erasure) RemoveFile(filename string) error {
//	baseFilename := filepath.Base(filename)
//	if _, ok := e.fileMap.Load(baseFilename); !ok {
//		return fmt.Errorf("the file %s does not exist in the file system",
//			baseFilename)
//	}
//	g := new(errgroup.Group)
//
//	for _, path := range e.diskInfos[:e.DiskNum] {
//		path := path
//		files, err := os.ReadDir(path.diskPath)
//		if err != nil {
//			return err
//		}
//		if len(files) == 0 {
//			continue
//		}
//		g.Go(func() error {
//
//			err = os.RemoveAll(filepath.Join(path.diskPath, baseFilename))
//			if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//	if err := g.Wait(); err != nil {
//		return err
//	}
//	e.fileMap.Delete(baseFilename)
//	// delete(e.fileMap, filename)
//	if !e.Quiet {
//		log.Printf("file %s successfully deleted.", baseFilename)
//	}
//	return nil
//}
//
////check if file exists both in config and storage blobs
//func (e *Erasure) checkIfFileExist(filename string) (bool, error) {
//	//1. first check the storage blobs if file still exists
//	baseFilename := filepath.Base(filename)
//
//	g := new(errgroup.Group)
//
//	for _, path := range e.diskInfos[:e.DiskNum] {
//		path := path
//		files, err := os.ReadDir(path.diskPath)
//		if err != nil {
//			return false, err
//		}
//		if len(files) == 0 {
//			continue
//		}
//		g.Go(func() error {
//
//			subpath := filepath.Join(path.diskPath, baseFilename)
//			if ok, err := pathExist(subpath); !ok && err == nil {
//				return errFileBlobNotFound
//			} else if err != nil {
//				return err
//			}
//			return nil
//		})
//	}
//	if err := g.Wait(); err != nil {
//		return false, err
//	}
//	//2. check if fileMap contains the file
//	if _, ok := e.fileMap.Load(baseFilename); !ok {
//		return false, nil
//	}
//	return true, nil
//}
//
