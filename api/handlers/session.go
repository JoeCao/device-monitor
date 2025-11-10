package handlers

import (
	"device-monitor-go/models"
	"device-monitor-go/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetSessions handles GET /api/sessions
func GetSessions(c *gin.Context) {
	// Parse query parameters
	filter := models.SessionFilter{
		DeviceID:  c.Query("deviceId"),
		Status:    c.Query("status"),
		StartDate: c.Query("startDate"),
		EndDate:   c.Query("endDate"),
	}

	// Parse pagination
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	} else {
		filter.Limit = 50 // Default limit (matching Node.js)
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = o
		}
	}

	// Get sessions
	sessions, total, err := models.GetSessions(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sessions: " + err.Error(),
		})
		return
	}

	// Match Node.js response format
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessions,
		"pagination": gin.H{
			"limit":  filter.Limit,
			"offset": filter.Offset,
			"total":  total,
		},
	})
}

// GetSessionByID handles GET /api/sessions/:id
func GetSessionByID(c *gin.Context) {
	sessionID := c.Param("id")

	session, err := models.GetSessionByID(sessionID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Session not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get session: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetSessionReport handles GET /api/sessions/:id/report
func GetSessionReport(c *gin.Context) {
	sessionID := c.Param("id")

	session, err := models.GetSessionByID(sessionID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Session not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get session: " + err.Error(),
			})
		}
		return
	}

	// Get IoT data points from database (like Node.js version)
	pointNames, err := models.GetIotDataPointNames(sessionID)
	if err != nil {
		// Log error but continue
		pointNames = []models.IotPointSummary{}
	}

	// Get aggregated data for each point
	aggregatedData := make(map[string]interface{})
	for _, point := range pointNames {
		timeSeries, _ := models.GetAggregatedIotData(sessionID, point.PointName, "minute")
		aggregatedData[point.PointName] = gin.H{
			"summary": gin.H{
				"point_name": point.PointName,
				"unit":       point.Unit,
				"count":      point.Count,
				"min_value":  point.MinValue,
				"max_value":  point.MaxValue,
				"avg_value":  point.AvgValue,
			},
			"timeSeries": timeSeries,
		}
	}

	// If no data in database, try to sync from IoT platform
	log.Printf("GetSessionReport: pointNames length: %d, session status: %s", len(pointNames), session.Status)
	if len(pointNames) == 0 {
		log.Printf("Syncing IoT data for session %s", sessionID)
		iotService := services.GetIotService()
		iotData, err := iotService.SyncSessionData(session)
		if err == nil && iotData != nil {
			log.Printf("Successfully synced IoT data, processing %d data points", len(iotData))
			// Convert IoT data to expected format
			for dataPoint, data := range iotData {
				log.Printf("Processing dataPoint: %s, data type: %T", dataPoint, data)
				if dataMap, ok := data.(map[string]interface{}); ok {
					log.Printf("DataMap keys for %s: %v", dataPoint, getMapKeys(dataMap))
					// Calculate summary from data array
					summary := calculateSummary(dataMap)
					
					// Convert data array to timeSeries format
					timeSeries := []gin.H{}
					if dataArray, ok := dataMap["data"].([]interface{}); ok {
						for _, item := range dataArray {
							if itemMap, ok := item.(map[string]interface{}); ok {
								// Use time as time_bucket and value as avg_value
								timeBucket := itemMap["time"]
								avgValue := itemMap["value"]
								
								// For non-Hilbert data, ensure value is numeric
								if dataPoint != "feature_hilbert_2_hb" {
									switch v := avgValue.(type) {
									case string:
										if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
											avgValue = floatVal
										}
									}
								}
								
								timeSeries = append(timeSeries, gin.H{
									"time_bucket": timeBucket,
									"avg_value":   avgValue,
								})
							}
						}
					}
					
					aggregatedData[dataPoint] = gin.H{
						"summary":    summary,
						"timeSeries": timeSeries,
					}
				}
			}
		}
	}

	// Get raw IoT data
	rawData, _ := models.GetIotDataBySessionId(sessionID)

	// Match Node.js response format exactly
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"session": session,
			"iotData": gin.H{
				"points":     pointNames,
				"aggregated": aggregatedData,
				"raw":        rawData,
			},
		},
	})
}

// getMapKeys returns the keys of a map
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// calculateSummary calculates summary statistics from IoT data
func calculateSummary(dataMap map[string]interface{}) gin.H {
	summary := gin.H{
		"point_name": "",
		"unit":       "",
		"count":      0,
		"min_value":  0.0,
		"max_value":  0.0,
		"avg_value":  0.0,
	}

	// Extract metadata
	if displayName, ok := dataMap["displayName"].(string); ok {
		summary["point_name"] = displayName
	}
	if unit, ok := dataMap["unit"].(string); ok {
		summary["unit"] = unit
	}
	
	log.Printf("calculateSummary: processing data for %s", summary["point_name"])

	// Calculate statistics from data array
	dataInterface := dataMap["data"]
	log.Printf("Data interface type: %T for %s", dataInterface, summary["point_name"])
	
	// Try to handle both []interface{} and []map[string]interface{}
	var dataArray []interface{}
	switch v := dataInterface.(type) {
	case []interface{}:
		dataArray = v
	case []map[string]interface{}:
		// Convert []map[string]interface{} to []interface{}
		dataArray = make([]interface{}, len(v))
		for i, item := range v {
			dataArray[i] = item
		}
	}
	
	if len(dataArray) > 0 {
		log.Printf("Data array length: %d for %s", len(dataArray), summary["point_name"])
		var sum, min, max float64
		count := 0
		
		for i, item := range dataArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Handle different numeric types
					var numValue float64
					handled := false
					
					valueInterface := itemMap["value"]
					if i == 0 {
						log.Printf("First value type for %s: %T, value: %v", summary["point_name"], valueInterface, valueInterface)
					}
					
					if value, ok := valueInterface.(float64); ok {
						numValue = value
						handled = true
					} else if intValue, ok := valueInterface.(int); ok {
						numValue = float64(intValue)
						handled = true
					} else if strValue, ok := valueInterface.(string); ok {
						// Try to parse string values as float
						if floatValue, err := strconv.ParseFloat(strValue, 64); err == nil {
							numValue = floatValue
							handled = true
						}
					}
					
					// Process the numeric value
					if handled {
						if i == 0 || count == 0 {
							min = numValue
							max = numValue
						} else {
							if numValue < min {
								min = numValue
							}
							if numValue > max {
								max = numValue
							}
						}
						sum += numValue
						count++
					}
				}
			}

			if count > 0 {
				summary["count"] = count
				summary["min_value"] = min
				summary["max_value"] = max
				summary["avg_value"] = sum / float64(count)
			}
		}

	return summary
}

// DeleteSession handles DELETE /api/sessions/:id
func DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")

	err := models.DeleteSession(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete session: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session deleted successfully",
	})
}

// GetStatistics handles GET /api/sessions/statistics
func GetStatistics(c *gin.Context) {
	deviceID := c.Query("deviceId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	stats, err := models.GetStatistics(deviceID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDeviceStatistics handles GET /api/sessions/device/:deviceId/statistics
func GetDeviceStatistics(c *gin.Context) {
	deviceID := c.Param("deviceId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	stats, err := models.GetStatistics(deviceID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics: " + err.Error(),
		})
		return
	}

	// Match Node.js response format
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}