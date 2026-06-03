// internal/faceclient/client.go
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

// ─── Response Structs ─────────────────────────────────────────────────────────

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

type VerifyFaceResponse struct {
	Matched      bool    `json:"matched"`
	Similarity   float64 `json:"similarity"`
	SpoofScore   float64 `json:"spoof_score"`
	FinalScore   float64 `json:"final_score"`
	Confidence   float64 `json:"confidence"`
	Threshold    float64 `json:"threshold"`
	Message      string  `json:"message"`
}

type AttendanceProcessResponse struct {
	Decision   string      `json:"decision"` // "approved" / "rejected_gps" / "rejected_face"
	Approved   bool        `json:"approved"`
	EmployeeID string      `json:"employee_id"`
	RecordType string      `json:"record_type"`
	Geo        GeoResult   `json:"geo"`
	Face       *FaceResult `json:"face"` // nil jika GPS sudah gagal
	ElapsedMs  float64     `json:"elapsed_ms"`
	Message    string      `json:"message"`
}

// ─── Client ───────────────────────────────────────────────────────────────────

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New membuat instance baru FaceClient dengan konfigurasi dari parameter
func New(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ─── 1. ExtractEmbedding ──────────────────────────────────────────────────────
func (c *Client) ExtractEmbedding(
	employeeID string,
	photoBytes []byte,
	filename string,
) ([]float32, error) {

	body, contentType, err := buildMultipartWithBytes(map[string]string{
		"employee_id": employeeID,
	}, "photo", filename, photoBytes)
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

// ─── 2. ProcessAttendance ─────────────────────────────────────────────────────
type ProcessAttendanceRequest struct {
	EmployeeID      string    `json:"employee_id"`
	StoredEmbedding []float32 `json:"stored_embedding"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	RecordType      string    `json:"record_type"` // "checkin" / "checkout"
	Threshold       float64   `json:"threshold,omitempty"`
	RadiusM         float64   `json:"radius_m,omitempty"`
}

func (c *Client) ProcessAttendance(
	req ProcessAttendanceRequest,
	photoBytes []byte,
	filename string,
) (*AttendanceProcessResponse, error) {

	dataJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	body, contentType, err := buildMultipartWithBytes(map[string]string{
		"data": string(dataJSON),
	}, "photo", filename, photoBytes)
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

func (c *Client) VerifyFace(
	employeeID string,
	storedEmbedding []float32,
	photoBytes []byte,
	filename string,
	liveness string,
	threshold *float64,
) (*VerifyFaceResponse, error) {

	reqData := map[string]interface{}{
		"employee_id":      employeeID,
		"stored_embedding": storedEmbedding,
	}
	if threshold != nil {
		reqData["threshold"] = *threshold
	}

	dataJSON, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("marshal request data: %w", err)
	}

	body, contentType, err := buildMultipartWithBytes(map[string]string{
		"data":     string(dataJSON),
		"liveness": liveness,
	}, "photo", filename, photoBytes)
	if err != nil {
		return nil, fmt.Errorf("build request body: %w", err)
	}

	respBytes, err := c.post("/face/verify", body, contentType)
	if err != nil {
		return nil, err
	}

	var result VerifyFaceResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}


// ─── 3. ValidateGeo (opsional, jika ingin cek GPS saja) ──────────────────────
type GeoRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	RadiusM   float64 `json:"radius_m,omitempty"`
}

func (c *Client) ValidateGeo(lat, lng float64, radiusM ...float64) (*GeoResult, error) {
	radius := 100.0
	if len(radiusM) > 0 {
		radius = radiusM[0]
	}

	body, err := json.Marshal(GeoRequest{
		Latitude:  lat,
		Longitude: lng,
		RadiusM:   radius,
	})
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

// ─── 4. Health Check ─────────────────────────────────────────────────────────
func (c *Client) HealthCheck() (bool, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/health", nil)
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200, nil
}

// ─── HTTP Helpers ─────────────────────────────────────────────────────────────
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
	if resp.StatusCode != 200 {
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
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("face service error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func buildMultipartWithBytes(
	fields map[string]string,
	fileField, filename string,
	fileBytes []byte,
) (io.Reader, string, error) {
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
	w.Close()

	return buf, w.FormDataContentType(), nil
}
