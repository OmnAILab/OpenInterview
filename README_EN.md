English | [中文](README.md)

# OpenInterview

> **AI-Powered Interview Assistant**: Listens to the interviewer's questions in real time, automatically recognises them, and streams generated answers — so you can stay calm and confident throughout your interview.

---

## Why OpenInterview?

Job interviews are stressful. Unexpected technical or behavioural questions can trip up even the most prepared candidates.  
OpenInterview acts as your **real-time AI co-pilot**, working silently in the background:

- 🎙️ **Real-time speech recognition**: Automatically converts the interviewer's speech to text — no manual action required.
- 🤖 **Instant AI answers**: As soon as a complete question is detected, a large language model (LLM) streams a reference answer.
- 🔄 **Streaming interruption**: When a new question arrives, the previous answer is interrupted and a fresh generation begins — no lag.
- 🏠 **Local deployment, privacy first**: STT (speech-to-text) can run entirely on-device; sensitive audio never leaves your machine.
- 🔌 **Open interface, easy to extend**: Fully open-source with support for custom LLM providers and STT backends.

---

## What Can It Do?

### Core Features

| Feature | Description |
|---------|-------------|
| Speech-to-Text | Converts audio in real time via a local [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) WebSocket service |
| Question Detection | The backend automatically determines whether the transcribed text forms a complete question |
| AI Answer Generation | Sends the question and candidate profile to Groq (or any OpenAI-compatible LLM) and streams the reply |
| Session Management | Supports multiple parallel sessions; stores candidate profiles and Q&A history |
| Manual Query | Skip STT and type a question directly in the UI to get an instant AI answer |
| Audio Source Selection | Capture audio from a microphone or system audio (speaker / screen share) |

### System Architecture

```
Browser Frontend
      │
      │  PCM16 audio chunks / HTTP REST
      ▼
interviewd (Go backend, :8080)
      │                       │
      │  float32 WebSocket    │  OpenAI-compatible API (streaming)
      ▼                       ▼
sherpa STT service        Groq / other LLM
(:6006)
```

---

## How to Use

### 1. Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js](https://nodejs.org/) (for the frontend dev server)
- A running [sherpa-onnx WebSocket STT service](https://github.com/k2-fsa/sherpa-onnx) (default: `ws://127.0.0.1:6006/`)
- A Groq API key (or any other OpenAI-compatible LLM service)

### 2. Configure Environment Variables

Copy `.env.example` to `.env` in the project root and fill in the values:

```bash
cp .env.example .env
```

Edit `.env` with the following key fields:

```env
# Server listen address
INTERVIEW_ADDR=:8080

# STT configuration
INTERVIEW_STT_PROVIDER=sherpa-websocket
INTERVIEW_STT_PORT=6006
INTERVIEW_STT_WS_URL=ws://127.0.0.1:6006/

# LLM configuration (Groq example)
INTERVIEW_LLM_PROVIDER=groq
INTERVIEW_LLM_BASE_URL=https://api.groq.com/openai/v1
INTERVIEW_LLM_API_KEY=your_api_key_here
INTERVIEW_LLM_MODEL=llama-3.3-70b-versatile
INTERVIEW_LLM_ENDPOINT=/chat/completions
INTERVIEW_LLM_TIMEOUT=90s
```

> **Note**: `.env` is loaded automatically, but process environment variables take precedence over `.env`.

### 3. Start the Backend

```bash
go run ./cmd/interviewd
```

The server listens on `http://localhost:8080`.

### 4. Start the Frontend

```bash
npm run dev
```

Open `http://localhost:5173` (or the port shown in your terminal).

### 5. Start Using It

1. **Create a session**: Click **New Session** in the left sidebar.
2. **Configure audio source**: Click **Settings** and choose a microphone or system audio.
3. **Start listening**: Click **Start Listen**. As the interviewer speaks, the transcription appears in the right panel in real time.
4. **View AI answers**: Once a complete question is detected, the **Interview Response** panel on the left streams the AI-generated reference answer.
5. **Manual query**: You can also type or edit text directly in the transcription editor and click **Send** to ask the LLM a question.
6. **Reset session**: Click **Reset** to clear the current session's Q&A history and start fresh.

---

## Test Mode

The project ships with a lightweight browser test console (`test-frontend/`) that provides three test modes:

| Mode | Pipeline |
|------|----------|
| `STT Direct` | Browser microphone → sherpa WebSocket STT |
| `LLM Only` | Browser → backend session → Groq streaming output |
| `Integrated` | Browser microphone → backend → sherpa STT → question detection → Groq streaming output |

Launch the test frontend:

```bash
npm run test-frontend
```

Default URL: `http://127.0.0.1:4173`

---

## API Reference

All frontend requests are handled by the `interviewd` backend:

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/sessions` | Create a new session |
| `GET` | `/api/sessions/{sessionID}` | Get a session snapshot |
| `GET` | `/api/sessions/{sessionID}/events` | Subscribe to real-time events (SSE) |
| `PUT` | `/api/sessions/{sessionID}/profile` | Update the candidate profile |
| `POST` | `/api/sessions/{sessionID}/listen/start` | Start audio listening |
| `POST` | `/api/sessions/{sessionID}/listen/stop` | Stop audio listening |
| `POST` | `/api/sessions/{sessionID}/audio` | Upload an audio data chunk |
| `POST` | `/api/sessions/{sessionID}/reset` | Reset the session context |
| `POST` | `/api/sessions/{sessionID}/ask` | Manually submit a question |

---

## Running Tests

```bash
go test ./...
```

LLM integration tests are skipped by default. To run them manually:

```bash
export RUN_LLM_INTEGRATION=1
export INTERVIEW_LLM_BASE_URL=...
export INTERVIEW_LLM_API_KEY=...
export INTERVIEW_LLM_MODEL=...
go test ./internal/llm -run Integration
```

---

## Tech Stack

- **Backend**: Go · gorilla/websocket
- **STT**: [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) (local WebSocket service)
- **LLM**: Groq API (compatible with OpenAI `/chat/completions`)
- **Frontend**: Vanilla HTML / CSS / JavaScript (no framework dependencies)

---

## License

This project is open-source. Contributions and forks are welcome.
