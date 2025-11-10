package handlers

import (
	"device-monitor-go/models"
	"device-monitor-go/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SyncIotData handles POST /api/iot/sync/:sessionId
func SyncIotData(c *gin.Context) {
	sessionID := c.Param("sessionId")

	// Get session
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

	// Sync IoT data
	iotService := services.GetIotService()
	iotData, err := iotService.SyncSessionData(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync IoT data: " + err.Error(),
		})
		return
	}

	// Convert iotData to the format expected by frontend
	dataArray := []gin.H{}
	dataCount := 0
	
	for pointName, pointData := range iotData {
		if dataMap, ok := pointData.(map[string]interface{}); ok {
			if dataList, ok := dataMap["data"].([]map[string]interface{}); ok {
				for _, item := range dataList {
					dataArray = append(dataArray, gin.H{
						"pointName":  pointName,
						"pointValue": item["value"],
						"timestamp":  item["time"],
						"unit":       dataMap["unit"],
					})
					dataCount++
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "IoT data synced successfully",
		"data":      dataArray,
		"dataCount": dataCount,
	})
}

// GetIotDataPoints handles GET /api/iot/data-points
func GetIotDataPoints(c *gin.Context) {
	dataPoints := models.GetIotDataPoints()
	c.JSON(http.StatusOK, gin.H{
		"dataPoints": dataPoints,
	})
}

// TestIotConnection handles GET /api/iot/test-connection
func TestIotConnection(c *gin.Context) {
	iotService := services.GetIotService()
	err := iotService.TestConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "IoT connection test failed: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IoT connection test successful",
		"success": true,
	})
}

// GetDevicePoints handles GET /api/iot/device/:deviceId/points
func GetDevicePoints(c *gin.Context) {
	// Return static list of known points based on the thing model
	points := []gin.H{
		{"name": "volume", "displayName": "噪音", "unit": "dB"},
		{"name": "shake", "displayName": "振动", "unit": "g"},
		{"name": "temperature", "displayName": "温度", "unit": "°C"},
		{"name": "feature_speed_1_speed", "displayName": "转速", "unit": "rpm"},
		{"name": "controlledvariable", "displayName": "是否在运行", "unit": ""},
		{"name": "controlledvolume", "displayName": "音量是否监控", "unit": ""},
		{"name": "feature_hilbert_2_hb", "displayName": "希尔伯特值", "unit": ""},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    points,
	})
}