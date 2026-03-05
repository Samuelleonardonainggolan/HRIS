package faceclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

type ExtractResponse struct {
	Success    bool      `json:"success"`
	EmployeeID string    `json:"employee_id"`
	Embedding  []float32 `json:"embedding"`
	Dimension  int       `json:"dimension"`
	ElapsedMs  float64   `json:"elapsed_ms"`
	Message    string    `json:"message"`
}

type GeoResult struct {
	IsValid   bool    `json:"is_valid"`
	DistanceM float64 `json:"distance_m"`
	RadiusM   float64 `json:"radius_m"`
	Message   string  `json:"message"`
}

type FaceResult struct {
	Matched    bool    `json:"matched"`
	Similarity float64 `json:"similarity"`
	Confidence float64 `json:"confidence"`
	Threshold  float64 `json:"threshold"`
	Message    string  `json:"message"`
}

type AttendanceProcessResponse struct {
	Decision   string      `json:"decision"`
	Approved   bool        `json:"approved"`
	EmployeeID string      `json:"employee_id"`
	RecordType string      `json:"record_type"`
	Geo        GeoResult   `json:"geo"`
	Face       *FaceResult `json:"face"`
	ElapsedMs  float64     `json:"elapsed_ms"`
	Message    string      `json:"message"`
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL, apiKey string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) ExtractEmbedding(employeeID string, photoBytes []byte, filename string) ([]float32, error) {
	body, contentType, err := buildMultipartWithBytes(
		map[string]string{"employee_id": employeeID},
		"photo",
		filename,
		photoBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.post("/face/extract", body, contentType)
	if err != nil {
		return nil, err
	}

	var result ExtractResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("extract failed: %s", result.Message)
	}
	return result.Embedding, nil
}

type ProcessAttendanceRequest struct {
	EmployeeID      string    `json:"employee_id"`
	StoredEmbedding []float32 `json:"stored_embedding"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	RecordType      string    `json:"record_type"`
	Threshold       float64   `json:"threshold,omitempty"`
	RadiusM         float64   `json:"radius_m,omitempty"`
}

func (c *Client) ProcessAttendance(req ProcessAttendanceRequest, photoBytes []byte, filename string) (*AttendanceProcessResponse, error) {
	dataJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	body, contentType, err := buildMultipartWithBytes(
		map[string]string{"data": string(dataJSON)},
		"photo",
		filename,
		photoBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	respBytes, err := c.post("/attendance/process", body, contentType)
	if err != nil {
		return nil, err
	}

	var result AttendanceProcessResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

type GeoRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	RadiusM   float64 `json:"radius_m,omitempty"`
}

func (c *Client) ValidateGeo(lat, lng float64) (*GeoResult, error) {
	body, err := json.Marshal(GeoRequest{Latitude: lat, Longitude: lng})
	if err != nil {
		return nil, err
	}
	respBytes, err := c.postJSON("/geo/validate", body)
	if err != nil {
		return nil, err
	}
	var result GeoResult
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) HealthCheck() (bool, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/health", nil)
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (c *Client) post(path string, body io.Reader, contentType string) ([]byte, error) {
	req, err := http.NewRequest("POST", c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("face service error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func (c *Client) postJSON(path string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("face service error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func buildMultipartWithBytes(fields map[string]string, fileField, filename string, fileBytes []byte) (io.Reader, string, error) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return nil, "", err
		}
	}

	part, err := w.CreateFormFile(fileField, filepath.Base(filename))
	if err != nil {
		return nil, "", err
	}
	if _, err = part.Write(fileBytes); err != nil {
		return nil, "", err
	}
	_ = w.Close()

	return buf, w.FormDataContentType(), nil
}

