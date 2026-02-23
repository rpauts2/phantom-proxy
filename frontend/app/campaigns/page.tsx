'use client'

import { useState, useEffect, useCallback } from 'react'
import { 
  Mail, Play, Pause, Trash2, Plus, Search, Filter,
  Users, FileText, Globe, Send, BarChart3, CheckCircle,
  Clock, XCircle, AlertCircle
} from 'lucide-react'

interface Campaign {
  id: string
  name: string
  status: string
  template: string
  page: string
  group: string
  sent: number
  opened: number
  clicked: number
  submitted: number
  created_at: string
}

interface Group {
  id: string
  name: string
  count: number
}

interface Template {
  id: string
  name: string
  subject: string
}

interface LandingPage {
  id: string
  name: string
}

interface SMTPProfile {
  id: string
  name: string
  host: string
}

const API_BASE = '/api/v1'

async function apiCall(endpoint: string, options?: RequestInit) {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options?.headers },
  })
  if (!response.ok) throw new Error(`Error: ${response.statusText}`)
  return response.json()
}

export default function CampaignsPage() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [groups, setGroups] = useState<Group[]>([])
  const [templates, setTemplates] = useState<Template[]>([])
  const [pages, setPages] = useState<LandingPage[]>([])
  const [profiles, setProfiles] = useState<SMTPProfile[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [activeTab, setActiveTab] = useState<'campaigns' | 'groups' | 'templates' | 'pages' | 'smtp'>('campaigns')

  const fetchData = useCallback(async () => {
    try {
      const [campsData, groupsData, tempsData, pagesData, profilesData] = await Promise.all([
        apiCall('/campaigns'),
        apiCall('/groups'),
        apiCall('/templates'),
        apiCall('/pages'),
        apiCall('/smtp'),
      ])
      
      setCampaigns(campsData.campaigns || [])
      setGroups(groupsData.groups || [])
      setTemplates(tempsData.templates || [])
      setPages(pagesData.pages || [])
      setProfiles(profilesData.profiles || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const startCampaign = async (id: string) => {
    try {
      await apiCall(`/campaigns/${id}/start`, { method: 'POST' })
      fetchData()
    } catch (err) {
      console.error(err)
    }
  }

  const pauseCampaign = async (id: string) => {
    try {
      await apiCall(`/campaigns/${id}/pause`, { method: 'POST' })
      fetchData()
    } catch (err) {
      console.error(err)
    }
  }

  const deleteCampaign = async (id: string) => {
    if (!confirm('Удалить кампанию?')) return
    try {
      await apiCall(`/campaigns/${id}`, { method: 'DELETE' })
      fetchData()
    } catch (err) {
      console.error(err)
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'running':
        return <span className="flex items-center gap-1 px-2 py-1 bg-green-900/50 text-green-400 rounded text-xs"><Play className="w-3 h-3" /> Активна</span>
      case 'paused':
        return <span className="flex items-center gap-1 px-2 py-1 bg-yellow-900/50 text-yellow-400 rounded text-xs"><Pause className="w-3 h-3" /> Приостановлена</span>
      case 'complete':
        return <span className="flex items-center gap-1 px-2 py-1 bg-blue-900/50 text-blue-400 rounded text-xs"><CheckCircle className="w-3 h-3" /> Завершена</span>
      case 'failed':
        return <span className="flex items-center gap-1 px-2 py-1 bg-red-900/50 text-red-400 rounded text-xs"><XCircle className="w-3 h-3" /> Ошибка</span>
      default:
        return <span className="flex items-center gap-1 px-2 py-1 bg-slate-700 text-slate-400 rounded text-xs"><Clock className="w-3 h-3" /> Ожидает</span>
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-900 flex items-center justify-center">
        <div className="text-center">
          <Mail className="w-12 h-12 text-blue-500 mx-auto mb-4 animate-pulse" />
          <p className="text-blue-400">Загрузка...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-900 text-white">
      {/* Header */}
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
        <div className="flex items-center justify-between max-w-7xl mx-auto">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-600 rounded-lg">
              <Mail className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold">Кампании</h1>
              <p className="text-xs text-slate-400">Управление email кампаниями</p>
            </div>
          </div>
          <button 
            onClick={() => setShowCreate(true)}
            className="flex items-center space-x-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
          >
            <Plus className="w-4 h-4" />
            <span>Создать кампанию</span>
          </button>
        </div>
      </header>

      {/* Tabs */}
      <div className="bg-slate-800/50 border-b border-slate-700">
        <div className="max-w-7xl mx-auto px-6">
          <div className="flex space-x-1">
            {[
              { id: 'campaigns', label: 'Кампании', icon: Send },
              { id: 'groups', label: 'Группы', icon: Users },
              { id: 'templates', label: 'Шаблоны', icon: FileText },
              { id: 'pages', label: 'Лендинги', icon: Globe },
              { id: 'smtp', label: 'SMTP', icon: Mail },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`flex items-center space-x-2 px-4 py-3 text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'text-blue-400 border-b-2 border-blue-400 bg-slate-800'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                <tab.icon className="w-4 h-4" />
                <span>{tab.label}</span>
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        {activeTab === 'campaigns' && (
          <div className="space-y-4">
            {campaigns.length > 0 ? campaigns.map((campaign) => (
              <div key={campaign.id} className="bg-slate-800 rounded-xl p-4">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-3">
                    <h3 className="font-semibold">{campaign.name}</h3>
                    {getStatusBadge(campaign.status)}
                  </div>
                  <div className="flex items-center space-x-2">
                    {campaign.status === 'pending' && (
                      <button 
                        onClick={() => startCampaign(campaign.id)}
                        className="p-2 hover:bg-green-600/20 rounded text-green-400"
                        title="Запустить"
                      >
                        <Play className="w-4 h-4" />
                      </button>
                    )}
                    {campaign.status === 'running' && (
                      <button 
                        onClick={() => pauseCampaign(campaign.id)}
                        className="p-2 hover:bg-yellow-600/20 rounded text-yellow-400"
                        title="Приостановить"
                      >
                        <Pause className="w-4 h-4" />
                      </button>
                    )}
                    <button 
                      onClick={() => deleteCampaign(campaign.id)}
                      className="p-2 hover:bg-red-600/20 rounded text-red-400"
                      title="Удалить"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
                
                <div className="grid grid-cols-4 gap-4">
                  <div className="text-center p-3 bg-slate-700/50 rounded-lg">
                    <p className="text-2xl font-bold text-blue-400">{campaign.sent || 0}</p>
                    <p className="text-xs text-slate-400">Отправлено</p>
                  </div>
                  <div className="text-center p-3 bg-slate-700/50 rounded-lg">
                    <p className="text-2xl font-bold text-green-400">{campaign.opened || 0}</p>
                    <p className="text-xs text-slate-400">Открыто</p>
                  </div>
                  <div className="text-center p-3 bg-slate-700/50 rounded-lg">
                    <p className="text-2xl font-bold text-yellow-400">{campaign.clicked || 0}</p>
                    <p className="text-xs text-slate-400">Кликнуто</p>
                  </div>
                  <div className="text-center p-3 bg-slate-700/50 rounded-lg">
                    <p className="text-2xl font-bold text-purple-400">{campaign.submitted || 0}</p>
                    <p className="text-xs text-slate-400">Отправлено данных</p>
                  </div>
                </div>
              </div>
            )) : (
              <div className="bg-slate-800 rounded-xl p-12 text-center">
                <Mail className="w-12 h-12 text-slate-600 mx-auto mb-4" />
                <p className="text-slate-400">Нет кампаний</p>
                <p className="text-sm text-slate-500 mt-1">Создайте первую кампанию</p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'groups' && (
          <div className="grid md:grid-cols-3 gap-4">
            {groups.map((group) => (
              <div key={group.id} className="bg-slate-800 rounded-xl p-4">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-semibold">{group.name}</h3>
                  <span className="text-xs text-slate-400">{group.count} целей</span>
                </div>
              </div>
            ))}
            <button className="border-2 border-dashed border-slate-700 rounded-xl p-4 hover:border-slate-600 transition-colors text-slate-500 hover:text-slate-400">
              <Plus className="w-8 h-8 mx-auto mb-2" />
              <p>Добавить группу</p>
            </button>
          </div>
        )}

        {activeTab === 'templates' && (
          <div className="grid md:grid-cols-3 gap-4">
            {templates.map((template) => (
              <div key={template.id} className="bg-slate-800 rounded-xl p-4">
                <h3 className="font-semibold mb-1">{template.name}</h3>
                <p className="text-sm text-slate-400">{template.subject}</p>
              </div>
            ))}
            <button className="border-2 border-dashed border-slate-700 rounded-xl p-4 hover:border-slate-600 transition-colors text-slate-500 hover:text-slate-400">
              <Plus className="w-8 h-8 mx-auto mb-2" />
              <p>Добавить шаблон</p>
            </button>
          </div>
        )}

        {activeTab === 'pages' && (
          <div className="grid md:grid-cols-3 gap-4">
            {pages.map((page) => (
              <div key={page.id} className="bg-slate-800 rounded-xl p-4">
                <h3 className="font-semibold">{page.name}</h3>
              </div>
            ))}
            <button className="border-2 border-dashed border-slate-700 rounded-xl p-4 hover:border-slate-600 transition-colors text-slate-500 hover:text-slate-400">
              <Plus className="w-8 h-8 mx-auto mb-2" />
              <p>Добавить лендинг</p>
            </button>
          </div>
        )}

        {activeTab === 'smtp' && (
          <div className="grid md:grid-cols-3 gap-4">
            {profiles.map((profile) => (
              <div key={profile.id} className="bg-slate-800 rounded-xl p-4">
                <h3 className="font-semibold mb-1">{profile.name}</h3>
                <p className="text-sm text-slate-400">{profile.host}</p>
              </div>
            ))}
            <button className="border-2 border-dashed border-slate-700 rounded-xl p-4 hover:border-slate-600 transition-colors text-slate-500 hover:text-slate-400">
              <Plus className="w-8 h-8 mx-auto mb-2" />
              <p>Добавить SMTP</p>
            </button>
          </div>
        )}
      </main>

      {/* Create Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-xl p-6 w-full max-w-md">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold">Создать кампанию</h2>
              <button onClick={() => setShowCreate(false)} className="text-slate-400 hover:text-white">
                <XCircle className="w-5 h-5" />
              </button>
            </div>
            
            <form className="space-y-4">
              <div>
                <label className="block text-sm text-slate-400 mb-2">Название</label>
                <input 
                  type="text" 
                  className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500"
                  placeholder="Моя кампания"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-2">Группа</label>
                <select className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500">
                  {groups.map(g => <option key={g.id} value={g.id}>{g.name}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-2">Шаблон</label>
                <select className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500">
                  {templates.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-2">Лендинг</label>
                <select className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500">
                  {pages.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-2">SMTP профиль</label>
                <select className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500">
                  {profiles.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-2">URL фишинга</label>
                <input 
                  type="text" 
                  className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg focus:outline-none focus:border-blue-500"
                  placeholder="https://login.company.com"
                />
              </div>
              <button 
                type="submit"
                className="w-full py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
              >
                Создать кампанию
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
