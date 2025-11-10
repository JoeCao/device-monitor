<template>
  <div class="page-container">
    <div class="page-header">
      <h2>设备运行记录</h2>
    </div>
    
    <!-- Filter Bar -->
    <div class="filter-bar">
      <el-form :inline="true" :model="filters">
        <el-form-item label="设备ID">
          <el-input
            v-model="filters.deviceId"
            placeholder="请输入设备ID"
            clearable
            @clear="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable>
            <el-option label="运行中" value="running" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>
    
    <!-- Data Table -->
    <el-table
      v-loading="loading"
      :data="sessions"
      stripe
      style="width: 100%"
    >
      <el-table-column prop="device_id" label="设备ID" width="180" />
      <el-table-column prop="session_id" label="会话ID" width="300" show-overflow-tooltip />
      <el-table-column prop="start_time" label="开始时间" width="180" />
      <el-table-column prop="end_time" label="结束时间" width="180">
        <template #default="{ row }">
          {{ row.end_time || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="duration" label="运行时长" width="120">
        <template #default="{ row }">
          {{ formatDuration(row.duration) }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag
            :type="row.status === 'running' ? 'success' : 'info'"
            size="small"
          >
            {{ row.status === 'running' ? '运行中' : '已完成' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="200">
        <template #default="{ row }">
          <el-button
            type="primary"
            link
            @click="viewDetail(row)"
          >
            查看报告
          </el-button>
          <el-button
            v-if="row.status === 'running'"
            type="warning"
            link
            @click="endSession(row)"
          >
            结束运行
          </el-button>
          <el-button
            type="danger"
            link
            @click="deleteSession(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- Pagination -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { sessionAPI } from '../api'
import dayjs from 'dayjs'

const router = useRouter()
const loading = ref(false)
const sessions = ref([])

const filters = reactive({
  deviceId: '',
  status: '',
  dateRange: null
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// Format duration
const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  return `${hours}小时${minutes}分${secs}秒`
}

// Fetch sessions
const fetchSessions = async () => {
  loading.value = true
  try {
    const params = {
      limit: pagination.pageSize,
      offset: (pagination.page - 1) * pagination.pageSize,
      deviceId: filters.deviceId || undefined,
      status: filters.status || undefined,
      startDate: filters.dateRange?.[0] || undefined,
      endDate: filters.dateRange?.[1] || undefined
    }
    
    const res = await sessionAPI.getList(params)
    sessions.value = res.data
    pagination.total = res.pagination.total
  } catch (error) {
    console.error('Failed to fetch sessions:', error)
  } finally {
    loading.value = false
  }
}

// Search
const handleSearch = () => {
  pagination.page = 1
  fetchSessions()
}

// Reset filters
const handleReset = () => {
  filters.deviceId = ''
  filters.status = ''
  filters.dateRange = null
  handleSearch()
}

// Pagination
const handlePageChange = () => {
  fetchSessions()
}

const handleSizeChange = () => {
  pagination.page = 1
  fetchSessions()
}

// View detail
const viewDetail = (row) => {
  router.push(`/session/${row.session_id}`)
}

// End session
const endSession = async (row) => {
  try {
    await ElMessageBox.confirm('确定要结束此设备运行吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await webhookAPI.end(row.device_id, row.session_id)
    ElMessage.success('设备运行已结束')
    fetchSessions()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to end session:', error)
    }
  }
}

// Delete session
const deleteSession = async (row) => {
  try {
    await ElMessageBox.confirm('删除后数据将无法恢复，确定要删除吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await sessionAPI.delete(row.session_id)
    ElMessage.success('删除成功')
    fetchSessions()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete session:', error)
    }
  }
}


onMounted(() => {
  fetchSessions()
})
</script>

<style lang="scss" scoped>
.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.test-controls {
  margin-top: 40px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 4px;
}
</style>