<template>
  <div class="page-container" v-loading.fullscreen.lock="loadingData" element-loading-text="正在加载数据...">
    <div class="page-header">
      <el-page-header @back="goBack">
        <template #content>
          <span class="text-large font-600">设备运行报告</span>
        </template>
      </el-page-header>
    </div>
    
    <el-skeleton :loading="loading" animated>
      <template #template>
        <el-skeleton-item variant="h3" style="width: 30%" />
        <el-skeleton-item variant="text" style="margin-top: 10px" />
        <el-skeleton-item variant="text" />
        <el-skeleton-item variant="rect" style="width: 100%; height: 400px; margin-top: 20px" />
      </template>
      
      <template #default>
        <div v-if="session" class="report-content">
          <!-- Session Info -->
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>基本信息</span>
                <div class="header-actions">
                  <el-tag v-if="session.status === 'running'" type="success" effect="plain">
                    <el-icon class="is-loading"><Loading /></el-icon>
                    自动刷新中
                  </el-tag>
                  <el-button
                    type="primary"
                    size="small"
                    :loading="refreshing"
                    @click="manualRefresh"
                  >
                    <el-icon><Refresh /></el-icon>
                    刷新数据
                  </el-button>
                </div>
              </div>
            </template>
            
            <el-descriptions :column="2" border>
              <el-descriptions-item label="设备ID">{{ session.device_id }}</el-descriptions-item>
              <el-descriptions-item label="会话ID">{{ session.session_id }}</el-descriptions-item>
              <el-descriptions-item label="开始时间">{{ session.start_time }}</el-descriptions-item>
              <el-descriptions-item label="结束时间">{{ session.end_time || '-' }}</el-descriptions-item>
              <el-descriptions-item label="运行时长">
                <span :class="{ 'status-running': session.status === 'running' }">
                  {{ formatDuration(session.duration) }}
                </span>
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="session.status === 'running' ? 'success' : 'info'">
                  {{ session.status === 'running' ? '运行中' : '已完成' }}
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
          </el-card>
          
          <!-- IoT Data Charts -->
          <div v-if="iotData && Object.keys(iotData.aggregated).length > 0" class="charts-container">
            <!-- 希尔伯特值热力图 -->
            <el-card v-if="iotData.aggregated['feature_hilbert_2_hb']" class="chart-card">
              <template #header>
                <span>振动功率谱密度热力图 (希尔伯特包络)</span>
              </template>
              <div class="stats-container">
                <el-row :gutter="20">
                  <el-col :span="8">
                    <el-statistic title="数据点数" :value="iotData.aggregated['feature_hilbert_2_hb'].summary.count" />
                  </el-col>
                  <el-col :span="8">
                    <div class="stat-info">
                      <div class="stat-title">频率范围</div>
                      <div class="stat-value">0-500 Hz <span class="stat-suffix">(128段)</span></div>
                    </div>
                  </el-col>
                  <el-col :span="8">
                    <div class="stat-info">
                      <div class="stat-title">分析类型</div>
                      <div class="stat-value">功率谱密度(PSD)</div>
                    </div>
                  </el-col>
                </el-row>
              </div>
              <div ref="hilbertChart" class="chart-container"></div>
            </el-card>
            
            <!-- 噪音图表 -->
            <el-card v-if="iotData.aggregated['volume']" class="chart-card">
              <template #header>
                <span>噪音趋势图</span>
              </template>
              <div class="stats-container">
                <el-row :gutter="20">
                  <el-col :span="6">
                    <el-statistic title="平均值" :value="parseFloat(iotData.aggregated['volume'].summary.avg_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['volume'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="最大值" :value="parseFloat(iotData.aggregated['volume'].summary.max_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['volume'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="最小值" :value="parseFloat(iotData.aggregated['volume'].summary.min_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['volume'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="数据点数" :value="iotData.aggregated['volume'].summary.count" />
                  </el-col>
                </el-row>
              </div>
              <div ref="volumeChart" class="chart-container"></div>
            </el-card>
            
            <!-- 振动图表 -->
            <el-card v-if="iotData.aggregated['shake']" class="chart-card">
              <template #header>
                <span>振动趋势图</span>
              </template>
              <div class="stats-container">
                <el-row :gutter="20">
                  <el-col :span="6">
                    <el-statistic title="平均值" :value="parseFloat(iotData.aggregated['shake'].summary.avg_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['shake'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="最大值" :value="parseFloat(iotData.aggregated['shake'].summary.max_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['shake'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="最小值" :value="parseFloat(iotData.aggregated['shake'].summary.min_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['shake'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="数据点数" :value="iotData.aggregated['shake'].summary.count" />
                  </el-col>
                </el-row>
              </div>
              <div ref="shakeChart" class="chart-container"></div>
            </el-card>
            
            <!-- 转速图表 -->
            <el-card v-if="iotData.aggregated['feature_speed_1_speed']" class="chart-card">
              <template #header>
                <span>转速趋势图</span>
              </template>
              <div class="stats-container">
                <el-row :gutter="20">
                  <el-col :span="6">
                    <el-statistic title="平均值" :value="parseFloat(iotData.aggregated['feature_speed_1_speed'].summary.avg_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['feature_speed_1_speed'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="最大值" :value="parseFloat(iotData.aggregated['feature_speed_1_speed'].summary.max_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['feature_speed_1_speed'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="最小值" :value="parseFloat(iotData.aggregated['feature_speed_1_speed'].summary.min_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['feature_speed_1_speed'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="数据点数" :value="iotData.aggregated['feature_speed_1_speed'].summary.count" />
                  </el-col>
                </el-row>
              </div>
              <div ref="speedChart" class="chart-container"></div>
            </el-card>
            
            <!-- 温度图表 -->
            <el-card v-if="iotData.aggregated['temperature']" class="chart-card">
              <template #header>
                <span>温度趋势图</span>
              </template>
              <div class="stats-container">
                <el-row :gutter="20">
                  <el-col :span="6">
                    <el-statistic title="平均值" :value="parseFloat(iotData.aggregated['temperature'].summary.avg_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['temperature'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="最大值" :value="parseFloat(iotData.aggregated['temperature'].summary.max_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['temperature'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="最小值" :value="parseFloat(iotData.aggregated['temperature'].summary.min_value.toFixed(2))">
                      <template #suffix>{{ iotData.aggregated['temperature'].summary.unit }}</template>
                    </el-statistic>
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="数据点数" :value="iotData.aggregated['temperature'].summary.count" />
                  </el-col>
                </el-row>
              </div>
              <div ref="temperatureChart" class="chart-container"></div>
            </el-card>
          </div>
          
          <!-- No Data -->
          <el-card v-else-if="!loading" class="empty-card">
            <el-empty description="暂无IoT数据">
              <el-button type="primary" @click="syncIotData">同步数据</el-button>
            </el-empty>
          </el-card>
        </div>
      </template>
    </el-skeleton>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Loading } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import dayjs from 'dayjs'
import { sessionAPI, iotAPI } from '../api'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const loadingData = ref(false)
const refreshing = ref(false)
const session = ref(null)
const iotData = ref(null)

// Chart refs
const hilbertChart = ref(null)
const volumeChart = ref(null)
const shakeChart = ref(null)
const speedChart = ref(null)
const temperatureChart = ref(null)

// Chart instances
const chartInstances = {
  hilbert: null,
  volume: null,
  shake: null,
  speed: null,
  temperature: null
}

// Auto refresh timer
let autoRefreshTimer = null

// Format duration
const formatDuration = (seconds) => {
  if (!seconds && session.value?.status === 'running') {
    // Calculate current duration for running sessions
    const start = dayjs(session.value.start_time)
    const now = dayjs()
    seconds = now.diff(start, 'second')
  }
  if (!seconds) return '-'
  
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  return `${hours}小时${minutes}分${secs}秒`
}

// Get all time points for x-axis alignment
const allTimePoints = computed(() => {
  if (!iotData.value || !iotData.value.aggregated) return []
  
  const timeSet = new Set()
  Object.values(iotData.value.aggregated).forEach(pointData => {
    if (pointData.timeSeries) {
      pointData.timeSeries.forEach(item => {
        timeSet.add(dayjs(item.time_bucket).format('HH:mm:ss'))
      })
    }
  })
  
  return Array.from(timeSet).sort()
})

// Go back
const goBack = () => {
  router.push('/')
}

// Fetch session report
const fetchSessionReport = async () => {
  loading.value = true
  try {
    const sessionId = route.params.id
    const res = await sessionAPI.getReport(sessionId)
    
    session.value = res.data.session
    iotData.value = res.data.iotData
    
    // Auto load IoT data after session info loaded
    await refreshIotData(false)
    
    // Setup auto refresh if session is running
    if (session.value?.status === 'running') {
      setupAutoRefresh()
    }
  } catch (error) {
    console.error('Failed to fetch session report:', error)
    ElMessage.error('获取报告失败')
  } finally {
    loading.value = false
  }
}

// Refresh IoT data (real-time query)
const refreshIotData = async (showMessage = true) => {
  if (!showMessage) {
    loadingData.value = true
  }
  refreshing.value = true
  
  try {
    const pointsResponse = await iotAPI.getDevicePoints(session.value.device_id)
    const points = pointsResponse.data || []
    
    if (!Array.isArray(points) || points.length === 0) {
      if (showMessage) {
        ElMessage.warning('未找到设备数据点')
      }
      return
    }
    
    // 只查询我们关心的数据点
    const targetPoints = ['feature_hilbert_2_hb', 'volume', 'shake', 'feature_speed_1_speed', 'temperature']
    const availablePoints = points
      .map(p => p.name || p.point_name || p)
      .filter(name => targetPoints.includes(name))
    
    const res = await iotAPI.sync(session.value.session_id, {
      pointNames: availablePoints
    })
    
    if (showMessage) {
      ElMessage.success(`查询成功，获取到 ${res.dataCount} 条实时数据`)
    }
    
    if (res.data && res.data.length > 0) {
      processRealtimeData(res.data)
    }
  } catch (error) {
    console.error('Failed to refresh IoT data:', error)
    if (showMessage) {
      ElMessage.error('刷新数据失败')
    }
  } finally {
    refreshing.value = false
    loadingData.value = false
  }
}

// Manual refresh
const manualRefresh = async () => {
  // Refresh session info first
  await refreshSessionInfo()
  
  // Then refresh IoT data
  await refreshIotData(true)
  
  // If session is running and auto refresh is not set up, set it up
  if (session.value?.status === 'running' && !autoRefreshTimer) {
    setupAutoRefresh()
  }
}

// Setup auto refresh
const setupAutoRefresh = () => {
  // Clear existing timer
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
  }
  
  // Set up new timer - refresh every 5 seconds
  autoRefreshTimer = setInterval(async () => {
    if (session.value?.status === 'running') {
      // Refresh session info first
      await refreshSessionInfo()
      
      // If session is still running after refresh, update IoT data
      if (session.value?.status === 'running') {
        await refreshIotData(false)
      } else {
        // Session has ended, stop auto refresh
        clearInterval(autoRefreshTimer)
        autoRefreshTimer = null
        ElMessage.info('会话已结束，停止自动刷新')
      }
    } else {
      // Stop auto refresh if session is no longer running
      clearInterval(autoRefreshTimer)
      autoRefreshTimer = null
    }
  }, 5000)
}

// Refresh session info only
const refreshSessionInfo = async () => {
  try {
    const sessionId = route.params.id
    const res = await sessionAPI.get(sessionId)
    
    if (res.data) {
      // Update session info
      session.value = res.data
      
      // If session has ended, update duration
      if (res.data.status === 'completed' && res.data.duration) {
        session.value.duration = res.data.duration
      }
    }
  } catch (error) {
    console.error('Failed to refresh session info:', error)
  }
}

// Process realtime data for charts
const processRealtimeData = (data) => {
  // 按数据点分组
  const groupedData = {}
  data.forEach(item => {
    if (!groupedData[item.pointName]) {
      groupedData[item.pointName] = {
        point_name: item.pointName,
        unit: item.unit,
        data: []
      }
    }
    groupedData[item.pointName].data.push({
      timestamp: item.timestamp,
      value: item.pointValue
    })
  })
  
  // 转换为 iotData 格式
  const points = Object.values(groupedData)
  const aggregated = {}
  
  points.forEach(point => {
    // 对于希尔伯特值，特殊处理
    if (point.point_name === 'feature_hilbert_2_hb') {
      // 希尔伯特值不计算统计信息，直接使用原始数据
      const timeSeries = point.data
        .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
        .map(d => ({
          time_bucket: d.timestamp,
          avg_value: d.value // 保持原始字符串格式
        }))
      
      aggregated[point.point_name] = {
        summary: {
          unit: point.unit,
          avg_value: 0,
          max_value: 0,
          min_value: 0,
          count: point.data.length
        },
        timeSeries
      }
    } else {
      // 其他数据点正常处理
      const values = point.data.map(d => typeof d.value === 'number' ? d.value : parseFloat(d.value))
      const avgValue = values.reduce((a, b) => a + b, 0) / values.length
      const maxValue = Math.max(...values)
      const minValue = Math.min(...values)
      
      // 按时间排序
      const timeSeries = point.data
        .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp))
        .map(d => ({
          time_bucket: d.timestamp,
          avg_value: typeof d.value === 'number' ? d.value : parseFloat(d.value)
        }))
      
      aggregated[point.point_name] = {
        summary: {
          unit: point.unit,
          avg_value: avgValue,
          max_value: maxValue,
          min_value: minValue,
          count: values.length
        },
        timeSeries
      }
    }
  })
  
  iotData.value = {
    points,
    aggregated
  }
  
  // 更新所有图表
  nextTick(() => {
    initAllCharts()
    updateAllCharts()
  })
}

// Initialize all charts
const initAllCharts = () => {
  // Initialize Hilbert chart
  if (hilbertChart.value && !chartInstances.hilbert) {
    chartInstances.hilbert = echarts.init(hilbertChart.value)
  }
  
  // Initialize Volume chart
  if (volumeChart.value && !chartInstances.volume) {
    chartInstances.volume = echarts.init(volumeChart.value)
  }
  
  // Initialize Shake chart
  if (shakeChart.value && !chartInstances.shake) {
    chartInstances.shake = echarts.init(shakeChart.value)
  }
  
  // Initialize Speed chart
  if (speedChart.value && !chartInstances.speed) {
    chartInstances.speed = echarts.init(speedChart.value)
  }
  
  // Initialize Temperature chart
  if (temperatureChart.value && !chartInstances.temperature) {
    chartInstances.temperature = echarts.init(temperatureChart.value)
  }
}

// Update all charts
const updateAllCharts = () => {
  if (!iotData.value) return
  
  // Update Hilbert heatmap
  if (chartInstances.hilbert && iotData.value.aggregated['feature_hilbert_2_hb']) {
    updateHilbertHeatmap(iotData.value.aggregated['feature_hilbert_2_hb'].timeSeries)
  }
  
  // Update Volume chart
  if (chartInstances.volume && iotData.value.aggregated['volume']) {
    updateLineChart('volume', '噪音', chartInstances.volume, iotData.value.aggregated['volume'])
  }
  
  // Update Shake chart
  if (chartInstances.shake && iotData.value.aggregated['shake']) {
    updateLineChart('shake', '振动', chartInstances.shake, iotData.value.aggregated['shake'])
  }
  
  // Update Speed chart
  if (chartInstances.speed && iotData.value.aggregated['feature_speed_1_speed']) {
    updateLineChart('feature_speed_1_speed', '转速', chartInstances.speed, iotData.value.aggregated['feature_speed_1_speed'])
  }
  
  // Update Temperature chart
  if (chartInstances.temperature && iotData.value.aggregated['temperature']) {
    updateLineChart('temperature', '温度', chartInstances.temperature, iotData.value.aggregated['temperature'])
  }
}

// Update line chart for normal data points
const updateLineChart = (pointName, displayName, chartInstance, pointData) => {
  const data = pointData.timeSeries
  const xData = data.map(item => dayjs(item.time_bucket).format('HH:mm:ss'))
  const yData = data.map(item => item.avg_value)
  
  // Define colors for different data points
  const colors = {
    volume: '#67C23A',
    shake: '#E6A23C',
    feature_speed_1_speed: '#409EFF',
    temperature: '#F56C6C'
  }
  
  const color = colors[pointName] || '#409EFF'
  
  const option = {
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        const point = params[0]
        return `${point.axisValue}<br>${point.seriesName}: ${point.value.toFixed(2)} ${pointData.summary.unit}`
      }
    },
    xAxis: {
      type: 'category',
      data: xData,
      axisLabel: {
        rotate: 45,
        interval: Math.floor(xData.length / 10) // Show about 10 labels
      }
    },
    yAxis: {
      type: 'value',
      name: pointData.summary.unit
    },
    series: [
      {
        name: displayName,
        type: 'line',
        data: yData,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        itemStyle: {
          color: color
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: `${color}33` }, // 20% opacity in hex
              { offset: 1, color: `${color}1A` }  // 10% opacity in hex
            ]
          }
        }
      }
    ],
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true
    }
  }
  
  try {
    chartInstance.setOption(option, true)
  } catch (e) {
    console.error('设置图表选项失败:', e)
  }
}

// Update Hilbert heatmap
const updateHilbertHeatmap = (data) => {
  const chartInstance = chartInstances.hilbert
  if (!chartInstance) return
  // 准备热力图数据
  const timePoints = data.map(item => dayjs(item.time_bucket).format('HH:mm:ss'))
  
  // 生成频率范围（0-500Hz分为128份，根据奈奎斯特定理）
  const frequencyBands = 128
  const maxFrequency = 500  // 奈奎斯特频率 = 采样率(1000Hz) / 2
  const frequencyStep = maxFrequency / frequencyBands
  const frequencies = Array.from({ length: frequencyBands }, (_, i) => 
    Math.round(i * frequencyStep)
  )
  
  // 将希尔伯特值转换为热力图数据
  const heatmapData = []
  let maxAmplitude = -10
  let minAmplitude = 10
  
  data.forEach((item, timeIndex) => {
    try {
      // 解析希尔伯特值（它是一个JSON字符串数组）
      const hilbertArray = typeof item.avg_value === 'string' 
        ? JSON.parse(item.avg_value) 
        : item.avg_value
      
      // 确保是数组且长度正确
      if (Array.isArray(hilbertArray)) {
        hilbertArray.forEach((amplitude, freqIndex) => {
          if (freqIndex < frequencyBands && typeof amplitude === 'number') {
            // 使用对数尺度来更好地显示小数值
            const logAmplitude = amplitude > 0 ? Math.log10(amplitude) : -10
            heatmapData.push([timeIndex, freqIndex, logAmplitude])
            maxAmplitude = Math.max(maxAmplitude, logAmplitude)
            minAmplitude = Math.min(minAmplitude, logAmplitude)
          }
        })
      }
    } catch (e) {
      console.error('解析希尔伯特值失败:', e, item.avg_value)
    }
  })
  
  // 确保有有效的范围
  if (heatmapData.length === 0) {
    console.warn('没有有效的希尔伯特数据')
    maxAmplitude = 0
    minAmplitude = -10
  }
  
  const option = {
    title: {
      text: '振动功率谱密度热力图 (希尔伯特包络)',
      left: 'center'
    },
    tooltip: {
      position: 'top',
      formatter: (params) => {
        const [timeIdx, freqIdx, logValue] = params.data
        const actualValue = Math.pow(10, logValue)
        return `时间: ${timePoints[timeIdx]}<br>频率: ${frequencies[freqIdx]}Hz<br>幅值: ${actualValue.toExponential(2)}<br>对数值: ${logValue.toFixed(2)}`
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: timePoints,
      axisLabel: {
        rotate: 45,
        interval: Math.floor(timePoints.length / 10) // 显示10个标签
      }
    },
    yAxis: {
      type: 'category',
      data: frequencies.map(f => `${f}Hz`),
      axisLabel: {
        interval: 4 // 每5个显示一个
      }
    },
    visualMap: {
      min: minAmplitude,
      max: maxAmplitude,
      calculable: true,
      orient: 'vertical',
      right: '2%',
      top: '20%',
      text: ['高', '低'],
      textStyle: {
        color: '#333'
      },
      inRange: {
        color: ['#313695', '#4575b4', '#74add1', '#abd9e9', '#e0f3f8', 
                '#ffffbf', '#fee090', '#fdae61', '#f46d43', '#d73027', '#a50026']
      },
      formatter: (value) => {
        // 将对数值转换回实际值显示
        return `10^${value.toFixed(1)}`
      }
    },
    series: [{
      name: '希尔伯特频谱',
      type: 'heatmap',
      data: heatmapData,
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      }
    }]
  }
  
  try {
    chartInstance.setOption(option, true)
  } catch (e) {
    console.error('设置图表选项失败:', e)
  }
}

// Get point display name
const getPointDisplayName = (pointName) => {
  const pointsMap = {
    'volume': '噪音',
    'shake': '振动',
    'temperature': '温度',
    'feature_speed_1_speed': '转速',
    'controlledvariable': '是否在运行',
    'controlledvolume': '音量是否监控',
    'feature_hilbert_2_hb': '希尔伯特值'
  }
  return pointsMap[pointName] || pointName
}

// Resize all charts
const resizeCharts = () => {
  Object.values(chartInstances).forEach(instance => {
    instance?.resize()
  })
}

// Update running session duration
let durationTimer = null
const startDurationTimer = () => {
  if (session.value?.status === 'running') {
    durationTimer = setInterval(() => {
      // Force re-render to update duration
      session.value = { ...session.value }
    }, 1000)
  }
}

onMounted(() => {
  fetchSessionReport()
  window.addEventListener('resize', resizeCharts)
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeCharts)
  
  // Dispose all chart instances
  Object.values(chartInstances).forEach(instance => {
    instance?.dispose()
  })
  
  // Clear timers
  if (durationTimer) {
    clearInterval(durationTimer)
  }
  
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
  }
})

// Start timer after data loaded
watch(session, (newVal) => {
  if (newVal) {
    startDurationTimer()
  }
})

// Watch for chart containers to be ready
watch([hilbertChart, volumeChart, shakeChart, speedChart, temperatureChart], () => {
  nextTick(() => {
    initAllCharts()
    if (iotData.value) {
      updateAllCharts()
    }
  })
})
</script>

<style lang="scss" scoped>
.report-content {
  .info-card {
    margin-bottom: 20px;
  }
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    
    .header-actions {
      display: flex;
      align-items: center;
      gap: 10px;
      
      .is-loading {
        animation: rotating 2s linear infinite;
      }
    }
  }
  
  .charts-container {
    .chart-card {
      margin-bottom: 20px;
    }
  }
  
  .stats-container {
    margin-bottom: 20px;
    text-align: center;
    
    .stat-info {
      text-align: center;
      
      .stat-title {
        color: #909399;
        font-size: 14px;
        margin-bottom: 8px;
      }
      
      .stat-value {
        color: #303133;
        font-size: 20px;
        font-weight: 500;
        
        .stat-suffix {
          font-size: 14px;
          color: #909399;
          font-weight: normal;
        }
      }
    }
  }
  
  .chart-container {
    width: 100%;
    height: 400px;
  }
  
  .empty-card {
    text-align: center;
    padding: 40px 0;
  }
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>