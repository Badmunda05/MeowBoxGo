# meowbox (Go)

> Permanent file hosting — files never expire.
> A Go library for uploading files to [MeowBox](https://files.tgvibes.online).

---

## Installation

### Install package

Run this inside your Go project:

```bash
go get github.com/Badmunda05/MeowBoxGo@v1.0.2
```

---

### Import package

```go
import "github.com/Badmunda05/MeowBoxGo"
```

---

### Or add directly in go.mod

```go
require github.com/Badmunda05/MeowBoxGo v1.0.2
```

> **Note:** The GitHub repository must be public and the module path must match.

---

## Usage

### Single file upload

```go
package main

import (
    "bytes"
    "fmt"
    "os"
    "time"

    "github.com/Badmunda05/meowbox"
)

func main() {
    file, err := os.Open("photo.jpg")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    var buf bytes.Buffer
    buf.ReadFrom(file)

    url, err := meowbox.UploadFile(&buf, "photo.jpg", 30*time.Second)
    if err != nil {
        fmt.Println("Upload failed:", err)
        return
    }
    fmt.Println("Uploaded:", url)
    // https://files.tgvibes.online/AbCdEfGh.jpg
}
```

### Multi-file upload

```go
entries := []meowbox.FileEntry{
    {Buffer: &buf1, FileName: "file1.jpg"},
    {Buffer: &buf2, FileName: "file2.mp4"},
}

urls, err := meowbox.UploadFiles(entries, 30*time.Second)
if err != nil {
    fmt.Println("Upload failed:", err)
    return
}
for _, u := range urls {
    fmt.Println(u)
}
```

### Run the example

```bash
git clone https://github.com/Badmunda05/meowbox
cd meowbox/example
go run main.go photo.jpg
# Multiple files:
go run main.go file1.jpg file2.mp4
```

---

Users can then install it with:

```bash
go get github.com/Badmunda05/meowbox@v1.0.0
# or always get the latest:
go get github.com/Badmunda05/meowbox
```

---

## API Reference

| Function | Description |
|----------|-------------|
| `UploadFile(buf, fileName, timeout)` | Uploads a single file, returns the direct URL |
| `UploadFiles([]FileEntry, timeout)` | Uploads multiple files, returns a slice of URLs |

### Types

```go
type FileEntry struct {
    Buffer   *bytes.Buffer
    FileName string
}
```

---

## Error Handling

```go
url, err := meowbox.UploadFile(&buf, "photo.jpg", 30*time.Second)
if err != nil {
    // Possible errors:
    // - "upload request timed out after N seconds"
    // - "rate limit hit — too many requests, try again later"
    // - "upload failed: <server message>"
    // - "HTTP error occurred: ..."
    fmt.Println("Upload failed:", err)
}
```

---

## Links

- Python library: [MeowBox (Python)](https://github.com/Badmunda05/MeowBox)
- Telegram: [@BadmundaXd](https://t.me/BadmundaXd)
