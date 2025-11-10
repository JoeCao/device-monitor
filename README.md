# Device Operation Monitor (Go Version)

设备运行记录监控系统的 Go 语言实现版本，提供与 Node.js 版本完全相同的功能，但具有更简单的部署方式和更好的性能。

## 项目完成情况总结

### ✅ 已完成功能

1. **核心功能实现**
   - 完整的 Webhook 接收设备启停事件功能
   - 设备运行会话管理（创建、结束、查询、删除）
   - SQLite 数据库集成，与 Node.js 版本数据库结构完全兼容
   - 支持数据库文件直接迁移使用

2. **IoT 平台集成**
   - OAuth2 认证机制实现
   - 实时数据查询（温度、振动、噪音、转速、希尔伯特包络等）
   - 支持运行中和已完成会话的数据查询
   - 数据不在本地存储，实时从 IoT 平台获取

3. **前端集成**
   - 完全复用原 Vue 3 前端代码
   - 生产模式下使用 embed 嵌入前端资源
   - 开发模式下代理到 Vite 开发服务器
   - 正确处理所有静态文件 MIME 类型

4. **API 完整性**
   - 所有 API 端点与 Node.js 版本保持一致
   - 响应格式完全兼容，前端无需任何修改
   - 新增了缺失的 `/api/iot/device/:deviceId/points` 端点

5. **统计功能**
   - 设备运行统计（总会话数、完成数、运行中数）
   - 运行时长统计（平均、最大、最小）
   - 每日运行分布数据

### 🔧 已解决的问题

1. **数据库兼容性问题**
   - 修复：NULL duration 字段导致的扫描错误
   - 解决方案：使用 `sql.NullInt64` 类型处理可空字段

2. **API 响应格式问题**
   - 修复：统计数据字段名不匹配（驼峰式改为下划线式）
   - 修复：IoT 数据同步响应格式不匹配
   - 修复：时间戳格式问题（Unix 毫秒改为格式化字符串）

3. **IoT 平台集成问题**
   - 修复：认证端点和请求格式
   - 修复：响应数据解析（`timestamp` → `time`）
   - 修复：移除了"仅已完成会话可同步"的限制

4. **静态文件服务问题**
   - 修复：JavaScript 模块的 MIME 类型设置
   - 修复：生产/开发模式的环境变量读取

### ⚠️ 已知问题和注意事项

1. **性能相关**
   - 前端自动刷新间隔（5秒）在数据量大时可能导致性能问题
   - 希尔伯特包络数据处理需要较多计算资源

2. **配置相关**
   - 设备 ID 必须与 IoT 平台上的实际设备匹配
   - 错误的设备 ID 会导致"设备不存在"错误
   - `.env` 文件中的 `NODE_ENV` 必须设置为 `production` 才能使用嵌入的前端资源

3. **部署相关**
   - 生产环境部署前必须先构建前端（`cd web && npm run build`）
   - 二进制文件需要 CGO 支持（SQLite 驱动需要）

## 特性

- ✅ 完全兼容原 Node.js 版本的所有功能
- 🚀 单二进制文件部署，无需安装运行时环境
- 💾 嵌入式前端资源，生产环境无需分离部署
- 🔧 保持相同的 API 接口，前端代码无需修改
- 📊 支持所有原有的 IoT 数据可视化功能
- 🔌 Webhook 接口完全兼容

## 技术栈

- **后端**: Go + Gin + SQLite
- **前端**: Vue 3 + Vite + Element Plus + ECharts（复用原项目）
- **数据库**: SQLite（嵌入式）

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+（仅开发时需要）
- Make（可选）

### 安装依赖

```bash
# 使用 Make
make deps

# 或手动安装
go mod download
cd web && npm install
```

### 配置环境变量

复制并修改 `.env` 文件：

```bash
cp .env.example .env
```

主要配置项：
- `IOT_APP_KEY` - IoT 平台应用密钥
- `IOT_APP_SECRET` - IoT 平台应用密钥
- `IOT_DEVICE_CODE` - 默认设备代码

### 开发模式

```bash
# 同时启动前后端开发服务器
make dev

# 或手动启动
# 终端1: 启动前端
cd web && npm run dev

# 终端2: 启动后端
go run main.go
```

- 前端访问: http://localhost:5173
- 后端 API: http://localhost:3000

### 生产构建

```bash
# 构建生产版本
make build

# 运行生产版本
make run

# 或直接运行二进制文件
NODE_ENV=production ./device-monitor
```

## API 兼容性

本项目保持与原 Node.js 版本完全相同的 API 接口：

### Webhook 接口
- `POST /api/webhooks/device/start?deviceName={deviceId}`
- `POST /api/webhooks/device/end?deviceName={deviceId}`

### 会话管理
- `GET /api/sessions` - 获取会话列表
- `GET /api/sessions/:id` - 获取会话详情
- `GET /api/sessions/:id/report` - 获取完整报告
- `DELETE /api/sessions/:id` - 删除会话
- `GET /api/sessions/statistics` - 获取统计信息

### IoT 集成
- `POST /api/iot/sync/:sessionId` - 同步 IoT 数据
- `GET /api/iot/data-points` - 获取数据点配置

## 部署优势

相比 Node.js 版本，Go 版本具有以下部署优势：

1. **单文件部署**: 编译后只需要一个二进制文件
2. **无需运行时**: 不需要安装 Node.js 或 npm
3. **内存占用小**: 相比 Node.js 减少约 50% 内存使用
4. **启动速度快**: 秒级启动，无需加载大量模块
5. **跨平台编译**: 可在任意平台编译目标平台的二进制文件

## 构建不同平台

```bash
# Linux
make build-linux

# macOS
make build-darwin

# Windows
make build-windows
```

## 项目结构

```
device-monitor-go/
├── main.go              # 主入口
├── api/                 # API 处理器
│   ├── handlers/        # 路由处理
│   └── middleware/      # 中间件
├── models/              # 数据模型
├── services/            # 业务逻辑
├── database/            # 数据库相关
├── config/              # 配置管理
├── web/                 # 前端代码（Vue 项目）
│   ├── src/
│   └── dist/            # 构建输出（嵌入到二进制）
└── Makefile            # 构建脚本
```

## 与原版本的差异

1. **部署方式**: 单二进制文件 vs Node.js + npm
2. **性能**: 更低的内存占用和 CPU 使用
3. **依赖管理**: Go modules vs npm packages
4. **构建过程**: 编译时嵌入前端资源

## 故障排查

### 数据库连接失败
确保 `database` 目录存在且有写入权限：
```bash
mkdir -p database
chmod 755 database
```

### 前端资源 404
确保已经构建前端：
```bash
cd web && npm run build
```

### IoT 连接失败
检查 `.env` 中的 IoT 配置是否正确，运行测试：
```bash
curl http://localhost:3000/api/iot/test-connection
```

## License

MIT