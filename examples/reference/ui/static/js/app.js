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
    statusEl.innerHTML = '<i class="fas fa-circle"></i><span>已连接</span>';
  } else {
    statusEl.className = 'connection-status connection-disconnected';
    statusEl.innerHTML = '<i class="fas fa-circle"></i><span>未连接</span>';
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
  
  entry.innerHTML = `
    <div class="log-timestamp">${timestamp}</div>
    <div class="log-content ${contentClass}">${message}</div>
  `;
  
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
    addLog('已经连接到 WebSocket', 'info');
    return;
  }

  try {
    ws = new WebSocket(`ws://${location.host}/ws`);
    connectionStartTime = Date.now();
    
    addLog('正在连接 WebSocket...', 'info');
    
    // 更新按钮状态
    document.getElementById('connectBtn').style.display = 'none';
    document.getElementById('disconnectBtn').style.display = 'inline-flex';
    
    ws.onopen = function() {
      updateConnectionStatus(true);
      addLog('WebSocket 连接已建立', 'received');
      // 发送自动连接测试消息
      setTimeout(() => {
        ws.send('hello plumego - 自动连接测试');
      }, 500);
    };

    ws.onmessage = function(event) {
      addLog(`接收: ${event.data}`, 'received');
    };

    ws.onclose = function(event) {
      updateConnectionStatus(false);
      addLog(`WebSocket 连接已关闭 (代码: ${event.code})`, 'error');
      connectionStartTime = null;
      // 恢复按钮状态
      document.getElementById('connectBtn').style.display = 'inline-flex';
      document.getElementById('disconnectBtn').style.display = 'none';
    };

    ws.onerror = function(error) {
      addLog('WebSocket 发生错误', 'error');
      console.error('WebSocket error:', error);
    };
  } catch (error) {
    addLog(`连接失败: ${error.message}`, 'error');
  }
}

function disconnectWebSocket() {
  if (ws) {
    ws.close();
    ws = null;
    updateConnectionStatus(false);
    connectionStartTime = null;
    // 恢复按钮状态
    document.getElementById('connectBtn').style.display = 'inline-flex';
    document.getElementById('disconnectBtn').style.display = 'none';
  }
}

function sendMessage() {
  const input = document.getElementById('messageInput');
  const message = input.value.trim();
  
  if (!message) {
    addLog('请输入要发送的消息', 'error');
    return;
  }
  
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    addLog('WebSocket 未连接', 'error');
    return;
  }
  
  ws.send(message);
  addLog(`发送: ${message}`, 'sent');
  input.value = '';
}

function clearLog() {
  const logEl = document.getElementById('messageLog');
  logEl.innerHTML = '<div class="log-entry"><div class="log-timestamp">系统</div><div class="log-content log-info">日志已清空</div></div>';
  stats = { sent: 0, received: 0, errors: 0 };
  updateStats();
}

// 高级测试功能
async function testAPI(event) {
  event.preventDefault();
  
  const path = document.getElementById('apiPath').value;
  const format = document.getElementById('apiFormat').value;
  const params = document.getElementById('apiParams').value;
  
  const responseEl = document.getElementById('apiResponse');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">正在执行 API 测试...</div>';
  
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
      responseEl.innerHTML = `<div style="color: var(--warning);">响应类型: ${contentType || 'unknown'}</div>`;
    }
    
    addLog(`API 测试完成: ${path}`, 'info');
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">错误: ${error.message}</div>`;
    addLog(`API 测试失败: ${error.message}`, 'error');
  }
}

async function checkHealth() {
  const responseEl = document.getElementById('healthMetrics');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">正在检查健康状态...</div>';
  
  try {
    const response = await fetch('/health/detailed');
    const data = await response.json();
    
    const healthStatus = data.status === 'healthy' ? '🟢 健康' : '🔴 异常';
    const uptime = data.system?.uptime || '未知';
    
    responseEl.innerHTML = `
      <div style="color: var(--success); font-weight: bold;">${healthStatus}</div>
      <div style="margin-top: 0.5rem;">运行时间: ${uptime}</div>
      <div>组件状态:</div>
      <ul style="margin: 0.5rem 0;">
        ${Object.entries(data.components || {}).map(([key, value]) => 
          `<li style="color: ${value === 'enabled' ? 'var(--success)' : 'var(--error)'};">${key}: ${value}</li>`
        ).join('')}
      </ul>
    `;
    
    addLog('健康检查完成', 'info');
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">健康检查失败: ${error.message}</div>`;
    addLog(`健康检查失败: ${error.message}`, 'error');
  }
}

async function loadMetrics() {
  const responseEl = document.getElementById('healthMetrics');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">正在加载指标...</div>';
  
  try {
    const response = await fetch('/metrics');
    const data = await response.text();
    
    responseEl.innerHTML = `<pre style="margin: 0; color: var(--text-secondary); font-size: 0.75rem; max-height: 300px; overflow-y: auto;">${data}</pre>`;
    addLog('指标加载完成', 'info');
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">指标加载失败: ${error.message}</div>`;
    addLog(`指标加载失败: ${error.message}`, 'error');
  }
}

async function testWebhook(event) {
  event.preventDefault();
  
  const type = document.getElementById('webhookType').value;
  const data = document.getElementById('webhookData').value;
  
  const responseEl = document.getElementById('webhookResponse');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">正在发送 Webhook...</div>';
  
  try {
    let payload;
    try {
      payload = data.trim() ? JSON.parse(data) : {};
    } catch (parseError) {
      throw new Error('JSON 格式错误: ' + parseError.message);
    }
    
    // 根据类型添加默认字段
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
      <div style="color: var(--success);">状态: ${response.status}</div>
      <pre style="margin: 0.5rem 0 0 0; color: var(--text-secondary);">${JSON.stringify(responseData, null, 2)}</pre>
    `;
    
    addLog(`Webhook 测试完成: ${type}`, 'info');
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">Webhook 测试失败: ${error.message}</div>`;
    addLog(`Webhook 测试失败: ${error.message}`, 'error');
  }
}

async function testPubSub(event) {
  event.preventDefault();
  
  const topic = document.getElementById('pubSubTopic').value;
  const message = document.getElementById('pubSubMessage').value;
  
  const responseEl = document.getElementById('pubSubResponse');
  responseEl.classList.add('show');
  responseEl.innerHTML = '<div style="color: var(--warning);">正在发布消息...</div>';
  
  try {
    if (!topic.trim()) {
      throw new Error('主题名称不能为空');
    }
    
    const url = `/test/pubsub?topic=${encodeURIComponent(topic)}`;
    const response = await fetch(url);
    const data = await response.json();
    
    responseEl.innerHTML = `
      <div style="color: var(--success);">消息发布成功</div>
      <div style="margin-top: 0.5rem;">主题: ${data.topic}</div>
      <div>消息: ${data.message}</div>
      <div>时间: ${data.timestamp}</div>
    `;
    
    addLog(`Pub/Sub 消息已发布到主题: ${topic}`, 'info');
    
    // 清空输入
    document.getElementById('pubSubMessage').value = '';
  } catch (error) {
    responseEl.innerHTML = `<div style="color: var(--error);">Pub/Sub 测试失败: ${error.message}</div>`;
    addLog(`Pub/Sub 测试失败: ${error.message}`, 'error');
  }
}

// 回车发送消息
document.getElementById('messageInput').addEventListener('keypress', function(e) {
  if (e.key === 'Enter') {
    sendMessage();
  }
});

// 自动更新运行时间
setInterval(updateUptime, 1000);

// 页面加载时自动连接
window.addEventListener('load', function() {
  // 延迟连接确保页面完全加载
  setTimeout(connectWebSocket, 1000);
});
