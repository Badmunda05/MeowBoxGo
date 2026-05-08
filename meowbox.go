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

// uploadResponse represents the API response.
type uploadResponse struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Files       []struct {
		URL string `json:"url"`
	} `json:"files"`
}

// FileEntry holds file data for multi-upload.
type FileEntry struct {
	Buffer   *bytes.Buffer
	FileName string
}

// UploadFile uploads a single file and returns the direct URL.
func UploadFile(fileBuffer *bytes.Buffer, fileName string, timeout time.Duration) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// IMPORTANT:
	// Backend expects "files" not "files[]"
	part, err := writer.CreateFormFile("files", fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, fileBuffer)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", meowboxURL, &body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error

		if errors.As(err, &netErr) && netErr.Timeout() {
			return "", fmt.Errorf("upload request timed out after %d seconds", int(timeout.Seconds()))
		}

		return "", fmt.Errorf("failed to connect to MeowBox: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("rate limit hit — too many requests")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"HTTP error occurred: %s - %s",
			resp.Status,
			string(respBody),
		)
	}

	var result uploadResponse

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if !result.Success {
		if result.Description == "" {
			result.Description = "unknown error"
		}

		return "", fmt.Errorf("upload failed: %s", result.Description)
	}

	if len(result.Files) == 0 {
		return "", fmt.Errorf("upload succeeded but no file URLs returned")
	}

	return result.Files[0].URL, nil
}

// UploadFiles uploads multiple files and returns all URLs.
func UploadFiles(files []FileEntry, timeout time.Duration) ([]string, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for _, file := range files {
		// IMPORTANT:
		// Backend expects "files" not "files[]"
		part, err := writer.CreateFormFile("files", file.FileName)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create form file for %q: %w",
				file.FileName,
				err,
			)
		}

		_, err = io.Copy(part, file.Buffer)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to copy file content for %q: %w",
				file.FileName,
				err,
			)
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", meowboxURL, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error

		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf(
				"upload request timed out after %d seconds",
				int(timeout.Seconds()),
			)
		}

		return nil, fmt.Errorf("failed to connect to MeowBox: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit hit — too many requests")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"HTTP error occurred: %s - %s",
			resp.Status,
			string(respBody),
		)
	}

	var result uploadResponse

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if !result.Success {
		if result.Description == "" {
			result.Description = "unknown error"
		}

		return nil, fmt.Errorf("upload failed: %s", result.Description)
	}

	if len(result.Files) == 0 {
		return nil, fmt.Errorf("upload succeeded but no file URLs returned")
	}

	var urls []string

	for _, file := range result.Files {
		urls = append(urls, file.URL)
	}

	return urls, nil
}
