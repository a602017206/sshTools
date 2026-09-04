package service

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LocalUploadItem is one file or directory to recreate under a remote folder.
type LocalUploadItem struct {
	LocalPath string `json:"localPath"`
	RelPath   string `json:"relPath"`
	IsDir     bool   `json:"isDir"`
}

// ExpandLocalUploadPaths turns dropped or picked local paths into a remote-relative
// tree. Files keep their basename. Directories include the folder name so the
// remote listing shows that folder, including empty subdirectories. Symlinks and
// non-regular files are skipped and not followed.
func ExpandLocalUploadPaths(paths []string) ([]LocalUploadItem, error) {
	items := make([]LocalUploadItem, 0)
	seen := make(map[string]struct{})

	for _, raw := range paths {
		localPath := strings.TrimSpace(raw)
		if localPath == "" {
			continue
		}

		info, err := os.Lstat(localPath)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		if info.IsDir() {
			dirItems, err := expandLocalDirectory(localPath, seen)
			if err != nil {
				return nil, err
			}
			items = append(items, dirItems...)
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}

		item := LocalUploadItem{
			LocalPath: localPath,
			RelPath:   filepath.ToSlash(filepath.Base(localPath)),
			IsDir:     false,
		}
		if !markSeen(seen, item.LocalPath) {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func expandLocalDirectory(root string, seen map[string]struct{}) ([]LocalUploadItem, error) {
	rootName := filepath.Base(root)
	items := make([]LocalUploadItem, 0)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relPath := filepath.ToSlash(rootName)
		if rel != "." {
			relPath = filepath.ToSlash(filepath.Join(rootName, rel))
		}

		if d.IsDir() {
			item := LocalUploadItem{LocalPath: path, RelPath: relPath, IsDir: true}
			if !markSeen(seen, item.LocalPath) {
				return nil
			}
			items = append(items, item)
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		item := LocalUploadItem{LocalPath: path, RelPath: relPath, IsDir: false}
		if !markSeen(seen, item.LocalPath) {
			return nil
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func markSeen(seen map[string]struct{}, key string) bool {
	if _, ok := seen[key]; ok {
		return false
	}
	seen[key] = struct{}{}
	return true
}

func joinRemoteUploadPath(remoteDir, relPath string) string {
	remoteDir = strings.TrimSuffix(filepath.ToSlash(remoteDir), "/")
	if remoteDir == "" {
		remoteDir = "/"
	}
	relPath = strings.TrimPrefix(filepath.ToSlash(relPath), "/")
	if relPath == "" {
		return remoteDir
	}
	if remoteDir == "/" {
		return "/" + relPath
	}
	return remoteDir + "/" + relPath
}
