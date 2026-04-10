const SESSION_STORAGE_KEY = "openinterview.frontend.sessions";
const BACKEND_STORAGE_KEY = "openinterview.frontend.backend";
const AUDIO_INPUT_STORAGE_KEY = "openinterview.frontend.audioInputId";
const CAPTURE_SOURCE_STORAGE_KEY = "openinterview.frontend.captureSource";

const elements = {
  backendBaseUrl: document.getElementById("backendBaseUrl"),
  createSessionBtn: document.getElementById("createSessionBtn"),
  sessionList: document.getElementById("sessionList"),
  listeningState: document.getElementById("listeningState"),
  answeringState: document.getElementById("answeringState"),
  runtimeSttProvider: document.getElementById("runtimeSttProvider"),
  runtimeLlmModel: document.getElementById("runtimeLlmModel"),
  startListeningBtn: document.getElementById("startListeningBtn"),
  stopListeningBtn: document.getElementById("stopListeningBtn"),
  resetSessionBtn: document.getElementById("resetSessionBtn"),
  openSettingsBtn: document.getElementById("openSettingsBtn"),
  settingsModal: document.getElementById("settingsModal"),
  settingsBackdrop: document.getElementById("settingsBackdrop"),
  closeSettingsBtn: document.getElementById("closeSettingsBtn"),
  saveSettingsBtn: document.getElementById("saveSettingsBtn"),
  captureSourceMicrophone: document.getElementById("captureSourceMicrophone"),
  captureSourceSpeaker: document.getElementById("captureSourceSpeaker"),
  microphoneSettingsSection: document.getElementById("microphoneSettingsSection"),
  audioInputSelect: document.getElementById("audioInputSelect"),
  refreshAudioDevicesBtn: document.getElementById("refreshAudioDevicesBtn"),
  audioDeviceHint: document.getElementById("audioDeviceHint"),
  deleteSessionBtn: document.getElementById("deleteSessionBtn"),
  activeSessionTitle: document.getElementById("activeSessionTitle"),
  activeSessionId: document.getElementById("activeSessionId"),
  audioStats: document.getElementById("audioStats"),
  conversation: document.getElementById("conversation"),
  partialTranscript: document.getElementById("partialTranscript"),
  pendingSegment: document.getElementById("pendingSegment"),
  transcriptEditor: document.getElementById("transcriptEditor"),
  sendCursorSegmentBtn: document.getElementById("sendCursorSegmentBtn"),
  sendTailSegmentBtn: document.getElementById("sendTailSegmentBtn"),
  dockHint: document.getElementById("dockHint"),
  dockStartBtn: document.getElementById("dockStartBtn"),
  dockStopBtn: document.getElementById("dockStopBtn"),
};

const state = {
  runtime: {
    sttProvider: "",
    llmModel: "",
  },
  audio: {
    devices: [],
    selectedDeviceId: localStorage.getItem(AUDIO_INPUT_STORAGE_KEY) || "",
    captureSource: localStorage.getItem(CAPTURE_SOURCE_STORAGE_KEY) || "microphone",
    settingsOpen: false,
    draftDeviceId: "",
    draftCaptureSource: "microphone",
  },
  sessions: loadStoredSessions(),
  activeSessionId: "",
  eventSource: null,
  capture: null,
  sendChain: Promise.resolve(),
  snapshot: null,
  textDeal: emptyTextDealState(),
  transcriptCursorRaw: 0,
};

boot();

function boot() {
  const storedBackend = localStorage.getItem(BACKEND_STORAGE_KEY);
  if (storedBackend) {
    elements.backendBaseUrl.value = storedBackend;
  }

  bindActions();
  initializeAudioSettings();
  renderRuntime();
  renderSessionList();
  renderEmptyStage();
  void refreshRuntime();

  if (state.sessions.length > 0) {
    void selectSession(state.sessions[0].id, { openEvents: true });
  }
}

function bindActions() {
  elements.backendBaseUrl.addEventListener("change", handleBackendChange);
  elements.createSessionBtn.addEventListener("click", () => runAction(createSession));
  elements.startListeningBtn.addEventListener("click", () => runAction(startListening));
  elements.stopListeningBtn.addEventListener("click", () => runAction(stopListening));
  elements.resetSessionBtn.addEventListener("click", () => runAction(resetSession));
  elements.openSettingsBtn.addEventListener("click", () => runAction(openSettingsModal));
  elements.settingsBackdrop.addEventListener("click", closeSettingsModal);
  elements.closeSettingsBtn.addEventListener("click", closeSettingsModal);
  elements.saveSettingsBtn.addEventListener("click", () => runAction(saveAudioSettings));
  elements.captureSourceMicrophone.addEventListener("change", handleCaptureSourceDraftChange);
  elements.captureSourceSpeaker.addEventListener("change", handleCaptureSourceDraftChange);
  elements.audioInputSelect.addEventListener("change", handleAudioDeviceDraftChange);
  elements.refreshAudioDevicesBtn.addEventListener("click", () => runAction(() => refreshAudioDevices({ requestPermission: true })));
  elements.deleteSessionBtn.addEventListener("click", () => runAction(deleteActiveSession));
  elements.dockStartBtn?.addEventListener("click", () => runAction(startListening));
  elements.dockStopBtn?.addEventListener("click", () => runAction(stopListening));
  elements.sendCursorSegmentBtn.addEventListener("click", () => runAction(sendSegmentToCursor));
  elements.sendTailSegmentBtn.addEventListener("click", () => runAction(sendPendingTail));
  elements.transcriptEditor.addEventListener("click", syncTranscriptCursorFromSelection);
  elements.transcriptEditor.addEventListener("keyup", syncTranscriptCursorFromSelection);
  elements.transcriptEditor.addEventListener("select", syncTranscriptCursorFromSelection);
  document.addEventListener("keydown", handleGlobalKeydown);
  window.addEventListener("beforeunload", closeEvents);
}

async function runAction(action) {
  try {
    await action();
  } catch (error) {
    renderSystemMessage(error instanceof Error ? error.message : String(error));
  }
}

function handleBackendChange() {
  const value = backendBaseURL();
  localStorage.setItem(BACKEND_STORAGE_KEY, value);
  closeEvents();
  void refreshRuntime();
  if (state.activeSessionId) {
    void selectSession(state.activeSessionId, { openEvents: true });
  }
}

function initializeAudioSettings() {
  if (state.audio.captureSource !== "speaker") {
    state.audio.captureSource = "microphone";
  }
  state.audio.draftCaptureSource = state.audio.captureSource;
  state.audio.draftDeviceId = state.audio.selectedDeviceId;
  renderAudioSettingsModal();
  void refreshAudioDevices({ requestPermission: false }).catch(() => {});

  if (navigator.mediaDevices?.addEventListener) {
    navigator.mediaDevices.addEventListener("devicechange", handleAudioDeviceChange);
  }
}

async function refreshRuntime() {
  try {
    const health = await requestJSON("GET", "/api/healthz");
    const runtime = health && typeof health === "object" ? health.runtime || {} : {};
    state.runtime.sttProvider = runtime?.stt?.provider || "unknown";
    state.runtime.llmModel = runtime?.llm?.model || runtime?.llm?.provider || "unknown";
  } catch {
    state.runtime.sttProvider = "offline";
    state.runtime.llmModel = "offline";
  }
  renderRuntime();
}

async function openSettingsModal() {
  state.audio.settingsOpen = true;
  state.audio.draftCaptureSource = state.audio.captureSource;
  state.audio.draftDeviceId = state.audio.selectedDeviceId;
  renderAudioSettingsModal();
  await refreshAudioDevices({ requestPermission: false });
}

function closeSettingsModal() {
  state.audio.settingsOpen = false;
  renderAudioSettingsModal();
}

function handleGlobalKeydown(event) {
  if (event.key === "Escape" && state.audio.settingsOpen) {
    closeSettingsModal();
  }
}

function handleCaptureSourceDraftChange() {
  state.audio.draftCaptureSource = elements.captureSourceSpeaker.checked ? "speaker" : "microphone";
  renderAudioSettingsModal();
}

function handleAudioDeviceDraftChange() {
  state.audio.draftDeviceId = elements.audioInputSelect.value;
}

async function saveAudioSettings() {
  state.audio.captureSource = state.audio.draftCaptureSource;
  state.audio.selectedDeviceId = state.audio.draftDeviceId;
  persistAudioSettings();
  closeSettingsModal();
  renderDockHint(
    state.audio.captureSource === "speaker"
      ? "Settings saved. Start listening to share speaker/system audio."
      : "Settings saved. Start listening to use the selected microphone.",
  );
}

async function refreshAudioDevices(options = {}) {
  const { requestPermission = false } = options;

  if (!navigator.mediaDevices?.enumerateDevices) {
    state.audio.devices = [];
    renderAudioSettingsModal("This browser does not support audio device enumeration.");
    return;
  }

  if (requestPermission) {
    await requestMicrophonePermission();
  }

  const devices = await navigator.mediaDevices.enumerateDevices();
  const audioInputs = devices.filter((device) => device.kind === "audioinput");
  state.audio.devices = audioInputs.map((device, index) => ({
    id: device.deviceId,
    label: device.label || `Microphone ${index + 1}`,
  }));

  syncAudioDeviceSelection();
  renderAudioSettingsModal();
}

async function requestMicrophonePermission() {
  if (!navigator.mediaDevices?.getUserMedia) {
    return;
  }

  const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
  stream.getTracks().forEach((track) => track.stop());
}

async function handleAudioDeviceChange() {
  try {
    await refreshAudioDevices({ requestPermission: false });
  } catch {
    // Keep the current draft and selection on passive refresh failure.
  }
}

function syncAudioDeviceSelection() {
  if (state.audio.devices.length === 0) {
    state.audio.selectedDeviceId = "";
    state.audio.draftDeviceId = "";
    persistAudioSettings();
    return;
  }

  const selectedExists = state.audio.devices.some((device) => device.id === state.audio.selectedDeviceId);
  if (!selectedExists) {
    state.audio.selectedDeviceId = state.audio.devices[0].id;
  }

  const draftExists = state.audio.devices.some((device) => device.id === state.audio.draftDeviceId);
  if (!draftExists) {
    state.audio.draftDeviceId = state.audio.selectedDeviceId;
  }

  persistAudioSettings();
}

function renderAudioSettingsModal(message) {
  elements.settingsModal.hidden = !state.audio.settingsOpen;
  elements.captureSourceMicrophone.checked = state.audio.draftCaptureSource !== "speaker";
  elements.captureSourceSpeaker.checked = state.audio.draftCaptureSource === "speaker";
  elements.microphoneSettingsSection.hidden = state.audio.draftCaptureSource === "speaker";

  const options = state.audio.devices.length > 0
    ? state.audio.devices
        .map((device) => {
          const selected = device.id === state.audio.draftDeviceId ? " selected" : "";
          return `<option value="${escapeHTML(device.id)}"${selected}>${escapeHTML(device.label)}</option>`;
        })
        .join("")
    : `<option value="">No microphones found</option>`;
  elements.audioInputSelect.innerHTML = options;
  elements.audioInputSelect.disabled = state.audio.devices.length === 0;

  const defaultMessage =
    state.audio.draftCaptureSource === "speaker"
      ? "Speaker capture uses the browser share dialog. Use Chrome or Edge, then choose a tab, window, or screen and enable audio sharing when prompted."
      : state.capture
        ? "Microphone changes apply the next time listening starts."
        : "Choose which microphone to use when listening starts.";
  elements.audioDeviceHint.textContent = message || defaultMessage;
}

async function createSession() {
  await refreshRuntime();
  const snapshot = await requestJSON("POST", "/api/sessions");
  upsertSessionMeta({
    id: snapshot.id,
    title: "New Session",
    preview: "Ready to listen",
    createdAt: new Date().toISOString(),
    listening: false,
    answering: false,
  });
  await selectSession(snapshot.id, { snapshot, openEvents: true });
}

async function selectSession(sessionId, options = {}) {
  const { snapshot = null, openEvents = true } = options;
  if (state.capture && sessionId !== state.activeSessionId) {
    await stopListening();
  }

  state.activeSessionId = sessionId;
  let nextSnapshot = snapshot;

  if (!nextSnapshot) {
    try {
      nextSnapshot = await requestJSON("GET", `/api/sessions/${encodeURIComponent(sessionId)}`);
    } catch (error) {
      if (error instanceof Error && error.message.includes("session not found")) {
        removeSessionMeta(sessionId);
        if (state.activeSessionId === sessionId) {
          state.activeSessionId = "";
          renderEmptyStage();
        }
      }
      throw error;
    }
  }

  applySnapshot(nextSnapshot);
  if (openEvents) {
    openEventsForActiveSession();
  }
}

function openEventsForActiveSession() {
  if (!state.activeSessionId) {
    return;
  }

  closeEvents();
  const source = new EventSource(`${backendBaseURL()}/api/sessions/${encodeURIComponent(state.activeSessionId)}/events`);
  state.eventSource = source;

  source.addEventListener("snapshot", (event) => {
    applySnapshot(parseJSON(event.data, "snapshot"));
  });

  for (const type of [
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
    "session.reset",
    "error",
  ]) {
    source.addEventListener(type, (event) => {
      handleEvent(type, parseJSON(event.data, type));
    });
  }

  source.onerror = () => {
    renderDockHint("Event stream disconnected. Browser will retry automatically.");
  };
}

function closeEvents() {
  if (state.eventSource) {
    state.eventSource.close();
    state.eventSource = null;
  }
}

async function startListening() {
  ensureActiveSession();
  ensureCaptureAvailable();

  await startAudioCapture((samples) => {
    const pcm = float32ToPCM16(samples);
    queueAudioChunk(pcm.buffer.slice(0));
  });

  try {
    await requestJSON("POST", `/api/sessions/${encodeURIComponent(state.activeSessionId)}/listen/start`);
  } catch (error) {
    await stopCaptureIfNeeded();
    throw error;
  }

  state.sendChain = Promise.resolve();

  renderDockHint(
    state.audio.captureSource === "speaker"
      ? "Listening for speaker/system audio. Make sure browser audio sharing is enabled."
      : "Listening for microphone audio. Stable transcript will accumulate below.",
  );
}

async function stopListening() {
  await stopCaptureIfNeeded();
  await state.sendChain.catch(() => {});

  if (!state.activeSessionId) {
    return;
  }

  await requestJSON("POST", `/api/sessions/${encodeURIComponent(state.activeSessionId)}/listen/stop`);
  renderDockHint("Listening stopped.");
}

async function resetSession() {
  ensureActiveSession();
  await stopCaptureIfNeeded();
  await requestJSON("POST", `/api/sessions/${encodeURIComponent(state.activeSessionId)}/reset`);
  renderDockHint("Session reset.");
}

async function deleteActiveSession() {
  ensureActiveSession();

  const sessionID = state.activeSessionId;
  const nextSession = state.sessions.find((item) => item.id !== sessionID) || null;

  await stopCaptureIfNeeded();
  closeEvents();

  await requestJSON("DELETE", `/api/sessions/${encodeURIComponent(sessionID)}`);
  removeSessionMeta(sessionID);

  if (nextSession) {
    await selectSession(nextSession.id, { openEvents: true });
    renderDockHint(`Deleted session ${sessionID.slice(0, 8)}.`);
    return;
  }

  state.activeSessionId = "";
  renderEmptyStage();
  renderDockHint("Session deleted.");
}

function queueAudioChunk(payload) {
  if (!state.activeSessionId) {
    return;
  }

  state.sendChain = state.sendChain
    .catch(() => {})
    .then(() => postAudioChunk(payload))
    .catch((error) => {
      renderSystemMessage(error instanceof Error ? error.message : String(error));
    });
}

async function postAudioChunk(payload) {
  const response = await fetch(
    `${backendBaseURL()}/api/sessions/${encodeURIComponent(state.activeSessionId)}/audio?sampleRate=16000&channels=1&encoding=pcm16`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
      },
      body: payload,
    },
  );

  if (!response.ok) {
    throw new Error(await responseError(response));
  }
}

async function sendSegmentToCursor() {
  ensureActiveSession();
  const stop = getTranscriptCursorRaw();
  if (stop <= state.textDeal.sentUntil) {
    throw new Error("Move the stop marker forward before sending.");
  }
  await submitTextDealStop(stop);
}

async function sendPendingTail() {
  ensureActiveSession();
  const stop = codePointLength(state.textDeal.stableText);
  if (stop <= state.textDeal.sentUntil) {
    throw new Error("There is no pending stable text to send.");
  }
  state.transcriptCursorRaw = stop;
  await submitTextDealStop(stop);
}

async function submitTextDealStop(stop) {
  await requestJSON("POST", `/api/sessions/${encodeURIComponent(state.activeSessionId)}/textdeal/segment`, {
    stop,
  });
  renderDockHint("Segment submitted to LLM. The answer stream will continue in the center panel.");
}

function handleEvent(type, record) {
  const payload = record && typeof record === "object" ? record.data || {} : {};

  switch (type) {
    case "state":
      if (state.snapshot) {
        state.snapshot.listening = Boolean(payload.listening);
        state.snapshot.answerInProgress = Boolean(payload.answerInProgress);
        renderSnapshotState();
      }
      break;
    case "stt.partial":
      if (state.snapshot) {
        state.snapshot.partialTranscript = payload.text || "";
        renderTranscripts();
      }
      break;
    case "stt.final":
      if (state.snapshot && payload.text) {
        state.snapshot.finalTranscripts = Array.isArray(state.snapshot.finalTranscripts)
          ? [...state.snapshot.finalTranscripts, payload.text].slice(-20)
          : [payload.text];
      }
      break;
    case "textdeal.updated":
      applyTextDealSnapshot(payload);
      break;
    case "question.detected":
    case "question.manual":
      if (state.snapshot) {
        state.snapshot.currentQuestion = payload.text || "";
        renderConversation();
      }
      break;
    case "llm.started":
      if (state.snapshot) {
        state.snapshot.currentQuestion = payload.question || state.snapshot.currentQuestion || "";
        state.snapshot.currentAnswer = "";
        state.snapshot.answerInProgress = true;
        renderConversation();
        renderSnapshotState();
      }
      renderDockHint("LLM is responding.");
      break;
    case "llm.token":
      if (state.snapshot) {
        state.snapshot.currentAnswer = `${state.snapshot.currentAnswer || ""}${payload.token || ""}`;
        renderConversation();
      }
      break;
    case "llm.done":
      if (state.snapshot) {
        state.snapshot.currentAnswer = payload.answer || state.snapshot.currentAnswer || "";
        state.snapshot.answerInProgress = false;
        renderConversation();
        renderSnapshotState();
      }
      renderDockHint("LLM finished answering.");
      void refreshActiveSnapshot();
      break;
    case "llm.cancelled":
    case "llm.interrupted":
      renderDockHint(type === "llm.cancelled" ? "LLM cancelled." : "Previous answer interrupted by a new segment.");
      break;
    case "session.reset":
      if (state.snapshot) {
        applySnapshot({
          ...state.snapshot,
          partialTranscript: "",
          finalTranscripts: [],
          currentQuestion: "",
          currentAnswer: "",
          history: [],
          textDeal: emptyTextDealState(),
        });
      }
      break;
    case "error":
      renderSystemMessage(payload.message || "Unknown backend error");
      break;
    default:
      break;
  }
}

async function refreshActiveSnapshot() {
  if (!state.activeSessionId) {
    return;
  }
  applySnapshot(await requestJSON("GET", `/api/sessions/${encodeURIComponent(state.activeSessionId)}`));
}

function applySnapshot(snapshot) {
  state.snapshot = snapshot;
  state.activeSessionId = snapshot.id || state.activeSessionId;
  applyTextDealSnapshot(snapshot.textDeal || emptyTextDealState());
  renderSnapshotState();
  renderConversation();
  renderTranscripts();

  upsertSessionMeta({
    id: state.activeSessionId,
    title: buildSessionTitle(snapshot),
    preview: snapshot.currentQuestion || snapshot.textDeal?.pendingText || "Awaiting input",
    createdAt: findSessionMeta(state.activeSessionId)?.createdAt || new Date().toISOString(),
    listening: Boolean(snapshot.listening),
    answering: Boolean(snapshot.answerInProgress),
  });
}

function applyTextDealSnapshot(textDeal) {
  const next = normalizeTextDeal(textDeal);
  const followTail = document.activeElement !== elements.transcriptEditor;

  state.textDeal = next;
  if (followTail) {
    state.transcriptCursorRaw = codePointLength(next.stableText);
  } else {
    state.transcriptCursorRaw = clamp(state.transcriptCursorRaw, next.sentUntil, codePointLength(next.stableText));
  }

  renderTranscriptEditor();
  renderTranscripts();
}

function renderSnapshotState() {
  const snapshot = state.snapshot;
  elements.activeSessionTitle.textContent = buildSessionTitle(snapshot);
  elements.activeSessionId.textContent = snapshot?.id || "none";
  elements.listeningState.textContent = String(Boolean(snapshot?.listening));
  elements.answeringState.textContent = String(Boolean(snapshot?.answerInProgress));
  elements.audioStats.textContent = `${snapshot?.audio?.chunks || 0} chunks`;
}

function renderConversation() {
  const snapshot = state.snapshot;
  if (!snapshot) {
    elements.conversation.innerHTML = "";
    return;
  }

  const parts = [];
  const history = Array.isArray(snapshot.history) ? snapshot.history : [];

  if (history.length === 0 && !snapshot.currentQuestion && !snapshot.currentAnswer) {
    parts.push(systemMessageMarkup("No answer yet. Start listening and send a stable segment to the LLM."));
  }

  for (const turn of history) {
    parts.push(messageMarkup("user", turn.question || ""));
    parts.push(messageMarkup("assistant", turn.answer || ""));
  }

  const lastTurn = history.length > 0 ? history[history.length - 1] : null;
  const showCurrent =
    snapshot.currentQuestion &&
    (!lastTurn ||
      lastTurn.question !== snapshot.currentQuestion ||
      lastTurn.answer !== snapshot.currentAnswer ||
      snapshot.answerInProgress);

  if (showCurrent) {
    parts.push(messageMarkup("user", snapshot.currentQuestion || ""));
    parts.push(messageMarkup("assistant", snapshot.currentAnswer || (snapshot.answerInProgress ? "..." : "")));
  }

  elements.conversation.innerHTML = parts.join("");
  elements.conversation.scrollTop = elements.conversation.scrollHeight;
}

function renderTranscripts() {
  const partial = state.snapshot?.partialTranscript || "";
  const pending = state.textDeal.pendingText || "";
  setText(elements.partialTranscript, partial, "Waiting for transcript.");
  setText(elements.pendingSegment, pending, "No unsent stable text.");
}

function renderTranscriptEditor() {
  const display = buildTranscriptDisplay(state.textDeal);
  const rawCursor = clamp(state.transcriptCursorRaw, state.textDeal.sentUntil, codePointLength(state.textDeal.stableText));
  const displayCursor = rawCursor + state.textDeal.markers.filter((marker) => marker <= rawCursor).length;
  const codeUnitOffset = toCodeUnitOffset(display, displayCursor);

  elements.transcriptEditor.value = display;
  elements.transcriptEditor.setSelectionRange(codeUnitOffset, codeUnitOffset);
}

function renderSessionList() {
  if (state.sessions.length === 0) {
    elements.sessionList.innerHTML = `
      <div class="session-item">
        <strong>No sessions</strong>
        <span>Create a session to begin.</span>
      </div>
    `;
    return;
  }

  elements.sessionList.innerHTML = state.sessions
    .map((session) => {
      const active = session.id === state.activeSessionId ? " active" : "";
      const status = session.answering ? "answering" : session.listening ? "listening" : "idle";
      return `
        <button class="session-item${active}" data-session-id="${escapeHTML(session.id)}" type="button">
          <strong>${escapeHTML(session.title || session.id)}</strong>
          <span>${escapeHTML(truncate(session.preview || "Awaiting input", 58))}</span>
          <small>${escapeHTML(status)} - ${escapeHTML(session.id.slice(0, 8))}</small>
        </button>
      `;
    })
    .join("");

  for (const button of elements.sessionList.querySelectorAll("[data-session-id]")) {
    button.addEventListener("click", () => {
      void runAction(() => selectSession(button.dataset.sessionId));
    });
  }
}

function renderRuntime() {
  elements.runtimeSttProvider.textContent = state.runtime.sttProvider || "unknown";
  elements.runtimeLlmModel.textContent = state.runtime.llmModel || "unknown";
}

function renderEmptyStage() {
  state.snapshot = null;
  state.textDeal = emptyTextDealState();
  state.transcriptCursorRaw = 0;
  elements.activeSessionTitle.textContent = "No Active Session";
  elements.activeSessionId.textContent = "none";
  elements.listeningState.textContent = "false";
  elements.answeringState.textContent = "false";
  elements.audioStats.textContent = "0 chunks";
  elements.conversation.innerHTML = systemMessageMarkup("Select a session from the left or create a new one.");
  setText(elements.partialTranscript, "", "Waiting for transcript.");
  setText(elements.pendingSegment, "", "No unsent stable text.");
  elements.transcriptEditor.value = "";
  renderDockHint("Create a session to begin.");
}

function renderSystemMessage(message) {
  renderDockHint(message);
  if (!state.snapshot) {
    elements.conversation.innerHTML = systemMessageMarkup(message);
    return;
  }
  elements.conversation.innerHTML = `${systemMessageMarkup(message)}${elements.conversation.innerHTML}`;
}

function renderDockHint(text) {
  elements.dockHint.textContent = text;
}

function messageMarkup(role, text) {
  const body = escapeHTML(text || (role === "assistant" ? "Waiting for answer." : ""));
  const roleLabel = role === "user" ? "Question" : "Answer";
  return `
    <div class="message-row ${role}">
      <article class="message-card">
        <span class="message-role">${roleLabel}</span>
        ${body}
      </article>
    </div>
  `;
}

function systemMessageMarkup(text) {
  return `
    <div class="message-row system">
      <article class="message-card">
        <span class="message-role">System</span>
        ${escapeHTML(text)}
      </article>
    </div>
  `;
}

function syncTranscriptCursorFromSelection() {
  const display = elements.transcriptEditor.value || "";
  const selectionStart = elements.transcriptEditor.selectionStart || 0;
  const rawCursor = codePointLength(display.slice(0, selectionStart).replaceAll("|", ""));
  state.transcriptCursorRaw = clamp(rawCursor, state.textDeal.sentUntil, codePointLength(state.textDeal.stableText));
}

function getTranscriptCursorRaw() {
  syncTranscriptCursorFromSelection();
  return state.transcriptCursorRaw;
}

function buildTranscriptDisplay(textDeal) {
  const chars = Array.from(textDeal.stableText || "");
  let offset = 0;
  for (const marker of textDeal.markers) {
    const position = clamp(marker + offset, 0, chars.length);
    chars.splice(position, 0, "|");
    offset += 1;
  }
  return chars.join("");
}

function normalizeTextDeal(textDeal) {
  const stableText = typeof textDeal?.stableText === "string" ? textDeal.stableText : "";
  const total = codePointLength(stableText);
  const sentUntil = clamp(Number.isFinite(textDeal?.sentUntil) ? Number(textDeal.sentUntil) : 0, 0, total);
  const markers = Array.isArray(textDeal?.markers)
    ? [...new Set(textDeal.markers.map(Number).filter((item) => Number.isInteger(item) && item > 0 && item <= total))].sort((left, right) => left - right)
    : [];
  const pendingText = typeof textDeal?.pendingText === "string" ? textDeal.pendingText : "";

  return {
    stableText,
    sentUntil,
    markers,
    pendingText,
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

function buildSessionTitle(snapshot) {
  if (!snapshot) {
    return "No Active Session";
  }
  if (snapshot.currentQuestion) {
    return truncate(snapshot.currentQuestion, 42);
  }
  return `Session ${String(snapshot.id || "").slice(0, 8)}`;
}

function ensureActiveSession() {
  if (!state.activeSessionId) {
    throw new Error("No active session. Create one first.");
  }
}

function ensureCaptureAvailable() {
  if (state.capture) {
    throw new Error("Audio capture is already running.");
  }
}

async function startAudioCapture(onSamples) {
  const stream =
    state.audio.captureSource === "speaker"
      ? await requestSpeakerAudioStream()
      : await requestMicrophoneStream();

  const audioTracks = stream.getAudioTracks();
  if (audioTracks.length === 0) {
    stream.getTracks().forEach((track) => track.stop());
    throw new Error("No audio track is available from the selected source.");
  }

  const processingStream = new MediaStream([audioTracks[0]]);

  const AudioContextClass = window.AudioContext || window.webkitAudioContext;
  if (!AudioContextClass) {
    stream.getTracks().forEach((track) => track.stop());
    throw new Error("Web Audio API is not available in this browser");
  }

  const audioContext = new AudioContextClass();
  await audioContext.resume();

  const source = audioContext.createMediaStreamSource(processingStream);
  const inputChannels = Math.max(1, Math.min(2, source.channelCount || audioTracks[0].getSettings?.().channelCount || 2));
  const processor = audioContext.createScriptProcessor(4096, inputChannels, 1);
  const mute = audioContext.createGain();
  mute.gain.value = 0;

  processor.onaudioprocess = (event) => {
    const mono = mixInputBufferToMono(event.inputBuffer);
    const resampled = resampleFloat32(mono, audioContext.sampleRate, 16000);
    if (resampled.length > 0) {
      onSamples(resampled);
    }
  };

  source.connect(processor);
  processor.connect(mute);
  mute.connect(audioContext.destination);

  state.capture = {
    stop: async () => {
      processor.disconnect();
      source.disconnect();
      mute.disconnect();
      stream.getTracks().forEach((track) => track.stop());
      await audioContext.close();
    },
  };
}

async function requestMicrophoneStream() {
  const audioConstraints = {
    channelCount: 1,
    echoCancellation: false,
    noiseSuppression: false,
    autoGainControl: false,
  };
  if (state.audio.selectedDeviceId) {
    audioConstraints.deviceId = { exact: state.audio.selectedDeviceId };
  }

  return navigator.mediaDevices.getUserMedia({
    audio: audioConstraints,
  });
}

async function requestSpeakerAudioStream() {
  if (!navigator.mediaDevices?.getDisplayMedia) {
    throw new Error("This browser does not support speaker/system audio capture.");
  }

  if (!isChromiumBrowser()) {
    throw new Error("Speaker/system audio capture currently requires Chrome or Edge. Firefox and Safari usually do not expose a shareable system-audio track.");
  }

  const stream = await navigator.mediaDevices.getDisplayMedia({
    video: true,
    audio: {
      echoCancellation: false,
      noiseSuppression: false,
      autoGainControl: false,
      suppressLocalAudioPlayback: false,
    },
  });

  if (stream.getAudioTracks().length === 0) {
    stream.getTracks().forEach((track) => track.stop());
    throw new Error("No shared audio track was provided. In the browser share dialog, select a tab/window/screen that supports audio and enable audio sharing.");
  }

  return stream;
}

function mixInputBufferToMono(inputBuffer) {
  const channelCount = inputBuffer.numberOfChannels || 1;
  const sampleCount = inputBuffer.length;
  const mixed = new Float32Array(sampleCount);

  if (channelCount === 1) {
    mixed.set(inputBuffer.getChannelData(0));
    return mixed;
  }

  for (let channelIndex = 0; channelIndex < channelCount; channelIndex += 1) {
    const channelData = inputBuffer.getChannelData(channelIndex);
    for (let sampleIndex = 0; sampleIndex < sampleCount; sampleIndex += 1) {
      mixed[sampleIndex] += channelData[sampleIndex];
    }
  }

  const scale = 1 / channelCount;
  for (let sampleIndex = 0; sampleIndex < sampleCount; sampleIndex += 1) {
    mixed[sampleIndex] *= scale;
  }

  return mixed;
}

function isChromiumBrowser() {
  if (Array.isArray(navigator.userAgentData?.brands)) {
    return navigator.userAgentData.brands.some((brand) =>
      /Chromium|Google Chrome|Microsoft Edge/i.test(brand.brand),
    );
  }

  const userAgent = navigator.userAgent || "";
  return /Chrome|Chromium|Edg\//.test(userAgent) && !/Firefox\//.test(userAgent);
}

async function stopCaptureIfNeeded() {
  if (!state.capture) {
    return;
  }
  const capture = state.capture;
  state.capture = null;
  await capture.stop();
}

function upsertSessionMeta(meta) {
  const next = state.sessions.filter((item) => item.id !== meta.id);
  next.unshift(meta);
  state.sessions = next.slice(0, 20);
  persistSessions();
  renderSessionList();
}

function removeSessionMeta(sessionId) {
  state.sessions = state.sessions.filter((item) => item.id !== sessionId);
  persistSessions();
  renderSessionList();
}

function findSessionMeta(sessionId) {
  return state.sessions.find((item) => item.id === sessionId) || null;
}

function persistSessions() {
  localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(state.sessions));
}

function persistAudioSettings() {
  if (state.audio.selectedDeviceId) {
    localStorage.setItem(AUDIO_INPUT_STORAGE_KEY, state.audio.selectedDeviceId);
  } else {
    localStorage.removeItem(AUDIO_INPUT_STORAGE_KEY);
  }

  localStorage.setItem(CAPTURE_SOURCE_STORAGE_KEY, state.audio.captureSource);
}

function loadStoredSessions() {
  try {
    const raw = localStorage.getItem(SESSION_STORAGE_KEY);
    const parsed = JSON.parse(raw || "[]");
    if (!Array.isArray(parsed)) {
      return [];
    }
    return parsed
      .filter((item) => item && typeof item.id === "string")
      .map((item) => ({
        id: item.id,
        title: typeof item.title === "string" ? item.title : item.id,
        preview: typeof item.preview === "string" ? item.preview : "",
        createdAt: typeof item.createdAt === "string" ? item.createdAt : new Date().toISOString(),
        listening: Boolean(item.listening),
        answering: Boolean(item.answering),
      }));
  } catch {
    return [];
  }
}

function setText(element, text, fallback) {
  const next = text && text.length > 0 ? text : fallback;
  element.textContent = next;
  element.classList.toggle("muted", !text);
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

function truncate(text, maxLength) {
  if (!text || text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, Math.max(0, maxLength - 3))}...`;
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

function backendBaseURL() {
  return elements.backendBaseUrl.value.trim().replace(/\/$/, "");
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

  if (response.status === 204) {
    return null;
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
