'use client'

import { useState, useEffect } from 'react'
import { Activity, Server, CheckCircle, XCircle, Wifi } from 'lucide-react'

interface C2Adapter {
  name: string
  enabled: boolean
  connected: boolean
  server_url: string
}

interface C2Health {
  sliver: { status: string; latency_ms: number; implants_count: number }
  empire: { status: string; latency_ms: number; agents_count: number }
}

export default function C2Tab() {
  const [adapters, setAdapters] = useState<C2Adapter[]>([])
  const [health, setHealth] = useState<C2Health | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadC2Data()
    const interval = setInterval(loadC2Data, 30000)
    return () => clearInterval(interval)
  }, [])

  const loadC2Data = async () => {
    try {
      const [adaptersResponse, healthResponse] = await Promise.all([
        fetch('/api/v1/c2/adapters'),
        fetch('/api/v1/c2/health')
      ])
      
      const adaptersData = await adaptersResponse.json()
      const healthData = await healthResponse.json()
      
      setAdapters(adaptersData.adapters || [])
      setHealth(healthData)
    } catch (err) {
      console.error('Failed to load C2 data:', err)
    } finally {
      setLoading(false)
    }
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
          <h2 className="text-2xl font-bold text-green-400">C2 INTEGRATION</h2>
          <p className="text-green-700">Command & Control framework integration</p>
        </div>
        <button className="flex items-center space-x-2 px-4 py-2 bg-green-600 hover:bg-green-700 rounded transition">
          <Wifi className="w-4 h-4" />
          <span>Refresh Status</span>
        </button>
      </div>

      {/* C2 Health Status */}
      <div className="grid grid-cols-2 gap-6">
        {health && Object.entries(health).map(([name, data]: [string, any]) => (
          <div key={name} className="bg-green-900/10 border border-green-800 rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-green-400 capitalize">{name}</h3>
              {data.status === 'healthy' ? (
                <CheckCircle className="w-5 h-5 text-green-500" />
              ) : (
                <XCircle className="w-5 h-5 text-red-500" />
              )}
            </div>
            
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-green-600">Status</span>
                <span className={`text-sm font-bold ${
                  data.status === 'healthy' ? 'text-green-400' : 'text-red-400'
                }`}>
                  {data.status}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-green-600">Latency</span>
                <span className="text-sm text-green-400">{data.latency_ms}ms</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-green-600">
                  {name === 'sliver' ? 'Implants' : 'Agents'}
                </span>
                <span className="text-sm text-green-400 font-bold">
                  {name === 'sliver' ? data.implants_count : data.agents_count}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Adapters List */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg overflow-hidden">
        <div className="px-6 py-4 bg-green-900/30 border-b border-green-800">
          <h3 className="text-lg font-bold text-green-400 flex items-center">
            <Server className="w-5 h-5 mr-2" />
            C2 Adapters
          </h3>
        </div>
        <table className="w-full">
          <thead className="bg-green-900/30">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Name</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Server URL</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Enabled</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Connected</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-green-900/30">
            {adapters.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-green-700">
                  No C2 adapters configured
                </td>
              </tr>
            ) : (
              adapters.map((adapter) => (
                <tr key={adapter.name} className="hover:bg-green-900/20 transition">
                  <td className="px-4 py-3 text-sm font-bold text-green-400 capitalize">
                    {adapter.name}
                  </td>
                  <td className="px-4 py-3 text-sm text-green-600 font-mono">
                    {adapter.server_url || 'Not configured'}
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${
                      adapter.enabled ? 'bg-green-900/50 text-green-400' : 'bg-gray-900/50 text-gray-400'
                    }`}>
                      {adapter.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${
                      adapter.connected ? 'bg-green-900/50 text-green-400' : 'bg-red-900/50 text-red-400'
                    }`}>
                      {adapter.connected ? 'Connected' : 'Disconnected'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <button className="px-3 py-1 bg-green-900/30 border border-green-700 rounded hover:bg-green-900/50 transition text-green-400">
                      Configure
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Supported Frameworks */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4">Supported C2 Frameworks</h3>
        <div className="grid grid-cols-3 gap-4">
          {[
            { name: 'Sliver', status: 'Full Support', icon: '🎯' },
            { name: 'Empire', status: 'Full Support', icon: '👑' },
            { name: 'Cobalt Strike', status: 'External C2', icon: '⚡' },
            { name: 'Metasploit', status: 'Coming Soon', icon: '🔨' },
            { name: 'DNS Tunnel', status: 'Full Support', icon: '📡' },
            { name: 'HTTP Callback', status: 'Full Support', icon: '🔄' }
          ].map((framework) => (
            <div key={framework.name} className="bg-green-900/20 rounded p-4">
              <div className="text-2xl mb-2">{framework.icon}</div>
              <div className="font-bold text-green-400">{framework.name}</div>
              <div className="text-xs text-green-600">{framework.status}</div>
            </div>
          ))}
        </div>
      </div>

      {/* Configuration Guide */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4">Quick Setup Guide</h3>
        <div className="space-y-3 text-sm text-green-400">
          <div className="flex items-start space-x-3">
            <span className="text-green-600 font-bold">1.</span>
            <div>
              <div className="font-bold">Configure Sliver C2</div>
              <div className="text-green-600">Edit config.yaml: v13.c2.sliver.server_url, operator_token</div>
            </div>
          </div>
          <div className="flex items-start space-x-3">
            <span className="text-green-600 font-bold">2.</span>
            <div>
              <div className="font-bold">Configure Empire C2</div>
              <div className="text-green-600">Edit config.yaml: v13.c2.empire.server_url, username, password</div>
            </div>
          </div>
          <div className="flex items-start space-x-3">
            <span className="text-green-600 font-bold">3.</span>
            <div>
              <div className="font-bold">Test Connection</div>
              <div className="text-green-600">Click "Refresh Status" to verify C2 connectivity</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
