<template>
  <div class="page-container">
    <div class="page-header">
      <h2>系统设置</h2>
    </div>
    
    <el-tabs v-model="activeTab">
      <el-tab-pane label="基本设置" name="basic">
        <el-form :model="settings" label-width="120px" style="max-width: 600px">
          <el-form-item label="系统名称">
            <el-input v-model="settings.systemName" disabled />
          </el-form-item>
          <el-form-item label="会话保留天数">
            <el-input-number v-model="settings.sessionRetentionDays" :min="7" :max="365" />
            <div class="text-gray-500 text-sm mt-1">
              设备运行会话记录的保留时间
            </div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary">保存设置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      
      <el-tab-pane label="IoT平台信息" name="iot">
        <el-alert
          title="IoT平台配置"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 20px"
        >
          IoT平台的连接信息在服务器端环境变量中配置，请联系管理员修改
        </el-alert>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="API地址">已在服务器配置</el-descriptions-item>
          <el-descriptions-item label="设备代码">已在服务器配置</el-descriptions-item>
          <el-descriptions-item label="连接状态">
            <el-tag type="success">正常</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <el-tab-pane label="Webhook接口" name="webhook">
        <el-alert
          title="Webhook接口说明"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 20px"
        >
          IoT平台可以通过以下接口发送设备状态变更通知
        </el-alert>
        
        <el-descriptions :column="1" border style="margin-bottom: 20px">
          <el-descriptions-item label="设备启动">
            <code>POST {{ webhookUrls.start }}?deviceName=[设备ID]</code>
            <div class="text-gray-500 text-sm mt-1">
              请求体: {"power": "on"}
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="设备停止">
            <code>POST {{ webhookUrls.end }}?deviceName=[设备ID]</code>
            <div class="text-gray-500 text-sm mt-1">
              请求体: {"power": "off"}
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="认证方式">
            <el-tag type="success">无需认证</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
      
      <el-tab-pane label="关于" name="about">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="系统版本">1.0.0</el-descriptions-item>
          <el-descriptions-item label="Node.js版本">{{ nodeVersion }}</el-descriptions-item>
          <el-descriptions-item label="数据库类型">SQLite</el-descriptions-item>
          <el-descriptions-item label="开发者">Device Monitor Team</el-descriptions-item>
        </el-descriptions>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'

const activeTab = ref('basic')

const settings = reactive({
  systemName: '设备运行记录监控系统',
  sessionRetentionDays: 30
})

const webhookUrls = computed(() => {
  const baseUrl = window.location.origin
  return {
    start: `${baseUrl}/api/webhooks/device/start`,
    end: `${baseUrl}/api/webhooks/device/end`
  }
})

const nodeVersion = import.meta.env.VITE_NODE_VERSION || 'Unknown'
</script>

<style lang="scss" scoped>
.el-tab-pane {
  padding: 20px 0;
}

code {
  background-color: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

.ml-2 {
  margin-left: 8px;
}

.text-gray-500 {
  color: #909399;
}

.text-sm {
  font-size: 12px;
}

.mt-1 {
  margin-top: 4px;
}
</style>