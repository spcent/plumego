let ws = null;
let connectionStartTime = null;
let stats = {
  sent: 0,
  received: 0,
  errors: 0
};

function updateConnectionStatus(connected) {
  const statusEl = document.getElementById('connectionStatus');
  if (connected) {
    statusEl.className = 'connection-status connection-connected';
    statusEl.innerHTML = '<i class="fas fa-circle"></i><span>Connected</span>';
  } else {
    statusEl.className = 'connection-status connection-disconnected';
    statusEl.innerHTML = '<i class="fas fa-circle"></i><span>Disconnected</span>';
  }
}

function addLog(message, type = 'info') {
  const logEl = document.getElementById('messageLog');
  const now = new Date();
  const timestamp = now.toLocaleTimeString();

  const entry = document.createElement('div');
  entry.className = 'log-entry';

  let contentClass = '';
  switch (type) {
    case 'sent': contentClass = 'log-sent'; stats.sent++; break;
    case 'received': contentClass = 'log-received'; stats.received++; break;
    case 'error': contentClass = 'log-error'; stats.errors++; break;
  }

  const timestampEl = document.createElement('div');
  timestampEl.className = 'log-timestamp';
  timestampEl.textContent = timestamp;

  const contentEl = document.createElement('div');
  contentEl.className = `log-content ${contentClass}`;
  contentEl.textContent = message;

  entry.appendChild(timestampEl);
  entry.appendChild(contentEl);

  logEl.appendChild(entry);
  logEl.scrollTop = logEl.scrollHeight;

  updateStats();
}

function updateStats() {
  document.getElementById('sentCount').textContent = stats.sent;
  document.getElementById('receivedCount').textContent = stats.received;
  document.getElementById('errorCount').textContent = stats.errors;
}

function updateUptime() {
  if (connectionStartTime) {
    const now = Date.now();
    const diff = Math.floor((now - connectionStartTime) / 1000);
    const minutes = Math.floor(diff / 60);
    const seconds = diff % 60;
    document.getElementById('uptime').textContent = 
      `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
  }
}

function connectWebSocket() {
  if (ws && ws.readyState === WebSocket.OPEN) {
    addLog('Already connected to WebSocket', 'info');
    return;
  }

  try {
    ws = new WebSocket(`ws://${location.host}/ws`);
    connectionStartTime = Date.now();
    
    addLog('Connecting to WebSocket...', 'info');
    
    // Update button states
    document.getElementById('connectBtn').style.display = 'none';
    document.getElementById('disconnectBtn').style.display = 'inline-flex';
    
    ws.onopen = function() {
      updateConnectionStatus(true);
      addLog('WebSocket connection established', 'received');
      // Send automatic connection test message
      setTimeout(() => {
        ws.send('hello plumego - automatic connection test');
      }, 500);
    };

    ws.onmessage = function(event) {
      addLog(`Received: ${event.data}`, 'received');
    };

    ws.onclose = function(event) {
      updateConnectionStatus(false);
      addLog(`WebSocket connection closed (code: ${event.code})`, 'error');
      connectionStartTime = null;
      // Restore button states
      document.getElementById('connectBtn').style.display = 'inline-flex';
      document.getElementById('disconnectBtn').style.display = 'none';
    };

    ws.onerror = function(error) {
      addLog('WebSocket error occurred', 'error');
      console.error('WebSocket error:', error);
    };
  } catch (error) {
    addLog(`Connection failed: ${error.message}`, 'error');
  }
}

function disconnectWebSocket() {
  if (ws) {
    ws.close();
    ws = null;
    updateConnectionStatus(false);
    connectionStartTime = null;
    // Restore button states
    document.getElementById('connectBtn').style.display = 'inline-flex';
    document.getElementById('disconnectBtn').style.display = 'none';
  }
}

function sendMessage() {
  const input = document.getElementById('messageInput');
  const message = input.value.trim();
  
  if (!message) {
    addLog('Please enter a message to send', 'error');
    return;
  }
  
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    addLog('WebSocket not connected', 'error');
    return;
  }
  
  ws.send(message);
  addLog(`Sent: ${message}`, 'sent');
  input.value = '';
}

function clearLog() {
  const logEl = document.getElementById('messageLog');
  logEl.innerHTML = '<div class="log-entry"><div class="log-timestamp">System</div><div class="log-content log-info">Log cleared</div></div>';
  stats = { sent: 0, received: 0, errors: 0 };
  updateStats();
}

// Advanced testing functions
async function testAPI(event) {
  event.preventDefault();
  
  const path = document.getElementById('apiPath').value;
  const format = document.getElementById('apiFormat').value;
  const params = document.getElementById('apiParams').value;
  
  const responseEl = document.getElementById('apiResponse');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">Executing API test...</div>';
  
  try {
    let url = path;
    if (params.trim()) {
      url += (url.includes('?') ? '&' : '?') + params;
    }
    if (format && format !== 'json') {
      url += (url.includes('?') ? '&' : '?') + `format=${format}`;
    }
    
    const response = await fetch(url);
    const contentType = response.headers.get('content-type') || '';
    let data;
    
    if (contentType.includes('application/json')) {
      data = await response.json();
      responseEl.innerHTML = `<pre style="margin: 0; color: var(--success);">${JSON.stringify(data, null, 2)}</pre>`;
    } else if (contentType.includes('text/')) {
      data = await response.text();
      responseEl.innerHTML = `<pre style="margin: 0; color: var(--text-secondary);">${data}</pre>`;
    } else {
      data = await response.blob();
      responseEl.innerHTML = `<div style="color: var(--warning);">Response type: ${contentType || 'unknown'}</div>`;
    }
    
    addLog(`API test completed: ${path}`, 'info');
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">Error: ${error.message}</div>`;
    addLog(`API test failed: ${error.message}`, 'error');
  }
}

async function checkHealth() {
  const responseEl = document.getElementById('healthMetrics');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">Checking health status...</div>';
  
  try {
    const response = await fetch('/health/detailed');
    const data = await response.json();
    
    const healthStatus = data.status === 'healthy' ? '🟢 Healthy' : '🔴 Abnormal';
    const uptime = data.system?.uptime || 'Unknown';
    
    responseEl.innerHTML = `
      <div style="color: var(--success); font-weight: bold;">${healthStatus}</div>
      <div style="margin-top: 0.5rem;">Uptime: ${uptime}</div>
      <div>Component Status:</div>
      <ul style="margin: 0.5rem 0;">
        ${Object.entries(data.components || {}).map(([key, value]) => 
          `<li style="color: ${value === 'enabled' ? 'var(--success)' : 'var(--error)'};">${key}: ${value}</li>`
        ).join('')}
      </ul>
    `;
    addLog('Health check completed', 'info');
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">Health check failed: ${error.message}</div>`;
    addLog(`Health check failed: ${error.message}`, 'error');
  }
}

async function loadMetrics() {
  const responseEl = document.getElementById('healthMetrics');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">Loading metrics...</div>';
  
  try {
    const response = await fetch('/metrics');
    const data = await response.text();
    
    responseEl.innerHTML = `<pre style="margin: 0; color: var(--text-secondary); font-size: 0.75rem; max-height: 300px; overflow-y: auto;">${data}</pre>`;
    addLog('Metrics loaded', 'info');
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">Failed to load metrics: ${error.message}</div>`;
    addLog(`Failed to load metrics: ${error.message}`, 'error');
  }
}

async function testWebhook(event) {
  event.preventDefault();
  
  const type = document.getElementById('webhookType').value;
  const data = document.getElementById('webhookData').value;
  
  const responseEl = document.getElementById('webhookResponse');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">Sending Webhook...</div>';
  
  try {
    let payload;
    try {
      payload = data.trim() ? JSON.parse(data) : {};
    } catch (parseError) {
      throw new Error('JSON format error: ' + parseError.message);
    }
    
    // Add default fields based on type
    if (type === 'github') {
      payload = {
        action: 'opened',
        pull_request: {
          id: 12345,
          number: 1,
          title: 'Test PR',
          user: { login: 'testuser' }
        },
        repository: { full_name: 'test/repo' },
        ...payload
      };
    } else if (type === 'stripe') {
      payload = {
        id: 'evt_test_webhook',
        object: 'event',
        type: 'test.event',
        data: {
          object: {
            id: 'cus_test',
            object: 'customer',
            email: 'test@example.com'
          }
        },
        ...payload
      };
    }
    
    const endpoint = type === 'github' ? '/webhooks/github' : 
                    type === 'stripe' ? '/webhooks/stripe' : '/test/webhook';
    
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'Plumego-Test-Tool/1.0'
      },
      body: JSON.stringify(payload)
    });
    
    const responseData = await response.json();
    
    responseEl.innerHTML = `
      <div style="color: var(--success);">Status: ${response.status}</div>
      <pre style="margin: 0.5rem 0 0 0; color: var(--text-secondary);">${JSON.stringify(responseData, null, 2)}</pre>
    `;
    
    addLog(`Webhook test completed: ${type}`, 'info');
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">Webhook test failed: ${error.message}</div>`;
    addLog(`Webhook test failed: ${error.message}`, 'error');
  }
}

async function testPubSub(event) {
  event.preventDefault();
  
  const topic = document.getElementById('pubSubTopic').value;
  const message = document.getElementById('pubSubMessage').value;
  
  const responseEl = document.getElementById('pubSubResponse');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">Publishing message...</div>';
  
  try {
    if (!topic.trim()) {
      throw new Error('Topic name cannot be empty');
    }
    
    const url = `/test/pubsub?topic=${encodeURIComponent(topic)}`;
    const response = await fetch(url);
    const data = await response.json();
    
    responseEl.innerHTML = `
      <div style="color: var(--success);">Message published successfully</div>
      <div style="margin-top: 0.5rem;">Topic: ${data.topic}</div>
      <div>Message: ${data.message}</div>
      <div>Time: ${data.timestamp}</div>
    `;
    
    addLog(`Pub/Sub message published to topic: ${topic}`, 'info');
    
    // Clear input
    document.getElementById('pubSubMessage').value = '';
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">Pub/Sub test failed: ${error.message}</div>`;
    addLog(`Pub/Sub test failed: ${error.message}`, 'error');
  }
}

// Send message on Enter key
document.getElementById('messageInput').addEventListener('keypress', function(e) {
  if (e.key === 'Enter') {
    sendMessage();
  }
});

// Auto-update uptime
setInterval(updateUptime, 1000);

// Auto-connect on page load
window.addEventListener('load', function() {
  // Delay connection to ensure page is fully loaded
  setTimeout(connectWebSocket, 1000);
});
