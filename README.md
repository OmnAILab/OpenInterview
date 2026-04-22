# OpenInterview

> **AI 驱动的面试助手**：实时将面试官的语音转为文字，由你决定何时将内容发送给 AI，流式生成参考答案，让你在面试中更加从容自信。

---

## RAG with sentence-transformers

项目目前支持两种知识库接入方式：

1. `interviewd` 直接读取本地 `.md` / `.txt` 文件，并调用 embedding 接口做检索。
2. `interviewd` 调用独立知识库服务，例如 `go run ./cmd/kbd`。

最简单的是本地向量检索模式。

### 1. 准备知识库文件

在项目根目录创建 `./knowledge`，将你的简历信息、项目总结、STAR 故事、架构说明、常见追问等内容整理成 `.md` 或 `.txt` 文件放进去。

### 2. 安装 Python 依赖并启动 embedding 服务

```bash
pip install -r requirements.txt
python tools/sentence_transformers_server.py
```

默认监听 `http://127.0.0.1:7008/embed`，默认模型为 `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`。

### 3. 配置 OpenInterview

将以下配置加入 `.env`：

```bash
INTERVIEW_KNOWLEDGE_LOCAL_PATH=./knowledge
INTERVIEW_KNOWLEDGE_EMBEDDING_ENDPOINT=http://127.0.0.1:7008/embed
INTERVIEW_KNOWLEDGE_EMBEDDING_MODEL=sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2
INTERVIEW_KNOWLEDGE_MAX_RESULTS=5
INTERVIEW_KNOWLEDGE_TIMEOUT=10s
```

然后正常启动主服务：

```bash
go run ./cmd/interviewd
```

后端会在每次生成答案前，先检索最相关的知识片段并注入到提示词中。

### 可选：独立运行知识库服务

如果你想把检索能力从 `interviewd` 中拆出来，可以单独启动：

```bash
go run ./cmd/kbd
```

然后在 `.env` 中配置：

```bash
INTERVIEW_KNOWLEDGE_ENDPOINT=http://127.0.0.1:7007/search
```

## Sherpa 本地部署

如果你打算使用本地 Sherpa 作为 STT，而不是腾讯云 ASR，可以按下面的方式准备环境。

### 1. 安装 Python 依赖

```bash
pip install -r requirements.txt
```

### 2. 准备模型

可以把模型目录复制到仓库内的以下位置：

```text
third_party/sherpa-models/sherpa-onnx-streaming-zipformer-bilingual-zh-en-2023-02-20
```

这个目录已经加入 `.gitignore`，模型文件不会提交到 Git。

### 3. 启动官方 `streaming_server.py`

推荐直接使用 `sherpa-onnx` 仓库自带的 `python-api-examples/streaming_server.py`。注意这个脚本要在 `sherpa-onnx` 仓库根目录运行，因为它会导入同目录下的 `http_server.py`。

Windows / PowerShell 示例：

```powershell
cd D:\.github\sherpa-onnx

$model = "D:\.github\OpenInterview\third_party\sherpa-models\sherpa-onnx-streaming-zipformer-bilingual-zh-en-2023-02-20"

python .\python-api-examples\streaming_server.py `
  --encoder "$model\encoder-epoch-99-avg-1.onnx" `
  --decoder "$model\decoder-epoch-99-avg-1.onnx" `
  --joiner "$model\joiner-epoch-99-avg-1.onnx" `
  --tokens "$model\tokens.txt" `
  --port 6006
```

### 4. 配置 OpenInterview 使用 Sherpa

在 `.env` 中确认以下配置：

```bash
INTERVIEW_STT_PROVIDER=sherpa
INTERVIEW_SHERPA_WS_URL=ws://127.0.0.1:6006/
```

确认 Sherpa 已经监听 `6006` 端口后，再启动 `interviewd` 即可。

## 为什么要用 OpenInterview？

求职面试往往充满压力——面对突如其来的技术题或行为题，即使准备充分的候选人也可能因紧张而发挥失常。  
同时，AI 的能力正在快速进化：候选人会用 AI 练题、整理项目与模拟问答，企业也开始默认“AI 辅助工作”会成为日常能力。传统只看记忆与套路回答的面试方式，正在被更强调**真实表达、临场拆解、追问深挖与协作能力**的新型面试替代。  
在这种变化下，候选人需要的不是“背答案工具”，而是一个能在高压对话中稳定输出、快速组织思路的实时辅助系统。  
OpenInterview 作为你的**实时 AI 副驾驶**，全程在后台默默工作：

- 🎙️ **实时语音识别**：接入腾讯云 ASR（实时流式语音识别），将面试官的语音自动转为文字，转录结果实时显示。
- ✋ **手动掌控节奏**：你来决定何时将转录内容发送给 AI——选中文字段落点击「Send」即可，完全由你主导。
- 🤖 **流式 AI 答题**：将问题发送给大语言模型（LLM）后，参考答案以流式方式逐字输出。
- 🔄 **流式打断响应**：新问题发送时，自动中断上一条回答并开始新的生成，响应灵活不卡顿。
- 🔒 **数据安全可控**：自部署运行，不经过 OpenInterview 托管服务器；转录、档案和问答历史默认保存在本机进程内存中，不内置数据库持久化。
- ☁️ **云端 STT，无需本地 GPU**：腾讯云 ASR 开箱即用，无需在本机部署模型或占用 GPU 资源。
- 🔌 **开放接口，易于扩展**：完全开源，支持自定义 LLM 提供商与 STT 后端。

### 🔐 数据安全：自部署、可审计、可替换

面试音频、简历信息、项目经历和回答内容都属于高度敏感数据。很多商业面试助手需要把这些内容上传到对方平台统一处理；OpenInterview 的设计重点之一，就是让数据流向尽量透明、可控：

- **自部署运行**：前端和 Go 后端运行在你自己的电脑或服务器上，没有 OpenInterview 官方中转服务，也不需要把账号、简历或完整面试记录交给第三方 SaaS 平台托管。
- **默认不落库**：会话状态、候选人档案、转录文本和问答历史默认保存在服务进程内存中，不内置数据库持久化；重置会话、删除会话或停止服务即可清理内置会话状态。
- **手动发送给 AI**：语音转写结果会先进入本地会话，你可以选择只把需要回答的片段发送给 LLM，避免整场面试内容被自动全量提交。
- **服务商可替换**：STT 和 LLM 都通过接口接入，可以按自己的隐私、成本和合规要求切换到腾讯云、Groq、OpenAI 兼容服务，或改造成私有化 / 本地模型方案。
- **开源可审计**：核心链路代码公开，音频、转写文本、候选人档案和 LLM 请求的处理方式可以直接检查，也便于你按公司或个人安全要求二次改造。

### 💰 费用对比：比市面上的竞品便宜得多

市面上主流面试 AI 工具，以及中文社区里常见的面试辅助产品，普遍按月订阅、时长包或点数包收费：

| 产品 | 定价 / 计费方式 |
|------|------|
| [OfferStar AI](https://docs.offerstar.cn/pricing) | 语音识别 3 点/分钟，面试问答 10 点/题；套餐约 ¥45–¥900 |
| [面灵AI](https://www.mianlingai.com/) | ¥29 / 60 分钟，¥89 / 240 分钟，¥159 / 600 分钟，或 ¥159 / 月 |
| [即答侠 / HireMe AI](https://interviewasssistant.com/zh) | ¥49–¥79 / 月，季卡 ¥109–¥179；按量付费低至 ¥7 / 次 |
| [鹅来面 / OfferGoose](https://apps.apple.com/cn/app/%E9%B9%85%E6%9D%A5%E9%9D%A2-ai%E6%A8%A1%E6%8B%9F%E9%9D%A2%E8%AF%95-%E7%AE%80%E5%8E%86%E4%BC%98%E5%8C%96%E5%8A%A9%E6%89%8B/id6504543050) | 30 分钟约 ¥68–¥108，120 分钟约 ¥188，300 分钟约 ¥378 |
| [面试牛牛](https://apps.apple.com/cn/app/%E9%9D%A2%E8%AF%95%E7%89%9B%E7%89%9B-ai%E8%BE%85%E5%8A%A9%E9%9D%A2%E8%AF%95%E5%8A%A9%E6%89%8B%E6%8F%90%E8%AF%8D%E5%99%A8%E6%99%BA%E8%83%BD%E7%AD%94%E9%A2%98%E5%AE%9A%E5%88%B6%E6%89%BE%E5%B7%A5%E4%BD%9C%E7%A5%9E%E5%99%A8/id6743431889) | 30 分钟约 ¥58，120 分钟约 ¥108，300 分钟约 ¥198，500 分钟约 ¥268 |
| [Offer蛙](https://apps.apple.com/cn/app/offer%E8%9B%99-%E5%B7%A5%E4%BD%9C%E6%B1%82%E8%81%8C%E7%95%99%E5%AD%A6%E9%9D%A2%E8%AF%95%E7%A5%9E%E5%99%A8/id6739964526) | 30 分钟约 ¥68，2 小时约 ¥78–¥168，4 小时约 ¥208–¥268，8 小时约 ¥328，月卡约 ¥398 |

> 以上价格来自公开官网或应用商店展示，可能随套餐活动调整，请以对应产品的官方页面为准。

**OpenInterview 完全开源，你只需为实际用量支付 API 费用：**

- **STT（语音转文字）**：使用[腾讯云实时语音识别](https://cloud.tencent.com/product/asr)，新用户有免费额度，超出后按用量计费，**每小时费用不到 ¥1**，一场面试几乎可以忽略不计。
- **LLM（AI 答题）**：以 [Groq](https://console.groq.com/) 为例，`openai/gpt-oss-120b` 等模型费用极低，一场 1 小时的面试通常花费**不到 $0.10**。
- **综合费用**：一个月密集面试（20 场）的 STT + LLM 总成本通常在 **¥20–¥30（约 $3–$5）** 以内，远低于竞品月订阅费用的零头。

---

## 这个项目有什么用？

### 核心功能

| 功能 | 说明 |
|------|------|
| 语音转文字 | 接入Sherpa-onnx/[腾讯云实时流式 ASR](https://cloud.tencent.com/product/asr)，将音频实时转为文字 |
| 手动发送问题 | 在转录编辑器中选中文字段落，点击「Send」手动将其发送给 LLM |
| AI 答案生成 | 将问题与候选人档案发送至 Groq（或兼容 OpenAI API 的其他 LLM），流式返回回答 |
| Markdown 消息渲染 | 支持在设置中开启/关闭 Markdown 渲染，让回答更易读（标题、列表、代码块、表格、链接等） |
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
Sherpa/腾讯云 ASR     Groq/其他 LLM
（实时流式语音识别）
```

---

## 怎么用？

### 1. 环境准备

**前置依赖：**
- [Go 1.21+](https://go.dev/dl/)
- [Node.js](https://nodejs.org/)（用于前端开发服务器）
- [Python 3.10+](https://www.python.org/)（如果你要启用 Sherpa 或 RAG）
- [腾讯云账号](https://cloud.tencent.com/)，并开通[实时语音识别](https://cloud.tencent.com/product/asr)，获取 `AppID`、`SecretID`、`SecretKey`（如果使用腾讯云 ASR）
- 本地 `sherpa-onnx` 环境和模型文件（如果使用 Sherpa，见上文“Sherpa 本地部署”）
- Groq API Key（或其他兼容 OpenAI API 的 LLM 服务）

### 2. 配置环境变量

在项目根目录复制 `.env.example` 并填写配置：

```bash
cp .env.example .env
```

> **注意**：`.env` 文件会被自动加载，其中的值会覆盖已有的环境变量。

### 3. 启动后端服务

```bash
go run ./cmd/interviewd
```

服务启动后监听 `http://localhost:8080`。

### 4. 启动前端界面

```bash
npm run frontend
```

在浏览器中打开 `http://localhost:5174`（或终端提示的端口）。

### 5. 开始使用

1. **确认后端地址**：左侧「Backend / 后端」默认是 `http://localhost:8080`。如果后端运行在其他地址或端口，请先在这里修改。
2. **新建会话**：点击左侧侧边栏的「New Session / 新建会话」按钮；多个会话会显示在左侧列表中，点击即可切换。
3. **配置前端设置**：点击「Settings / 设置」打开设置弹窗，按需调整语言、主题、Markdown 渲染和音频采集来源，然后点击「Save Settings / 保存设置」保存。
4. **开始监听**：点击「Start Listen / 开始监听」，浏览器获得授权后会开始采集音频；点击「Stop / 停止」可停止当前监听。
5. **发送问题给 LLM**：面试官说话后，稳态转录会显示在右侧「Stable Transcript / 稳态文本」。把光标放到要发送的位置，点击「Send To LLM / 发送到LLM」发送从上次截止位置到当前光标之间的文本；点击「Insert Stop / 插入截止符」只插入截止符，不会发送给 LLM。
6. **查看 AI 回答**：中间「Interview Response / LLM 回答」面板会流式展示 AI 生成的参考答案；新片段发送时会自动打断上一轮未完成回答并生成新回答。
7. **管理会话**：点击「Reset / 重置」清空当前会话的转录、问答历史和音频统计；点击「Delete Session / 删除会话」删除当前会话。
8. **直接提问**：也可以通过 `/api/sessions/{id}/ask` 接口直接提交文字问题，跳过 STT 流程。

常用前端设置说明：

| 位置 / 按钮 | 说明 |
|------|------|
| 「Settings / 设置」→「Language / 语言」 | 切换界面语言，当前支持中文和英文。 |
| 「Settings / 设置」→「Theme Mode / 主题模式」 | 切换主题模式：跟随系统、浅色、深色。选择「深色」即可启用黑暗模式。 |
| 「Settings / 设置」→「Enable Markdown / 启用 Markdown」 | 开启后，AI 回答会按标题、列表、代码块、表格和链接等 Markdown 格式渲染；关闭后按纯文本展示。 |
| 「Settings / 设置」→「Microphone / 麦克风」 | 采集指定麦克风输入，适合线下面试、耳机麦克风或虚拟音频设备。 |
| 「Settings / 设置」→「Speaker / System Audio / 扬声器 / 系统音频」 | 通过浏览器共享标签页、窗口或屏幕音频，适合线上会议面试。该能力通常需要 Chrome 或 Edge，并且要在浏览器共享弹窗里勾选共享音频。 |
| 「Settings / 设置」→「Device / 设备」 | 在麦克风模式下选择具体输入设备；切换设备会在下一次点击「Start Listen / 开始监听」时生效。 |
| 「Refresh Devices / 刷新设备」 | 重新扫描麦克风设备。插拔耳机、声卡或虚拟音频设备后可以点击刷新。 |
| 「Close / 关闭」 | 关闭设置弹窗，不会保存尚未点击「Save Settings / 保存设置」的改动。 |
| 「Listening / 监听中」「Answering / 回答中」状态 | 左下角显示当前是否正在监听音频、是否正在生成回答。 |
| 「Audio / 音频」计数 | 顶部显示已上传的音频块数量，用于确认浏览器是否正在持续发送音频。 |

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

