# OpenInterview

This repository now contains the backend orchestration layer for an interview copilot.

Current architecture:

- `interviewd` is the only service exposed to the frontend
- STT runs as an external sherpa WebSocket server on `ws://127.0.0.1:6006/`
- LLM runs through the Groq OpenAI-compatible streaming API

What the backend does:

- accepts `16kHz / mono / PCM16` audio chunks from the frontend
- converts PCM16 audio to `float32` frames for the remote sherpa STT server
- maps sherpa `text + segment` messages into internal `partial` and `final` transcript events
- keeps candidate profile and recent QA context
- decides whether a final transcript looks like a complete question
- sends the selected question to the LLM and streams tokens back to the frontend
- interrupts the previous answer when a new question arrives

## Run

```bash
go run ./cmd/interviewd
```

If a `.env` file exists in the project root, it is loaded automatically. Process environment variables still win over `.env`.

## Environment

See [`.env.example`](d:/.github/OpenInterview/.env.example).

Important values:

- `INTERVIEW_ADDR=:8080`
- `INTERVIEW_STT_PROVIDER=sherpa-websocket`
- `INTERVIEW_STT_WS_URL=ws://127.0.0.1:6006/`
- `INTERVIEW_AUDIO_SAMPLE_RATE=16000`
- `INTERVIEW_AUDIO_CHANNELS=1`
- `INTERVIEW_AUDIO_ENCODING=pcm16`
- `INTERVIEW_LLM_PROVIDER=groq`
- `INTERVIEW_LLM_BASE_URL=https://api.groq.com/openai/v1`
- `INTERVIEW_LLM_API_KEY=...`
- `INTERVIEW_LLM_MODEL=...`

## Frontend API

The frontend still talks only to `interviewd`.

- `POST /api/sessions`
- `GET /api/sessions/{sessionID}`
- `GET /api/sessions/{sessionID}/events`
- `PUT /api/sessions/{sessionID}/profile`
- `POST /api/sessions/{sessionID}/listen/start`
- `POST /api/sessions/{sessionID}/listen/stop`
- `POST /api/sessions/{sessionID}/audio`
- `POST /api/sessions/{sessionID}/reset`
- `POST /api/sessions/{sessionID}/ask`

## Test Frontend

A lightweight browser test console lives in `test-frontend/`.

It provides three test modes:

- `STT Direct`: browser microphone -> sherpa websocket STT
- `LLM Only`: browser -> backend session -> Groq stream
- `Integrated`: browser microphone -> backend -> sherpa STT -> question detection -> Groq stream

Run it with:

```bash
npm run test-frontend
```

It serves the page on `http://127.0.0.1:4173` by default.

You can override the frontend host and port with:

```bash
set TEST_FRONTEND_HOST=127.0.0.1
set TEST_FRONTEND_PORT=4173
```

## STT Protocol Notes

The sherpa server you are using accepts:

- binary WebSocket frames containing `float32` audio samples
- a text frame containing `Done` to flush the last result

It returns JSON messages like:

```json
{"text":"hello world","segment":0}
```

Because the server does not explicitly label `partial` vs `final`, the backend uses this rule:

- same `segment`: treat changed text as `partial`
- larger `segment`: treat the previous segment text as `final`
- connection close / `Done`: flush the current text as `final`

## Tests

```bash
go test ./...
```

The external LLM integration test is opt-in. To run it manually:

```bash
set RUN_LLM_INTEGRATION=1
set INTERVIEW_LLM_BASE_URL=...
set INTERVIEW_LLM_ENDPOINT=...
set INTERVIEW_LLM_API_KEY=...
set INTERVIEW_LLM_MODEL=...
go test ./internal/llm -run Integration
```
