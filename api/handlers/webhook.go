package handlers

import (
	"device-monitor-go/config"
	"device-monitor-go/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// WebhookRequest represents the webhook request body
type WebhookRequest struct {
	Power     string                 `json:"power"`
	DeviceID  string                 `json:"deviceId"`
	SessionID string                 `json:"sessionId"`
	Timestamp string                 `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DeviceStart handles POST /api/webhooks/device/start
func DeviceStart(c *gin.Context) {
	var req WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Check power status
	if req.Power != "on" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid power status",
		})
		return
	}

	// Get device ID from query parameter or request body
	deviceID := c.Query("deviceName")
	if deviceID == "" {
		deviceID = req.DeviceID
	}
	if deviceID == "" {
		deviceID = config.AppConfig.IotDeviceCode
	}

	// Parse timestamp or use current time
	var startTime time.Time
	if req.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, req.Timestamp)
		if err == nil {
			startTime = parsed
		} else {
			startTime = time.Now()
		}
	} else {
		startTime = time.Now()
	}

	// Create new session
	session, err := models.CreateSession(deviceID, startTime, req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create session: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Device started successfully",
		"sessionId": session.SessionID,
		"deviceId":  deviceID,
		"startTime": startTime,
	})
}

// DeviceEnd handles POST /api/webhooks/device/end
func DeviceEnd(c *gin.Context) {
	var req WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Check power status
	if req.Power != "off" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid power status",
		})
		return
	}

	// Get device ID from query parameter or request body
	deviceID := c.Query("deviceName")
	if deviceID == "" {
		deviceID = req.DeviceID
	}
	if deviceID == "" {
		deviceID = config.AppConfig.IotDeviceCode
	}

	// Parse timestamp or use current time
	var endTime time.Time
	if req.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, req.Timestamp)
		if err == nil {
			endTime = parsed
		} else {
			endTime = time.Now()
		}
	} else {
		endTime = time.Now()
	}

	// Find session to end
	var sessionID string
	if req.SessionID != "" {
		sessionID = req.SessionID
	} else {
		// Find the latest running session for this device
		sessions, err := models.GetRunningSessions(deviceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to find running session: " + err.Error(),
			})
			return
		}

		if len(sessions) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No running session found for device",
			})
			return
		}

		// Use the most recent running session
		sessionID = sessions[0].SessionID
	}

	// End the session
	err := models.EndSession(sessionID, endTime, req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to end session: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Device stopped successfully",
		"sessionId": sessionID,
		"deviceId":  deviceID,
		"endTime":   endTime,
	})
}

// TestWebhookStart handles POST /api/webhooks/test/start
func TestWebhookStart(c *gin.Context) {
	deviceID := c.Query("deviceId")
	if deviceID == "" {
		deviceID = config.AppConfig.IotDeviceCode
	}

	// Create test session
	session, err := models.CreateSession(deviceID, time.Now(), map[string]interface{}{
		"test": true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create test session: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Test device started successfully",
		"sessionId": session.SessionID,
		"deviceId":  deviceID,
	})
}

// TestWebhookEnd handles POST /api/webhooks/test/end
func TestWebhookEnd(c *gin.Context) {
	sessionID := c.Query("sessionId")
	deviceID := c.Query("deviceId")

	if sessionID == "" && deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either sessionId or deviceId is required",
		})
		return
	}

	// If no session ID provided, find the latest running session
	if sessionID == "" {
		sessions, err := models.GetRunningSessions(deviceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to find running session: " + err.Error(),
			})
			return
		}

		if len(sessions) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No running session found for device",
			})
			return
		}

		sessionID = sessions[0].SessionID
	}

	// End the session
	err := models.EndSession(sessionID, time.Now(), map[string]interface{}{
		"test": true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to end test session: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Test device stopped successfully",
		"sessionId": sessionID,
	})
}