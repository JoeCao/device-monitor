package models

// IotPointSummary represents aggregated statistics for a data point
type IotPointSummary struct {
	PointName string  `db:"point_name" json:"point_name"`
	Unit      string  `db:"unit" json:"unit"`
	Count     int     `db:"count" json:"count"`
	MinValue  float64 `db:"min_value" json:"min_value"`
	MaxValue  float64 `db:"max_value" json:"max_value"`
	AvgValue  float64 `db:"avg_value" json:"avg_value"`
}

// IotTimeSeries represents time-bucketed data
type IotTimeSeries struct {
	TimeBucket string  `db:"time_bucket" json:"time_bucket"`
	AvgValue   float64 `db:"avg_value" json:"avg_value"`
	MinValue   float64 `db:"min_value" json:"min_value"`
	MaxValue   float64 `db:"max_value" json:"max_value"`
	DataCount  int     `db:"data_count" json:"data_count"`
}

// GetIotDataPointNames returns summary statistics for each unique point name in a session
// Note: In the Go version, we don't store IoT data locally, so this returns empty results
func GetIotDataPointNames(sessionID string) ([]IotPointSummary, error) {
	// Return empty array as we don't store IoT data locally
	return []IotPointSummary{}, nil
}

// GetAggregatedIotData returns time-bucketed aggregated data for a specific point
// Note: In the Go version, we don't store IoT data locally, so this returns empty results
func GetAggregatedIotData(sessionID string, pointName string, interval string) ([]IotTimeSeries, error) {
	// Return empty array as we don't store IoT data locally
	return []IotTimeSeries{}, nil
}

// GetIotDataBySessionId returns all raw IoT data points for a session
// Note: In the Go version, we don't store IoT data locally, so this returns empty results
func GetIotDataBySessionId(sessionID string) ([]IotDataPoint, error) {
	// Return empty array as we don't store IoT data locally
	return []IotDataPoint{}, nil
}