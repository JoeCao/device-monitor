<template>
  <div class="page-container">
    <div class="page-header">
      <h2>统计分析</h2>
      <p class="page-subtitle">查看设备运行数据的统计分析和趋势</p>
    </div>
    
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters" class="filter-bar">
        <el-form-item label="设备选择">
          <el-select
            v-model="filters.deviceId"
            placeholder="请选择设备"
            clearable
            filterable
            style="width: 200px"
          >
            <el-option
              v-for="device in deviceList"
              :key="device"
              :label="device"
              :value="device"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            :shortcuts="dateShortcuts"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchStatistics" :loading="loading">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="resetFilters">
            <el-icon><RefreshLeft /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
    
    <!-- 统计卡片 -->
    <div v-if="hasData">
      <el-row :gutter="20" class="stat-cards">
        <el-col :span="6">
          <el-card class="stat-card">
            <el-statistic title="总运行次数" :value="statistics.total_sessions">
              <template #prefix>
                <el-icon><Timer /></el-icon>
              </template>
            </el-statistic>
            <div class="stat-footer">
              <span class="stat-label">正在运行</span>
              <span class="stat-value">{{ statistics.running_sessions }}</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <el-statistic title="已完成次数" :value="statistics.completed_sessions">
              <template #prefix>
                <el-icon style="color: #67C23A"><CircleCheck /></el-icon>
              </template>
            </el-statistic>
            <div class="stat-footer">
              <span class="stat-label">完成率</span>
              <span class="stat-value">{{ completionRate }}%</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <el-statistic title="平均运行时长" :value="avgDurationHours">
              <template #suffix>小时</template>
              <template #prefix>
                <el-icon style="color: #409EFF"><Clock /></el-icon>
              </template>
            </el-statistic>
            <div class="stat-footer">
              <span class="stat-label">最长运行</span>
              <span class="stat-value">{{ maxDurationHours.toFixed(2) }}小时</span>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <el-statistic title="总运行时长" :value="totalDurationHours">
              <template #suffix>小时</template>
              <template #prefix>
                <el-icon style="color: #E6A23C"><TrendCharts /></el-icon>
              </template>
            </el-statistic>
            <div class="stat-footer">
              <span class="stat-label">平均每天</span>
              <span class="stat-value">{{ avgDailyHours.toFixed(2) }}小时</span>
            </div>
          </el-card>
        </el-col>
      </el-row>
    
      <!-- 图表区域 -->
      <el-row :gutter="20" class="chart-row">
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span>运行趋势图</span>
            </template>
            <div ref="trendChart" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span>运行时长分布</span>
            </template>
            <div ref="distributionChart" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>
      
      <el-row :gutter="20" class="chart-row">
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span>每日运行统计</span>
            </template>
            <div ref="dailyChart" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span>运行时段分析</span>
            </template>
            <div ref="hourlyChart" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>
    </div>
    
    <!-- 无数据提示 -->
    <el-empty v-else description="暂无统计数据" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { sessionAPI } from '../api'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import dayjs from 'dayjs'
import { 
  Search, 
  RefreshLeft, 
  Timer, 
  CircleCheck, 
  Clock, 
  TrendCharts 
} from '@element-plus/icons-vue'

const loading = ref(false)
const statistics = ref({
  total_sessions: 0,
  completed_sessions: 0,
  running_sessions: 0,
  avg_duration: 0,
  max_duration: 0,
  min_duration: 0,
  total_duration: 0
})

const filters = reactive({
  deviceId: '',
  dateRange: null
})

// Chart refs
const trendChart = ref(null)
const distributionChart = ref(null)
const dailyChart = ref(null)
const hourlyChart = ref(null)

// Chart instances
const chartInstances = {
  trend: null,
  distribution: null,
  daily: null,
  hourly: null
}

// Device list
const deviceList = ref(['6661840236a2fea3cf78b6dd']) // Default device

// Date shortcuts
const dateShortcuts = [
  {
    text: '今天',
    value: () => {
      const today = new Date()
      today.setHours(0, 0, 0, 0)
      const end = new Date(today)
      end.setHours(23, 59, 59, 999)
      return [today, end]
    }
  },
  {
    text: '昨天',
    value: () => {
      const yesterday = new Date()
      yesterday.setDate(yesterday.getDate() - 1)
      yesterday.setHours(0, 0, 0, 0)
      const end = new Date(yesterday)
      end.setHours(23, 59, 59, 999)
      return [yesterday, end]
    }
  },
  {
    text: '过去7天',
    value: () => {
      const end = new Date()
      end.setHours(23, 59, 59, 999)
      const start = new Date()
      start.setDate(start.getDate() - 6)
      start.setHours(0, 0, 0, 0)
      return [start, end]
    }
  },
  {
    text: '过去30天',
    value: () => {
      const end = new Date()
      end.setHours(23, 59, 59, 999)
      const start = new Date()
      start.setDate(start.getDate() - 29)
      start.setHours(0, 0, 0, 0)
      return [start, end]
    }
  },
  {
    text: '本月',
    value: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), 1)
      start.setHours(0, 0, 0, 0)
      const end = new Date(now.getFullYear(), now.getMonth() + 1, 0)
      end.setHours(23, 59, 59, 999)
      return [start, end]
    }
  },
  {
    text: '上月',
    value: () => {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth() - 1, 1)
      start.setHours(0, 0, 0, 0)
      const end = new Date(now.getFullYear(), now.getMonth(), 0)
      end.setHours(23, 59, 59, 999)
      return [start, end]
    }
  }
]

// Chart data storage
const chartData = ref({
  trend: [],
  daily: [],
  hourly: []
})

// Computed values
const hasData = computed(() => statistics.value.total_sessions > 0)

const avgDurationHours = computed(() => {
  return statistics.value.avg_duration ? parseFloat((statistics.value.avg_duration / 3600).toFixed(2)) : 0
})

const totalDurationHours = computed(() => {
  return statistics.value.total_duration ? parseFloat((statistics.value.total_duration / 3600).toFixed(2)) : 0
})

const maxDurationHours = computed(() => {
  return statistics.value.max_duration ? parseFloat((statistics.value.max_duration / 3600).toFixed(2)) : 0
})

const completionRate = computed(() => {
  if (statistics.value.total_sessions === 0) return 0
  return Math.round((statistics.value.completed_sessions / statistics.value.total_sessions) * 100)
})

const avgDailyHours = computed(() => {
  if (!filters.dateRange || totalDurationHours.value === 0) return 0
  
  // 计算实际天数（包含起始和结束日期）
  const startDate = dayjs(filters.dateRange[0]).startOf('day')
  const endDate = dayjs(filters.dateRange[1]).endOf('day')
  const days = endDate.diff(startDate, 'day') + 1
  
  return parseFloat((totalDurationHours.value / days).toFixed(2))
})

const fetchStatistics = async () => {
  if (!filters.deviceId) {
    ElMessage.warning('请选择设备')
    return
  }
  
  loading.value = true
  try {
    const params = {}
    if (filters.dateRange) {
      params.startDate = filters.dateRange[0]
      params.endDate = filters.dateRange[1]
    }
    
    // Fetch basic statistics
    const res = await sessionAPI.getStatistics(filters.deviceId, params)
    statistics.value = res.data
    
    // Fetch detailed data for charts
    await fetchChartData(params)
    
    // Update all charts
    updateAllCharts()
  } catch (error) {
    console.error('Failed to fetch statistics:', error)
    ElMessage.error('获取统计数据失败')
  } finally {
    loading.value = false
  }
}

// Reset filters
const resetFilters = () => {
  filters.deviceId = ''
  filters.dateRange = null
  statistics.value = {
    total_sessions: 0,
    completed_sessions: 0,
    running_sessions: 0,
    avg_duration: 0,
    max_duration: 0,
    min_duration: 0,
    total_duration: 0
  }
  clearAllCharts()
}

// Fetch chart data
const fetchChartData = async (params) => {
  try {
    // Fetch sessions for detailed analysis
    const sessionsRes = await sessionAPI.getList({
      deviceId: filters.deviceId,
      ...params,
      limit: 1000
    })
    
    const sessions = sessionsRes.data || []
    
    // Process data for different charts
    processChartData(sessions)
  } catch (error) {
    console.error('Failed to fetch chart data:', error)
  }
}

// Process chart data
const processChartData = (sessions) => {
  // Initialize data structures
  const dailyData = {}
  const hourlyData = new Array(24).fill(0)
  const durationRanges = {
    '<1h': 0,
    '1-3h': 0,
    '3-6h': 0,
    '6-12h': 0,
    '>12h': 0
  }
  
  // Process each session
  sessions.forEach(session => {
    // Daily statistics
    const date = dayjs(session.start_time).format('YYYY-MM-DD')
    if (!dailyData[date]) {
      dailyData[date] = { count: 0, duration: 0 }
    }
    dailyData[date].count++
    if (session.duration) {
      dailyData[date].duration += session.duration
    }
    
    // Hourly distribution
    const hour = dayjs(session.start_time).hour()
    hourlyData[hour]++
    
    // Duration distribution
    if (session.duration) {
      const hours = session.duration / 3600
      if (hours < 1) durationRanges['<1h']++
      else if (hours <= 3) durationRanges['1-3h']++
      else if (hours <= 6) durationRanges['3-6h']++
      else if (hours <= 12) durationRanges['6-12h']++
      else durationRanges['>12h']++
    }
  })
  
  // Store processed data
  chartData.value = {
    daily: Object.entries(dailyData).map(([date, data]) => ({
      date,
      count: data.count,
      duration: (data.duration / 3600).toFixed(2)
    })).sort((a, b) => a.date.localeCompare(b.date)),
    hourly: hourlyData,
    distribution: durationRanges
  }
}

// Initialize all charts
const initAllCharts = () => {
  console.log('Initializing charts...', {
    trendChart: !!trendChart.value,
    distributionChart: !!distributionChart.value,
    dailyChart: !!dailyChart.value,
    hourlyChart: !!hourlyChart.value
  })
  
  if (trendChart.value && !chartInstances.trend) {
    chartInstances.trend = echarts.init(trendChart.value)
    console.log('Trend chart initialized')
  }
  if (distributionChart.value && !chartInstances.distribution) {
    chartInstances.distribution = echarts.init(distributionChart.value)
    console.log('Distribution chart initialized')
  }
  if (dailyChart.value && !chartInstances.daily) {
    chartInstances.daily = echarts.init(dailyChart.value)
    console.log('Daily chart initialized')
  }
  if (hourlyChart.value && !chartInstances.hourly) {
    chartInstances.hourly = echarts.init(hourlyChart.value)
    console.log('Hourly chart initialized')
  }
}

// Update all charts
const updateAllCharts = () => {
  console.log('Updating all charts with data:', {
    hasDaily: chartData.value.daily.length > 0,
    hasHourly: chartData.value.hourly.length > 0,
    hasDistribution: !!chartData.value.distribution
  })
  
  updateTrendChart()
  updateDistributionChart()
  updateDailyChart()
  updateHourlyChart()
}

// Update trend chart
const updateTrendChart = () => {
  if (!chartInstances.trend || !chartData.value.daily.length) return
  
  const dates = chartData.value.daily.map(d => d.date)
  const counts = chartData.value.daily.map(d => d.count)
  
  const option = {
    title: {
      text: '最近' + dates.length + '天运行趋势',
      left: 'center'
    },
    tooltip: {
      trigger: 'axis',
      formatter: '{b}<br/>运行次数: {c}'
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: {
        rotate: 45
      }
    },
    yAxis: {
      type: 'value',
      name: '运行次数',
      minInterval: 1
    },
    series: [{
      data: counts,
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      itemStyle: {
        color: '#409EFF'
      },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: '#409EFF33' },
            { offset: 1, color: '#409EFF1A' }
          ]
        }
      }
    }],
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true
    }
  }
  
  chartInstances.trend.setOption(option)
}

// Update distribution chart
const updateDistributionChart = () => {
  if (!chartInstances.distribution) return
  
  const data = chartData.value.distribution || {}
  
  const option = {
    title: {
      text: '运行时长分布',
      left: 'center'
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      bottom: '0%',
      orient: 'horizontal'
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: false,
        position: 'center'
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 20,
          fontWeight: 'bold'
        }
      },
      labelLine: {
        show: false
      },
      data: Object.entries(data).map(([name, value]) => ({
        name,
        value
      }))
    }]
  }
  
  chartInstances.distribution.setOption(option)
}

// Update daily chart
const updateDailyChart = () => {
  if (!chartInstances.daily || !chartData.value.daily.length) return
  
  const dates = chartData.value.daily.map(d => d.date)
  const durations = chartData.value.daily.map(d => parseFloat(d.duration))
  
  const option = {
    title: {
      text: '每日运行时长',
      left: 'center'
    },
    tooltip: {
      trigger: 'axis',
      formatter: '{b}<br/>运行时长: {c} 小时'
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLabel: {
        rotate: 45
      }
    },
    yAxis: {
      type: 'value',
      name: '小时'
    },
    series: [{
      data: durations,
      type: 'bar',
      itemStyle: {
        color: '#67C23A',
        borderRadius: [4, 4, 0, 0]
      }
    }],
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true
    }
  }
  
  chartInstances.daily.setOption(option)
}

// Update hourly chart
const updateHourlyChart = () => {
  if (!chartInstances.hourly) return
  
  const hours = Array.from({ length: 24 }, (_, i) => i + '时')
  const data = chartData.value.hourly || []
  
  const option = {
    title: {
      text: '运行时段分布',
      left: 'center'
    },
    tooltip: {
      trigger: 'axis',
      formatter: '{b}<br/>启动次数: {c}'
    },
    xAxis: {
      type: 'category',
      data: hours
    },
    yAxis: {
      type: 'value',
      name: '启动次数',
      minInterval: 1
    },
    series: [{
      data: data,
      type: 'bar',
      itemStyle: {
        color: '#E6A23C',
        borderRadius: [4, 4, 0, 0]
      }
    }],
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    }
  }
  
  chartInstances.hourly.setOption(option)
}

// Clear all charts
const clearAllCharts = () => {
  Object.values(chartInstances).forEach(instance => {
    instance?.clear()
  })
}

// Resize all charts
const resizeCharts = () => {
  Object.values(chartInstances).forEach(instance => {
    instance?.resize()
  })
}

// Load device list
const loadDeviceList = async () => {
  try {
    const res = await sessionAPI.getList({ limit: 100 })
    const sessions = res.data || []
    const devices = [...new Set(sessions.map(s => s.device_id))]
    if (devices.length > 0) {
      deviceList.value = devices
    }
  } catch (error) {
    console.error('Failed to load device list:', error)
  }
}

// Watch for chart container availability
const watchChartsReady = () => {
  const checkInterval = setInterval(() => {
    if (hasData.value && trendChart.value && distributionChart.value && dailyChart.value && hourlyChart.value) {
      clearInterval(checkInterval)
      initAllCharts()
      updateAllCharts()
    }
  }, 100)
  
  // Clear after 5 seconds to prevent memory leak
  setTimeout(() => clearInterval(checkInterval), 5000)
}

onMounted(async () => {
  // Load device list
  await loadDeviceList()
  
  // Add resize listener
  window.addEventListener('resize', resizeCharts)
  
  // Set default date range to last 7 days
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - 6)
  filters.dateRange = [
    dayjs(start).format('YYYY-MM-DD'),
    dayjs(end).format('YYYY-MM-DD')
  ]
  
  // Auto load if has default device
  if (deviceList.value.length > 0) {
    filters.deviceId = deviceList.value[0]
    await fetchStatistics()
    
    // Watch for charts to be ready
    watchChartsReady()
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeCharts)
  Object.values(chartInstances).forEach(instance => {
    instance?.dispose()
  })
})
</script>

<style lang="scss" scoped>
.page-container {
  .page-subtitle {
    color: #909399;
    font-size: 14px;
    margin-top: 5px;
  }
}

.filter-card {
  margin-bottom: 20px;
  
  .filter-bar {
    margin-bottom: 0;
  }
}

.stat-cards {
  margin-bottom: 20px;
  
  .stat-card {
    .el-statistic {
      text-align: center;
    }
    
    .stat-footer {
      margin-top: 10px;
      padding-top: 10px;
      border-top: 1px solid #EBEEF5;
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 14px;
      
      .stat-label {
        color: #909399;
      }
      
      .stat-value {
        color: #606266;
        font-weight: 500;
      }
    }
  }
}

.chart-row {
  margin-bottom: 20px;
  
  &:last-child {
    margin-bottom: 0;
  }
}

.chart-card {
  height: 100%;
}

.chart-container {
  width: 100%;
  height: 400px;
}

.el-empty {
  padding: 40px 0;
}
</style>