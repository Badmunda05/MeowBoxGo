package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/Badmunda05/meowbox"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file1> [file2] ...")
		return
	}

	timeout := 30 * time.Second

	// ── Single file upload ──────────────────────────────────────────
	if len(os.Args) == 2 {
		filePath := os.Args[1]
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		var buf bytes.Buffer
		if _, err = buf.ReadFrom(file); err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		url, err := meowbox.UploadFile(&buf, filePath, timeout)
		if err != nil {
			fmt.Printf("Upload failed: %v\n", err)
			return
		}
		fmt.Println("Uploaded:", url)
		return
	}

	// ── Multi-file upload ───────────────────────────────────────────
	var entries []meowbox.FileEntry
	for _, filePath := range os.Args[1:] {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", filePath, err)
			return
		}
		var buf bytes.Buffer
		if _, err = buf.ReadFrom(file); err != nil {
			file.Close()
			fmt.Printf("Error reading %s: %v\n", filePath, err)
			return
		}
		file.Close()
		entries = append(entries, meowbox.FileEntry{Buffer: &buf, FileName: filePath})
	}

	urls, err := meowbox.UploadFiles(entries, timeout)
	if err != nil {
		fmt.Printf("Multi-file upload failed: %v\n", err)
		return
	}
	for i, u := range urls {
		fmt.Printf("File %d: %s\n", i+1, u)
	}
}
