const elements = {
  backendBaseUrl: document.getElementById("backendBaseUrl"),
  refreshRuntimeBtn: document.getElementById("refreshRuntimeBtn"),
  createSessionBtn: document.getElementById("createSessionBtn"),
  openEventsBtn: document.getElementById("openEventsBtn"),
  resetSessionBtn: document.getElementById("resetSessionBtn"),
  runtimeSttProvider: document.getElementById("runtimeSttProvider"),
  runtimeLlmProvider: document.getElementById("runtimeLlmProvider"),
  runtimeLlmModel: document.getElementById("runtimeLlmModel"),
  saveProfileBtn: document.getElementById("saveProfileBtn"),
  targetRole: document.getElementById("targetRole"),
  techStack: document.getElementById("techStack"),
  projectSummary: document.getElementById("projectSummary"),
  strengths: document.getElementById("strengths"),
  answerStyle: document.getElementById("answerStyle"),
  sessionId: document.getElementById("sessionId"),
  listeningState: document.getElementById("listeningState"),
  answeringState: document.getElementById("answeringState"),
  audioChunks: document.getElementById("audioChunks"),
  audioBytes: document.getElementById("audioBytes"),
  backendQuestion: document.getElementById("backendQuestion"),
  backendAnswer: document.getElementById("backendAnswer"),
  historyList: document.getElementById("historyList"),
  backendLog: document.getElementById("backendLog"),
  startSttOnlyBtn: document.getElementById("startSttOnlyBtn"),
  stopSttOnlyBtn: document.getElementById("stopSttOnlyBtn"),
  clearSttOnlyBtn: document.getElementById("clearSttOnlyBtn"),
  sttOnlyPartial: document.getElementById("sttOnlyPartial"),
  sttOnlyFinals: document.getElementById("sttOnlyFinals"),
  sttOnlyLog: document.getElementById("sttOnlyLog"),
  askQuestion: document.getElementById("askQuestion"),
  askQuestionBtn: document.getElementById("askQuestionBtn"),
  clearLlmOnlyBtn: document.getElementById("clearLlmOnlyBtn"),
  llmOnlyAnswer: document.getElementById("llmOnlyAnswer"),
  llmOnlyLog: document.getElementById("llmOnlyLog"),
  startIntegratedBtn: document.getElementById("startIntegratedBtn"),
  stopIntegratedBtn: document.getElementById("stopIntegratedBtn"),
  clearIntegratedBtn: document.getElementById("clearIntegratedBtn"),
  integratedPartial: document.getElementById("integratedPartial"),
  integratedFinals: document.getElementById("integratedFinals"),
  textDealTranscript: document.getElementById("textDealTranscript"),
  sendTextDealSegmentBtn: document.getElementById("sendTextDealSegmentBtn"),
  textDealCursorEndBtn: document.getElementById("textDealCursorEndBtn"),
  textDealPending: document.getElementById("textDealPending"),
  integratedQuestion: document.getElementById("integratedQuestion"),
  integratedAnswer: document.getElementById("integratedAnswer"),
  integratedLog: document.getElementById("integratedLog"),
};

const backendEventTypes = [
  "state",
  "stt.partial",
  "stt.final",
  "textdeal.updated",
  "question.detected",
  "question.manual",
  "llm.started",
  "llm.token",
  "llm.done",
  "llm.cancelled",
  "llm.interrupted",
  "profile.updated",
  "session.reset",
  "log",
  "error",
];

const state = {
  sessionId: "",
  eventSource: null,
  capture: null,
  backendSnapshot: null,
  runtime: emptyRuntimeState(),
  sttOnly: {
    sessionId: "",
    eventSource: null,
    finals: [],
    sendChain: Promise.resolve(),
  },
  integrated: {
    finals: [],
    sendChain: Promise.resolve(),
    textDeal: emptyTextDealState(),
    textCursorRaw: 0,
  },
};

boot();

function boot() {
  bindActions();
  renderEmptyState();
  appendLog(elements.backendLog, "Ready. Configure the backend URL, refresh runtime, then create a session.");
  void refreshBackendRuntime({ quiet: true });
  window.addEventListener("beforeunload", () => {
    closeBackendEvents();
    closeSTTOnlyEvents();
  });
}

function bindActions() {
  elements.backendBaseUrl.addEventListener("change", () => {
    void refreshBackendRuntime({ quiet: true });
  });
  elements.refreshRuntimeBtn.addEventListener("click", () => runAction(refreshBackendRuntime));
  elements.createSessionBtn.addEventListener("click", () => runAction(createBackendSession));
  elements.openEventsBtn.addEventListener("click", () => runAction(openBackendEvents));
  elements.resetSessionBtn.addEventListener("click", () => runAction(resetBackendSession));
  elements.saveProfileBtn.addEventListener("click", () => runAction(saveProfile));
  elements.startSttOnlyBtn.addEventListener("click", () => runAction(startSTTOnly));
  elements.stopSttOnlyBtn.addEventListener("click", () => runAction(stopSTTOnly));
  elements.clearSttOnlyBtn.addEventListener("click", clearSTTOnlyPanel);
  elements.askQuestionBtn.addEventListener("click", () => runAction(askManualQuestion));
  elements.clearLlmOnlyBtn.addEventListener("click", clearLlmOnlyPanel);
  elements.startIntegratedBtn.addEventListener("click", () => runAction(startIntegrated));
  elements.stopIntegratedBtn.addEventListener("click", () => runAction(stopIntegrated));
  elements.clearIntegratedBtn.addEventListener("click", clearIntegratedPanel);
  elements.sendTextDealSegmentBtn.addEventListener("click", () => runAction(submitTextDealSegment));
  elements.textDealCursorEndBtn.addEventListener("click", moveTextDealCursorToEnd);
  elements.textDealTranscript.addEventListener("click", syncTextDealCursorFromSelection);
  elements.textDealTranscript.addEventListener("keyup", syncTextDealCursorFromSelection);
  elements.textDealTranscript.addEventListener("select", syncTextDealCursorFromSelection);
}

async function runAction(action) {
  try {
    await action();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    appendLog(elements.backendLog, `ERROR: ${message}`);
    appendLog(elements.sttOnlyLog, `ERROR: ${message}`);
    appendLog(elements.integratedLog, `ERROR: ${message}`);
    appendLog(elements.llmOnlyLog, `ERROR: ${message}`);
  }
}

async function createBackendSession() {
  await refreshBackendRuntime({ quiet: true });
  const snapshot = await requestJSON("POST", "/api/sessions");
  state.sessionId = snapshot.id;
  appendLog(elements.backendLog, `Created session ${state.sessionId}`);
  applySnapshot(snapshot);
  await openBackendEvents();
}

async function refreshBackendRuntime(options = {}) {
  const { quiet = false } = options;

  try {
    const health = await requestJSON("GET", "/api/healthz");
    applyBackendRuntime(health);
    if (!quiet) {
      appendLog(elements.backendLog, "Loaded backend runtime configuration.");
    }
  } catch (error) {
    state.runtime = emptyRuntimeState();
    renderRuntimeState();
    if (!quiet) {
      throw error;
    }
  }
}

async function openBackendEvents() {
  ensureSessionId();
  closeBackendEvents();

  const eventsUrl = `${backendBaseURL()}/api/sessions/${encodeURIComponent(state.sessionId)}/events`;
  const source = new EventSource(eventsUrl);
  state.eventSource = source;

  source.addEventListener("snapshot", (event) => {
    try {
      const snapshot = parseJSON(event.data, "backend snapshot");
      applySnapshot(snapshot);
      appendLog(elements.backendLog, "Received snapshot event.");
    } catch (error) {
      appendLog(elements.backendLog, error instanceof Error ? error.message : String(error));
    }
  });

  for (const type of backendEventTypes) {
    source.addEventListener(type, (event) => {
      try {
        const record = parseJSON(event.data, type);
        handleBackendEvent(type, record);
      } catch (error) {
        appendLog(elements.backendLog, error instanceof Error ? error.message : String(error));
      }
    });
  }

  source.onopen = () => {
    appendLog(elements.backendLog, `SSE connected for ${state.sessionId}`);
  };

  source.onerror = () => {
    appendLog(elements.backendLog, "SSE connection reported an error. Browser will retry automatically.");
  };
}

function closeBackendEvents() {
  if (state.eventSource) {
    state.eventSource.close();
    state.eventSource = null;
  }
}

async function ensureSTTOnlySession() {
  if (!state.sttOnly.sessionId) {
    const snapshot = await requestJSON("POST", "/api/sessions");
    state.sttOnly.sessionId = snapshot.id;
    applySTTOnlySnapshot(snapshot);
    appendLog(elements.sttOnlyLog, `Created dedicated STT session ${state.sttOnly.sessionId}`);
  }

  if (!state.sttOnly.eventSource) {
    openSTTOnlyEvents();
  }

  return state.sttOnly.sessionId;
}

function openSTTOnlyEvents() {
  if (!state.sttOnly.sessionId) {
    throw new Error("No STT-only session. Start the STT test first.");
  }

  closeSTTOnlyEvents();

  const eventsUrl = `${backendBaseURL()}/api/sessions/${encodeURIComponent(state.sttOnly.sessionId)}/events`;
  const source = new EventSource(eventsUrl);
  state.sttOnly.eventSource = source;

  source.addEventListener("snapshot", (event) => {
    try {
      const snapshot = parseJSON(event.data, "stt-only snapshot");
      applySTTOnlySnapshot(snapshot);
      appendLog(elements.sttOnlyLog, "Received STT-only snapshot event.");
    } catch (error) {
      appendLog(elements.sttOnlyLog, error instanceof Error ? error.message : String(error));
    }
  });

  for (const type of ["stt.partial", "stt.final", "error", "log"]) {
    source.addEventListener(type, (event) => {
      try {
        const record = parseJSON(event.data, `stt-only ${type}`);
        handleSTTOnlyEvent(type, record);
      } catch (error) {
        appendLog(elements.sttOnlyLog, error instanceof Error ? error.message : String(error));
      }
    });
  }

  source.onopen = () => {
    appendLog(elements.sttOnlyLog, `SSE connected for dedicated STT session ${state.sttOnly.sessionId}`);
  };

  source.onerror = () => {
    appendLog(elements.sttOnlyLog, "STT-only SSE connection reported an error. Browser will retry automatically.");
  };
}

function closeSTTOnlyEvents() {
  if (state.sttOnly.eventSource) {
    state.sttOnly.eventSource.close();
    state.sttOnly.eventSource = null;
  }
}

async function resetBackendSession() {
  ensureSessionId();
  if (state.capture?.mode === "integrated") {
    await stopIfCapturing("integrated");
  }
  const snapshot = await requestJSON("POST", `/api/sessions/${state.sessionId}/reset`);
  state.integrated.finals = [];
  applySnapshot(snapshot);
  clearIntegratedPanel();
  clearLlmOnlyPanel();
  appendLog(elements.backendLog, "Session reset.");
}

async function saveProfile() {
  ensureSessionId();
  const body = {
    targetRole: elements.targetRole.value.trim(),
    techStack: splitTechStack(elements.techStack.value),
    projectSummary: elements.projectSummary.value.trim(),
    strengths: elements.strengths.value.trim(),
    answerStyle: elements.answerStyle.value.trim(),
  };
  const snapshot = await requestJSON("PUT", `/api/sessions/${state.sessionId}/profile`, body);
  applySnapshot(snapshot);
  appendLog(elements.backendLog, "Profile saved.");
}

async function askManualQuestion() {
  const question = elements.askQuestion.value.trim();
  if (!question) {
    throw new Error("Question is empty");
  }
  if (!state.sessionId) {
    await createBackendSession();
  } else if (!state.eventSource) {
    await openBackendEvents();
  }

  setStreamText(elements.llmOnlyAnswer, "", "Waiting for streamed answer.");
  setStreamText(elements.integratedAnswer, "", "Waiting for streamed answer.");

  await requestJSON("POST", `/api/sessions/${state.sessionId}/ask`, { question });
  appendLog(elements.llmOnlyLog, `Asked question: ${question}`);
}

async function submitTextDealSegment() {
  ensureSessionId();

  const rawCursor = getTextDealCursorRaw();
  if (rawCursor <= state.integrated.textDeal.sentUntil) {
    throw new Error("Move the stop marker forward before sending.");
  }

  await requestJSON("POST", `/api/sessions/${state.sessionId}/textdeal/segment`, {
    stop: rawCursor,
  });
  appendLog(elements.integratedLog, `Submitted textdeal segment up to ${rawCursor}.`);
}

async function startSTTOnly() {
  ensureCaptureAvailable("stt-only");
  clearSTTOnlyPanel();
  const sessionID = await ensureSTTOnlySession();

  await requestJSON("POST", `/api/sessions/${sessionID}/reset`);
  await requestJSON("POST", `/api/sessions/${sessionID}/listen/start`);
  appendLog(elements.sttOnlyLog, "Dedicated STT listening started.");

  state.sttOnly.sendChain = Promise.resolve();

  await startMicCapture("stt-only", (samples) => {
    const pcm = float32ToPCM16(samples);
    queueSTTOnlyChunk(pcm.buffer.slice(0));
  });
}

async function stopSTTOnly() {
  await stopIfCapturing("stt-only");
  await state.sttOnly.sendChain.catch(() => {});

  if (state.sttOnly.sessionId) {
    await requestJSON("POST", `/api/sessions/${state.sttOnly.sessionId}/listen/stop`);
    appendLog(elements.sttOnlyLog, "Dedicated STT listening stopped.");
  }
}

async function startIntegrated() {
  ensureCaptureAvailable("integrated");
  if (!state.sessionId) {
    await createBackendSession();
  } else if (!state.eventSource) {
    await openBackendEvents();
  }

  clearIntegratedPanel();
  await requestJSON("POST", `/api/sessions/${state.sessionId}/listen/start`);
  appendLog(elements.integratedLog, "Backend listening started.");

  state.integrated.sendChain = Promise.resolve();

  await startMicCapture("integrated", (samples) => {
    const pcm = float32ToPCM16(samples);
    queueIntegratedChunk(pcm.buffer.slice(0));
  });
}

async function stopIntegrated() {
  await stopIfCapturing("integrated");
  await state.integrated.sendChain.catch(() => {});

  if (state.sessionId) {
    await requestJSON("POST", `/api/sessions/${state.sessionId}/listen/stop`);
    appendLog(elements.integratedLog, "Backend listening stopped.");
  }
}

function moveTextDealCursorToEnd() {
  const total = codePointLength(state.integrated.textDeal.stableText);
  state.integrated.textCursorRaw = total;
  renderTextDealEditor();
}

function syncTextDealCursorFromSelection() {
  const display = elements.textDealTranscript.value || "";
  const selectionStart = elements.textDealTranscript.selectionStart || 0;
  const rawCursor = codePointLength(display.slice(0, selectionStart).replaceAll("|", ""));
  state.integrated.textCursorRaw = clamp(rawCursor, state.integrated.textDeal.sentUntil, codePointLength(state.integrated.textDeal.stableText));
}

function getTextDealCursorRaw() {
  syncTextDealCursorFromSelection();
  return state.integrated.textCursorRaw;
}

function queueIntegratedChunk(payload) {
  if (!state.sessionId) {
    return;
  }

  state.integrated.sendChain = state.integrated.sendChain
    .catch(() => {})
    .then(() => postAudioChunk(payload))
    .catch((error) => {
      const message = error instanceof Error ? error.message : String(error);
      appendLog(elements.integratedLog, `Audio upload error: ${message}`);
    });
}

function queueSTTOnlyChunk(payload) {
  if (!state.sttOnly.sessionId) {
    return;
  }

  state.sttOnly.sendChain = state.sttOnly.sendChain
    .catch(() => {})
    .then(() => postSTTOnlyChunk(payload))
    .catch((error) => {
      const message = error instanceof Error ? error.message : String(error);
      appendLog(elements.sttOnlyLog, `Audio upload error: ${message}`);
    });
}

async function postAudioChunk(payload) {
  ensureSessionId();
  const url = `${backendBaseURL()}/api/sessions/${encodeURIComponent(state.sessionId)}/audio?sampleRate=16000&channels=1&encoding=pcm16`;
  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/octet-stream",
    },
    body: payload,
  });

  if (!response.ok) {
    throw new Error(await responseError(response));
  }
}

async function postSTTOnlyChunk(payload) {
  if (!state.sttOnly.sessionId) {
    throw new Error("No STT-only session. Start the STT test first.");
  }

  const url = `${backendBaseURL()}/api/sessions/${encodeURIComponent(state.sttOnly.sessionId)}/audio?sampleRate=16000&channels=1&encoding=pcm16`;
  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/octet-stream",
    },
    body: payload,
  });

  if (!response.ok) {
    throw new Error(await responseError(response));
  }
}

async function startMicCapture(mode, onSamples) {
  if (state.capture) {
    throw new Error(`Capture already active in mode ${state.capture.mode}`);
  }

  const stream = await navigator.mediaDevices.getUserMedia({
    audio: {
      channelCount: 1,
      echoCancellation: false,
      noiseSuppression: false,
      autoGainControl: false,
    },
  });

  const AudioContextClass = window.AudioContext || window.webkitAudioContext;
  if (!AudioContextClass) {
    throw new Error("Web Audio API is not available in this browser");
  }

  const audioContext = new AudioContextClass();
  await audioContext.resume();

  const source = audioContext.createMediaStreamSource(stream);
  const processor = audioContext.createScriptProcessor(4096, 1, 1);
  const mute = audioContext.createGain();
  mute.gain.value = 0;

  processor.onaudioprocess = (event) => {
    const input = event.inputBuffer.getChannelData(0);
    const copied = new Float32Array(input.length);
    copied.set(input);

    const resampled = resampleFloat32(copied, audioContext.sampleRate, 16000);
    if (resampled.length > 0) {
      onSamples(resampled);
    }
  };

  source.connect(processor);
  processor.connect(mute);
  mute.connect(audioContext.destination);

  state.capture = {
    mode,
    stop: async () => {
      processor.disconnect();
      source.disconnect();
      mute.disconnect();
      stream.getTracks().forEach((track) => track.stop());
      await audioContext.close();
    },
  };
}

async function stopIfCapturing(mode) {
  if (!state.capture) {
    return;
  }
  if (state.capture.mode !== mode) {
    throw new Error(`Capture is active in mode ${state.capture.mode}`);
  }

  const capture = state.capture;
  state.capture = null;
  await capture.stop();
}

function ensureCaptureAvailable(mode) {
  if (state.capture && state.capture.mode !== mode) {
    throw new Error(`Capture is active in mode ${state.capture.mode}`);
  }
}

function handleSTTOnlyEvent(type, record) {
  const payload = record && typeof record === "object" ? record.data || {} : {};
  switch (type) {
    case "stt.partial":
      setStreamText(elements.sttOnlyPartial, payload.text || "", "Waiting for backend transcript.");
      appendLog(elements.sttOnlyLog, `Partial: ${payload.text || ""}`);
      break;
    case "stt.final":
      if (payload.text) {
        state.sttOnly.finals.push(payload.text);
        renderFinalList(elements.sttOnlyFinals, state.sttOnly.finals, "No final transcripts yet.");
      }
      setStreamText(elements.sttOnlyPartial, "", "Waiting for backend transcript.");
      appendLog(elements.sttOnlyLog, `Final: ${payload.text || ""}`);
      break;
    case "error":
      appendLog(elements.sttOnlyLog, `Backend error: ${payload.message || "unknown error"}`);
      break;
    case "log":
      if (payload.scope === "stt" || payload.scope === "audio") {
        appendLog(elements.sttOnlyLog, JSON.stringify(payload));
      }
      break;
    default:
      break;
  }
}

function handleBackendEvent(type, record) {
  appendLog(elements.backendLog, JSON.stringify(record));

  const payload = record && typeof record === "object" ? record.data || {} : {};
  switch (type) {
    case "state":
      if (typeof payload.listening === "boolean") {
        elements.listeningState.textContent = String(payload.listening);
      }
      if (typeof payload.answerInProgress === "boolean") {
        elements.answeringState.textContent = String(payload.answerInProgress);
      }
      break;
    case "stt.partial":
      setStreamText(elements.integratedPartial, payload.text || "", "Waiting for backend transcript.");
      appendLog(elements.integratedLog, `Partial: ${payload.text || ""}`);
      break;
    case "stt.final":
      if (payload.text) {
        state.integrated.finals.push(payload.text);
        renderFinalList(elements.integratedFinals, state.integrated.finals, "No final transcripts yet.");
      }
      appendLog(elements.integratedLog, `Final: ${payload.text || ""}`);
      break;
    case "textdeal.updated":
      applyTextDealSnapshot(payload);
      appendLog(elements.integratedLog, `Stable text updated. Pending length: ${codePointLength(payload.pendingText || "")}`);
      break;
    case "question.detected":
    case "question.manual":
      setStreamText(elements.integratedQuestion, payload.text || "", "No question detected yet.");
      setStreamText(elements.backendQuestion, payload.text || "", "No question yet.");
      appendLog(elements.integratedLog, `Question: ${payload.text || ""}`);
      appendLog(elements.llmOnlyLog, `Question: ${payload.text || ""}`);
      break;
    case "llm.started":
      setStreamText(elements.backendQuestion, payload.question || "", "No question yet.");
      setStreamText(elements.backendAnswer, "", "No answer yet.");
      setStreamText(elements.integratedAnswer, "", "Waiting for backend answer stream.");
      setStreamText(elements.llmOnlyAnswer, "", "Waiting for streamed answer.");
      appendLog(elements.integratedLog, `LLM started: ${payload.question || ""}`);
      appendLog(elements.llmOnlyLog, `LLM started: ${payload.question || ""}`);
      break;
    case "llm.token":
      appendStreamToken(elements.integratedAnswer, payload.token || "");
      appendStreamToken(elements.llmOnlyAnswer, payload.token || "");
      appendStreamToken(elements.backendAnswer, payload.token || "");
      break;
    case "llm.done":
      setStreamText(elements.integratedAnswer, payload.answer || "", "Waiting for backend answer stream.");
      setStreamText(elements.llmOnlyAnswer, payload.answer || "", "Waiting for streamed answer.");
      setStreamText(elements.backendAnswer, payload.answer || "", "No answer yet.");
      appendLog(elements.integratedLog, "LLM finished.");
      appendLog(elements.llmOnlyLog, "LLM finished.");
      void refreshSnapshot();
      break;
    case "llm.cancelled":
    case "llm.interrupted":
      appendLog(elements.integratedLog, type);
      appendLog(elements.llmOnlyLog, type);
      break;
    case "profile.updated":
      appendLog(elements.backendLog, "Profile updated.");
      break;
    case "session.reset":
      state.integrated.finals = [];
      clearIntegratedPanel();
      clearLlmOnlyPanel();
      void refreshSnapshot();
      break;
    case "log":
      appendLog(elements.integratedLog, JSON.stringify(payload));
      break;
    case "error":
      appendLog(elements.integratedLog, `Backend error: ${payload.message || "unknown error"}`);
      appendLog(elements.llmOnlyLog, `Backend error: ${payload.message || "unknown error"}`);
      break;
    default:
      break;
  }
}

async function refreshSnapshot() {
  if (!state.sessionId) {
    return;
  }
  const snapshot = await requestJSON("GET", `/api/sessions/${state.sessionId}`);
  applySnapshot(snapshot);
}

function applySnapshot(snapshot) {
  state.backendSnapshot = snapshot;
  state.sessionId = snapshot.id || state.sessionId;
  state.integrated.finals = Array.isArray(snapshot.finalTranscripts) ? [...snapshot.finalTranscripts] : [];

  elements.sessionId.textContent = state.sessionId || "none";
  elements.listeningState.textContent = String(Boolean(snapshot.listening));
  elements.answeringState.textContent = String(Boolean(snapshot.answerInProgress));
  elements.audioChunks.textContent = String(snapshot.audio?.chunks || 0);
  elements.audioBytes.textContent = String(snapshot.audio?.bytes || 0);

  setStreamText(elements.backendQuestion, snapshot.currentQuestion || "", "No question yet.");
  setStreamText(elements.backendAnswer, snapshot.currentAnswer || "", "No answer yet.");
  setStreamText(elements.integratedPartial, snapshot.partialTranscript || "", "Waiting for backend transcript.");
  applyTextDealSnapshot(snapshot.textDeal || {});
  setStreamText(elements.integratedQuestion, snapshot.currentQuestion || "", "No question detected yet.");
  setStreamText(elements.integratedAnswer, snapshot.currentAnswer || "", "Waiting for backend answer stream.");
  setStreamText(elements.llmOnlyAnswer, snapshot.currentAnswer || "", "Waiting for a manual question.");

  renderFinalList(elements.integratedFinals, state.integrated.finals, "No final transcripts yet.");
  renderHistory(snapshot.history || []);
  fillProfileForm(snapshot.profile || {});
}

function applySTTOnlySnapshot(snapshot) {
  state.sttOnly.sessionId = snapshot.id || state.sttOnly.sessionId;
  state.sttOnly.finals = Array.isArray(snapshot.finalTranscripts) ? [...snapshot.finalTranscripts] : [];

  setStreamText(elements.sttOnlyPartial, snapshot.partialTranscript || "", "Waiting for backend transcript.");
  renderFinalList(elements.sttOnlyFinals, state.sttOnly.finals, "No final transcripts yet.");
}

function applyTextDealSnapshot(textDeal) {
  const next = normalizeTextDealState(textDeal);
  const shouldFollowTail = document.activeElement !== elements.textDealTranscript;

  state.integrated.textDeal = next;
  if (shouldFollowTail) {
    state.integrated.textCursorRaw = codePointLength(next.stableText);
  } else {
    state.integrated.textCursorRaw = clamp(state.integrated.textCursorRaw, next.sentUntil, codePointLength(next.stableText));
  }

  renderTextDealEditor();
  setStreamText(elements.textDealPending, next.pendingText || "", "No unsent stable text.");
}

function applyBackendRuntime(health) {
  const runtime = health && typeof health === "object" ? health.runtime || {} : {};
  const stt = runtime && typeof runtime === "object" ? runtime.stt || {} : {};
  const llm = runtime && typeof runtime === "object" ? runtime.llm || {} : {};

  state.runtime = {
    stt: {
      provider: typeof stt.provider === "string" ? stt.provider : "",
    },
    llm: {
      provider: typeof llm.provider === "string" ? llm.provider : "",
      model: typeof llm.model === "string" ? llm.model : "",
    },
  };

  renderRuntimeState();
}

function renderRuntimeState() {
  elements.runtimeSttProvider.textContent = state.runtime.stt.provider || "unknown";
  elements.runtimeLlmProvider.textContent = state.runtime.llm.provider || "unknown";
  elements.runtimeLlmModel.textContent = state.runtime.llm.model || "unknown";
}

function fillProfileForm(profile) {
  elements.targetRole.value = profile.targetRole || "";
  elements.techStack.value = Array.isArray(profile.techStack) ? profile.techStack.join(", ") : "";
  elements.projectSummary.value = profile.projectSummary || "";
  elements.strengths.value = profile.strengths || "";
  elements.answerStyle.value = profile.answerStyle || "";
}

function renderHistory(history) {
  if (!Array.isArray(history) || history.length === 0) {
    elements.historyList.textContent = "No turns yet.";
    elements.historyList.classList.add("muted");
    return;
  }

  elements.historyList.classList.remove("muted");
  elements.historyList.innerHTML = history
    .map(
      (turn) => `
        <div class="final-item">
          <strong>Q:</strong> ${escapeHTML(turn.question || "")}<br />
          <strong>A:</strong> ${escapeHTML(turn.answer || "")}
        </div>
      `,
    )
    .join("");
}

function renderFinalList(container, items, emptyText) {
  if (!items || items.length === 0) {
    container.textContent = emptyText;
    container.classList.add("muted");
    return;
  }

  container.classList.remove("muted");
  container.innerHTML = items
    .map((item) => `<div class="final-item">${escapeHTML(item)}</div>`)
    .join("");
}

function clearSTTOnlyPanel() {
  state.sttOnly.finals = [];
  setStreamText(elements.sttOnlyPartial, "", "Waiting for backend transcript.");
  renderFinalList(elements.sttOnlyFinals, [], "No final transcripts yet.");
  elements.sttOnlyLog.textContent = "";
}

function clearIntegratedPanel() {
  state.integrated.finals = [];
  state.integrated.textDeal = emptyTextDealState();
  state.integrated.textCursorRaw = 0;
  setStreamText(elements.integratedPartial, "", "Waiting for backend transcript.");
  renderTextDealEditor();
  setStreamText(elements.textDealPending, "", "No unsent stable text.");
  setStreamText(elements.integratedQuestion, "", "No question detected yet.");
  setStreamText(elements.integratedAnswer, "", "Waiting for backend answer stream.");
  renderFinalList(elements.integratedFinals, [], "No final transcripts yet.");
  elements.integratedLog.textContent = "";
}

function clearLlmOnlyPanel() {
  setStreamText(elements.llmOnlyAnswer, "", "Waiting for a manual question.");
  elements.llmOnlyLog.textContent = "";
}

function renderEmptyState() {
  state.runtime = emptyRuntimeState();
  clearSTTOnlyPanel();
  clearIntegratedPanel();
  clearLlmOnlyPanel();
  renderRuntimeState();
  elements.sessionId.textContent = "none";
  elements.listeningState.textContent = "false";
  elements.answeringState.textContent = "false";
  elements.audioChunks.textContent = "0";
  elements.audioBytes.textContent = "0";
  setStreamText(elements.backendQuestion, "", "No question yet.");
  setStreamText(elements.backendAnswer, "", "No answer yet.");
  elements.historyList.textContent = "No turns yet.";
  elements.historyList.classList.add("muted");
}

function emptyRuntimeState() {
  return {
    stt: {
      provider: "",
    },
    llm: {
      provider: "",
      model: "",
    },
  };
}

function emptyTextDealState() {
  return {
    stableText: "",
    sentUntil: 0,
    markers: [],
    pendingText: "",
  };
}

function setStreamText(element, text, fallback) {
  const next = text && text.length > 0 ? text : fallback;
  element.textContent = next;
  element.classList.toggle("muted", !text);
}

function appendStreamToken(element, token) {
  if (!token) {
    return;
  }
  if (element.classList.contains("muted")) {
    element.textContent = "";
    element.classList.remove("muted");
  }
  element.textContent += token;
}

function appendLog(element, line) {
  const stamp = new Date().toLocaleTimeString();
  const nextLine = `[${stamp}] ${line}`;
  element.textContent = element.textContent
    ? `${element.textContent}\n${nextLine}`
    : nextLine;

  const lines = element.textContent.split("\n");
  if (lines.length > 120) {
    element.textContent = lines.slice(lines.length - 120).join("\n");
  }
  element.scrollTop = element.scrollHeight;
}

function escapeHTML(input) {
  return String(input)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function parseJSON(raw, label) {
  try {
    return JSON.parse(raw);
  } catch (error) {
    throw new Error(`Failed to parse ${label}: ${error instanceof Error ? error.message : String(error)}`);
  }
}

function splitTechStack(raw) {
  return raw
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function normalizeTextDealState(textDeal) {
  const stableText = typeof textDeal?.stableText === "string" ? textDeal.stableText : "";
  const total = codePointLength(stableText);
  const sentUntil = clamp(Number.isFinite(textDeal?.sentUntil) ? Number(textDeal.sentUntil) : 0, 0, total);
  const markers = Array.isArray(textDeal?.markers)
    ? [...new Set(textDeal.markers.map((item) => Number(item)).filter((item) => Number.isInteger(item) && item > 0 && item <= total))].sort((left, right) => left - right)
    : [];
  const pendingText = typeof textDeal?.pendingText === "string" ? textDeal.pendingText : "";

  return {
    stableText,
    sentUntil,
    markers,
    pendingText,
  };
}

function renderTextDealEditor() {
  const textDeal = state.integrated.textDeal;
  const display = buildTextDealDisplay(textDeal);
  const rawCursor = clamp(state.integrated.textCursorRaw, textDeal.sentUntil, codePointLength(textDeal.stableText));
  const displayCursor = rawCursor + textDeal.markers.filter((marker) => marker <= rawCursor).length;
  const displayOffset = toCodeUnitOffset(display, displayCursor);

  elements.textDealTranscript.value = display;
  elements.textDealTranscript.placeholder = "Waiting for stable transcript.";
  elements.textDealTranscript.setSelectionRange(displayOffset, displayOffset);
}

function buildTextDealDisplay(textDeal) {
  const chars = Array.from(textDeal.stableText || "");
  let offset = 0;
  for (const marker of textDeal.markers || []) {
    const position = clamp(marker + offset, 0, chars.length);
    chars.splice(position, 0, "|");
    offset += 1;
  }
  return chars.join("");
}

function codePointLength(text) {
  return Array.from(text || "").length;
}

function toCodeUnitOffset(text, codePointIndex) {
  return Array.from(text || "")
    .slice(0, codePointIndex)
    .join("").length;
}

function clamp(value, min, max) {
  return Math.min(Math.max(value, min), max);
}

function resampleFloat32(input, inputRate, outputRate) {
  if (inputRate === outputRate) {
    return input;
  }

  const ratio = inputRate / outputRate;
  const outputLength = Math.max(1, Math.round(input.length / ratio));
  const output = new Float32Array(outputLength);

  for (let i = 0; i < outputLength; i += 1) {
    const position = i * ratio;
    const left = Math.floor(position);
    const right = Math.min(left + 1, input.length - 1);
    const weight = position - left;
    output[i] = input[left] * (1 - weight) + input[right] * weight;
  }

  return output;
}

function float32ToPCM16(input) {
  const output = new Int16Array(input.length);
  for (let i = 0; i < input.length; i += 1) {
    const sample = Math.max(-1, Math.min(1, input[i]));
    output[i] = sample < 0 ? sample * 32768 : sample * 32767;
  }
  return output;
}

function backendBaseURL() {
  return elements.backendBaseUrl.value.trim().replace(/\/$/, "");
}

function ensureSessionId() {
  if (!state.sessionId) {
    throw new Error("No backend session. Create one first.");
  }
}

async function requestJSON(method, path, body) {
  const url = path.startsWith("http") ? path : `${backendBaseURL()}${path}`;
  const headers = {};
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  const response = await fetch(url, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  if (!response.ok) {
    throw new Error(await responseError(response));
  }

  return response.json();
}

async function responseError(response) {
  const text = await response.text();
  try {
    const parsed = JSON.parse(text);
    return parsed.error || text || `${response.status} ${response.statusText}`;
  } catch {
    return text || `${response.status} ${response.statusText}`;
  }
}
