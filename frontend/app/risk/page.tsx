'use client'

import { useState, useEffect } from 'react'
import { TrendingUp, AlertTriangle, Shield, Activity, Users, Download } from 'lucide-react'

interface RiskDistribution {
  low: number
  medium: number
  high: number
  critical: number
}

interface HighRiskUser {
  user_id: string
  email: string
  overall_score: number
  risk_level: string
  trend: string
  last_updated: string
}

export default function RiskTab() {
  const [distribution, setDistribution] = useState<RiskDistribution>({ low: 0, medium: 0, high: 0, critical: 0 })
  const [highRiskUsers, setHighRiskUsers] = useState<HighRiskUser[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedUser, setSelectedUser] = useState<HighRiskUser | null>(null)

  useEffect(() => {
    loadRiskData()
    const interval = setInterval(loadRiskData, 60000) // Refresh every minute
    return () => clearInterval(interval)
  }, [])

  const loadRiskData = async () => {
    try {
      const [distResponse, usersResponse] = await Promise.all([
        fetch('/api/v1/risk/distribution'),
        fetch('/api/v1/risk/high-risk?limit=20')
      ])
      
      const distData = await distResponse.json()
      const usersData = await usersResponse.json()
      
      setDistribution(distData.distribution || { low: 0, medium: 0, high: 0, critical: 0 })
      setHighRiskUsers(usersData.users || [])
    } catch (err) {
      console.error('Failed to load risk data:', err)
    } finally {
      setLoading(false)
    }
  }

  const exportRiskData = () => {
    const csv = highRiskUsers.map(u => 
      `${u.user_id},${u.email},${u.overall_score},${u.risk_level},${u.trend},${u.last_updated}`
    ).join('\n')
    
    const blob = new Blob([`ID,Email,Score,Risk Level,Trend,Last Updated\n${csv}`], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `risk_report_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
  }

  const totalUsers = Object.values(distribution).reduce((a, b) => a + b, 0)

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
          <h2 className="text-2xl font-bold text-green-400">RISK ANALYSIS</h2>
          <p className="text-green-700">Human Risk Score distribution and high-risk users</p>
        </div>
        <button
          onClick={exportRiskData}
          className="flex items-center space-x-2 px-4 py-2 bg-green-900/30 border border-green-700 rounded hover:bg-green-900/50 transition"
        >
          <Download className="w-4 h-4" />
          <span>Export Report</span>
        </button>
      </div>

      {/* Distribution Chart */}
      <div className="grid grid-cols-4 gap-4">
        {[
          { level: 'low', color: 'green', label: 'Low Risk' },
          { level: 'medium', color: 'yellow', label: 'Medium Risk' },
          { level: 'high', color: 'orange', label: 'High Risk' },
          { level: 'critical', color: 'red', label: 'Critical Risk' }
        ].map((item) => {
          const count = distribution[item.level as keyof RiskDistribution]
          const percentage = totalUsers > 0 ? (count / totalUsers * 100).toFixed(1) : 0
          
          return (
            <div key={item.level} className="bg-green-900/10 border border-green-800 rounded-lg p-6">
              <div className="flex items-center justify-between mb-2">
                <span className={`text-sm font-bold uppercase text-${item.color}-500`}>
                  {item.label}
                </span>
                {item.level === 'critical' && <AlertTriangle className="w-4 h-4 text-red-500" />}
              </div>
              <div className="text-3xl font-bold text-green-400 mb-2">{count}</div>
              <div className="h-2 bg-green-900/30 rounded-full overflow-hidden">
                <div
                  className={`h-full bg-${item.color}-500 transition-all duration-500`}
                  style={{ width: `${percentage}%` }}
                />
              </div>
              <div className="text-xs text-green-600 mt-2">{percentage}% of total</div>
            </div>
          )
        })}
      </div>

      {/* Risk Factors Info */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4 flex items-center">
          <Shield className="w-5 h-5 mr-2" />
          Risk Factors (8 Behavioral Indicators)
        </h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {[
            { name: 'Click Speed', weight: '15%' },
            { name: 'Form Submission', weight: '20%' },
            { name: 'Hover Patterns', weight: '10%' },
            { name: 'Time on Page', weight: '10%' },
            { name: 'Mouse Movement', weight: '10%' },
            { name: 'Keyboard Patterns', weight: '15%' },
            { name: 'Previous Clicks', weight: '10%' },
            { name: 'Device Fingerprint', weight: '10%' }
          ].map((factor) => (
            <div key={factor.name} className="bg-green-900/20 rounded p-3">
              <div className="text-xs text-green-600">{factor.name}</div>
              <div className="text-sm text-green-400 font-bold">{factor.weight}</div>
            </div>
          ))}
        </div>
      </div>

      {/* High Risk Users Table */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg overflow-hidden">
        <div className="px-6 py-4 bg-green-900/30 border-b border-green-800">
          <h3 className="text-lg font-bold text-green-400 flex items-center">
            <AlertTriangle className="w-5 h-5 mr-2 text-red-500" />
            High Risk Users ({highRiskUsers.length})
          </h3>
        </div>
        <table className="w-full">
          <thead className="bg-green-900/30">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Email</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Risk Score</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Level</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Trend</th>
              <th className="px-4 py-3 text-left text-xs font-bold text-green-500 uppercase">Last Updated</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-green-900/30">
            {highRiskUsers.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-green-700">
                  No high-risk users detected
                </td>
              </tr>
            ) : (
              highRiskUsers.map((user) => (
                <tr
                  key={user.user_id}
                  className="hover:bg-green-900/20 transition cursor-pointer"
                  onClick={() => setSelectedUser(user)}
                >
                  <td className="px-4 py-3 text-sm text-green-400">{user.email}</td>
                  <td className="px-4 py-3 text-sm">
                    <div className="flex items-center space-x-2">
                      <div className="flex-1 h-2 w-24 bg-green-900/30 rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full ${
                            user.overall_score >= 80 ? 'bg-red-500' :
                            user.overall_score >= 60 ? 'bg-orange-500' :
                            'bg-yellow-500'
                          }`}
                          style={{ width: `${user.overall_score}%` }}
                        />
                      </div>
                      <span className="text-green-400 font-bold">{user.overall_score}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${
                      user.risk_level === 'critical' ? 'bg-red-900/50 text-red-400' :
                      user.risk_level === 'high' ? 'bg-orange-900/50 text-orange-400' :
                      'bg-yellow-900/50 text-yellow-400'
                    }`}>
                      {user.risk_level}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <span className={`px-2 py-1 rounded text-xs ${
                      user.trend === 'worsening' ? 'bg-red-900/30 text-red-400' :
                      user.trend === 'improving' ? 'bg-green-900/30 text-green-400' :
                      'bg-gray-900/30 text-gray-400'
                    }`}>
                      {user.trend}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-green-600">
                    {new Date(user.last_updated).toLocaleString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* User Details Modal */}
      {selectedUser && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
          <div className="bg-green-900/20 border border-green-700 rounded-lg max-w-lg w-full">
            <div className="p-6 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-xl font-bold text-green-400">User Risk Details</h3>
                <button
                  onClick={() => setSelectedUser(null)}
                  className="text-green-600 hover:text-green-400"
                >
                  ✕
                </button>
              </div>

              <div className="space-y-3">
                <div>
                  <div className="text-xs text-green-600 uppercase">Email</div>
                  <div className="text-green-400">{selectedUser.email}</div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">Risk Score</div>
                  <div className="text-3xl font-bold text-green-400">{selectedUser.overall_score}/100</div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <div className="text-xs text-green-600 uppercase">Risk Level</div>
                    <div className="text-green-400">{selectedUser.risk_level}</div>
                  </div>
                  <div>
                    <div className="text-xs text-green-600 uppercase">Trend</div>
                    <div className="text-green-400">{selectedUser.trend}</div>
                  </div>
                </div>
                <div>
                  <div className="text-xs text-green-600 uppercase">Last Updated</div>
                  <div className="text-green-400">{new Date(selectedUser.last_updated).toLocaleString()}</div>
                </div>
              </div>

              <div className="border-t border-green-800 pt-4">
                <div className="text-sm text-green-600 mb-2">Recommendations:</div>
                <ul className="text-sm text-green-400 space-y-1">
                  <li>• Monitor user activity closely</li>
                  <li>• Consider additional security training</li>
                  <li>• Review recent login attempts</li>
                  <li>• Enable MFA if not already enabled</li>
                </ul>
              </div>

              <div className="flex items-center justify-end pt-4 border-t border-green-800">
                <button
                  onClick={() => setSelectedUser(null)}
                  className="px-4 py-2 bg-green-900/30 border border-green-700 rounded hover:bg-green-900/50 transition"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
