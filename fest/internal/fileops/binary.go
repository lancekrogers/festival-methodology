package fileops

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// Common binary file extensions
var binaryExtensions = map[string]bool{
	// Images
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".bmp":  true,
	".ico":  true,
	".svg":  true,
	".webp": true,
	".tiff": true,
	".tif":  true,

	// Documents
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,

	// Archives
	".zip": true,
	".tar": true,
	".gz":  true,
	".bz2": true,
	".7z":  true,
	".rar": true,

	// Executables
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
	".app":   true,
	".bin":   true,

	// Audio/Video
	".mp3":  true,
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".wav":  true,
	".flac": true,
	".ogg":  true,

	// Fonts
	".ttf":   true,
	".otf":   true,
	".woff":  true,
	".woff2": true,

	// Other
	".pyc":   true,
	".class": true,
	".o":     true,
	".a":     true,
}

// IsBinaryFile checks if a file is likely binary
func IsBinaryFile(path string) (bool, error) {
	// Check extension first (fast check)
	ext := strings.ToLower(filepath.Ext(path))
	if binaryExtensions[ext] {
		return true, nil
	}

	// Check file content for null bytes (more accurate but slower)
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return false, err
	}

	// Check for null bytes (common indicator of binary data)
	if bytes.Contains(buf[:n], []byte{0}) {
		return true, nil
	}

	return false, nil
}
