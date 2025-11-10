package services

import (
	"bytes"
	"crypto/tls"
	"device-monitor-go/config"
	"device-monitor-go/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type IotService struct {
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

var iotServiceInstance *IotService
var once sync.Once

// GetIotService returns singleton instance of IotService
func GetIotService() *IotService {
	once.Do(func() {
		// Create HTTP client with proxy support if configured
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For development/testing
			},
		}

		// Configure proxy if set
		if config.AppConfig.HttpProxy != "" || config.AppConfig.HttpsProxy != "" {
			proxyURL, err := url.Parse(config.AppConfig.HttpsProxy)
			if err == nil {
				transport.Proxy = http.ProxyURL(proxyURL)
			}
		}

		iotServiceInstance = &IotService{
			httpClient: &http.Client{
				Transport: transport,
				Timeout:   30 * time.Second,
			},
		}
	})
	return iotServiceInstance
}

// getAccessToken retrieves or refreshes OAuth access token
func (s *IotService) getAccessToken() (string, error) {
	s.tokenMutex.RLock()
	if s.accessToken != "" && time.Now().Before(s.tokenExpiry) {
		token := s.accessToken
		s.tokenMutex.RUnlock()
		return token, nil
	}
	s.tokenMutex.RUnlock()

	// Need to refresh token
	s.tokenMutex.Lock()
	defer s.tokenMutex.Unlock()

	// Double check after acquiring write lock
	if s.accessToken != "" && time.Now().Before(s.tokenExpiry) {
		return s.accessToken, nil
	}

	// Get new token - matching Node.js implementation
	tokenURL := fmt.Sprintf("%s/api/v1/oauth/auth", config.AppConfig.IotApiBaseURL)
	
	// Create JSON payload
	payload := map[string]string{
		"appId":     config.AppConfig.IotAppKey,
		"appSecret": config.AppConfig.IotAppSecret,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	req, err := http.NewRequest("POST", tokenURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp models.IotAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if !authResp.Success || authResp.Code != 200 {
		return "", fmt.Errorf("authentication failed: %s", authResp.ErrorMessage)
	}

	// Store token with default expiry (24 hours as per original code)
	s.accessToken = authResp.Data
	s.tokenExpiry = time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	log.Printf("Successfully obtained IoT access token")
	return s.accessToken, nil
}

// QueryDeviceData queries device data from IoT platform
func (s *IotService) QueryDeviceData(deviceCode, dataPoint string, startTime, endTime time.Time) (*models.IotDataResponse, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Build query URL - matching Node.js implementation
	queryURL := fmt.Sprintf("%s/api/v1/thing/queryDevicePropertiesData", config.AppConfig.IotApiBaseURL)

	// Prepare request body - use formatted strings like Node.js version
	requestBody := map[string]interface{}{
		"deviceName": deviceCode,
		"identifier": []string{dataPoint},
		"startTime":  startTime.Format("2006-01-02 15:04:05"),
		"endTime":    endTime.Format("2006-01-02 15:04:05"),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	
	log.Printf("Query request URL: %s", queryURL)
	log.Printf("Query request body: %s", string(jsonBody))

	req, err := http.NewRequest("POST", queryURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query device data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read query response: %w", err)
	}

	// Check for auth errors
	if resp.StatusCode == http.StatusUnauthorized {
		// Clear token to force refresh on next request
		s.tokenMutex.Lock()
		s.accessToken = ""
		s.tokenMutex.Unlock()
		
		var errResp models.IotErrorResponse
		json.Unmarshal(body, &errResp)
		return nil, fmt.Errorf("authentication failed: %s", errResp.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Log response for debugging
	log.Printf("IoT query response status: %d, body length: %d", resp.StatusCode, len(body))
	
	// Parse response
	var dataResp map[string]interface{}
	if err := json.Unmarshal(body, &dataResp); err != nil {
		return nil, fmt.Errorf("failed to parse query response: %w", err)
	}

	// Log the response structure
	log.Printf("IoT response structure: %+v", dataResp)

	// Extract data for the specific data point
	result := &models.IotDataResponse{
		Code:    0,
		Message: "success",
		Data: struct {
			List []models.IotDataItem `json:"list"`
		}{
			List: []models.IotDataItem{},
		},
	}

	// Check response structure - matching Node.js implementation
	// Response format: { success: bool, data: [{dataList: [...], point: {...}}, ...] }
	if data, ok := dataResp["data"].([]interface{}); ok {
		log.Printf("Found data array with %d items", len(data))
		for _, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// Check if this item has the right point
				if point, ok := itemMap["point"].(map[string]interface{}); ok {
					if identifier, ok := point["identifier"].(string); ok {
						log.Printf("Found point: %s (looking for %s)", identifier, dataPoint)
						if identifier == dataPoint {
							// Extract dataList
							if dataList, ok := itemMap["dataList"].([]interface{}); ok {
								log.Printf("Found dataList with %d items for %s", len(dataList), dataPoint)
								for _, dataItem := range dataList {
									if di, ok := dataItem.(map[string]interface{}); ok {
										// Use "time" field like Node.js version
										item := models.IotDataItem{
											Time:  di["time"],
											Value: di["value"],
										}
										result.Data.List = append(result.Data.List, item)
									}
								}
							}
						}
					}
				}
			}
		}
	} else {
		log.Printf("Response data is not an array, type: %T", dataResp["data"])
	}

	return result, nil
}

// SyncSessionData queries all IoT data for a session period
func (s *IotService) SyncSessionData(session *models.DeviceSession) (map[string]interface{}, error) {
	deviceCode := session.DeviceID
	if deviceCode == "" {
		deviceCode = config.AppConfig.IotDeviceCode
	}

	// For running sessions, use current time as end time
	endTime := time.Now()
	if session.Status == "completed" && session.EndTime != nil {
		endTime = *session.EndTime
	}

	// Log the query parameters
	log.Printf("Syncing IoT data for device %s, session %s, time range: %s to %s", 
		deviceCode, session.SessionID, session.StartTime, endTime)

	// Query all data points
	dataPoints := models.GetIotDataPoints()
	results := make(map[string]interface{})
	
	// Use goroutines to query multiple data points concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(dataPoints))

	for _, dp := range dataPoints {
		wg.Add(1)
		go func(dataPoint models.IotDeviceDataPoint) {
			defer wg.Done()

			log.Printf("Querying data point: %s", dataPoint.Name)
			resp, err := s.QueryDeviceData(deviceCode, dataPoint.Name, session.StartTime, endTime)
			if err != nil {
				log.Printf("Error querying %s: %v", dataPoint.Name, err)
				errChan <- fmt.Errorf("failed to query %s: %w", dataPoint.Name, err)
				
				// Still add empty data for this point
				mu.Lock()
				results[dataPoint.Name] = map[string]interface{}{
					"displayName": dataPoint.DisplayName,
					"unit":        dataPoint.Unit,
					"type":        dataPoint.Type,
					"data":        []map[string]interface{}{},
				}
				mu.Unlock()
				return
			}

			log.Printf("Got %d data items for %s", len(resp.Data.List), dataPoint.Name)

			// Process data
			processedData := s.processDataPoints(resp.Data.List, dataPoint)
			
			mu.Lock()
			results[dataPoint.Name] = map[string]interface{}{
				"displayName": dataPoint.DisplayName,
				"unit":        dataPoint.Unit,
				"type":        dataPoint.Type,
				"data":        processedData,
			}
			mu.Unlock()
		}(dp)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []string
	for err := range errChan {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		log.Printf("Errors during data sync: %v", errors)
	}

	log.Printf("IoT data sync completed with %d data points", len(results))
	return results, nil
}

// processDataPoints processes raw IoT data points
func (s *IotService) processDataPoints(items []models.IotDataItem, dataPoint models.IotDeviceDataPoint) []map[string]interface{} {
	processed := []map[string]interface{}{}

	for _, item := range items {
		// Convert timestamp
		var timestamp time.Time
		switch t := item.Time.(type) {
		case float64:
			timestamp = time.UnixMilli(int64(t))
		case string:
			if ts, err := strconv.ParseInt(t, 10, 64); err == nil {
				if len(t) == 13 { // Milliseconds
					timestamp = time.UnixMilli(ts)
				} else { // Seconds
					timestamp = time.Unix(ts, 0)
				}
			}
		}

		// Process value based on type
		var value interface{}
		if dataPoint.Type == "array" && dataPoint.Name == "feature_hilbert_2_hb" {
			// Keep as string for Hilbert envelope data
			value = item.Value
		} else {
			// Convert to appropriate type
			switch v := item.Value.(type) {
			case string:
				if dataPoint.Type == "number" {
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						value = f
					} else {
						value = 0.0
					}
				} else {
					value = v
				}
			case float64:
				value = v
			case bool:
				value = v
			default:
				value = v
			}
		}

		processed = append(processed, map[string]interface{}{
			"time":  timestamp.Format(time.RFC3339),
			"value": value,
		})
	}

	return processed
}

// TestConnection tests the IoT platform connection
func (s *IotService) TestConnection() error {
	_, err := s.getAccessToken()
	return err
}