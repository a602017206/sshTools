//go:build darwin

package service

import (
	"context"
	"os/exec"
	"time"
)

const macClipboardFilesScript = `ObjC.import('AppKit');
ObjC.import('Foundation');
function run() {
  const pb = $.NSPasteboard.generalPasteboard;
  const paths = [];
  const names = pb.propertyListForType('NSFilenamesPboardType');
  if (names) {
    for (let i = 0; i < names.count; i++) {
      paths.push(ObjC.unwrap(names.objectAtIndex(i)));
    }
  }
  if (paths.length === 0) {
    const items = pb.pasteboardItems;
    if (items) {
      for (let i = 0; i < items.count; i++) {
        const item = items.objectAtIndex(i);
        let value = item.stringForType('public.file-url');
        if (value) {
          const url = $.NSURL.URLWithString(value);
          if (url) paths.push(ObjC.unwrap(url.path));
        }
      }
    }
  }
  return paths.join('\n');
}`

func readClipboardFilePaths() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "osascript", "-l", "JavaScript", "-e", macClipboardFilesScript)
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}
	return splitClipboardLines(string(out)), nil
}
