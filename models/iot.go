package models

import (
	"time"
)

// IotDataPoint represents an IoT data point
type IotDataPoint struct {
	ID         int       `db:"id" json:"id"`
	SessionID  string    `db:"session_id" json:"session_id"`
	PointName  string    `db:"point_name" json:"point_name"`
	PointValue float64   `db:"point_value" json:"point_value"`
	Unit       string    `db:"unit" json:"unit"`
	Timestamp  time.Time `db:"timestamp" json:"timestamp"`
	RawData    string    `db:"raw_data" json:"raw_data"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// IotDataResponse represents the response from IoT platform
type IotDataResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []IotDataItem `json:"list"`
	} `json:"data"`
}

// IotDataItem represents a single data item from IoT platform
type IotDataItem struct {
	Time  interface{} `json:"time"`  // Can be string or number
	Value interface{} `json:"value"` // Can be various types
}

// IotDeviceDataPoint defines IoT data point configuration
type IotDeviceDataPoint struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Unit        string `json:"unit"`
	Type        string `json:"type"`
}

// GetIotDataPoints returns the configuration for all IoT data points
func GetIotDataPoints() []IotDeviceDataPoint {
	return []IotDeviceDataPoint{
		{
			Name:        "volume",
			DisplayName: "噪音",
			Unit:        "dB",
			Type:        "number",
		},
		{
			Name:        "shake",
			DisplayName: "振动",
			Unit:        "g",
			Type:        "number",
		},
		{
			Name:        "temperature",
			DisplayName: "温度",
			Unit:        "°C",
			Type:        "number",
		},
		{
			Name:        "feature_speed_1_speed",
			DisplayName: "转速",
			Unit:        "rpm",
			Type:        "number",
		},
		{
			Name:        "feature_hilbert_2_hb",
			DisplayName: "希尔伯特包络",
			Unit:        "",
			Type:        "array",
		},
		{
			Name:        "controlledvariable",
			DisplayName: "是否在运行",
			Unit:        "",
			Type:        "boolean",
		},
		{
			Name:        "controlledvolume",
			DisplayName: "音量是否监控",
			Unit:        "",
			Type:        "boolean",
		},
	}
}

// IoT auth response
type IotAuthResponse struct {
	Success      bool   `json:"success"`
	Code         int    `json:"code"`
	Data         string `json:"data"`
	ErrorMessage string `json:"errorMessage"`
}

// IoT error response
type IotErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}