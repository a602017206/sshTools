package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxExtractedArchiveSize int64 = 2 << 30

func ExtractArchive(archivePath, targetDir string) (err error) {
	if _, statErr := os.Stat(targetDir); statErr == nil {
		return fmt.Errorf("归档目标目录已存在: %s", targetDir)
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("检查归档目标目录失败: %w", statErr)
	}
	if err := os.MkdirAll(targetDir, 0o700); err != nil {
		return fmt.Errorf("创建归档目标目录失败: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(targetDir)
		}
	}()

	lowerPath := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lowerPath, ".zip"):
		err = extractZipArchive(archivePath, targetDir)
	case strings.HasSuffix(lowerPath, ".tar.gz"), strings.HasSuffix(lowerPath, ".tgz"):
		err = extractTarGzArchive(archivePath, targetDir)
	default:
		err = fmt.Errorf("不支持的归档格式: %s", filepath.Ext(archivePath))
	}
	if err != nil {
		return err
	}
	committed = true
	return nil
}

func extractZipArchive(archivePath, targetDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("打开 zip 归档失败: %w", err)
	}
	defer reader.Close()

	var extracted int64
	for _, entry := range reader.File {
		if entry.UncompressedSize64 > uint64(maxExtractedArchiveSize) || extracted > maxExtractedArchiveSize-int64(entry.UncompressedSize64) {
			return fmt.Errorf("归档展开内容超过限制")
		}
		extracted += int64(entry.UncompressedSize64)
		target, err := safeArchiveTarget(targetDir, entry.Name)
		if err != nil {
			return err
		}
		mode := entry.Mode()
		switch {
		case mode.IsDir():
			if err := os.MkdirAll(target, secureDirMode(mode)); err != nil {
				return fmt.Errorf("创建归档目录失败: %w", err)
			}
		case mode.IsRegular():
			if err := extractZipRegularFile(entry, target, secureFileMode(mode)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("归档包含不允许的文件类型: %s", entry.Name)
		}
	}
	return nil
}

func extractZipRegularFile(entry *zip.File, target string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return fmt.Errorf("创建归档文件目录失败: %w", err)
	}
	reader, err := entry.Open()
	if err != nil {
		return fmt.Errorf("读取 zip 文件失败: %w", err)
	}
	defer reader.Close()
	return writeArchiveFile(target, mode, reader)
}

func extractTarGzArchive(archivePath, targetDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("打开 tar.gz 归档失败: %w", err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("读取 gzip 数据失败: %w", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	var extracted int64
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar 归档失败: %w", err)
		}
		if header.Size < 0 || header.Size > maxExtractedArchiveSize || extracted > maxExtractedArchiveSize-header.Size {
			return fmt.Errorf("归档展开内容超过限制")
		}
		extracted += header.Size
		target, err := safeArchiveTarget(targetDir, header.Name)
		if err != nil {
			return err
		}
		mode := os.FileMode(header.Mode)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, secureDirMode(mode)); err != nil {
				return fmt.Errorf("创建归档目录失败: %w", err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return fmt.Errorf("创建归档文件目录失败: %w", err)
			}
			if err := writeArchiveFile(target, secureFileMode(mode), io.LimitReader(tarReader, header.Size)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("归档包含不允许的文件类型: %s", header.Name)
		}
	}
	return nil
}

func safeArchiveTarget(targetDir, name string) (string, error) {
	normalized := strings.ReplaceAll(name, "\\", "/")
	cleaned := filepath.Clean(filepath.FromSlash(normalized))
	if cleaned == "." || filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("归档路径不安全: %s", name)
	}
	target := filepath.Join(targetDir, cleaned)
	relative, err := filepath.Rel(targetDir, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("归档路径超出目标目录: %s", name)
	}
	return target, nil
}

func writeArchiveFile(target string, mode os.FileMode, reader io.Reader) error {
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("创建归档文件失败: %w", err)
	}
	if _, err := io.Copy(file, reader); err != nil {
		_ = file.Close()
		return fmt.Errorf("写入归档文件失败: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("关闭归档文件失败: %w", err)
	}
	return nil
}

func secureDirMode(mode os.FileMode) os.FileMode {
	permissions := mode.Perm() &^ 0o022
	if permissions == 0 {
		permissions = 0o700
	}
	return permissions
}

func secureFileMode(mode os.FileMode) os.FileMode {
	permissions := mode.Perm() &^ 0o022
	if permissions == 0 {
		permissions = 0o600
	}
	return permissions
}
