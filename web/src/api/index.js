import axios from 'axios'
import { ElMessage } from 'element-plus'

// Create axios instance
const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  config => {
    // Add auth token if needed
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    const message = error.response?.data?.error || '请求失败'
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// Session APIs
export const sessionAPI = {
  // Get session list
  getList(params) {
    return api.get('/sessions', { params })
  },
  
  // Get session by ID
  get(sessionId) {
    return api.get(`/sessions/${sessionId}`)
  },
  
  // Get session report
  getReport(sessionId) {
    return api.get(`/sessions/${sessionId}/report`)
  },
  
  // Get device statistics
  getStatistics(deviceId, params) {
    return api.get(`/sessions/device/${deviceId}/statistics`, { params })
  },
  
  // Delete session
  delete(sessionId) {
    return api.delete(`/sessions/${sessionId}`)
  }
}

// IoT APIs
export const iotAPI = {
  // Query IoT data
  query(data) {
    return api.post('/iot/query', data)
  },
  
  // Get IoT data for session
  getSessionData(sessionId, params) {
    return api.get(`/iot/session/${sessionId}`, { params })
  },
  
  // Get device points
  getDevicePoints(deviceId) {
    return api.get(`/iot/device/${deviceId}/points`)
  },
  
  // Sync IoT data
  sync(sessionId, data) {
    return api.post(`/iot/sync/${sessionId}`, data)
  }
}

// Webhook APIs (for testing)
export const webhookAPI = {
  // Trigger device start
  start(deviceId, data = {}) {
    return api.post('/webhooks/device/start', {
      deviceId,
      ...data
    }, {
      headers: {
        'X-Webhook-Token': 'test-token'
      }
    })
  },
  
  // Trigger device end
  end(deviceId, sessionId, data = {}) {
    return api.post('/webhooks/device/end', {
      deviceId,
      sessionId,
      ...data
    }, {
      headers: {
        'X-Webhook-Token': 'test-token'
      }
    })
  }
}

export default api