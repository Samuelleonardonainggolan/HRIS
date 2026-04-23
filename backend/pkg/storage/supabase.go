package storage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// SupabaseUploader handles file uploads to Supabase Storage
type SupabaseUploader struct {
	supabaseURL string
	apiKey      string
	bucketName  string
	httpClient  *http.Client
}

// NewSupabaseUploader creates a new Supabase uploader instance
func NewSupabaseUploader(supabaseURL, apiKey, bucketName string) *SupabaseUploader {
	return &SupabaseUploader{
		supabaseURL: strings.TrimSuffix(supabaseURL, "/"),
		apiKey:      apiKey,
		bucketName:  bucketName,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// UploadFile uploads a file to Supabase and returns the public URL
func (s *SupabaseUploader) UploadFile(fileBytes []byte, fileName, folder string) (string, error) {
	if s.supabaseURL == "" || s.apiKey == "" {
		return "", fmt.Errorf("supabase credentials not configured")
	}

	// Generate unique filename with timestamp to avoid conflicts
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(filepath.Base(fileName), ext)
	timestamp := time.Now().UnixMilli()
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, timestamp, ext)

	// Construct the path: folder/filename
	filePath := uniqueName
	if folder != "" {
		filePath = fmt.Sprintf("%s/%s", strings.Trim(folder, "/"), uniqueName)
	}

	// Upload endpoint: POST /storage/v1/object/{bucket}/{path}
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.supabaseURL, s.bucketName, filePath)

	// Create request
	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("apikey", s.apiKey)
	req.Header.Set("Content-Type", "application/octet-stream")

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyText := string(bodyBytes)
		if strings.Contains(strings.ToLower(bodyText), "row-level security policy") {
			return "", fmt.Errorf("upload ditolak oleh Supabase RLS. Gunakan SUPABASE_SERVICE_ROLE_KEY di backend atau buat policy insert bucket '%s'. detail: %s", s.bucketName, bodyText)
		}

		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, bodyText)
	}

	// Pastikan response body habis dibaca agar koneksi reusable.
	_, _ = io.Copy(io.Discard, resp.Body)

	// Construct public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.supabaseURL, s.bucketName, filePath)

	return publicURL, nil
}

// DeleteFile deletes a file from Supabase Storage
func (s *SupabaseUploader) DeleteFile(filePath string) error {
	if s.supabaseURL == "" || s.apiKey == "" {
		return fmt.Errorf("supabase credentials not configured")
	}

	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.supabaseURL, s.bucketName, filePath)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("apikey", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
