package service

import "errors"

var errFolderUploadCancelled = errors.New("folder upload cancelled")

type folderUploadRunner struct {
	mkdir     func(remotePath string) error
	upload    func(localPath, remotePath string) error
	cancelled func() bool
}

func partitionLocalUploadItems(items []LocalUploadItem) (dirs, files []LocalUploadItem) {
	dirs = make([]LocalUploadItem, 0)
	files = make([]LocalUploadItem, 0)
	for _, item := range items {
		if item.IsDir {
			dirs = append(dirs, item)
			continue
		}
		files = append(files, item)
	}
	return dirs, files
}

func runFolderUpload(runner folderUploadRunner, remotePath string, items []LocalUploadItem) error {
	dirs, files := partitionLocalUploadItems(items)
	for _, dir := range dirs {
		if isFolderUploadCancelled(runner) {
			return errFolderUploadCancelled
		}
		if err := runner.mkdir(joinRemoteUploadPath(remotePath, dir.RelPath)); err != nil {
			return err
		}
	}
	for _, file := range files {
		if isFolderUploadCancelled(runner) {
			return errFolderUploadCancelled
		}
		if err := runner.upload(file.LocalPath, joinRemoteUploadPath(remotePath, file.RelPath)); err != nil {
			return err
		}
	}
	return nil
}

func isFolderUploadCancelled(runner folderUploadRunner) bool {
	return runner.cancelled != nil && runner.cancelled()
}
