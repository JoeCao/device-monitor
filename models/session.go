package models

import (
	"database/sql"
	"device-monitor-go/database"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DeviceSession struct {
	ID         int                `db:"id" json:"id"`
	DeviceID   string             `db:"device_id" json:"device_id"`
	SessionID  string             `db:"session_id" json:"session_id"`
	StartTime  time.Time          `db:"start_time" json:"start_time"`
	EndTime    *time.Time         `db:"end_time" json:"end_time"`
	Duration   sql.NullInt64      `db:"duration" json:"-"`
	DurationInt *int64            `json:"duration"`
	Status     string             `db:"status" json:"status"`
	Metadata   sql.NullString     `db:"metadata" json:"-"`
	MetadataObj map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `db:"updated_at" json:"updated_at"`
}

type SessionFilter struct {
	DeviceID  string
	Status    string
	StartDate string
	EndDate   string
	Limit     int
	Offset    int
}

// BeforeSave processes metadata before saving
func (s *DeviceSession) BeforeSave() error {
	if s.MetadataObj != nil {
		data, err := json.Marshal(s.MetadataObj)
		if err != nil {
			return err
		}
		s.Metadata = sql.NullString{String: string(data), Valid: true}
	}
	return nil
}

// AfterFind processes metadata after loading
func (s *DeviceSession) AfterFind() error {
	// Process metadata
	if s.Metadata.Valid && s.Metadata.String != "" {
		s.MetadataObj = make(map[string]interface{})
		if err := json.Unmarshal([]byte(s.Metadata.String), &s.MetadataObj); err != nil {
			return err
		}
	}
	
	// Process duration
	if s.Duration.Valid {
		s.DurationInt = &s.Duration.Int64
	} else {
		s.DurationInt = nil
	}
	
	return nil
}

// CreateSession creates a new device session
func CreateSession(deviceID string, startTime time.Time, metadata map[string]interface{}) (*DeviceSession, error) {
	session := &DeviceSession{
		DeviceID:    deviceID,
		SessionID:   uuid.New().String(),
		StartTime:   startTime,
		Status:      "running",
		MetadataObj: metadata,
	}

	if err := session.BeforeSave(); err != nil {
		return nil, err
	}

	query := `
		INSERT INTO device_sessions (device_id, session_id, start_time, status, metadata)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := database.DB.Exec(query, session.DeviceID, session.SessionID, 
		session.StartTime, session.Status, session.Metadata)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	session.ID = int(id)
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	return session, nil
}

// EndSession ends a running session
func EndSession(sessionID string, endTime time.Time, metadata map[string]interface{}) error {
	// First get the session to calculate duration
	session, err := GetSessionByID(sessionID)
	if err != nil {
		return err
	}

	if session.Status != "running" {
		return fmt.Errorf("session is not running")
	}

	duration := int64(endTime.Sub(session.StartTime).Seconds())
	
	// Merge metadata
	if metadata != nil {
		if session.MetadataObj == nil {
			session.MetadataObj = make(map[string]interface{})
		}
		for k, v := range metadata {
			session.MetadataObj[k] = v
		}
	}

	session.EndTime = &endTime
	session.Duration = sql.NullInt64{Int64: duration, Valid: true}
	session.DurationInt = &duration
	session.Status = "completed"

	if err := session.BeforeSave(); err != nil {
		return err
	}

	query := `
		UPDATE device_sessions 
		SET end_time = ?, duration = ?, status = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP
		WHERE session_id = ?
	`

	_, err = database.DB.Exec(query, endTime, duration, session.Status, session.Metadata, sessionID)
	return err
}

// GetSessionByID retrieves a session by ID
func GetSessionByID(sessionID string) (*DeviceSession, error) {
	session := &DeviceSession{}
	query := `SELECT * FROM device_sessions WHERE session_id = ?`
	
	err := database.DB.Get(session, query, sessionID)
	if err != nil {
		return nil, err
	}

	if err := session.AfterFind(); err != nil {
		return nil, err
	}

	return session, nil
}

// GetRunningSessions gets all running sessions for a device
func GetRunningSessions(deviceID string) ([]*DeviceSession, error) {
	sessions := []*DeviceSession{}
	query := `SELECT * FROM device_sessions WHERE device_id = ? AND status = 'running' ORDER BY start_time DESC`
	
	err := database.DB.Select(&sessions, query, deviceID)
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if err := session.AfterFind(); err != nil {
			return nil, err
		}
	}

	return sessions, nil
}

// GetSessions retrieves sessions with filtering
func GetSessions(filter SessionFilter) ([]*DeviceSession, int, error) {
	// Build query
	query := `SELECT * FROM device_sessions WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM device_sessions WHERE 1=1`
	args := []interface{}{}

	if filter.DeviceID != "" {
		query += " AND device_id = ?"
		countQuery += " AND device_id = ?"
		args = append(args, filter.DeviceID)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		countQuery += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.StartDate != "" {
		query += " AND DATE(start_time) >= ?"
		countQuery += " AND DATE(start_time) >= ?"
		args = append(args, filter.StartDate)
	}

	if filter.EndDate != "" {
		query += " AND DATE(start_time) <= ?"
		countQuery += " AND DATE(start_time) <= ?"
		args = append(args, filter.EndDate)
	}

	// Get total count
	var total int
	err := database.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Add ordering and pagination
	query += " ORDER BY start_time DESC"
	
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	// Get sessions
	sessions := []*DeviceSession{}
	err = database.DB.Select(&sessions, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// Process metadata
	for _, session := range sessions {
		if err := session.AfterFind(); err != nil {
			return nil, 0, err
		}
	}

	return sessions, total, nil
}

// DeleteSession deletes a session
func DeleteSession(sessionID string) error {
	query := `DELETE FROM device_sessions WHERE session_id = ?`
	_, err := database.DB.Exec(query, sessionID)
	return err
}

// GetStatistics retrieves session statistics
func GetStatistics(deviceID string, startDate, endDate string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total sessions count
	var totalSessions int
	query := `SELECT COUNT(*) FROM device_sessions WHERE 1=1`
	args := []interface{}{}

	if deviceID != "" {
		query += " AND device_id = ?"
		args = append(args, deviceID)
	}

	if startDate != "" {
		query += " AND DATE(start_time) >= ?"
		args = append(args, startDate)
	}

	if endDate != "" {
		query += " AND DATE(start_time) <= ?"
		args = append(args, endDate)
	}

	err := database.DB.Get(&totalSessions, query, args...)
	if err != nil {
		return nil, err
	}
	stats["total_sessions"] = totalSessions

	// Completed sessions
	var completedSessions int
	completedQuery := query + " AND status = 'completed'"
	err = database.DB.Get(&completedSessions, completedQuery, args...)
	if err != nil {
		return nil, err
	}
	stats["completed_sessions"] = completedSessions

	// Running sessions
	var runningSessions int
	runningQuery := query + " AND status = 'running'"
	err = database.DB.Get(&runningSessions, runningQuery, args...)
	if err != nil {
		return nil, err
	}
	stats["running_sessions"] = runningSessions

	// Total duration
	var totalDuration sql.NullInt64
	durationQuery := strings.Replace(query, "COUNT(*)", "SUM(duration)", 1)
	err = database.DB.Get(&totalDuration, durationQuery, args...)
	if err != nil {
		return nil, err
	}
	if totalDuration.Valid {
		stats["total_duration"] = totalDuration.Int64
	} else {
		stats["total_duration"] = 0
	}

	// Average duration
	var avgDuration sql.NullFloat64
	avgQuery := strings.Replace(query, "COUNT(*)", "AVG(duration)", 1) + " AND status = 'completed'"
	err = database.DB.Get(&avgDuration, avgQuery, args...)
	if err != nil {
		return nil, err
	}
	if avgDuration.Valid {
		stats["avg_duration"] = avgDuration.Float64
	} else {
		stats["avg_duration"] = 0
	}

	// Max duration
	var maxDuration sql.NullInt64
	maxQuery := strings.Replace(query, "COUNT(*)", "MAX(duration)", 1) + " AND status = 'completed'"
	err = database.DB.Get(&maxDuration, maxQuery, args...)
	if err != nil {
		return nil, err
	}
	if maxDuration.Valid {
		stats["max_duration"] = maxDuration.Int64
	} else {
		stats["max_duration"] = 0
	}

	// Min duration
	var minDuration sql.NullInt64
	minQuery := strings.Replace(query, "COUNT(*)", "MIN(duration)", 1) + " AND status = 'completed' AND duration > 0"
	err = database.DB.Get(&minDuration, minQuery, args...)
	if err != nil {
		return nil, err
	}
	if minDuration.Valid {
		stats["min_duration"] = minDuration.Int64
	} else {
		stats["min_duration"] = 0
	}

	// Daily distribution
	dailyQuery := `
		SELECT DATE(start_time) as date, COUNT(*) as count, SUM(duration) as total_duration
		FROM device_sessions
		WHERE 1=1
	`
	if deviceID != "" {
		dailyQuery += " AND device_id = ?"
	}
	if startDate != "" {
		dailyQuery += " AND DATE(start_time) >= ?"
	}
	if endDate != "" {
		dailyQuery += " AND DATE(start_time) <= ?"
	}
	dailyQuery += " GROUP BY DATE(start_time) ORDER BY date"

	rows, err := database.DB.Query(dailyQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dailyStats := []map[string]interface{}{}
	for rows.Next() {
		var date string
		var count int
		var totalDuration sql.NullInt64
		
		err := rows.Scan(&date, &count, &totalDuration)
		if err != nil {
			return nil, err
		}

		stat := map[string]interface{}{
			"date":  date,
			"count": count,
		}
		if totalDuration.Valid {
			stat["total_duration"] = totalDuration.Int64
		} else {
			stat["total_duration"] = 0
		}
		dailyStats = append(dailyStats, stat)
	}
	stats["daily_distribution"] = dailyStats

	return stats, nil
}