//go:build !windows && !darwin && !linux

package service

func readClipboardFilePaths() ([]string, error) {
	return nil, nil
}
