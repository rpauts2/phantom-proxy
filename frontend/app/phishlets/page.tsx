'use client'

import { useState, useEffect } from 'react'
import { FileText, CheckCircle, XCircle, Activity, Download, Plus, Trash2 } from 'lucide-react'

interface Phishlet {
  id: string
  name: string
  version: string
  status: string
  features: string
  author: string
}

export default function PhishletsTab() {
  const [phishlets, setPhishlets] = useState<Phishlet[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadPhishlets()
  }, [])

  const loadPhishlets = async () => {
    try {
      const response = await fetch('/api/v1/phishlets')
      const data = await response.json()
      setPhishlets(data.phishlets || [])
    } catch (err) {
      console.error('Failed to load phishlets:', err)
    } finally {
      setLoading(false)
    }
  }

  const togglePhishlet = async (id: string, currentStatus: string) => {
    try {
      const endpoint = currentStatus === 'ACTIVE' ? 'disable' : 'enable'
      await fetch(`/api/v1/phishlets/${id}/${endpoint}`, { method: 'POST' })
      loadPhishlets()
    } catch (err) {
      console.error('Failed to toggle phishlet:', err)
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
          <h2 className="text-2xl font-bold text-green-400">PHISHLETS</h2>
          <p className="text-green-700">Manage phishing templates and configurations</p>
        </div>
        <button className="flex items-center space-x-2 px-4 py-2 bg-green-600 hover:bg-green-700 rounded transition">
          <Plus className="w-4 h-4" />
          <span>New Phishlet</span>
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
          <div className="text-2xl font-bold text-green-400">{phishlets.length}</div>
          <div className="text-xs text-green-600 uppercase">Total Phishlets</div>
        </div>
        <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
          <div className="text-2xl font-bold text-green-400">
            {phishlets.filter(p => p.status === 'ACTIVE').length}
          </div>
          <div className="text-xs text-green-600 uppercase">Active</div>
        </div>
        <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
          <div className="text-2xl font-bold text-green-400">
            {phishlets.filter(p => p.status === 'INACTIVE').length}
          </div>
          <div className="text-xs text-green-600 uppercase">Inactive</div>
        </div>
      </div>

      {/* Phishlets Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {phishlets.map((phishlet) => (
          <div
            key={phishlet.id}
            className="bg-green-900/10 border border-green-800 rounded-lg p-6 hover:border-green-600 transition"
          >
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center space-x-3">
                <div className="p-2 bg-green-900/30 rounded">
                  <FileText className="w-6 h-6 text-green-400" />
                </div>
                <div>
                  <h3 className="font-bold text-green-400">{phishlet.name}</h3>
                  <p className="text-xs text-green-600">v{phishlet.version}</p>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                {phishlet.status === 'ACTIVE' ? (
                  <CheckCircle className="w-5 h-5 text-green-500" />
                ) : (
                  <XCircle className="w-5 h-5 text-red-500" />
                )}
              </div>
            </div>

            <div className="space-y-2 mb-4">
              <div className="text-xs text-green-600">Author: <span className="text-green-400">{phishlet.author}</span></div>
              <div className="text-xs text-green-600">Features: <span className="text-green-400">{phishlet.features}</span></div>
            </div>

            <div className="flex items-center space-x-2">
              <button
                onClick={() => togglePhishlet(phishlet.id, phishlet.status)}
                className={`flex-1 px-3 py-2 rounded text-sm transition ${
                  phishlet.status === 'ACTIVE'
                    ? 'bg-red-900/30 border border-red-700 text-red-400 hover:bg-red-900/50'
                    : 'bg-green-600 text-white hover:bg-green-700'
                }`}
              >
                {phishlet.status === 'ACTIVE' ? 'Disable' : 'Enable'}
              </button>
              <button className="p-2 bg-green-900/30 border border-green-700 rounded hover:bg-green-900/50 transition">
                <Download className="w-4 h-4 text-green-400" />
              </button>
              <button className="p-2 bg-red-900/30 border border-red-700 rounded hover:bg-red-900/50 transition">
                <Trash2 className="w-4 h-4 text-red-400" />
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Available Templates */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4">Available Templates</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {[
            'Microsoft 365',
            'Google Workspace',
            'Сбербанк Бизнес',
            'Тинькофф Бизнес',
            'Госуслуги',
            'Office 365',
            'GitHub',
            'AWS Console'
          ].map((template) => (
            <div key={template} className="bg-green-900/20 rounded p-3 text-center">
              <div className="text-sm text-green-400">{template}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
