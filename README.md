# OpenInterview

> **AI 驱动的面试助手**：实时监听面试官的提问，自动识别问题并流式生成答案，让你在面试中更加从容自信。

---

## 为什么要用 OpenInterview？

求职面试往往充满压力——面对突如其来的技术题或行为题，即使准备充分的候选人也可能因紧张而发挥失常。  
OpenInterview 作为你的**实时 AI 副驾驶（Copilot）**，全程在后台默默工作：

- 🎙️ **实时语音识别**：自动将面试官的语音转换为文字，无需手动操作。
- 🤖 **即时 AI 答题**：检测到完整问题后，立即调用大语言模型（LLM）流式输出参考答案。
- 🔄 **流式打断响应**：新问题到来时，自动中断上一条回答并开始新的生成，响应灵活不卡顿。
- 🏠 **本地部署，隐私安全**：STT（语音转文字）可在本地运行，敏感音频数据不离开你的机器。
- 🔌 **开放接口，易于扩展**：完全开源，支持自定义 LLM 提供商与 STT 后端。

---

## 这个项目有什么用？

### 核心功能

| 功能 | 说明 |
|------|------|
| 语音转文字 | 通过本地 [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) WebSocket 服务将音频实时转为文字 |
| 问题检测 | 后端自动判断转录文本是否构成一个完整问题 |
| AI 答案生成 | 将问题与候选人档案发送至 Groq（或兼容 OpenAI API 的其他 LLM），流式返回回答 |
| 会话管理 | 支持多会话并行，保存候选人档案与历史问答上下文 |
| 手动提问 | 支持跳过 STT，直接在界面输入问题，快速获得 AI 答案 |
| 音频来源选择 | 可捕获麦克风或系统音频（扬声器/屏幕共享） |

### 系统架构

```
浏览器前端
    │
    │  PCM16 音频块 / HTTP REST
    ▼
interviewd（Go 后端，:8080）
    │                    │
    │ float32 WebSocket  │  OpenAI 兼容 API（流式）
    ▼                    ▼
sherpa STT 服务       Groq / 其他 LLM
(:6006)
```

---

## 怎么用？

### 1. 环境准备

**前置依赖：**
- [Go 1.25+](https://go.dev/dl/)
- [Node.js](https://nodejs.org/)（用于前端开发服务器）
- 运行中的 [sherpa-onnx WebSocket STT 服务](https://github.com/k2-fsa/sherpa-onnx)（默认监听 `ws://127.0.0.1:6006/`）
- Groq API Key（或其他兼容 OpenAI API 的 LLM 服务）

### 2. 配置环境变量

在项目根目录复制 `.env.example` 并填写配置：

```bash
cp .env.example .env
```

编辑 `.env`，重点填写以下字段：

```env
# 服务监听地址
INTERVIEW_ADDR=:8080

# STT 服务地址
INTERVIEW_STT_PROVIDER=sherpa-websocket
INTERVIEW_STT_WS_URL=ws://127.0.0.1:6006/

# LLM 配置（以 Groq 为例）
INTERVIEW_LLM_PROVIDER=groq
INTERVIEW_LLM_BASE_URL=https://api.groq.com/openai/v1
INTERVIEW_LLM_API_KEY=你的_API_KEY
INTERVIEW_LLM_MODEL=llama-3.3-70b-versatile
```

> **注意**：`.env` 文件会被自动加载，但进程环境变量的优先级高于 `.env`。

### 3. 启动后端服务

```bash
go run ./cmd/interviewd
```

服务启动后监听 `http://localhost:8080`。

### 4. 启动前端界面

```bash
npm run dev
```

在浏览器中打开 `http://localhost:5173`（或终端提示的端口）。

### 5. 开始使用

1. **新建会话**：点击左侧侧边栏的「New Session」按钮。
2. **配置音频来源**：点击「Settings」，选择麦克风或系统音频。
3. **开始监听**：点击「Start Listen」，面试官开始说话后，转录文字会实时显示在右侧面板。
4. **查看 AI 回答**：检测到完整问题后，左侧「Interview Response」面板会流式展示 AI 生成的参考答案。
5. **手动提问**：也可以在转录编辑器中输入或修改文字，点击「Send」直接向 LLM 提问。
6. **重置会话**：点击「Reset」清除当前会话的问答历史，重新开始。

---

## 测试模式

项目内置轻量级浏览器测试控制台（`test-frontend/`），提供三种测试模式：

| 模式 | 链路 |
|------|------|
| `STT Direct` | 浏览器麦克风 → sherpa WebSocket STT |
| `LLM Only` | 浏览器 → 后端会话 → Groq 流式输出 |
| `Integrated` | 浏览器麦克风 → 后端 → sherpa STT → 问题检测 → Groq 流式输出 |

启动测试前端：

```bash
npm run test-frontend
```

默认访问地址：`http://127.0.0.1:4173`

---

## API 接口一览

所有前端请求均通过 `interviewd` 后端处理：

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/sessions` | 创建新会话 |
| `GET` | `/api/sessions/{sessionID}` | 获取会话快照 |
| `GET` | `/api/sessions/{sessionID}/events` | 订阅实时事件（SSE） |
| `PUT` | `/api/sessions/{sessionID}/profile` | 更新候选人档案 |
| `POST` | `/api/sessions/{sessionID}/listen/start` | 开始监听音频 |
| `POST` | `/api/sessions/{sessionID}/listen/stop` | 停止监听音频 |
| `POST` | `/api/sessions/{sessionID}/audio` | 上传音频数据块 |
| `POST` | `/api/sessions/{sessionID}/reset` | 重置会话上下文 |
| `POST` | `/api/sessions/{sessionID}/ask` | 手动提交问题 |

---

## 运行测试

```bash
go test ./...
```

LLM 集成测试默认跳过，如需手动运行：

```bash
export RUN_LLM_INTEGRATION=1
export INTERVIEW_LLM_BASE_URL=...
export INTERVIEW_LLM_API_KEY=...
export INTERVIEW_LLM_MODEL=...
go test ./internal/llm -run Integration
```

---

## 技术栈

- **后端**：Go · gorilla/websocket
- **STT**：[sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx)（本地 WebSocket 服务）
- **LLM**：Groq API（兼容 OpenAI `/chat/completions`）
- **前端**：原生 HTML / CSS / JavaScript（无框架依赖）

---

## 开源协议

本项目采用开源许可证，欢迎贡献与二次开发。
