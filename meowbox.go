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

// uploadFile represents a single uploaded file response.
type uploadFile struct {
	URL string `json:"url"`
	Src string `json:"src"`
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
			return "", fmt.Errorf(
				"upload request timed out after %d seconds",
				int(timeout.Seconds()),
			)
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

	// API returns ARRAY not object
	var result []uploadFile

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("upload succeeded but no file URLs returned")
	}

	url := result[0].URL

	if url == "" {
		url = result[0].Src
	}

	if url == "" {
		return "", fmt.Errorf("response missing file URL")
	}

	return url, nil
}

// UploadFiles uploads multiple files and returns all URLs.
func UploadFiles(files []FileEntry, timeout time.Duration) ([]string, error) {

	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	for _, file := range files {

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

	// API returns ARRAY not object
	var result []uploadFile

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("upload succeeded but no file URLs returned")
	}

	var urls []string

	for _, file := range result {

		url := file.URL

		if url == "" {
			url = file.Src
		}

		if url != "" {
			urls = append(urls, url)
		}
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("response missing file URLs")
	}

	return urls, nil
}
