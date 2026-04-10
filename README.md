[English](README_EN.md) | 中文

# OpenInterview

> **AI 驱动的面试助手**：实时将面试官的语音转为文字，由你决定何时将内容发送给 AI，流式生成参考答案，让你在面试中更加从容自信。

---

## 为什么要用 OpenInterview？

求职面试往往充满压力——面对突如其来的技术题或行为题，即使准备充分的候选人也可能因紧张而发挥失常。  
OpenInterview 作为你的**实时 AI 副驾驶（Copilot）**，全程在后台默默工作：

- 🎙️ **实时语音识别**：接入腾讯云 ASR（实时流式语音识别），将面试官的语音自动转为文字，转录结果实时显示。
- ✋ **手动掌控节奏**：你来决定何时将转录内容发送给 AI——选中文字段落点击「Send」即可，完全由你主导。
- 🤖 **流式 AI 答题**：将问题发送给大语言模型（LLM）后，参考答案以流式方式逐字输出。
- 🔄 **流式打断响应**：新问题发送时，自动中断上一条回答并开始新的生成，响应灵活不卡顿。
- ☁️ **云端 STT，无需本地 GPU**：腾讯云 ASR 开箱即用，无需在本机部署模型或占用 GPU 资源。
- 🔌 **开放接口，易于扩展**：完全开源，支持自定义 LLM 提供商与 STT 后端。

### 💰 费用对比：比市面上的竞品便宜得多

市面上主流面试 AI 工具普遍收取高额订阅费：

| 产品 | 定价 |
|------|------|
| Interview Copilot | 约 $29–$49 / 月 |
| Final Round AI | 约 $24–$74 / 月 |
| Cluely | 约 $20–$40 / 月 |
| Sensei AI | 约 $20–$50 / 月 |

**OpenInterview 完全开源，你只需为实际用量支付 API 费用：**

- **STT（语音转文字）**：使用[腾讯云实时语音识别](https://cloud.tencent.com/product/asr)，新用户有免费额度，超出后按用量计费，**每小时费用不到 ¥1**，一场面试几乎可以忽略不计。
- **LLM（AI 答题）**：以 [Groq](https://console.groq.com/) 为例，`openai/gpt-oss-120b` 等模型费用极低，一场 1 小时的面试通常花费**不到 $0.10**。
- **综合费用**：一个月密集面试（20 场）的 STT + LLM 总成本通常在 **¥20–¥30（约 $3–$5）** 以内，远低于竞品月订阅费用的零头。

---

## 这个项目有什么用？

### 核心功能

| 功能 | 说明 |
|------|------|
| 语音转文字 | 接入[腾讯云实时流式 ASR](https://cloud.tencent.com/product/asr)，将音频实时转为文字 |
| 手动发送问题 | 在转录编辑器中选中文字段落，点击「Send」手动将其发送给 LLM |
| AI 答案生成 | 将问题与候选人档案发送至 Groq（或兼容 OpenAI API 的其他 LLM），流式返回回答 |
| 会话管理 | 支持多会话并行，保存候选人档案与历史问答上下文 |
| 直接提问 | 支持跳过 STT，直接在界面输入问题，快速获得 AI 答案 |
| 音频来源选择 | 可捕获麦克风或系统音频（扬声器/屏幕共享） |

### 系统架构

```
浏览器前端
    │
    │  PCM16 音频块 / HTTP REST
    ▼
interviewd（Go 后端，:8080）
    │                    │
    │ WebSocket（PCM16）  │  OpenAI 兼容 API（流式）
    ▼                    ▼
腾讯云 ASR            Groq / 其他 LLM
（实时流式语音识别）
```

---

## 怎么用？

### 1. 环境准备

**前置依赖：**
- [Go 1.21+](https://go.dev/dl/)
- [Node.js](https://nodejs.org/)（用于前端开发服务器）
- [腾讯云账号](https://cloud.tencent.com/)，并开通[实时语音识别](https://cloud.tencent.com/product/asr)，获取 `AppID`、`SecretID`、`SecretKey`
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

# STT 配置（使用腾讯云实时语音识别）
INTERVIEW_STT_PROVIDER=tencent

# 腾讯云 ASR
INTERVIEW_TENCENT_WS_URL=wss://asr.cloud.tencent.com/asr/v2/
INTERVIEW_TENCENT_APP_ID=你的_AppID
INTERVIEW_TENCENT_SECRET_ID=你的_SecretID
INTERVIEW_TENCENT_SECRET_KEY=你的_SecretKey
INTERVIEW_TENCENT_ENGINE_TYPE=16k_zh
INTERVIEW_TENCENT_NEED_VAD=1
INTERVIEW_TENCENT_NO_EMPTY_RESULT=1

# LLM 配置（以 Groq 为例）
INTERVIEW_LLM_PROVIDER=openai-compatible
INTERVIEW_LLM_BASE_URL=https://api.groq.com
INTERVIEW_LLM_API_KEY=你的_API_KEY
INTERVIEW_LLM_MODEL=openai/gpt-oss-120b
INTERVIEW_LLM_ENDPOINT=/openai/v1/chat/completions
INTERVIEW_LLM_TIMEOUT=90s
```

> **注意**：`.env` 文件会被自动加载，其中的值会覆盖已有的环境变量。

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
4. **发送问题**：在转录编辑器中，将光标定位到问题末尾，点击「Send to Cursor」将光标之前的文字发送给 LLM；或点击「Send Tail」发送最新的待处理内容。
5. **查看 AI 回答**：左侧「Interview Response」面板会流式展示 AI 生成的参考答案。
6. **直接提问**：也可以通过 `/api/sessions/{id}/ask` 接口直接提交文字问题，跳过 STT 流程。
7. **重置会话**：点击「Reset」清除当前会话的问答历史，重新开始。

---

## 测试模式

项目内置轻量级浏览器测试控制台（`test-frontend/`），提供三种测试模式：

| 模式 | 链路 |
|------|------|
| `LLM Only` | 浏览器 → 后端会话 → Groq 流式输出 |
| `Integrated` | 浏览器麦克风 → 后端 → 腾讯云 ASR → 手动发送 → Groq 流式输出 |

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
- **STT**：[腾讯云实时语音识别 ASR](https://cloud.tencent.com/product/asr)（云端 WebSocket 流式服务）
- **LLM**：Groq API（兼容 OpenAI `/chat/completions`）
- **前端**：原生 HTML / CSS / JavaScript（无框架依赖）

---

## 开源协议

本项目采用开源许可证，欢迎贡献与二次开发。
