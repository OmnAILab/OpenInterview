const SESSION_STORAGE_KEY = "openinterview.frontend.sessions";
const BACKEND_STORAGE_KEY = "openinterview.frontend.backend";
const AUDIO_INPUT_STORAGE_KEY = "openinterview.frontend.audioInputId";
const CAPTURE_SOURCE_STORAGE_KEY = "openinterview.frontend.captureSource";
const LANGUAGE_STORAGE_KEY = "openinterview.frontend.language";
const THEME_STORAGE_KEY = "openinterview.frontend.theme";
const MARKDOWN_STORAGE_KEY = "openinterview.frontend.markdownEnabled";

const translations = {
  "zh-CN": {
    "app.title": "OpenInterview",
    "sidebar.sessions": "会话",
    "sidebar.backend": "后端",
    "sidebar.newSession": "新建会话",
    "status.listening": "监听中",
    "status.answering": "回答中",
    "status.true": "是",
    "status.false": "否",
    "status.idle": "空闲",
    "status.answeringValue": "回答中",
    "status.listeningValue": "监听中",
    "controls.startListen": "开始监听",
    "controls.stop": "停止",
    "controls.reset": "重置",
    "controls.settings": "设置",
    "controls.deleteSession": "删除会话",
    "controls.sendToCursor": "发送到LLM",
    "controls.sendPendingTail": "截止符",
    "controls.close": "关闭",
    "controls.refreshDevices": "刷新设备",
    "controls.saveSettings": "保存设置",
    "stage.eyebrow": "面试助手",
    "stage.session": "会话",
    "stage.audio": "音频",
    "stage.interviewResponse": "LLM 回答",
    "stage.stableTranscript": "稳态文本",
    "stage.partial": "临时结果",
    "stage.pending": "待发送",
    "stage.placeStop": "点击放置下一个停止符",
    "stage.transcriptPlaceholder": "稳态文本会显示在这里。",
    "stage.noActiveSession": "当前没有会话",
    "stage.createSessionHint": "从左侧创建或选择一个会话。",
    "stage.awaitingTranscript": "等待转写结果。",
    "stage.noPendingStableText": "暂无未发送的稳态文本。",
    "stage.noAnswerYet": "还没有回答。开始监听并手动发送一段稳态文本给 LLM。",
    "stage.waitingAnswer": "等待回答。",
    "stage.question": "问题",
    "stage.answer": "回答",
    "stage.system": "系统",
    "session.defaultTitle": "会话 {id}",
    "session.newTitle": "新会话",
    "session.readyToListen": "准备开始监听",
    "session.awaitingInput": "等待输入",
    "session.none": "无",
    "session.noneFound": "暂无会话",
    "session.createToBegin": "创建一个会话开始使用。",
    "session.deletedShort": "已删除会话 {id}。",
    "session.deleted": "会话已删除。",
    "settings.eyebrow": "设置",
    "settings.title": "采集与界面",
    "settings.languageEyebrow": "语言",
    "settings.languageTitle": "界面语言",
    "settings.languageLabel": "语言",
    "settings.themeEyebrow": "主题",
    "settings.themeTitle": "外观",
    "settings.themeLabel": "主题模式",
    "settings.themeSystem": "跟随系统",
    "settings.themeLight": "浅色",
    "settings.themeDark": "深色",
    "settings.markdownEyebrow": "Markdown",
    "settings.markdownTitle": "消息渲染",
    "settings.markdownEnabled": "启用 Markdown",
    "settings.markdownDesc": "用标题、列表、代码块和链接来渲染消息内容。",
    "settings.sourceEyebrow": "音频来源",
    "settings.sourceTitle": "选择采集对象",
    "settings.microphoneTitle": "麦克风",
    "settings.microphoneDesc": "采集指定的音频输入设备。",
    "settings.speakerTitle": "扬声器 / 系统音频",
    "settings.speakerDesc": "通过浏览器共享窗口，采集标签页、窗口或屏幕中的音频。",
    "settings.microphoneEyebrow": "麦克风",
    "settings.inputDeviceTitle": "输入设备",
    "settings.deviceLabel": "设备",
    "settings.notesEyebrow": "说明",
    "settings.notesTitle": "浏览器行为",
    "settings.noMicrophones": "未找到麦克风设备",
    "settings.microphoneFallback": "麦克风 {index}",
    "settings.savedSpeaker": "设置已保存。开始监听后会请求共享扬声器/系统音频。",
    "settings.savedMicrophone": "设置已保存。开始监听后会使用已选择的麦克风。",
    "settings.noEnumeration": "当前浏览器不支持音频设备枚举。",
    "settings.noteSpeaker": "扬声器采集依赖浏览器共享音频。请使用 Chrome 或 Edge，并在弹窗中选择标签页、窗口或屏幕后启用音频共享。",
    "settings.noteMicrophoneActive": "麦克风切换会在下次开始监听时生效。",
    "settings.noteMicrophoneIdle": "选择开始监听时要使用的麦克风。",
    "runtime.unknown": "未知",
    "runtime.offline": "离线",
    "runtime.audioChunks": "{count} 个音频块",
    "hint.eventsDisconnected": "事件流已断开，浏览器会自动重连。",
    "hint.listeningSpeaker": "正在采集扬声器/系统音频，请确认浏览器共享音频已开启。",
    "hint.listeningMicrophone": "正在采集麦克风音频，稳态文本会持续累积在右侧。",
    "hint.listeningStopped": "已停止监听。",
    "hint.sessionReset": "会话已重置。",
    "hint.segmentSubmitted": "片段已提交给 LLM，回答会继续显示在中间面板。",
    "hint.llmResponding": "LLM 正在生成回答。",
    "hint.llmDone": "LLM 已完成回答。",
    "hint.llmCancelled": "LLM 已取消。",
    "hint.llmInterrupted": "上一轮回答已被新的片段打断。",
    "error.moveStopForward": "请将停止符继续向后移动后再发送。",
    "error.noPendingStableText": "当前没有可发送的剩余稳态文本。",
    "error.noActiveSession": "当前没有会话，请先创建一个。",
    "error.captureRunning": "音频采集已经在运行中。",
    "error.noAudioTrack": "所选来源没有可用的音轨。",
    "error.noWebAudio": "当前浏览器不支持 Web Audio API。",
    "error.noDisplayMedia": "当前浏览器不支持扬声器/系统音频采集。",
    "error.speakerRequiresChromium": "扬声器/系统音频采集目前需要 Chrome 或 Edge，Firefox 和 Safari 通常不会暴露可共享的系统音轨。",
    "error.noSharedAudioTrack": "未拿到共享音轨。请在浏览器共享弹窗中选择支持音频的标签页、窗口或屏幕，并启用音频共享。",
    "error.failedToParse": "解析 {label} 失败：{message}",
    "error.unknownBackend": "未知后端错误",
  },
  "en-US": {
    "app.title": "OpenInterview",
    "sidebar.sessions": "Sessions",
    "sidebar.backend": "Backend",
    "sidebar.newSession": "New Session",
    "status.listening": "Listening",
    "status.answering": "Answering",
    "status.true": "true",
    "status.false": "false",
    "status.idle": "idle",
    "status.answeringValue": "answering",
    "status.listeningValue": "listening",
    "controls.startListen": "Start Listen",
    "controls.stop": "Stop",
    "controls.reset": "Reset",
    "controls.settings": "Settings",
    "controls.deleteSession": "Delete Session",
    "controls.sendToCursor": "Send To LLM",
    "controls.sendPendingTail": "Pending Tail",
    "controls.close": "Close",
    "controls.refreshDevices": "Refresh Devices",
    "controls.saveSettings": "Save Settings",
    "stage.eyebrow": "Interview Copilot",
    "stage.session": "Session",
    "stage.audio": "Audio",
    "stage.interviewResponse": "Interview Response",
    "stage.stableTranscript": "Stable Transcript",
    "stage.partial": "Partial",
    "stage.pending": "Pending",
    "stage.placeStop": "Click to place the next stop marker",
    "stage.transcriptPlaceholder": "Stable transcript will appear here.",
    "stage.noActiveSession": "No Active Session",
    "stage.createSessionHint": "Create or select a session from the left.",
    "stage.awaitingTranscript": "Waiting for transcript.",
    "stage.noPendingStableText": "No unsent stable text.",
    "stage.noAnswerYet": "No answer yet. Start listening and send a stable segment to the LLM.",
    "stage.waitingAnswer": "Waiting for answer.",
    "stage.question": "Question",
    "stage.answer": "Answer",
    "stage.system": "System",
    "session.defaultTitle": "Session {id}",
    "session.newTitle": "New Session",
    "session.readyToListen": "Ready to listen",
    "session.awaitingInput": "Awaiting input",
    "session.none": "none",
    "session.noneFound": "No sessions",
    "session.createToBegin": "Create a session to begin.",
    "session.deletedShort": "Deleted session {id}.",
    "session.deleted": "Session deleted.",
    "settings.eyebrow": "Settings",
    "settings.title": "Capture & Interface",
    "settings.languageEyebrow": "Language",
    "settings.languageTitle": "Display Language",
    "settings.languageLabel": "Language",
    "settings.themeEyebrow": "Theme",
    "settings.themeTitle": "Appearance",
    "settings.themeLabel": "Theme Mode",
    "settings.themeSystem": "System",
    "settings.themeLight": "Light",
    "settings.themeDark": "Dark",
    "settings.markdownEyebrow": "Markdown",
    "settings.markdownTitle": "Message Rendering",
    "settings.markdownEnabled": "Enable Markdown",
    "settings.markdownDesc": "Render message content with headings, lists, code blocks, and links.",
    "settings.sourceEyebrow": "Source",
    "settings.sourceTitle": "Choose What To Capture",
    "settings.microphoneTitle": "Microphone",
    "settings.microphoneDesc": "Capture a selected audio input device.",
    "settings.speakerTitle": "Speaker / System Audio",
    "settings.speakerDesc": "Capture shared tab, window, or screen audio through the browser share dialog.",
    "settings.microphoneEyebrow": "Microphone",
    "settings.inputDeviceTitle": "Input Device",
    "settings.deviceLabel": "Device",
    "settings.notesEyebrow": "Notes",
    "settings.notesTitle": "Browser Behavior",
    "settings.noMicrophones": "No microphones found",
    "settings.microphoneFallback": "Microphone {index}",
    "settings.savedSpeaker": "Settings saved. Start listening to share speaker/system audio.",
    "settings.savedMicrophone": "Settings saved. Start listening to use the selected microphone.",
    "settings.noEnumeration": "This browser does not support audio device enumeration.",
    "settings.noteSpeaker": "Speaker capture uses the browser share dialog. Use Chrome or Edge, then choose a tab, window, or screen and enable audio sharing when prompted.",
    "settings.noteMicrophoneActive": "Microphone changes apply the next time listening starts.",
    "settings.noteMicrophoneIdle": "Choose which microphone to use when listening starts.",
    "runtime.unknown": "unknown",
    "runtime.offline": "offline",
    "runtime.audioChunks": "{count} chunks",
    "hint.eventsDisconnected": "Event stream disconnected. Browser will retry automatically.",
    "hint.listeningSpeaker": "Listening for speaker/system audio. Make sure browser audio sharing is enabled.",
    "hint.listeningMicrophone": "Listening for microphone audio. Stable transcript will accumulate below.",
    "hint.listeningStopped": "Listening stopped.",
    "hint.sessionReset": "Session reset.",
    "hint.segmentSubmitted": "Segment submitted to LLM. The answer stream will continue in the center panel.",
    "hint.llmResponding": "LLM is responding.",
    "hint.llmDone": "LLM finished answering.",
    "hint.llmCancelled": "LLM cancelled.",
    "hint.llmInterrupted": "Previous answer interrupted by a new segment.",
    "error.moveStopForward": "Move the stop marker forward before sending.",
    "error.noPendingStableText": "There is no pending stable text to send.",
    "error.noActiveSession": "No active session. Create one first.",
    "error.captureRunning": "Audio capture is already running.",
    "error.noAudioTrack": "No audio track is available from the selected source.",
    "error.noWebAudio": "Web Audio API is not available in this browser",
    "error.noDisplayMedia": "This browser does not support speaker/system audio capture.",
    "error.speakerRequiresChromium": "Speaker/system audio capture currently requires Chrome or Edge. Firefox and Safari usually do not expose a shareable system-audio track.",
    "error.noSharedAudioTrack": "No shared audio track was provided. In the browser share dialog, select a tab/window/screen that supports audio and enable audio sharing.",
    "error.failedToParse": "Failed to parse {label}: {message}",
    "error.unknownBackend": "Unknown backend error",
  },
};

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
  languageSelect: document.getElementById("languageSelect"),
  themeSelect: document.getElementById("themeSelect"),
  markdownEnabledToggle: document.getElementById("markdownEnabledToggle"),
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
  preferences: {
    language: localStorage.getItem(LANGUAGE_STORAGE_KEY) || inferInitialLanguage(),
    theme: localStorage.getItem(THEME_STORAGE_KEY) || "system",
    markdownEnabled: localStorage.getItem(MARKDOWN_STORAGE_KEY) !== "false",
    draftLanguage: "",
    draftTheme: "system",
    draftMarkdownEnabled: true,
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

  initializePreferences();
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
  elements.languageSelect.addEventListener("change", handleLanguageDraftChange);
  elements.themeSelect.addEventListener("change", handleThemeDraftChange);
  elements.markdownEnabledToggle.addEventListener("change", handleMarkdownDraftChange);
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

function initializePreferences() {
  if (!translations[state.preferences.language]) {
    state.preferences.language = "en-US";
  }
  if (!["system", "light", "dark"].includes(state.preferences.theme)) {
    state.preferences.theme = "system";
  }

  state.preferences.draftLanguage = state.preferences.language;
  state.preferences.draftTheme = state.preferences.theme;
  state.preferences.draftMarkdownEnabled = state.preferences.markdownEnabled;

  applyTheme();
  renderTranslations();
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
  renderSettingsModal();
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
  state.preferences.draftLanguage = state.preferences.language;
  state.preferences.draftTheme = state.preferences.theme;
  state.preferences.draftMarkdownEnabled = state.preferences.markdownEnabled;
  renderSettingsModal();
  await refreshAudioDevices({ requestPermission: false });
}

function closeSettingsModal() {
  state.audio.settingsOpen = false;
  renderSettingsModal();
}

function handleGlobalKeydown(event) {
  if (event.key === "Escape" && state.audio.settingsOpen) {
    closeSettingsModal();
  }
}

function handleCaptureSourceDraftChange() {
  state.audio.draftCaptureSource = elements.captureSourceSpeaker.checked ? "speaker" : "microphone";
  renderSettingsModal();
}

function handleAudioDeviceDraftChange() {
  state.audio.draftDeviceId = elements.audioInputSelect.value;
}

function handleLanguageDraftChange() {
  state.preferences.draftLanguage = elements.languageSelect.value;
}

function handleThemeDraftChange() {
  state.preferences.draftTheme = elements.themeSelect.value;
}

function handleMarkdownDraftChange() {
  state.preferences.draftMarkdownEnabled = elements.markdownEnabledToggle.checked;
}

async function saveAudioSettings() {
  state.preferences.language = normalizeLanguage(elements.languageSelect.value || state.preferences.draftLanguage);
  state.preferences.theme = normalizeTheme(elements.themeSelect.value || state.preferences.draftTheme);
  state.preferences.markdownEnabled = Boolean(state.preferences.draftMarkdownEnabled);
  persistPreferences();
  applyTheme();
  renderTranslations();

  state.audio.captureSource = state.audio.draftCaptureSource;
  state.audio.selectedDeviceId = state.audio.draftDeviceId;
  persistAudioSettings();
  closeSettingsModal();
  renderDockHint(
    state.audio.captureSource === "speaker"
      ? t("settings.savedSpeaker")
      : t("settings.savedMicrophone"),
  );
}

async function refreshAudioDevices(options = {}) {
  const { requestPermission = false } = options;

  if (!navigator.mediaDevices?.enumerateDevices) {
    state.audio.devices = [];
    renderSettingsModal(t("settings.noEnumeration"));
    return;
  }

  if (requestPermission) {
    await requestMicrophonePermission();
  }

  const devices = await navigator.mediaDevices.enumerateDevices();
  const audioInputs = devices.filter((device) => device.kind === "audioinput");
  state.audio.devices = audioInputs.map((device, index) => ({
    id: device.deviceId,
    label: device.label || t("settings.microphoneFallback", { index: String(index + 1) }),
  }));

  syncAudioDeviceSelection();
  renderSettingsModal();
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

function renderSettingsModal(message) {
  elements.settingsModal.hidden = !state.audio.settingsOpen;
  elements.captureSourceMicrophone.checked = state.audio.draftCaptureSource !== "speaker";
  elements.captureSourceSpeaker.checked = state.audio.draftCaptureSource === "speaker";
  elements.microphoneSettingsSection.hidden = state.audio.draftCaptureSource === "speaker";
  elements.languageSelect.innerHTML = buildLanguageOptions();
  elements.themeSelect.innerHTML = buildThemeOptions();
  elements.languageSelect.value = normalizeLanguage(state.preferences.draftLanguage);
  elements.themeSelect.value = normalizeTheme(state.preferences.draftTheme);
  elements.markdownEnabledToggle.checked = Boolean(state.preferences.draftMarkdownEnabled);

  const options = state.audio.devices.length > 0
    ? state.audio.devices
        .map((device) => {
          const selected = device.id === state.audio.draftDeviceId ? " selected" : "";
          return `<option value="${escapeHTML(device.id)}"${selected}>${escapeHTML(device.label)}</option>`;
        })
        .join("")
    : `<option value="">${escapeHTML(t("settings.noMicrophones"))}</option>`;
  elements.audioInputSelect.innerHTML = options;
  elements.audioInputSelect.disabled = state.audio.devices.length === 0;

  const defaultMessage =
    state.audio.draftCaptureSource === "speaker"
      ? t("settings.noteSpeaker")
      : state.capture
        ? t("settings.noteMicrophoneActive")
        : t("settings.noteMicrophoneIdle");
  elements.audioDeviceHint.textContent = message || defaultMessage;
}

async function createSession() {
  await refreshRuntime();
  const snapshot = await requestJSON("POST", "/api/sessions");
  upsertSessionMeta({
    id: snapshot.id,
    title: t("session.newTitle"),
    preview: t("session.readyToListen"),
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
    renderDockHint(t("hint.eventsDisconnected"));
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
      ? t("hint.listeningSpeaker")
      : t("hint.listeningMicrophone"),
  );
}

async function stopListening() {
  await stopCaptureIfNeeded();
  await state.sendChain.catch(() => {});

  if (!state.activeSessionId) {
    return;
  }

  await requestJSON("POST", `/api/sessions/${encodeURIComponent(state.activeSessionId)}/listen/stop`);
  renderDockHint(t("hint.listeningStopped"));
}

async function resetSession() {
  ensureActiveSession();
  await stopCaptureIfNeeded();
  await requestJSON("POST", `/api/sessions/${encodeURIComponent(state.activeSessionId)}/reset`);
  renderDockHint(t("hint.sessionReset"));
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
    renderDockHint(t("session.deletedShort", { id: sessionID.slice(0, 8) }));
    return;
  }

  state.activeSessionId = "";
  renderEmptyStage();
  renderDockHint(t("session.deleted"));
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
    throw new Error(t("error.moveStopForward"));
  }
  await submitTextDealStop(stop);
}

async function sendPendingTail() {
  ensureActiveSession();
  const stop = codePointLength(state.textDeal.stableText);
  if (stop <= state.textDeal.sentUntil) {
    throw new Error(t("error.noPendingStableText"));
  }
  state.transcriptCursorRaw = stop;
  await submitTextDealStop(stop);
}

async function submitTextDealStop(stop) {
  await requestJSON("POST", `/api/sessions/${encodeURIComponent(state.activeSessionId)}/textdeal/segment`, {
    stop,
  });
  renderDockHint(t("hint.segmentSubmitted"));
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
      renderDockHint(t("hint.llmResponding"));
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
      renderDockHint(t("hint.llmDone"));
      void refreshActiveSnapshot();
      break;
    case "llm.cancelled":
    case "llm.interrupted":
      renderDockHint(type === "llm.cancelled" ? t("hint.llmCancelled") : t("hint.llmInterrupted"));
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
      renderSystemMessage(payload.message || t("error.unknownBackend"));
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
    preview: snapshot.currentQuestion || snapshot.textDeal?.pendingText || t("session.awaitingInput"),
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
  elements.activeSessionId.textContent = snapshot?.id || t("session.none");
  elements.listeningState.textContent = t(`status.${Boolean(snapshot?.listening)}`);
  elements.answeringState.textContent = t(`status.${Boolean(snapshot?.answerInProgress)}`);
  elements.audioStats.textContent = formatAudioChunks(snapshot?.audio?.chunks || 0);
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
    parts.push(systemMessageMarkup(t("stage.noAnswerYet")));
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
  setText(elements.partialTranscript, partial, t("stage.awaitingTranscript"));
  setText(elements.pendingSegment, pending, t("stage.noPendingStableText"));
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
        <strong>${escapeHTML(t("session.noneFound"))}</strong>
        <span>${escapeHTML(t("session.createToBegin"))}</span>
      </div>
    `;
    return;
  }

  elements.sessionList.innerHTML = state.sessions
    .map((session) => {
      const active = session.id === state.activeSessionId ? " active" : "";
      const status = session.answering ? "answeringValue" : session.listening ? "listeningValue" : "idle";
      return `
        <button class="session-item${active}" data-session-id="${escapeHTML(session.id)}" type="button">
          <strong>${escapeHTML(session.title || session.id)}</strong>
          <span>${escapeHTML(truncate(session.preview || t("session.awaitingInput"), 58))}</span>
          <small>${escapeHTML(t(`status.${status}`))} - ${escapeHTML(session.id.slice(0, 8))}</small>
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
  elements.runtimeSttProvider.textContent = localizeRuntimeValue(state.runtime.sttProvider);
  elements.runtimeLlmModel.textContent = localizeRuntimeValue(state.runtime.llmModel);
}

function renderEmptyStage() {
  state.snapshot = null;
  state.textDeal = emptyTextDealState();
  state.transcriptCursorRaw = 0;
  elements.activeSessionTitle.textContent = t("stage.noActiveSession");
  elements.activeSessionId.textContent = t("session.none");
  elements.listeningState.textContent = t("status.false");
  elements.answeringState.textContent = t("status.false");
  elements.audioStats.textContent = formatAudioChunks(0);
  elements.conversation.innerHTML = systemMessageMarkup(t("stage.createSessionHint"));
  setText(elements.partialTranscript, "", t("stage.awaitingTranscript"));
  setText(elements.pendingSegment, "", t("stage.noPendingStableText"));
  elements.transcriptEditor.value = "";
  renderDockHint(t("stage.createSessionHint"));
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
  const body = renderMessageBody(text || (role === "assistant" ? t("stage.waitingAnswer") : ""), role);
  const roleLabel = role === "user" ? t("stage.question") : t("stage.answer");
  return `
    <div class="message-row ${role}">
      <article class="message-card">
        <span class="message-role">${roleLabel}</span>
        <div class="message-body">${body}</div>
      </article>
    </div>
  `;
}

function systemMessageMarkup(text) {
  return `
    <div class="message-row system">
      <article class="message-card">
        <span class="message-role">${escapeHTML(t("stage.system"))}</span>
        <div class="message-body">${renderPlainText(text)}</div>
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
    return t("stage.noActiveSession");
  }
  if (snapshot.currentQuestion) {
    return truncate(snapshot.currentQuestion, 42);
  }
  return t("session.defaultTitle", { id: String(snapshot.id || "").slice(0, 8) });
}

function ensureActiveSession() {
  if (!state.activeSessionId) {
    throw new Error(t("error.noActiveSession"));
  }
}

function ensureCaptureAvailable() {
  if (state.capture) {
    throw new Error(t("error.captureRunning"));
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
    throw new Error(t("error.noAudioTrack"));
  }

  const processingStream = new MediaStream([audioTracks[0]]);

  const AudioContextClass = window.AudioContext || window.webkitAudioContext;
  if (!AudioContextClass) {
    stream.getTracks().forEach((track) => track.stop());
    throw new Error(t("error.noWebAudio"));
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
    throw new Error(t("error.noDisplayMedia"));
  }

  if (!isChromiumBrowser()) {
    throw new Error(t("error.speakerRequiresChromium"));
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
    throw new Error(t("error.noSharedAudioTrack"));
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

function persistPreferences() {
  localStorage.setItem(LANGUAGE_STORAGE_KEY, state.preferences.language);
  localStorage.setItem(THEME_STORAGE_KEY, state.preferences.theme);
  localStorage.setItem(MARKDOWN_STORAGE_KEY, String(state.preferences.markdownEnabled));
}

function persistAudioSettings() {
  if (state.audio.selectedDeviceId) {
    localStorage.setItem(AUDIO_INPUT_STORAGE_KEY, state.audio.selectedDeviceId);
  } else {
    localStorage.removeItem(AUDIO_INPUT_STORAGE_KEY);
  }

  localStorage.setItem(CAPTURE_SOURCE_STORAGE_KEY, state.audio.captureSource);
}

function applyTheme() {
  document.documentElement.dataset.theme = normalizeTheme(state.preferences.theme);
}

function renderTranslations() {
  document.documentElement.lang = normalizeLanguage(state.preferences.language);
  document.title = t("app.title");

  for (const node of document.querySelectorAll("[data-i18n]")) {
    const key = node.dataset.i18n;
    node.textContent = t(key);
  }

  for (const node of document.querySelectorAll("[data-i18n-placeholder]")) {
    const key = node.dataset.i18nPlaceholder;
    node.setAttribute("placeholder", t(key));
  }

  relocalizeSessionMeta();
  renderSettingsModal();
  renderRuntime();
  renderSessionList();

  if (state.snapshot) {
    renderSnapshotState();
    renderConversation();
    renderTranscripts();
    renderTranscriptEditor();
  } else {
    renderEmptyStage();
  }
}

function relocalizeSessionMeta() {
  state.sessions = state.sessions.map((session) => ({
    ...session,
    title: relocalizeSessionTitle(session),
    preview: relocalizeSessionPreview(session.preview),
  }));
  persistSessions();
}

function relocalizeSessionTitle(session) {
  if (!session.title) {
    return session.id;
  }

  const knownTitles = new Set(Object.values(translations).map((entry) => entry["session.newTitle"]));
  if (knownTitles.has(session.title)) {
    return t("session.newTitle");
  }

  const knownPrefixes = Object.values(translations).map((entry) => entry["session.defaultTitle"].split("{id}")[0]);
  if (knownPrefixes.some((prefix) => prefix && session.title.startsWith(prefix))) {
    return t("session.defaultTitle", { id: session.id.slice(0, 8) });
  }

  return session.title;
}

function relocalizeSessionPreview(preview) {
  const knownReady = new Set(Object.values(translations).map((entry) => entry["session.readyToListen"]));
  if (knownReady.has(preview)) {
    return t("session.readyToListen");
  }

  const knownAwaiting = new Set(Object.values(translations).map((entry) => entry["session.awaitingInput"]));
  if (knownAwaiting.has(preview)) {
    return t("session.awaitingInput");
  }

  return preview;
}

function buildLanguageOptions() {
  return [
    { value: "zh-CN", label: "中文" },
    { value: "en-US", label: "English" },
  ]
    .map((option) => `<option value="${option.value}">${escapeHTML(option.label)}</option>`)
    .join("");
}

function buildThemeOptions() {
  return [
    { value: "system", label: t("settings.themeSystem") },
    { value: "light", label: t("settings.themeLight") },
    { value: "dark", label: t("settings.themeDark") },
  ]
    .map((option) => `<option value="${option.value}">${escapeHTML(option.label)}</option>`)
    .join("");
}

function inferInitialLanguage() {
  const browserLanguage = navigator.language || "en-US";
  return browserLanguage.toLowerCase().startsWith("zh") ? "zh-CN" : "en-US";
}

function normalizeLanguage(language) {
  return translations[language] ? language : "en-US";
}

function normalizeTheme(theme) {
  return ["system", "light", "dark"].includes(theme) ? theme : "system";
}

function localizeRuntimeValue(value) {
  if (!value || value === "unknown") {
    return t("runtime.unknown");
  }
  if (value === "offline") {
    return t("runtime.offline");
  }
  return value;
}

function formatAudioChunks(count) {
  return t("runtime.audioChunks", { count: String(count) });
}

function t(key, vars = {}) {
  const language = normalizeLanguage(state.preferences.language);
  const catalog = translations[language] || translations["en-US"];
  const fallback = translations["en-US"];
  const template = catalog[key] ?? fallback[key] ?? key;

  return template.replace(/\{(\w+)\}/g, (_, name) => String(vars[name] ?? `{${name}}`));
}

function renderMessageBody(text, role) {
  if (!text) {
    return "";
  }

  const shouldUseMarkdown = state.preferences.markdownEnabled && (role === "assistant" || role === "user");
  return shouldUseMarkdown ? renderMarkdown(text) : renderPlainText(text);
}

function renderPlainText(text) {
  return escapeHTML(text).replace(/\n/g, "<br>");
}

function renderMarkdown(text) {
  const normalized = String(text || "").replace(/\r\n/g, "\n");
  const lines = normalized.split("\n");
  const blocks = [];
  let index = 0;

  while (index < lines.length) {
    const line = lines[index];
    const trimmed = line.trim();

    if (trimmed === "") {
      index += 1;
      continue;
    }

    if (/^```/.test(trimmed)) {
      const language = trimmed.slice(3).trim();
      const codeLines = [];
      index += 1;
      while (index < lines.length && !/^```/.test(lines[index].trim())) {
        codeLines.push(lines[index]);
        index += 1;
      }
      if (index < lines.length) {
        index += 1;
      }
      const languageClass = language ? ` class="language-${escapeHTML(language)}"` : "";
      blocks.push(`<pre><code${languageClass}>${escapeHTML(codeLines.join("\n"))}</code></pre>`);
      continue;
    }

    const headingMatch = line.match(/^(#{1,6})\s+(.*)$/);
    if (headingMatch) {
      const level = headingMatch[1].length;
      blocks.push(`<h${level}>${renderInlineMarkdown(headingMatch[2])}</h${level}>`);
      index += 1;
      continue;
    }

    if (/^---+$|^\*\*\*+$|^___+$/.test(trimmed)) {
      blocks.push("<hr>");
      index += 1;
      continue;
    }

    if (isMarkdownTableStart(lines, index)) {
      const table = parseMarkdownTable(lines, index);
      blocks.push(table.html);
      index = table.nextIndex;
      continue;
    }

    if (/^>\s?/.test(trimmed)) {
      const quoteLines = [];
      while (index < lines.length && /^>\s?/.test(lines[index].trim())) {
        quoteLines.push(lines[index].trim().replace(/^>\s?/, ""));
        index += 1;
      }
      blocks.push(`<blockquote>${quoteLines.map((quoteLine) => renderPlainText(quoteLine)).join("<br>")}</blockquote>`);
      continue;
    }

    if (parseListMarker(line)) {
      const list = parseMarkdownList(lines, index, getIndent(line));
      blocks.push(list.html);
      index = list.nextIndex;
      continue;
    }

    const paragraphLines = [];
    while (index < lines.length) {
      const current = lines[index];
      const currentTrimmed = current.trim();
      if (
        currentTrimmed === "" ||
        /^```/.test(currentTrimmed) ||
        isMarkdownTableStart(lines, index) ||
        /^(#{1,6})\s+/.test(current) ||
        /^>\s?/.test(currentTrimmed) ||
        parseListMarker(current) ||
        /^---+$|^\*\*\*+$|^___+$/.test(currentTrimmed)
      ) {
        break;
      }
      paragraphLines.push(current);
      index += 1;
    }
    blocks.push(`<p>${paragraphLines.map((paragraphLine) => renderInlineMarkdown(paragraphLine)).join("<br>")}</p>`);
  }

  return blocks.join("");
}

function parseMarkdownList(lines, startIndex, baseIndent) {
  const firstMarker = parseListMarker(lines[startIndex]);
  const ordered = firstMarker?.ordered || false;
  const tag = ordered ? "ol" : "ul";
  const items = [];
  let index = startIndex;

  while (index < lines.length) {
    const line = lines[index];
    const marker = parseListMarker(line);
    const indent = getIndent(line);

    if (!marker || indent !== baseIndent || marker.ordered !== ordered) {
      break;
    }

    const itemLines = [marker.content];
    index += 1;

    while (index < lines.length) {
      const nextLine = lines[index];
      const nextTrimmed = nextLine.trim();
      const nextMarker = parseListMarker(nextLine);
      const nextIndent = getIndent(nextLine);

      if (nextTrimmed === "") {
        itemLines.push("");
        index += 1;
        continue;
      }

      if (nextMarker && nextIndent === baseIndent) {
        break;
      }

      if (nextIndent > baseIndent) {
        itemLines.push(nextLine.slice(Math.min(nextLine.length, baseIndent + 2)));
        index += 1;
        continue;
      }

      break;
    }

    items.push(renderMarkdownListItem(itemLines));
  }

  return {
    html: `<${tag}>${items.join("")}</${tag}>`,
    nextIndex: index,
  };
}

function renderMarkdownListItem(lines) {
  const cleanLines = trimBlankEdges(lines);
  if (cleanLines.length === 0) {
    return "<li></li>";
  }

  const task = parseTaskMarker(cleanLines[0]);
  if (task) {
    cleanLines[0] = task.content;
  }

  let index = 0;
  const parts = [];
  const paragraphLines = [];

  while (index < cleanLines.length) {
    const line = cleanLines[index];
    const trimmed = line.trim();

    if (trimmed === "") {
      if (paragraphLines.length > 0) {
        parts.push(renderListParagraph(paragraphLines));
        paragraphLines.length = 0;
      }
      index += 1;
      continue;
    }

    if (parseListMarker(line)) {
      if (paragraphLines.length > 0) {
        parts.push(renderListParagraph(paragraphLines));
        paragraphLines.length = 0;
      }
      const nested = parseMarkdownList(cleanLines, index, getIndent(line));
      parts.push(nested.html);
      index = nested.nextIndex;
      continue;
    }

    paragraphLines.push(line);
    index += 1;
  }

  if (paragraphLines.length > 0) {
    parts.push(renderListParagraph(paragraphLines));
  }

  const taskCheckbox = task
    ? `<input class="task-list-checkbox" type="checkbox" disabled${task.checked ? " checked" : ""}>`
    : "";
  const taskClass = task ? " class=\"task-list-item\"" : "";
  return `<li${taskClass}>${taskCheckbox}${parts.join("")}</li>`;
}

function renderListParagraph(lines) {
  return lines.map((line) => renderInlineMarkdown(line.trim())).join("<br>");
}

function parseListMarker(line) {
  const match = String(line || "").match(/^(\s*)([-*+]|\d+\.)\s+(.*)$/);
  if (!match) {
    return null;
  }

  return {
    indent: match[1].length,
    ordered: /\d+\./.test(match[2]),
    content: match[3],
  };
}

function parseTaskMarker(text) {
  const match = String(text || "").match(/^\s*\[([ xX])\]\s+(.*)$/);
  if (!match) {
    return null;
  }

  return {
    checked: match[1].toLowerCase() === "x",
    content: match[2],
  };
}

function getIndent(line) {
  const match = String(line || "").match(/^\s*/);
  return match ? match[0].replace(/\t/g, "    ").length : 0;
}

function trimBlankEdges(lines) {
  let start = 0;
  let end = lines.length;
  while (start < end && String(lines[start]).trim() === "") {
    start += 1;
  }
  while (end > start && String(lines[end - 1]).trim() === "") {
    end -= 1;
  }
  return lines.slice(start, end);
}

function isMarkdownTableStart(lines, index) {
  if (index + 1 >= lines.length) {
    return false;
  }

  const header = lines[index].trim();
  const separator = lines[index + 1].trim();
  return header.includes("|") && isMarkdownTableSeparator(separator);
}

function isMarkdownTableSeparator(line) {
  if (!line.includes("|")) {
    return false;
  }

  const cells = splitMarkdownTableRow(line);
  return cells.length > 0 && cells.every((cell) => /^:?-{3,}:?$/.test(cell.trim()));
}

function parseMarkdownTable(lines, startIndex) {
  const headers = splitMarkdownTableRow(lines[startIndex]);
  const alignments = splitMarkdownTableRow(lines[startIndex + 1]).map(parseTableAlignment);
  const rows = [];
  let index = startIndex + 2;

  while (index < lines.length && lines[index].trim().includes("|") && lines[index].trim() !== "") {
    if (isMarkdownTableSeparator(lines[index].trim())) {
      break;
    }
    rows.push(splitMarkdownTableRow(lines[index]));
    index += 1;
  }

  const headerHTML = headers
    .map((header, cellIndex) => renderTableCell("th", header, alignments[cellIndex]))
    .join("");
  const bodyHTML = rows
    .map((row) => {
      const cells = headers.map((_, cellIndex) => row[cellIndex] || "");
      return `<tr>${cells.map((cell, cellIndex) => renderTableCell("td", cell, alignments[cellIndex])).join("")}</tr>`;
    })
    .join("");

  return {
    html: `<div class="table-wrap"><table><thead><tr>${headerHTML}</tr></thead><tbody>${bodyHTML}</tbody></table></div>`,
    nextIndex: index,
  };
}

function splitMarkdownTableRow(line) {
  let value = String(line || "").trim();
  if (value.startsWith("|")) {
    value = value.slice(1);
  }
  if (value.endsWith("|")) {
    value = value.slice(0, -1);
  }

  const cells = [];
  let current = "";
  let escaped = false;
  for (const char of value) {
    if (escaped) {
      current += char;
      escaped = false;
      continue;
    }
    if (char === "\\") {
      escaped = true;
      continue;
    }
    if (char === "|") {
      cells.push(current.trim());
      current = "";
      continue;
    }
    current += char;
  }
  cells.push(current.trim());
  return cells;
}

function parseTableAlignment(separatorCell) {
  const cell = String(separatorCell || "").trim();
  if (cell.startsWith(":") && cell.endsWith(":")) {
    return "center";
  }
  if (cell.endsWith(":")) {
    return "right";
  }
  if (cell.startsWith(":")) {
    return "left";
  }
  return "";
}

function renderTableCell(tag, content, alignment) {
  const align = alignment ? ` style="text-align:${alignment}"` : "";
  return `<${tag}${align}>${renderInlineMarkdown(content)}</${tag}>`;
}

function renderInlineMarkdown(text) {
  const codeSpans = [];
  let html = escapeHTML(String(text || "")).replace(/`([^`]+)`/g, (_, code) => {
    const token = `@@CODE${codeSpans.length}@@`;
    codeSpans.push(`<code>${code}</code>`);
    return token;
  });

  html = html.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (_, label, url) => {
    const safeURL = sanitizeURL(url);
    if (!safeURL) {
      return label;
    }
    return `<a href="${safeURL}" target="_blank" rel="noreferrer">${label}</a>`;
  });
  html = html.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
  html = html.replace(/(^|[\s(])\*([^*]+)\*(?=[\s).,!?:;]|$)/g, "$1<em>$2</em>");
  html = html.replace(/~~([^~]+)~~/g, "<del>$1</del>");

  for (let index = 0; index < codeSpans.length; index += 1) {
    html = html.replace(`@@CODE${index}@@`, codeSpans[index]);
  }

  return html;
}

function sanitizeURL(url) {
  try {
    const parsed = new URL(url, window.location.origin);
    if (["http:", "https:", "mailto:"].includes(parsed.protocol)) {
      return escapeHTML(parsed.href);
    }
  } catch {
    return "";
  }

  return "";
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
    throw new Error(t("error.failedToParse", {
      label,
      message: error instanceof Error ? error.message : String(error),
    }));
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
