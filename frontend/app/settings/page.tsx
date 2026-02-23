'use client'

import { useState } from 'react'
import { Settings, Save, RotateCcw, Shield, Key, Database, Bell } from 'lucide-react'

export default function SettingsTab() {
  const [saving, setSaving] = useState(false)
  const [formData, setFormData] = useState({
    // Network
    bind_ip: '0.0.0.0',
    https_port: '8443',
    domain: 'phantom.local',
    
    // API
    api_enabled: true,
    api_port: '8080',
    api_key: '',
    
    // Security
    mtls_enabled: false,
    fstec_enabled: false,
    
    // Features
    multi_tenant_enabled: false,
    risk_score_enabled: true,
    vishing_enabled: false,
    
    // Logging
    debug: false,
    log_level: 'info'
  })

  const handleChange = (key: string, value: any) => {
    setFormData(prev => ({ ...prev, [key]: value }))
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await fetch('/api/v1/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      })
      alert('Settings saved successfully!')
    } catch (err) {
      console.error('Failed to save settings:', err)
      alert('Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  const handleReset = () => {
    if (confirm('Reset all settings to defaults?')) {
      // Reset logic here
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-green-400">SETTINGS</h2>
          <p className="text-green-700">Configure PhantomProxy system settings</p>
        </div>
        <div className="flex items-center space-x-3">
          <button
            onClick={handleReset}
            className="flex items-center space-x-2 px-4 py-2 bg-green-900/30 border border-green-700 rounded hover:bg-green-900/50 transition"
          >
            <RotateCcw className="w-4 h-4" />
            <span>Reset</span>
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center space-x-2 px-4 py-2 bg-green-600 hover:bg-green-700 rounded transition disabled:opacity-50"
          >
            <Save className="w-4 h-4" />
            <span>{saving ? 'Saving...' : 'Save Changes'}</span>
          </button>
        </div>
      </div>

      {/* Network Settings */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4 flex items-center">
          <Settings className="w-5 h-5 mr-2" />
          Network Settings
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-xs text-green-600 uppercase mb-1">Bind IP</label>
            <input
              type="text"
              value={formData.bind_ip}
              onChange={(e) => handleChange('bind_ip', e.target.value)}
              className="w-full px-3 py-2 bg-green-900/20 border border-green-800 rounded focus:outline-none focus:border-green-600 text-green-400"
            />
          </div>
          <div>
            <label className="block text-xs text-green-600 uppercase mb-1">HTTPS Port</label>
            <input
              type="number"
              value={formData.https_port}
              onChange={(e) => handleChange('https_port', e.target.value)}
              className="w-full px-3 py-2 bg-green-900/20 border border-green-800 rounded focus:outline-none focus:border-green-600 text-green-400"
            />
          </div>
          <div>
            <label className="block text-xs text-green-600 uppercase mb-1">Domain</label>
            <input
              type="text"
              value={formData.domain}
              onChange={(e) => handleChange('domain', e.target.value)}
              className="w-full px-3 py-2 bg-green-900/20 border border-green-800 rounded focus:outline-none focus:border-green-600 text-green-400"
            />
          </div>
        </div>
      </div>

      {/* API Settings */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4 flex items-center">
          <Key className="w-5 h-5 mr-2" />
          API Settings
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={formData.api_enabled}
              onChange={(e) => handleChange('api_enabled', e.target.checked)}
              className="w-4 h-4 bg-green-900/20 border border-green-800 rounded"
            />
            <label className="ml-2 text-sm text-green-400">Enable API</label>
          </div>
          <div>
            <label className="block text-xs text-green-600 uppercase mb-1">API Port</label>
            <input
              type="number"
              value={formData.api_port}
              onChange={(e) => handleChange('api_port', e.target.value)}
              className="w-full px-3 py-2 bg-green-900/20 border border-green-800 rounded focus:outline-none focus:border-green-600 text-green-400"
            />
          </div>
          <div>
            <label className="block text-xs text-green-600 uppercase mb-1">API Key</label>
            <input
              type="password"
              value={formData.api_key}
              onChange={(e) => handleChange('api_key', e.target.value)}
              placeholder="Enter API key"
              className="w-full px-3 py-2 bg-green-900/20 border border-green-800 rounded focus:outline-none focus:border-green-600 text-green-400"
            />
          </div>
        </div>
      </div>

      {/* Security Settings */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4 flex items-center">
          <Shield className="w-5 h-5 mr-2" />
          Security Settings
        </h3>
        <div className="space-y-3">
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={formData.mtls_enabled}
              onChange={(e) => handleChange('mtls_enabled', e.target.checked)}
              className="w-4 h-4 bg-green-900/20 border border-green-800 rounded"
            />
            <label className="ml-2 text-sm text-green-400">Enable Zero-Trust mTLS</label>
          </div>
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={formData.fstec_enabled}
              onChange={(e) => handleChange('fstec_enabled', e.target.checked)}
              className="w-4 h-4 bg-green-900/20 border border-green-800 rounded"
            />
            <label className="ml-2 text-sm text-green-400">Enable FSTEC Compliance (GOST Encryption)</label>
          </div>
        </div>
      </div>

      {/* Feature Toggles */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4 flex items-center">
          <Database className="w-5 h-5 mr-2" />
          Feature Toggles
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={formData.multi_tenant_enabled}
              onChange={(e) => handleChange('multi_tenant_enabled', e.target.checked)}
              className="w-4 h-4 bg-green-900/20 border border-green-800 rounded"
            />
            <label className="ml-2 text-sm text-green-400">Multi-Tenant Architecture</label>
          </div>
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={formData.risk_score_enabled}
              onChange={(e) => handleChange('risk_score_enabled', e.target.checked)}
              className="w-4 h-4 bg-green-900/20 border border-green-800 rounded"
            />
            <label className="ml-2 text-sm text-green-400">Risk Score Analysis</label>
          </div>
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={formData.vishing_enabled}
              onChange={(e) => handleChange('vishing_enabled', e.target.checked)}
              className="w-4 h-4 bg-green-900/20 border border-green-800 rounded"
            />
            <label className="ml-2 text-sm text-green-400">Vishing/Smishing</label>
          </div>
        </div>
      </div>

      {/* Logging Settings */}
      <div className="bg-green-900/10 border border-green-800 rounded-lg p-6">
        <h3 className="text-lg font-bold text-green-400 mb-4 flex items-center">
          <Bell className="w-5 h-5 mr-2" />
          Logging Settings
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="flex items-center">
            <input
              type="checkbox"
              checked={formData.debug}
              onChange={(e) => handleChange('debug', e.target.checked)}
              className="w-4 h-4 bg-green-900/20 border border-green-800 rounded"
            />
            <label className="ml-2 text-sm text-green-400">Debug Mode</label>
          </div>
          <div>
            <label className="block text-xs text-green-600 uppercase mb-1">Log Level</label>
            <select
              value={formData.log_level}
              onChange={(e) => handleChange('log_level', e.target.value)}
              className="w-full px-3 py-2 bg-green-900/20 border border-green-800 rounded focus:outline-none focus:border-green-600 text-green-400"
            >
              <option value="debug">DEBUG</option>
              <option value="info">INFO</option>
              <option value="warn">WARN</option>
              <option value="error">ERROR</option>
            </select>
          </div>
        </div>
      </div>
    </div>
  )
}
