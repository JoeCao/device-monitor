<template>
  <el-container class="app-container">
    <el-header class="app-header">
      <div class="header-content">
        <h1 class="app-title">设备运行记录监控系统</h1>
        <div class="header-info">
          <el-tag type="success">系统运行中</el-tag>
        </div>
      </div>
    </el-header>
    
    <el-container>
      <el-aside width="200px" class="app-aside">
        <el-menu
          :default-active="$route.path"
          router
          class="aside-menu"
        >
          <el-menu-item index="/">
            <el-icon><Monitor /></el-icon>
            <span>运行记录</span>
          </el-menu-item>
          <el-menu-item index="/statistics">
            <el-icon><DataAnalysis /></el-icon>
            <span>统计分析</span>
          </el-menu-item>
          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { Monitor, DataAnalysis, Setting } from '@element-plus/icons-vue'
</script>

<style lang="scss">
.app-container {
  height: 100vh;
  
  .app-header {
    background-color: #fff;
    border-bottom: 1px solid #e6e6e6;
    padding: 0 20px;
    
    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      height: 100%;
      
      .app-title {
        margin: 0;
        font-size: 20px;
        color: #303133;
      }
    }
  }
  
  .app-aside {
    background-color: #f5f7fa;
    border-right: 1px solid #e6e6e6;
    
    .aside-menu {
      border: none;
      background-color: transparent;
    }
  }
  
  .app-main {
    background-color: #f5f7fa;
    padding: 20px;
  }
}
</style>