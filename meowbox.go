package meowbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"time"
)

const meowboxURL = "https://files.tgvibes.online/upload"

// uploadResponse is the JSON response from the MeowBox API.
type uploadResponse struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Files       []struct {
		URL string `json:"url"`
	} `json:"files"`
}

// UploadFile uploads a file from a bytes.Buffer to MeowBox.
// The fileName parameter specifies the name of the file to be uploaded.
// The timeout parameter specifies how long the client should wait for the server to respond.
// The function returns the direct URL of the uploaded file or an error if the upload failed.
func UploadFile(fileBuffer *bytes.Buffer, fileName string, timeout time.Duration) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if part, err := writer.CreateFormFile("files[]", fileName); err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	} else if _, err = io.Copy(part, fileBuffer); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	_ = writer.Close()

	req, err := http.NewRequest("POST", meowboxURL, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return "", fmt.Errorf("upload request timed out after %d seconds", int(timeout.Seconds()))
		}
		return "", fmt.Errorf("failed to connect to MeowBox: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return "", fmt.Errorf("rate limit hit — too many requests, try again later")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP error occurred: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result uploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if !result.Success {
		desc := result.Description
		if desc == "" {
			desc = "unknown error"
		}
		return "", fmt.Errorf("upload failed: %s", desc)
	}

	if len(result.Files) == 0 {
		return "", fmt.Errorf("upload succeeded but no file URLs returned")
	}

	return result.Files[0].URL, nil
}

// UploadFiles uploads multiple files to MeowBox in a single request.
// Each file is provided as a FileEntry with its buffer and name.
// The function returns a slice of direct URLs or an error.
func UploadFiles(files []FileEntry, timeout time.Duration) ([]string, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for _, f := range files {
		part, err := writer.CreateFormFile("files[]", f.FileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create form file for %q: %w", f.FileName, err)
		}
		if _, err = io.Copy(part, f.Buffer); err != nil {
			return nil, fmt.Errorf("failed to copy file content for %q: %w", f.FileName, err)
		}
	}

	_ = writer.Close()

	req, err := http.NewRequest("POST", meowboxURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return nil, fmt.Errorf("upload request timed out after %d seconds", int(timeout.Seconds()))
		}
		return nil, fmt.Errorf("failed to connect to MeowBox: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limit hit — too many requests, try again later")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error occurred: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result uploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if !result.Success {
		desc := result.Description
		if desc == "" {
			desc = "unknown error"
		}
		return nil, fmt.Errorf("upload failed: %s", desc)
	}

	if len(result.Files) == 0 {
		return nil, fmt.Errorf("upload succeeded but no file URLs returned")
	}

	urls := make([]string, 0, len(result.Files))
	for _, f := range result.Files {
		urls = append(urls, f.URL)
	}
	return urls, nil
}

// FileEntry holds a file buffer and its name for multi-file uploads.
type FileEntry struct {
	Buffer   *bytes.Buffer
	FileName string
}
