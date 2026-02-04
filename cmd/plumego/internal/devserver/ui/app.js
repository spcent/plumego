// State
let ws = null;
let logs = [];
let events = [];
let logCount = 0;
let errorCount = 0;
let lastRequestSnapshot = null;
let lastDBSlow = [];
let lastAlertThresholds = {};
let alertOverrides = {};
let depsSnapshot = null;
let projectKey = window.location.host;
let configState = {
    entries: [],
    path: '',
    exists: false,
    updatedAt: ''
};
let filters = {
    info: true,
    warn: true,
    error: true
};

// DOM Elements
const elements = {
    logs: document.getElementById('logs'),
    buildOutput: document.getElementById('build-output'),
    eventsList: document.getElementById('events'),
    appStatus: document.getElementById('app-status'),
    wsStatus: document.getElementById('ws-status'),
    uptime: document.getElementById('uptime'),
    logCount: document.getElementById('log-count'),
    errorCount: document.getElementById('error-count'),
    buildStatus: document.getElementById('build-status'),
    btnRestart: document.getElementById('btn-restart'),
    btnBuild: document.getElementById('btn-build'),
    btnStop: document.getElementById('btn-stop'),
    btnClear: document.getElementById('btn-clear'),
    filterInfo: document.getElementById('filter-info'),
    filterWarn: document.getElementById('filter-warn'),
    filterError: document.getElementById('filter-error'),
    requestMetrics: document.getElementById('request-metrics'),
    requestRoutes: document.getElementById('request-routes'),
    btnClearMetrics: document.getElementById('btn-clear-metrics'),
    metricsAlerts: document.getElementById('metrics-alerts'),
    dbMetrics: document.getElementById('db-metrics'),
    dbSlow: document.getElementById('db-slow'),
    dbDetail: document.getElementById('db-detail'),
    dbRedaction: document.getElementById('db-redaction'),
    alertTotalP95: document.getElementById('alert-total-p95'),
    alertTotalP99: document.getElementById('alert-total-p99'),
    alertTotalError: document.getElementById('alert-total-error'),
    alertRouteP95: document.getElementById('alert-route-p95'),
    alertRouteError: document.getElementById('alert-route-error'),
    btnAlertApply: document.getElementById('btn-alert-apply'),
    btnAlertReset: document.getElementById('btn-alert-reset'),
    btnAlertExport: document.getElementById('btn-alert-export'),
    btnAlertImport: document.getElementById('btn-alert-import'),
    alertImportFile: document.getElementById('alert-import-file'),
    profileType: document.getElementById('profile-type'),
    profileSeconds: document.getElementById('profile-seconds'),
    profileSecondsField: document.getElementById('profile-seconds-field'),
    btnProfileDownload: document.getElementById('btn-profile-download'),
    btnProfilePreview: document.getElementById('btn-profile-preview'),
    btnProfileOpen: document.getElementById('btn-profile-open'),
    profileStatus: document.getElementById('profile-status'),
    profilePreview: document.getElementById('profile-preview'),
    apiMethod: document.getElementById('api-method'),
    apiPath: document.getElementById('api-path'),
    apiQuery: document.getElementById('api-query'),
    apiHeaders: document.getElementById('api-headers'),
    apiBody: document.getElementById('api-body'),
    btnApiSend: document.getElementById('btn-api-send'),
    btnApiSave: document.getElementById('btn-api-save'),
    btnApiExport: document.getElementById('btn-api-export'),
    btnApiImport: document.getElementById('btn-api-import'),
    apiImportFile: document.getElementById('api-import-file'),
    btnApiClearHistory: document.getElementById('btn-api-clear-history'),
    apiHistory: document.getElementById('api-history'),
    apiSaved: document.getElementById('api-saved'),
    apiResponse: document.getElementById('api-response'),
    configMeta: document.getElementById('config-meta'),
    configTable: document.getElementById('config-table'),
    configRuntime: document.getElementById('config-runtime'),
    btnConfigAdd: document.getElementById('btn-config-add'),
    btnConfigSave: document.getElementById('btn-config-save'),
    btnConfigSaveRestart: document.getElementById('btn-config-save-restart'),
    btnConfigRefresh: document.getElementById('btn-config-refresh'),
    depsSummary: document.getElementById('deps-summary'),
    depsGraph: document.getElementById('deps-graph'),
    depsList: document.getElementById('deps-list'),
    depsEdges: document.getElementById('deps-edges'),
    depsIncludeStd: document.getElementById('deps-include-std'),
    depsMaxNodes: document.getElementById('deps-max-nodes'),
    btnDepsRefresh: document.getElementById('btn-deps-refresh')
};

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initProjectContext().finally(() => {
        initWebSocket();
        initTabs();
        initControls();
        initFilters();
        initProfiling();
        initApiTester();
        initAlertControls();
        initDeps();
        initConfigEditor();
    });
});

function initProjectContext() {
    return fetch('/api/status')
        .then(res => res.json())
        .then(data => {
            if (data && data.project && data.project.dir) {
                projectKey = data.project.dir;
            }
        })
        .catch(() => {})
        .finally(() => {
            alertOverrides = loadAlertOverrides();
        });
}

// WebSocket
function initWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        updateConnectionStatus('connected');
    };

    ws.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data);
            handleEvent(message);
        } catch (err) {
            console.error('Failed to parse message:', err);
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        updateConnectionStatus('error');
    };

    ws.onclose = () => {
        updateConnectionStatus('disconnected');
        // Attempt reconnect after 3 seconds
        setTimeout(initWebSocket, 3000);
    };
}

function updateConnectionStatus(status) {
    const statusMap = {
        connected: { text: 'Connected', class: 'badge-success' },
        disconnected: { text: 'Disconnected', class: 'badge-error' },
        error: { text: 'Error', class: 'badge-error' },
        connecting: { text: 'Connecting...', class: 'badge-secondary' }
    };

    const config = statusMap[status] || statusMap.connecting;
    elements.wsStatus.textContent = config.text;
    elements.wsStatus.className = `badge ${config.class}`;

    // Update connection indicator
    const indicator = document.getElementById('connection-indicator');
    if (indicator) {
        indicator.style.background = status === 'connected' ? 'var(--success)' :
                                     status === 'error' ? 'var(--error)' :
                                     'var(--warning)';
        indicator.style.boxShadow = status === 'connected' ? '0 0 10px var(--success)' :
                                    status === 'error' ? '0 0 10px var(--error)' :
                                    '0 0 10px var(--warning)';
    }
}

// Event Handlers
function handleEvent(message) {
    const { type, data } = message;

    // Add to events list
    events.unshift({ type, data, timestamp: new Date() });
    if (events.length > 100) events.pop();
    updateEventsDisplay();

    // Handle specific event types
    switch (type) {
        case 'app.log':
            handleLogEvent(data);
            break;
        case 'app.start':
            handleAppStart(data);
            break;
        case 'app.stop':
            handleAppStop(data);
            break;
        case 'build.start':
            handleBuildStart(data);
            break;
        case 'build.success':
            handleBuildSuccess(data);
            break;
        case 'build.fail':
            handleBuildFail(data);
            break;
        case 'dashboard.info':
            handleDashboardInfo(data);
            break;
    }
}

function handleLogEvent(data) {
    const { level, message, source } = data;

    logCount++;
    if (level === 'error') errorCount++;

    const logEntry = {
        level: level || 'info',
        message,
        source,
        timestamp: new Date()
    };

    logs.push(logEntry);
    if (logs.length > 1000) logs.shift();

    addLogToDisplay(logEntry);
    updateLogStats();
}

function handleAppStart(data) {
    const { state, pid } = data;
    if (state === 'running') {
        updateAppStatus('running', pid);
    }
}

function handleAppStop(data) {
    const { state } = data;
    updateAppStatus(state === 'crashed' ? 'crashed' : 'stopped');
}

function handleBuildStart() {
    updateBuildStatus('building', 'Building...');
}

function handleBuildSuccess(data) {
    const { duration, output } = data;
    updateBuildStatus('success', `Build successful (${formatDuration(duration)})`);
    if (output) {
        elements.buildOutput.innerHTML = `<div class="log-entry">${escapeHtml(output)}</div>`;
    }
}

function handleBuildFail(data) {
    const { error, output } = data;
    updateBuildStatus('error', `Build failed: ${error}`);
    if (output) {
        elements.buildOutput.innerHTML = `<div class="log-entry error">${escapeHtml(output)}</div>`;
    }
}

function handleDashboardInfo(data) {
    const { uptime } = data;
    if (uptime) {
        elements.uptime.textContent = uptime;
    }
}

// UI Updates
function addLogToDisplay(logEntry) {
    const { level, message, timestamp } = logEntry;

    const div = document.createElement('div');
    div.className = `log-entry ${level}`;
    if (!filters[level]) {
        div.classList.add('hidden');
    }

    const time = formatTime(timestamp);

    const timeSpan = document.createElement('span');
    timeSpan.className = 'log-time';
    timeSpan.textContent = time;

    const levelSpan = document.createElement('span');
    levelSpan.className = 'log-level';
    levelSpan.style.color = getLevelColor(level);
    levelSpan.textContent = (level || '').toUpperCase();

    const messageSpan = document.createElement('span');
    messageSpan.className = 'log-message';
    messageSpan.textContent = message;

    div.appendChild(timeSpan);
    div.appendChild(levelSpan);
    div.appendChild(messageSpan);

    elements.logs.appendChild(div);

    // Auto-scroll to bottom
    elements.logs.scrollTop = elements.logs.scrollHeight;

    // Limit displayed logs
    while (elements.logs.children.length > 500) {
        elements.logs.removeChild(elements.logs.firstChild);
    }
}

function updateLogStats() {
    elements.logCount.textContent = logCount;
    elements.errorCount.textContent = errorCount;
}

function updateAppStatus(status, pid) {
    const statusMap = {
        running: { text: `Running (PID: ${pid})`, class: 'badge-success' },
        stopped: { text: 'Stopped', class: 'badge-secondary' },
        crashed: { text: 'Crashed', class: 'badge-error' }
    };

    const config = statusMap[status] || statusMap.stopped;
    elements.appStatus.textContent = config.text;
    elements.appStatus.className = `badge ${config.class}`;
}

function updateBuildStatus(status, message) {
    const statusMap = {
        building: { class: 'badge-info' },
        success: { class: 'badge-success' },
        error: { class: 'badge-error' }
    };

    const config = statusMap[status] || statusMap.building;
    elements.buildStatus.innerHTML = `<span class="badge ${config.class}">${escapeHtml(message)}</span>`;
}

function updateEventsDisplay() {
    elements.eventsList.innerHTML = events.map(event => `
        <div class="event-item">
            <div class="event-type">${event.type}</div>
            <div class="event-time">${formatTime(event.timestamp)}</div>
            ${event.data ? `<div class="event-data">${JSON.stringify(event.data, null, 2)}</div>` : ''}
        </div>
    `).join('');
}

// Controls
function initControls() {
    elements.btnRestart.addEventListener('click', () => {
        fetch('/api/restart', { method: 'POST' })
            .then(res => res.json())
            .then(data => {
                if (!data.success) {
                    alert(`Restart failed: ${data.error}`);
                }
            })
            .catch(err => alert(`Restart failed: ${err.message}`));
    });

    elements.btnBuild.addEventListener('click', () => {
        fetch('/api/build', { method: 'POST' })
            .then(res => res.json())
            .then(data => {
                if (!data.success) {
                    alert(`Build failed: ${data.error}`);
                }
            })
            .catch(err => alert(`Build failed: ${err.message}`));
    });

    elements.btnStop.addEventListener('click', () => {
        if (confirm('Stop the application?')) {
            fetch('/api/stop', { method: 'POST' })
                .then(res => res.json())
                .then(data => {
                    if (!data.success) {
                        alert(`Stop failed: ${data.error}`);
                    }
                })
                .catch(err => alert(`Stop failed: ${err.message}`));
        }
    });

    elements.btnClear.addEventListener('click', () => {
        logs = [];
        logCount = 0;
        errorCount = 0;
        elements.logs.innerHTML = '';
        updateLogStats();
    });

    if (elements.btnClearMetrics) {
        elements.btnClearMetrics.addEventListener('click', () => {
            fetch('/api/metrics/clear', { method: 'POST' })
                .then(res => res.json())
                .then(data => {
                    if (!data.success) {
                        alert(`Clear failed: ${data.error}`);
                        return;
                    }
                    loadMetrics();
                })
                .catch(err => alert(`Clear failed: ${err.message}`));
        });
    }
}

// Tabs
function initTabs() {
    const tabs = document.querySelectorAll('.tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const targetTab = tab.dataset.tab;

            // Update tab buttons
            tabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            // Update tab panes
            document.querySelectorAll('.tab-pane').forEach(pane => {
                pane.classList.remove('active');
            });
            document.getElementById(`${targetTab}-tab`).classList.add('active');
        });
    });
}

// Filters
function initFilters() {
    elements.filterInfo.addEventListener('change', (e) => {
        filters.info = e.target.checked;
        applyFilters();
    });

    elements.filterWarn.addEventListener('change', (e) => {
        filters.warn = e.target.checked;
        applyFilters();
    });

    elements.filterError.addEventListener('change', (e) => {
        filters.error = e.target.checked;
        applyFilters();
    });
}

// Profiling
function initProfiling() {
    if (!elements.profileType) return;

    const fallbackProfiles = [
        { id: 'cpu', label: 'CPU', supports_seconds: true, default_seconds: 10 },
        { id: 'heap', label: 'Heap' },
        { id: 'allocs', label: 'Allocs' },
        { id: 'goroutine', label: 'Goroutine' },
        { id: 'block', label: 'Block' },
        { id: 'mutex', label: 'Mutex' },
        { id: 'threadcreate', label: 'Thread Create' },
        { id: 'trace', label: 'Trace', supports_seconds: true, default_seconds: 5 }
    ];

    populateProfileTypes(fallbackProfiles);

    fetch('/api/pprof/types')
        .then(res => res.json())
        .then(data => {
            if (data.types && Array.isArray(data.types) && data.types.length > 0) {
                populateProfileTypes(data.types);
            }
        })
        .catch(() => {});

    elements.profileType.addEventListener('change', updateProfilingSecondsVisibility);

    if (elements.btnProfileDownload) {
        elements.btnProfileDownload.addEventListener('click', () => {
            const query = buildProfileQuery();
            setProfileStatus('Downloading profile...');
            window.location = `/api/pprof/raw${query}`;
        });
    }

    if (elements.btnProfilePreview) {
        elements.btnProfilePreview.addEventListener('click', () => {
            const query = buildProfileQuery();
            setProfileStatus('Fetching preview...');
            fetch(`/api/pprof/raw${query}&download=0`)
                .then(res => res.json())
                .then(data => renderProfilePreview(data))
                .catch(err => {
                    setProfileStatus(`Preview failed: ${err.message}`);
                });
        });
    }

    if (elements.btnProfileOpen) {
        elements.btnProfileOpen.addEventListener('click', () => {
            const query = buildProfileQuery();
            const profileURL = `${window.location.origin}/api/pprof/raw${query}`;
            const speedscopeURL = `https://www.speedscope.app/#profileURL=${encodeURIComponent(profileURL)}`;
            window.open(speedscopeURL, '_blank');
            setProfileStatus('Opened flamegraph in Speedscope.');
        });
    }

    updateProfilingSecondsVisibility();
}

function initApiTester() {
    if (!elements.btnApiSend) return;

    loadApiState();

    elements.btnApiSend.addEventListener('click', () => {
        const payload = buildApiTestPayload();
        if (!payload.path) {
            alert('Path is required');
            return;
        }

        elements.apiResponse.innerHTML = '<div class="text-secondary">Sending request...</div>';

        fetch('/api/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        })
            .then(res => res.json())
            .then(data => {
                renderApiTestResponse(data);
                if (data && data.success !== false) {
                    addHistoryEntry(payload);
                }
            })
            .catch(err => {
                elements.apiResponse.innerHTML = `<div class="text-error">Request failed: ${escapeHtml(err.message)}</div>`;
            });
    });

    if (elements.btnApiSave) {
        elements.btnApiSave.addEventListener('click', () => {
            const payload = buildApiTestPayload();
            if (!payload.path) {
                alert('Path is required');
                return;
            }
            const label = prompt('Name for saved request', `${payload.method} ${payload.path}`) || '';
            addSavedEntry(payload, label.trim());
        });
    }

    if (elements.btnApiExport) {
        elements.btnApiExport.addEventListener('click', () => {
            exportSavedRequests();
        });
    }

    if (elements.btnApiImport && elements.apiImportFile) {
        elements.btnApiImport.addEventListener('click', () => {
            elements.apiImportFile.click();
        });
        elements.apiImportFile.addEventListener('change', handleImportFile);
    }

    if (elements.btnApiClearHistory) {
        elements.btnApiClearHistory.addEventListener('click', () => {
            apiState.history = [];
            persistApiState();
            renderApiLists();
        });
    }
}

function buildApiTestPayload() {
    return {
        method: elements.apiMethod.value,
        path: elements.apiPath.value.trim(),
        query: elements.apiQuery.value.trim(),
        headers: parseHeaders(elements.apiHeaders.value),
        body: elements.apiBody.value
    };
}

function parseHeaders(raw) {
    const headers = {};
    raw.split('\n').forEach(line => {
        const trimmed = line.trim();
        if (!trimmed) return;
        const idx = trimmed.indexOf(':');
        if (idx === -1) return;
        const key = trimmed.slice(0, idx).trim();
        const value = trimmed.slice(idx + 1).trim();
        if (key) {
            headers[key] = value;
        }
    });
    return headers;
}

const apiState = {
    history: [],
    saved: []
};

function loadApiState() {
    try {
        apiState.history = JSON.parse(localStorage.getItem('plumego.api.history') || '[]');
        apiState.saved = JSON.parse(localStorage.getItem('plumego.api.saved') || '[]');
    } catch {
        apiState.history = [];
        apiState.saved = [];
    }
    renderApiLists();
}

function persistApiState() {
    localStorage.setItem('plumego.api.history', JSON.stringify(apiState.history));
    localStorage.setItem('plumego.api.saved', JSON.stringify(apiState.saved));
}

function addHistoryEntry(payload) {
    const entry = {
        id: Date.now().toString(),
        method: payload.method,
        path: payload.path,
        query: payload.query,
        headers: payload.headers,
        body: payload.body
    };
    apiState.history.unshift(entry);
    if (apiState.history.length > 20) {
        apiState.history = apiState.history.slice(0, 20);
    }
    persistApiState();
    renderApiLists();
}

function addSavedEntry(payload, label) {
    const entry = {
        id: Date.now().toString(),
        label: label || `${payload.method} ${payload.path}`,
        method: payload.method,
        path: payload.path,
        query: payload.query,
        headers: payload.headers,
        body: payload.body
    };
    apiState.saved.unshift(entry);
    persistApiState();
    renderApiLists();
}

function exportSavedRequests() {
    const payload = {
        version: 1,
        saved: apiState.saved
    };
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `plumego-api-saved-${new Date().toISOString().slice(0, 10)}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
}

function handleImportFile(event) {
    const file = event.target.files && event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = () => {
        try {
            const parsed = JSON.parse(reader.result);
            const items = Array.isArray(parsed) ? parsed : parsed.saved;
            if (!Array.isArray(items)) {
                throw new Error('Invalid file format');
            }
            importSavedRequests(items);
        } catch (err) {
            alert(`Import failed: ${err.message}`);
        }
    };
    reader.readAsText(file);
    event.target.value = '';
}

function importSavedRequests(items) {
    const normalized = items.map(item => ({
        id: item.id || Date.now().toString() + Math.random().toString(16).slice(2),
        label: item.label || `${item.method || 'GET'} ${item.path || ''}`.trim(),
        method: item.method || 'GET',
        path: item.path || '',
        query: item.query || '',
        headers: item.headers || {},
        body: item.body || ''
    }));

    apiState.saved = [...normalized, ...apiState.saved];
    persistApiState();
    renderApiLists();
}

function renderApiLists() {
    if (elements.apiHistory) {
        elements.apiHistory.innerHTML = renderApiList(apiState.history, true);
        bindApiListButtons(elements.apiHistory, apiState.history, true);
    }
    if (elements.apiSaved) {
        elements.apiSaved.innerHTML = renderApiList(apiState.saved, false);
        bindApiListButtons(elements.apiSaved, apiState.saved, false);
    }
}

function renderApiList(items, canSave) {
    if (!items || items.length === 0) {
        return '<div class="text-secondary">No items yet.</div>';
    }
    return items.map(item => `
        <div class="api-list-item" data-id="${item.id}">
            <span>${escapeHtml(item.label || `${item.method} ${item.path}`)}</span>
            <button data-action="load" class="btn btn-outline">Load</button>
            ${canSave ? '<button data-action="save" class="btn btn-secondary">Save</button>' : '<button data-action="delete" class="btn btn-danger">Delete</button>'}
        </div>
    `).join('');
}

function bindApiListButtons(container, items, canSave) {
    container.querySelectorAll('.api-list-item button').forEach(btn => {
        btn.addEventListener('click', () => {
            const itemEl = btn.closest('.api-list-item');
            const id = itemEl ? itemEl.dataset.id : '';
            const item = items.find(entry => entry.id === id);
            if (!item) return;

            if (btn.dataset.action === 'load') {
                loadApiEntry(item);
                return;
            }

            if (btn.dataset.action === 'save' && canSave) {
                addSavedEntry(item, item.label);
                return;
            }

            if (btn.dataset.action === 'delete' && !canSave) {
                apiState.saved = apiState.saved.filter(entry => entry.id !== id);
                persistApiState();
                renderApiLists();
            }
        });
    });
}

function loadApiEntry(item) {
    elements.apiMethod.value = item.method || 'GET';
    elements.apiPath.value = item.path || '';
    elements.apiQuery.value = item.query || '';
    elements.apiHeaders.value = headersToText(item.headers || {});
    elements.apiBody.value = item.body || '';
}

function headersToText(headers) {
    return Object.entries(headers).map(([key, value]) => `${key}: ${value}`).join('\n');
}

function renderApiTestResponse(data) {
    if (!data || data.success === false) {
        const err = data && data.error ? data.error : 'Unknown error';
        elements.apiResponse.innerHTML = `<div class="text-error">Request failed: ${escapeHtml(err)}</div>`;
        return;
    }

    const statusClass = data.status >= 200 && data.status < 400 ? 'status-ok' : 'status-error';
    const body = data.body ? data.body : '';
    const headers = data.headers ? Object.entries(data.headers).map(([k, v]) => `${k}: ${v}`).join('\n') : '';
    const json = tryParseJson(body, data.headers);
    const bodyHtml = json ? renderJsonViewer(json) : renderRawBody(body, data.body_base64);

    elements.apiResponse.innerHTML = `
        <div><span class="${statusClass}">Status: ${data.status}</span> | ${data.duration_ms}ms | ${data.bytes} bytes</div>
        ${data.body_truncated ? '<div class="text-secondary">Response truncated</div>' : ''}
        <div class="text-secondary">Headers</div>
        <div>${escapeHtml(headers)}</div>
        <div class="text-secondary" style="margin-top: 8px;">Body</div>
        <div>${bodyHtml}</div>
    `;
}

function renderRawBody(body, base64) {
    if (body) {
        return `<div>${escapeHtml(body)}</div>`;
    }
    if (base64) {
        return `<div>Base64:\n${escapeHtml(base64)}</div>`;
    }
    return '<div class="text-secondary">No response body.</div>';
}

function tryParseJson(body, headers) {
    if (!body) return null;
    if (headers) {
        const contentType = headers['Content-Type'] || headers['content-type'] || '';
        if (contentType.includes('application/json')) {
            return safeParseJson(body);
        }
    }
    const trimmed = body.trim();
    if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
        return safeParseJson(trimmed);
    }
    return null;
}

function safeParseJson(raw) {
    try {
        return JSON.parse(raw);
    } catch {
        return null;
    }
}

function renderJsonViewer(value) {
    return `<div class="json-viewer">${renderJsonNode(value)}</div>`;
}

function renderJsonNode(value, key) {
    if (value === null) {
        return formatJsonLeaf(key, 'null', 'json-null');
    }

    if (Array.isArray(value)) {
        const summary = key ? `${escapeHtml(key)} [${value.length}]` : `[${value.length}]`;
        const children = value.map((item, idx) => renderJsonNode(item, String(idx))).join('');
        return `<details open><summary>${summary}</summary><div class="json-node">${children}</div></details>`;
    }

    if (typeof value === 'object') {
        const entries = Object.entries(value);
        const summary = key ? `${escapeHtml(key)} {${entries.length}}` : `{${entries.length}}`;
        const children = entries.map(([k, v]) => renderJsonNode(v, k)).join('');
        return `<details open><summary>${summary}</summary><div class="json-node">${children}</div></details>`;
    }

    if (typeof value === 'string') {
        return formatJsonLeaf(key, `"${escapeHtml(value)}"`, 'json-string');
    }

    if (typeof value === 'number') {
        return formatJsonLeaf(key, String(value), 'json-number');
    }

    if (typeof value === 'boolean') {
        return formatJsonLeaf(key, String(value), 'json-boolean');
    }

    return formatJsonLeaf(key, escapeHtml(String(value)), 'json-null');
}

function formatJsonLeaf(key, value, className) {
    const label = key ? `<span class="json-key">${escapeHtml(key)}:</span>` : '';
    return `<div class="json-leaf">${label}<span class="${className}">${value}</span></div>`;
}

function populateProfileTypes(profiles) {
    elements.profileType.innerHTML = profiles.map(profile => {
        const supports = profile.supports_seconds ? 'true' : 'false';
        const defSeconds = profile.default_seconds ? String(profile.default_seconds) : '';
        return `<option value="${profile.id}" data-supports="${supports}" data-default-seconds="${defSeconds}">${profile.label}</option>`;
    }).join('');
}

function updateProfilingSecondsVisibility() {
    const selected = elements.profileType.selectedOptions[0];
    if (!selected) return;

    const supports = selected.dataset.supports === 'true';
    const defaultSeconds = selected.dataset.defaultSeconds;
    if (supports && defaultSeconds) {
        elements.profileSeconds.value = defaultSeconds;
    }

    elements.profileSecondsField.style.display = supports ? 'flex' : 'none';
}

function buildProfileQuery() {
    const selected = elements.profileType.selectedOptions[0];
    const type = elements.profileType.value || 'cpu';
    let query = `?type=${encodeURIComponent(type)}`;

    if (selected && selected.dataset.supports === 'true') {
        const seconds = parseInt(elements.profileSeconds.value, 10);
        if (!Number.isNaN(seconds) && seconds > 0) {
            query += `&seconds=${seconds}`;
        }
    }

    return query;
}

function setProfileStatus(message) {
    if (elements.profileStatus) {
        elements.profileStatus.textContent = message;
    }
}

function renderProfilePreview(data) {
    if (!elements.profilePreview) return;

    if (!data || data.error) {
        elements.profilePreview.textContent = data && data.error ? data.error : 'No preview available';
        return;
    }

    const lines = [
        `Type: ${data.type}`,
        `Seconds: ${data.seconds || 0}`,
        `Content-Type: ${data.content_type}`,
        `Size: ${data.size_bytes} bytes`,
        `Preview: ${data.preview_hex || 'n/a'}`
    ];
    elements.profilePreview.textContent = lines.join('\n');
    setProfileStatus('Preview loaded.');
}

function applyFilters() {
    const logEntries = elements.logs.querySelectorAll('.log-entry');
    logEntries.forEach(entry => {
        const level = entry.classList.contains('error') ? 'error' :
                     entry.classList.contains('warn') ? 'warn' : 'info';

        if (filters[level]) {
            entry.classList.remove('hidden');
        } else {
            entry.classList.add('hidden');
        }
    });
}

// Utilities
function formatTime(date) {
    return date.toLocaleTimeString('en-US', { hour12: false });
}

function formatDuration(ms) {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
}

function getLevelColor(level) {
    const colors = {
        info: '#2196f3',
        warn: '#ff9800',
        error: '#f44336',
        debug: '#9c27b0'
    };
    return colors[level] || colors.info;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Routes and Metrics
function loadRoutes() {
    fetch('/api/routes')
        .then(res => res.json())
        .then(data => {
            const container = document.getElementById('routes-container');

            if (!data.routes || data.routes.length === 0) {
                container.innerHTML = '<p class="text-secondary">No routes found. App may not be running or routes endpoint not available.</p>';
                if (data.error) {
                    container.innerHTML += `<p class="text-error">${data.error}</p>`;
                }
                return;
            }

            container.innerHTML = data.routes.map(route => `
                <div class="route-item">
                    <span class="route-method ${route.method}">${route.method}</span>
                    <span class="route-path">${escapeHtml(route.path)}</span>
                    ${route.description ? `<div class="route-description">${escapeHtml(route.description)}</div>` : ''}
                </div>
            `).join('');
        })
        .catch(err => {
            document.getElementById('routes-container').innerHTML =
                `<p class="text-error">Failed to load routes: ${err.message}</p>`;
        });
}

function loadMetrics() {
    fetch('/api/metrics')
        .then(res => res.json())
        .then(data => {
            // Dashboard metrics
            const dashboardMetrics = document.getElementById('dashboard-metrics');
            if (data.dashboard) {
                dashboardMetrics.innerHTML = `
                    <div class="metric-item">
                        <span class="metric-label">Uptime</span>
                        <span class="metric-value">${formatUptime(data.dashboard.uptime)}</span>
                    </div>
                    <div class="metric-item">
                        <span class="metric-label">Start Time</span>
                        <span class="metric-value">${new Date(data.dashboard.startTime).toLocaleString()}</span>
                    </div>
                `;
            }

            // App metrics
            const appMetrics = document.getElementById('app-metrics');
            if (data.app) {
                let html = `
                    <div class="metric-item">
                        <span class="metric-label">Status</span>
                        <span class="metric-value">${data.app.running ? '✅ Running' : '⏸️ Stopped'}</span>
                    </div>
                    <div class="metric-item">
                        <span class="metric-label">PID</span>
                        <span class="metric-value">${data.app.pid || 'N/A'}</span>
                    </div>
                `;

                if (data.app.healthy !== undefined) {
                    html += `
                        <div class="metric-item">
                            <span class="metric-label">Health</span>
                            <span class="metric-value">${data.app.healthy ? '✅ Healthy' : '❌ Unhealthy'}</span>
                        </div>
                    `;
                }

                appMetrics.innerHTML = html;
            }

            if (data.app && data.app.requests) {
                renderRequestMetrics(data.app.requests);
                lastRequestSnapshot = data.app.requests;
            } else {
                const err = data.app && data.app.requests_error ? `: ${escapeHtml(data.app.requests_error)}` : '';
                elements.requestMetrics.innerHTML = `<p class="text-secondary">Request metrics unavailable${err}</p>`;
                elements.requestRoutes.innerHTML = '';
                lastRequestSnapshot = null;
            }

            if (data.app && data.app.db) {
                renderDBMetrics(data.app.db);
            } else if (elements.dbMetrics) {
                elements.dbMetrics.innerHTML = '<p class="text-secondary">DB metrics unavailable.</p>';
                elements.dbSlow.innerHTML = '';
            }

            lastAlertThresholds = data.thresholds || {};
            applyAlertThresholds();
        })
        .catch(err => {
            console.error('Failed to load metrics:', err);
        });
}

function formatUptime(seconds) {
    if (seconds < 60) return `${seconds.toFixed(1)}s`;
    if (seconds < 3600) return `${(seconds / 60).toFixed(1)}m`;
    return `${(seconds / 3600).toFixed(1)}h`;
}

function renderRequestMetrics(snapshot) {
    if (!snapshot || !snapshot.total) {
        elements.requestMetrics.innerHTML = '<p class="text-secondary">No request metrics yet.</p>';
        elements.requestRoutes.innerHTML = '';
        return;
    }

    const total = snapshot.total;
    const duration = total.duration_ms || {};
    const count = total.count || 0;
    const errorCount = total.error_count || 0;
    const status = total.status || {};
    const errorRate = count > 0 ? ((errorCount / count) * 100).toFixed(1) : '0.0';

    elements.requestMetrics.innerHTML = `
        <div class="metric-item">
            <span class="metric-label">Requests</span>
            <span class="metric-value">${count}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">Error Rate</span>
            <span class="metric-value">${errorRate}%</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">P50</span>
            <span class="metric-value">${formatMs(duration.p50)}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">P95</span>
            <span class="metric-value">${formatMs(duration.p95)}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">P99</span>
            <span class="metric-value">${formatMs(duration.p99)}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">2xx</span>
            <span class="metric-value">${status['2xx'] || 0}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">3xx</span>
            <span class="metric-value">${status['3xx'] || 0}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">4xx</span>
            <span class="metric-value">${status['4xx'] || 0}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">5xx</span>
            <span class="metric-value">${status['5xx'] || 0}</span>
        </div>
    `;

    const routes = (snapshot.routes || []).slice();
    routes.sort((a, b) => {
        const ap95 = (a.duration_ms && a.duration_ms.p95) || 0;
        const bp95 = (b.duration_ms && b.duration_ms.p95) || 0;
        return bp95 - ap95;
    });

    const top = routes.slice(0, 5);
    if (top.length === 0) {
        elements.requestRoutes.innerHTML = '<p class="text-secondary">No route metrics yet.</p>';
        return;
    }

    const rows = top.map(route => {
        const p95 = route.duration_ms ? route.duration_ms.p95 : 0;
        const count = route.count || 0;
        const errorCount = route.error_count || 0;
        const errRate = count > 0 ? (errorCount / count) * 100 : 0;
        const status = route.status || {};
        const className = errRate >= 5 || status['5xx'] > 0 ? 'metric-table-row danger' :
                          errRate > 0 ? 'metric-table-row warn' : 'metric-table-row';

        return `
            <div class="${className}">
                <div class="metric-table-cell">${escapeHtml(route.method)} ${escapeHtml(route.path)}</div>
                <div class="metric-table-cell">${formatMs(p95)}</div>
                <div class="metric-table-cell">${count}</div>
                <div class="metric-table-cell">${status['4xx'] || 0}</div>
                <div class="metric-table-cell">${status['5xx'] || 0}</div>
                <div class="metric-table-cell">${errRate.toFixed(1)}%</div>
            </div>
        `;
    }).join('');

    elements.requestRoutes.innerHTML = `
        <div class="metric-table-header">
            <div>Route</div>
            <div>P95</div>
            <div>Count</div>
            <div>4xx</div>
            <div>5xx</div>
            <div>Err%</div>
        </div>
        ${rows}
    `;
}

function renderDBMetrics(snapshot) {
    if (!elements.dbMetrics || !elements.dbSlow) return;

    if (!snapshot || !snapshot.total) {
        elements.dbMetrics.innerHTML = '<p class="text-secondary">No DB metrics yet.</p>';
        elements.dbSlow.innerHTML = '';
        lastDBSlow = [];
        if (elements.dbRedaction) {
            elements.dbRedaction.textContent = '';
        }
        renderDBDetail(null);
        return;
    }

    const total = snapshot.total;
    const duration = total.duration_ms || {};
    const count = total.count || 0;
    const errorCount = total.error_count || 0;
    const errorRate = count > 0 ? ((errorCount / count) * 100).toFixed(1) : '0.0';

    elements.dbMetrics.innerHTML = `
        <div class="metric-item">
            <span class="metric-label">Queries</span>
            <span class="metric-value">${count}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">Error Rate</span>
            <span class="metric-value">${errorRate}%</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">P95</span>
            <span class="metric-value">${formatMs(duration.p95)}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">P99</span>
            <span class="metric-value">${formatMs(duration.p99)}</span>
        </div>
    `;

    const slow = (snapshot.slow || []).slice().reverse().slice(0, 5);
    lastDBSlow = slow;

    if (elements.dbRedaction) {
        const redaction = snapshot.redaction || {};
        if (redaction.enabled) {
            const rules = Array.isArray(redaction.rules) ? redaction.rules : [];
            elements.dbRedaction.textContent = rules.length
                ? `Redaction rules: ${rules.join(' | ')}`
                : 'Redaction enabled for captured queries.';
        } else {
            elements.dbRedaction.textContent = 'Redaction disabled for captured queries.';
        }
    }

    if (slow.length === 0) {
        elements.dbSlow.innerHTML = '<p class="text-secondary">No slow queries.</p>';
        lastDBSlow = [];
        renderDBDetail(null);
        return;
    }

    const rows = slow.map((entry, idx) => {
        const label = `${entry.driver || 'db'} ${entry.operation || ''}`;
        const preview = entry.query_preview || entry.query || '';
        return `
            <div class="metric-table-row clickable" data-index="${idx}">
                <div class="metric-table-cell">${escapeHtml(label)}</div>
                <div class="metric-table-cell">${formatMs(entry.duration_ms)}</div>
                <div class="metric-table-cell">${escapeHtml(preview)}</div>
                <div class="metric-table-cell">${escapeHtml(entry.error || '')}</div>
                <div class="metric-table-cell">${new Date(entry.timestamp).toLocaleTimeString()}</div>
            </div>
        `;
    }).join('');

    elements.dbSlow.innerHTML = `
        <div class="metric-table-header">
            <div>Op</div>
            <div>Duration</div>
            <div>Query</div>
            <div>Error</div>
            <div>Time</div>
        </div>
        ${rows}
    `;

    elements.dbSlow.querySelectorAll('.metric-table-row').forEach(row => {
        row.addEventListener('click', () => {
            const idx = Number(row.dataset.index);
            if (!Number.isNaN(idx) && lastDBSlow[idx]) {
                renderDBDetail(lastDBSlow[idx]);
            }
        });
    });
}

function renderDBDetail(entry) {
    if (!elements.dbDetail) return;

    if (!entry) {
        elements.dbDetail.textContent = 'Select a slow query to see details.';
        return;
    }

    const label = `${entry.driver || 'db'} ${entry.operation || ''}`.trim();
    const meta = `${formatMs(entry.duration_ms)} | ${new Date(entry.timestamp).toLocaleString()}`;
    const error = entry.error ? `<div class="text-error">Error: ${escapeHtml(entry.error)}</div>` : '';
    const query = entry.query || entry.query_preview || '';
    const redactionNote = entry.redacted ? 'Query text was redacted to protect parameters.' : 'Query captured without redaction.';

    elements.dbDetail.innerHTML = `
        <div class="db-detail-title">${escapeHtml(label)}</div>
        <div class="db-detail-meta">${escapeHtml(meta)}</div>
        ${error}
        <div class="db-detail-query">${escapeHtml(query)}</div>
        <div class="text-secondary">${escapeHtml(redactionNote)}</div>
    `;
}

function renderAlerts(alerts, thresholds) {
    if (!elements.metricsAlerts) return;

    if (!alerts || alerts.length === 0) {
        elements.metricsAlerts.innerHTML = '<p class="text-secondary">No alerts.</p>';
        return;
    }

    elements.metricsAlerts.innerHTML = alerts.map(alert => {
        const level = alert.level || 'warn';
        const target = alert.method && alert.path ? ` ${escapeHtml(alert.method)} ${escapeHtml(alert.path)}` : '';
        const threshold = alert.threshold !== undefined ? ` (threshold: ${alert.threshold})` : '';
        return `
            <div class="alert-item ${level}">
                <div><strong>${escapeHtml(alert.message || alert.code || 'Alert')}</strong>${target}</div>
                <div class="text-secondary">${escapeHtml(alert.metric || '')}: ${alert.value}${threshold}</div>
            </div>
        `;
    }).join('');
}

function initAlertControls() {
    if (!elements.btnAlertApply) return;

    alertOverrides = loadAlertOverrides();

    elements.btnAlertApply.addEventListener('click', () => {
        alertOverrides = readAlertInputs();
        persistAlertOverrides();
        applyAlertThresholds();
    });

    elements.btnAlertReset.addEventListener('click', () => {
        alertOverrides = {};
        persistAlertOverrides();
        applyAlertThresholds(true);
    });

    if (elements.btnAlertExport) {
        elements.btnAlertExport.addEventListener('click', exportAlertOverrides);
    }

    if (elements.btnAlertImport && elements.alertImportFile) {
        elements.btnAlertImport.addEventListener('click', () => {
            elements.alertImportFile.click();
        });
        elements.alertImportFile.addEventListener('change', handleAlertImportFile);
    }
}

function applyAlertThresholds(forceReset) {
    const thresholds = { ...lastAlertThresholds, ...alertOverrides };
    if (forceReset) {
        setAlertInputs(lastAlertThresholds);
    } else if (!isEditingAlertControls()) {
        if (!alertOverrides || Object.keys(alertOverrides).length === 0) {
            setAlertInputs(lastAlertThresholds);
        } else {
            setAlertInputs(thresholds);
        }
    }

    const alerts = computeAlerts(lastRequestSnapshot, thresholds);
    renderAlerts(alerts, thresholds);
}

function loadAlertOverrides() {
    try {
        const raw = JSON.parse(localStorage.getItem(alertOverridesKey()) || '{}');
        return normalizeAlertOverrides(raw);
    } catch {
        return {};
    }
}

function persistAlertOverrides() {
    alertOverrides = normalizeAlertOverrides(alertOverrides);
    localStorage.setItem(alertOverridesKey(), JSON.stringify(alertOverrides));
}

function alertOverridesKey() {
    return `plumego.alert.overrides::${projectKey}`;
}

function exportAlertOverrides() {
    const payload = {
        version: 1,
        project: projectKey,
        exported_at: new Date().toISOString(),
        overrides: normalizeAlertOverrides(alertOverrides)
    };
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `plumego-alert-thresholds-${new Date().toISOString().slice(0, 10)}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
}

function handleAlertImportFile(event) {
    const file = event.target.files && event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = () => {
        try {
            const parsed = JSON.parse(reader.result);
            const overrides = parsed && parsed.overrides ? parsed.overrides : parsed;
            alertOverrides = normalizeAlertOverrides(overrides);
            persistAlertOverrides();
            setAlertInputs({ ...lastAlertThresholds, ...alertOverrides });
            applyAlertThresholds();
        } catch (err) {
            alert(`Import failed: ${err.message}`);
        }
    };
    reader.readAsText(file);
    event.target.value = '';
}

function normalizeAlertOverrides(raw) {
    if (!raw || typeof raw !== 'object') {
        return {};
    }
    return {
        total_p95_ms: toNumber(raw.total_p95_ms),
        total_p99_ms: toNumber(raw.total_p99_ms),
        total_error_rate_pct: toNumber(raw.total_error_rate_pct),
        route_p95_ms: toNumber(raw.route_p95_ms),
        route_error_rate_pct: toNumber(raw.route_error_rate_pct)
    };
}

function setAlertInputs(thresholds) {
    if (!elements.alertTotalP95) return;
    elements.alertTotalP95.value = thresholds.total_p95_ms ?? '';
    elements.alertTotalP99.value = thresholds.total_p99_ms ?? '';
    elements.alertTotalError.value = thresholds.total_error_rate_pct ?? '';
    elements.alertRouteP95.value = thresholds.route_p95_ms ?? '';
    elements.alertRouteError.value = thresholds.route_error_rate_pct ?? '';
}

function readAlertInputs() {
    return {
        total_p95_ms: toNumber(elements.alertTotalP95.value),
        total_p99_ms: toNumber(elements.alertTotalP99.value),
        total_error_rate_pct: toNumber(elements.alertTotalError.value),
        route_p95_ms: toNumber(elements.alertRouteP95.value),
        route_error_rate_pct: toNumber(elements.alertRouteError.value)
    };
}

function computeAlerts(snapshot, thresholds) {
    if (!snapshot || !snapshot.total) {
        return [];
    }

    const alerts = [];
    const total = snapshot.total;
    const minTotal = thresholds.min_total_count ?? 0;
    const minRoute = thresholds.min_route_count ?? 0;

    if (total.count >= minTotal) {
        if (thresholds.total_p95_ms && total.duration_ms && total.duration_ms.p95 > thresholds.total_p95_ms) {
            alerts.push(makeAlert('warn', 'total_p95_high', 'Total P95 latency is above threshold', 'total.p95_ms', total.duration_ms.p95, thresholds.total_p95_ms));
        }
        if (thresholds.total_p99_ms && total.duration_ms && total.duration_ms.p99 > thresholds.total_p99_ms) {
            alerts.push(makeAlert('error', 'total_p99_high', 'Total P99 latency is above threshold', 'total.p99_ms', total.duration_ms.p99, thresholds.total_p99_ms));
        }
        const totalErrRate = percent(total.error_count || 0, total.count || 0);
        if (thresholds.total_error_rate_pct !== undefined && totalErrRate > thresholds.total_error_rate_pct) {
            alerts.push(makeAlert('error', 'total_error_rate_high', 'Total error rate is above threshold', 'total.error_rate_pct', totalErrRate, thresholds.total_error_rate_pct));
        }
    }

    (snapshot.routes || []).forEach(route => {
        if ((route.count || 0) < minRoute) return;
        if (thresholds.route_p95_ms && route.duration_ms && route.duration_ms.p95 > thresholds.route_p95_ms) {
            alerts.push(makeAlert('warn', 'route_p95_high', 'Route P95 latency is above threshold', 'route.p95_ms', route.duration_ms.p95, thresholds.route_p95_ms, route));
        }
        const routeErrRate = percent(route.error_count || 0, route.count || 0);
        if (thresholds.route_error_rate_pct !== undefined && routeErrRate > thresholds.route_error_rate_pct) {
            alerts.push(makeAlert('warn', 'route_error_rate_high', 'Route error rate is above threshold', 'route.error_rate_pct', routeErrRate, thresholds.route_error_rate_pct, route));
        }
    });

    return alerts;
}

function makeAlert(level, code, message, metric, value, threshold, route) {
    const alert = { level, code, message, metric, value, threshold };
    if (route) {
        alert.method = route.method;
        alert.path = route.path;
    }
    return alert;
}

function percent(count, total) {
    if (!total) return 0;
    return (count / total) * 100;
}

function initDeps() {
    if (!elements.btnDepsRefresh || !elements.depsSummary) return;

    elements.btnDepsRefresh.addEventListener('click', () => {
        loadDeps(true);
    });

    if (elements.depsIncludeStd) {
        elements.depsIncludeStd.addEventListener('change', () => {
            loadDeps();
        });
    }

    if (elements.depsMaxNodes) {
        elements.depsMaxNodes.addEventListener('change', () => {
            loadDeps();
        });
    }
}

function initConfigEditor() {
    if (!elements.configTable) return;

    if (elements.btnConfigAdd) {
        elements.btnConfigAdd.addEventListener('click', () => {
            addConfigRow();
        });
    }

    if (elements.btnConfigSave) {
        elements.btnConfigSave.addEventListener('click', () => {
            saveConfig(false);
        });
    }

    if (elements.btnConfigSaveRestart) {
        elements.btnConfigSaveRestart.addEventListener('click', () => {
            saveConfig(true);
        });
    }

    if (elements.btnConfigRefresh) {
        elements.btnConfigRefresh.addEventListener('click', () => {
            loadConfigEditor();
        });
    }
}

function loadDeps(forceRefresh) {
    if (!elements.depsSummary) return;

    const includeStd = elements.depsIncludeStd ? elements.depsIncludeStd.checked : true;
    const maxNodes = elements.depsMaxNodes ? parseInt(elements.depsMaxNodes.value, 10) : 0;
    let query = `?include_std=${includeStd ? 1 : 0}`;
    if (!Number.isNaN(maxNodes) && maxNodes > 0) {
        query += `&max_nodes=${maxNodes}`;
    }
    if (forceRefresh) {
        query += '&refresh=1';
    }

    elements.depsSummary.innerHTML = '<div class="text-secondary">Loading dependency graph...</div>';
    if (elements.depsGraph) {
        elements.depsGraph.innerHTML = '';
    }
    if (elements.depsList) {
        elements.depsList.innerHTML = '';
    }
    if (elements.depsEdges) {
        elements.depsEdges.innerHTML = '';
    }

    fetch(`/api/deps${query}`)
        .then(res => res.json().then(data => ({ ok: res.ok, data })))
        .then(({ ok, data }) => {
            if (!ok) {
                throw new Error(data && data.error ? data.error : 'Request failed');
            }
            depsSnapshot = data;
            renderDeps(data);
        })
        .catch(err => {
            const msg = `Failed to load dependencies: ${err.message}`;
            if (elements.depsSummary) {
                elements.depsSummary.innerHTML = `<div class="text-error">${escapeHtml(msg)}</div>`;
            }
            if (elements.depsList) {
                elements.depsList.innerHTML = '';
            }
            if (elements.depsEdges) {
                elements.depsEdges.innerHTML = '';
            }
    });
}

function loadConfigEditor() {
    if (!elements.configTable) return;

    if (elements.configMeta) {
        elements.configMeta.textContent = 'Loading config file...';
    }
    elements.configTable.innerHTML = '';

    fetch('/api/config/edit')
        .then(res => res.json().then(data => ({ ok: res.ok, data })))
        .then(({ ok, data }) => {
            if (!ok) {
                throw new Error(data && data.error ? data.error : 'Request failed');
            }
            configState = {
                entries: Array.isArray(data.entries) ? data.entries : [],
                path: data.path || '.env',
                exists: data.exists || false,
                updatedAt: data.updated_at || ''
            };
            renderConfigEditor();
            loadRuntimeConfig();
        })
        .catch(err => {
            if (elements.configMeta) {
                elements.configMeta.textContent = `Failed to load config: ${err.message}`;
            }
        });
}

function renderConfigEditor() {
    if (!elements.configTable) return;

    const rows = (configState.entries || []).map(entry => renderConfigRow(entry)).join('');
    elements.configTable.innerHTML = rows || '<div class="text-secondary">No entries yet.</div>';

    if (elements.configMeta) {
        const existsText = configState.exists ? 'file exists' : 'file not found';
        const updated = configState.updatedAt ? ` Updated: ${new Date(configState.updatedAt).toLocaleString()}.` : '';
        elements.configMeta.textContent = `Editing ${configState.path} (${existsText}).${updated}`;
    }

    bindConfigRowHandlers();
}

function renderConfigRow(entry) {
    const key = entry.key || entry.Key || '';
    const value = entry.value || entry.Value || '';
    return `
        <div class="config-row">
            <input class="config-key" type="text" placeholder="KEY" value="${escapeHtml(key)}">
            <input class="config-value" type="text" placeholder="value" value="${escapeHtml(value)}">
            <button class="btn btn-outline config-remove">Remove</button>
        </div>
    `;
}

function bindConfigRowHandlers() {
    elements.configTable.querySelectorAll('.config-remove').forEach(btn => {
        btn.addEventListener('click', () => {
            const row = btn.closest('.config-row');
            if (row) {
                row.remove();
            }
        });
    });
}

function addConfigRow() {
    if (!elements.configTable) return;
    if (elements.configTable.querySelector('.text-secondary')) {
        elements.configTable.innerHTML = '';
    }
    elements.configTable.insertAdjacentHTML('beforeend', renderConfigRow({}));
    bindConfigRowHandlers();
}

function collectConfigEntries() {
    const entries = [];
    if (!elements.configTable) return entries;

    elements.configTable.querySelectorAll('.config-row').forEach(row => {
        const keyInput = row.querySelector('.config-key');
        const valueInput = row.querySelector('.config-value');
        const key = keyInput ? keyInput.value.trim() : '';
        const value = valueInput ? valueInput.value : '';
        if (!key) return;
        entries.push({ key, value });
    });

    return entries;
}

function saveConfig(restart) {
    const entries = collectConfigEntries();
    const payload = { entries, restart: !!restart };

    fetch('/api/config/edit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    })
        .then(res => res.json().then(data => ({ ok: res.ok, data })))
        .then(({ ok, data }) => {
            if (!ok) {
                throw new Error(data && data.error ? data.error : 'Save failed');
            }
            if (elements.configMeta) {
                const restartText = data.restarted ? 'Restarted.' : 'Saved.';
                elements.configMeta.textContent = `Saved to ${data.path}. ${restartText}`;
            }
            loadConfigEditor();
        })
        .catch(err => {
            if (elements.configMeta) {
                elements.configMeta.textContent = `Save failed: ${err.message}`;
            }
        });
}

function loadRuntimeConfig() {
    if (!elements.configRuntime) return;
    elements.configRuntime.innerHTML = '<div class="text-secondary">Loading runtime config...</div>';

    fetch('/api/config')
        .then(res => res.json().then(data => ({ ok: res.ok, data })))
        .then(({ ok, data }) => {
            if (!ok) {
                throw new Error(data && data.error ? data.error : 'Runtime config unavailable');
            }
            elements.configRuntime.innerHTML = renderJsonViewer(data);
        })
        .catch(err => {
            elements.configRuntime.innerHTML = `<div class="text-secondary">${escapeHtml(err.message)}</div>`;
        });
}

function renderDeps(graph) {
    if (!graph || !graph.summary) {
        if (elements.depsSummary) {
            elements.depsSummary.innerHTML = '<div class="text-secondary">No dependency data.</div>';
        }
        if (elements.depsList) {
            elements.depsList.innerHTML = '';
        }
        if (elements.depsEdges) {
            elements.depsEdges.innerHTML = '';
        }
        return;
    }

    const summary = graph.summary;
    const generated = graph.generated_at ? new Date(graph.generated_at).toLocaleString() : 'n/a';

    if (elements.depsSummary) {
        elements.depsSummary.innerHTML = `
            <div class="metric-item">
                <span class="metric-label">Modules</span>
                <span class="metric-value">${summary.total_modules}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Direct</span>
                <span class="metric-value">${summary.direct_modules}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Indirect</span>
                <span class="metric-value">${summary.indirect_modules}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Packages</span>
                <span class="metric-value">${summary.total_packages}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Edges</span>
                <span class="metric-value">${summary.total_edges}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Updated</span>
                <span class="metric-value">${escapeHtml(generated)}</span>
            </div>
        `;
    }

    renderDepsGraph(graph);
    renderDepsList(graph);
    renderDepsEdges(graph);
}

function renderDepsGraph(graph) {
    if (!elements.depsGraph) return;

    const svg = elements.depsGraph;
    svg.innerHTML = '';

    const nodes = Array.isArray(graph.nodes) ? graph.nodes.slice() : [];
    if (nodes.length === 0) {
        const text = createSvgElement('text', {
            x: 400,
            y: 225,
            'text-anchor': 'middle',
            fill: 'var(--text-secondary)',
            'font-size': '14'
        });
        text.textContent = 'No dependency graph data.';
        svg.appendChild(text);
        return;
    }

    const main = graph.summary ? graph.summary.main : '';
    const root = nodes.find(node => node.id === main) || nodes[0];
    const others = nodes.filter(node => node.id !== root.id);
    const width = 800;
    const height = 450;
    const center = { x: width / 2, y: height / 2 };
    const radius = Math.min(width, height) / 2 - 70;
    const positions = new Map();

    positions.set(root.id, center);
    const count = Math.max(1, others.length);
    others.forEach((node, idx) => {
        const angle = (idx / count) * Math.PI * 2;
        positions.set(node.id, {
            x: center.x + radius * Math.cos(angle),
            y: center.y + radius * Math.sin(angle)
        });
    });

    const edges = (graph.edges || []).filter(edge => positions.has(edge.from) && positions.has(edge.to));
    const maxCount = edges.reduce((max, edge) => Math.max(max, edge.count || 0), 1);

    edges.forEach(edge => {
        const from = positions.get(edge.from);
        const to = positions.get(edge.to);
        if (!from || !to) return;
        const weight = 1 + (edge.count / maxCount) * 2;
        const line = createSvgElement('line', {
            x1: from.x,
            y1: from.y,
            x2: to.x,
            y2: to.y,
            stroke: 'rgba(0, 212, 255, 0.2)',
            'stroke-width': weight
        });
        svg.appendChild(line);
    });

    nodes.forEach(node => {
        const pos = positions.get(node.id);
        if (!pos) return;
        const radiusSize = node.main ? 10 : node.stdlib ? 8 : 7;
        const fill = node.main ? 'var(--accent-primary)' : node.stdlib ? 'var(--warning)' : node.direct ? 'var(--info)' : 'var(--text-tertiary)';
        const circle = createSvgElement('circle', {
            cx: pos.x,
            cy: pos.y,
            r: radiusSize,
            fill: fill,
            stroke: 'rgba(255, 255, 255, 0.5)',
            'stroke-width': 1
        });
        const title = createSvgElement('title', {});
        title.textContent = `${node.id} (${node.packages} packages)`;
        circle.appendChild(title);
        svg.appendChild(circle);

        const label = createSvgElement('text', {
            x: pos.x,
            y: pos.y - radiusSize - 6,
            'text-anchor': 'middle',
            fill: 'var(--text-secondary)',
            'font-size': '10'
        });
        label.textContent = formatDepLabel(node.id);
        svg.appendChild(label);
    });
}

function renderDepsList(graph) {
    if (!elements.depsList) return;

    const nodes = (graph.nodes || []).slice();
    if (nodes.length === 0) {
        elements.depsList.innerHTML = '<p class="text-secondary">No modules found.</p>';
        return;
    }

    nodes.sort((a, b) => {
        if (a.packages === b.packages) {
            return a.id.localeCompare(b.id);
        }
        return b.packages - a.packages;
    });

    const rows = nodes.map(node => {
        const version = node.version || (node.stdlib ? 'stdlib' : '');
        const versionText = node.replace ? `${version || 'replace'} -> ${node.replace}` : version;
        return `
        <div class="metric-table-row">
            <div class="metric-table-cell">${escapeHtml(node.id)}</div>
            <div class="metric-table-cell">${escapeHtml(versionText)}</div>
            <div class="metric-table-cell">${node.packages || 0}</div>
            <div class="metric-table-cell">${node.stdlib ? 'stdlib' : node.main ? 'main' : node.direct ? 'direct' : 'indirect'}</div>
        </div>
    `;
    }).join('');

    elements.depsList.innerHTML = `
        <div class="metric-table-header deps-table">
            <div>Module</div>
            <div>Version</div>
            <div>Pkgs</div>
            <div>Type</div>
        </div>
        ${rows}
    `;
}

function renderDepsEdges(graph) {
    if (!elements.depsEdges) return;

    const edges = (graph.edges || []).slice(0, 12);
    if (edges.length === 0) {
        elements.depsEdges.innerHTML = '<p class="text-secondary">No edges to display.</p>';
        return;
    }

    const rows = edges.map(edge => `
        <div class="metric-table-row">
            <div class="metric-table-cell">${escapeHtml(edge.from)}</div>
            <div class="metric-table-cell">${escapeHtml(edge.to)}</div>
            <div class="metric-table-cell">${edge.count}</div>
        </div>
    `).join('');

    elements.depsEdges.innerHTML = `
        <div class="metric-table-header deps-edge-table">
            <div>From</div>
            <div>To</div>
            <div>Count</div>
        </div>
        ${rows}
    `;
}

function formatDepLabel(value) {
    if (!value) return '';
    const parts = value.split('/');
    if (parts.length <= 2) {
        return value;
    }
    return `${parts[0]}/.../${parts[parts.length - 1]}`;
}

function createSvgElement(tag, attrs) {
    const el = document.createElementNS('http://www.w3.org/2000/svg', tag);
    Object.entries(attrs || {}).forEach(([key, val]) => {
        el.setAttribute(key, val);
    });
    return el;
}

function toNumber(value) {
    if (value === '' || value === null || value === undefined) {
        return undefined;
    }
    const num = Number(value);
    return Number.isFinite(num) ? num : undefined;
}

function isEditingAlertControls() {
    const active = document.activeElement;
    if (!active) return false;
    return active.tagName === 'INPUT' && active.closest('.alert-controls');
}

function formatMs(value) {
    const ms = Number(value || 0);
    if (ms < 1) return `${ms.toFixed(2)}ms`;
    if (ms < 1000) return `${ms.toFixed(1)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
}

// Tab change handler - load data when switching tabs
document.addEventListener('DOMContentLoaded', () => {
    const tabs = document.querySelectorAll('.tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const targetTab = tab.dataset.tab;

            // Load data when switching to specific tabs
            if (targetTab === 'routes') {
                loadRoutes();
            } else if (targetTab === 'metrics') {
                loadMetrics();
            } else if (targetTab === 'deps') {
                loadDeps();
            } else if (targetTab === 'config') {
                loadConfigEditor();
            }
        });
    });

    // Refresh routes button
    const btnRefreshRoutes = document.getElementById('btn-refresh-routes');
    if (btnRefreshRoutes) {
        btnRefreshRoutes.addEventListener('click', loadRoutes);
    }

    // Auto-refresh metrics every 5 seconds
    setInterval(() => {
        const metricsTab = document.getElementById('metrics-tab');
        if (metricsTab && metricsTab.classList.contains('active')) {
            loadMetrics();
        }
    }, 5000);
});
