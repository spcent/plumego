// State
let ws = null;
let logs = [];
let events = [];
let logCount = 0;
let errorCount = 0;
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
    filterError: document.getElementById('filter-error')
};

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initWebSocket();
    initTabs();
    initControls();
    initFilters();
});

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
    div.innerHTML = `
        <span class="log-time">${time}</span>
        <span class="log-level" style="color: ${getLevelColor(level)}">${level.toUpperCase()}</span>
        <span class="log-message">${escapeHtml(message)}</span>
    `;

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
