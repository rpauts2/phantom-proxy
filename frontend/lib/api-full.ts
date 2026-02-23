// Расширенный API клиент для Evingix
const API_BASE = '/api/v1'

// Stats
export async function getStats() {
  const res = await fetch(`${API_BASE}/stats`)
  return res.json()
}

// Sessions
export async function getSessions(limit = 50, offset = 0) {
  const res = await fetch(`${API_BASE}/sessions?limit=${limit}&offset=${offset}`)
  return res.json()
}

export async function deleteSession(id: string) {
  const res = await fetch(`${API_BASE}/sessions/${id}`, { method: 'DELETE' })
  return res.json()
}

// Credentials
export async function getCredentials(limit = 50, offset = 0) {
  const res = await fetch(`${API_BASE}/credentials?limit=${limit}&offset=${offset}`)
  return res.json()
}

// Phishlets
export async function getPhishlets() {
  const res = await fetch(`${API_BASE}/phishlets`)
  return res.json()
}

export async function enablePhishlet(id: string) {
  const res = await fetch(`${API_BASE}/phishlets/${id}/enable`, { method: 'POST' })
  return res.json()
}

export async function disablePhishlet(id: string) {
  const res = await fetch(`${API_BASE}/phishlets/${id}/disable`, { method: 'POST' })
  return res.json()
}

// Campaigns
export async function getCampaigns() {
  const res = await fetch(`${API_BASE}/campaigns`)
  return res.json()
}

export async function createCampaign(data: {
  name: string
  template: string
  page: string
  group: string
  smtp: string
  url: string
}) {
  const res = await fetch(`${API_BASE}/campaigns`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function startCampaign(id: string) {
  const res = await fetch(`${API_BASE}/campaigns/${id}/start`, { method: 'POST' })
  return res.json()
}

export async function pauseCampaign(id: string) {
  const res = await fetch(`${API_BASE}/campaigns/${id}/pause`, { method: 'POST' })
  return res.json()
}

export async function stopCampaign(id: string) {
  const res = await fetch(`${API_BASE}/campaigns/${id}/stop`, { method: 'POST' })
  return res.json()
}

export async function deleteCampaign(id: string) {
  const res = await fetch(`${API_BASE}/campaigns/${id}`, { method: 'DELETE' })
  return res.json()
}

export async function getCampaignStats(id: string) {
  const res = await fetch(`${API_BASE}/campaigns/${id}/stats`)
  return res.json()
}

// Groups
export async function getGroups() {
  const res = await fetch(`${API_BASE}/groups`)
  return res.json()
}

export async function createGroup(data: { name: string; targets: Array<{ email: string; first_name?: string; last_name?: string }> }) {
  const res = await fetch(`${API_BASE}/groups`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function deleteGroup(id: string) {
  const res = await fetch(`${API_BASE}/groups/${id}`, { method: 'DELETE' })
  return res.json()
}

// Templates
export async function getTemplates() {
  const res = await fetch(`${API_BASE}/templates`)
  return res.json()
}

export async function createTemplate(data: { name: string; subject: string; text: string; html: string }) {
  const res = await fetch(`${API_BASE}/templates`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

// Landing Pages
export async function getPages() {
  const res = await fetch(`${API_BASE}/pages`)
  return res.json()
}

export async function createPage(data: { name: string; html: string }) {
  const res = await fetch(`${API_BASE}/pages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

// SMTP Profiles
export async function getSMTPProfiles() {
  const res = await fetch(`${API_BASE}/smtp`)
  return res.json()
}

export async function createSMTPProfile(data: {
  name: string
  host: string
  port: number
  username: string
  password: string
  from: string
  use_tls: boolean
}) {
  const res = await fetch(`${API_BASE}/smtp`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

// AI
export async function generatePhishlet(data: { target_url: string; target_name: string }) {
  const res = await fetch(`${API_BASE}/ai/generate-phishlet`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function analyzeTarget(url: string) {
  const res = await fetch(`${API_BASE}/ai/analyze/${encodeURIComponent(url)}`)
  return res.json()
}

// C2
export async function getC2Adapters() {
  const res = await fetch(`${API_BASE}/c2/adapters`)
  return res.json()
}

export async function configureC2Adapter(name: string, data: { server_url: string; operator_token?: string }) {
  const res = await fetch(`${API_BASE}/c2/adapters/${name}/configure`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function toggleC2Adapter(name: string, enabled: boolean) {
  const res = await fetch(`${API_BASE}/c2/adapters/${name}/toggle`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ enabled }),
  })
  return res.json()
}

// Vishing
export async function makeVishingCall(data: { phone_number: string; voice_profile: string; scenario: string }) {
  const res = await fetch(`${API_BASE}/vishing/call`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function getVishingCallStatus(id: string) {
  const res = await fetch(`${API_BASE}/vishing/call/${id}`)
  return res.json()
}

export async function generateVishingScenario(target_service: string, goal: string) {
  const res = await fetch(`${API_BASE}/vishing/generate-scenario`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ target_service, goal }),
  })
  return res.json()
}

// Domains
export async function getDomains() {
  const res = await fetch(`${API_BASE}/domains`)
  return res.json()
}

export async function registerDomain(data: { domain: string; provider: string; auto_ssl: boolean }) {
  const res = await fetch(`${API_BASE}/domains/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function rotateDomain() {
  const res = await fetch(`${API_BASE}/domains/rotate`, { method: 'POST' })
  return res.json()
}

// Risk
export async function getRiskDistribution() {
  const res = await fetch(`${API_BASE}/risk/distribution`)
  return res.json()
}

export async function getHighRiskUsers(limit = 20) {
  const res = await fetch(`${API_BASE}/risk/high-risk?limit=${limit}`)
  return res.json()
}

// Logs
export async function getLogs(limit = 100, level?: string) {
  const params = new URLSearchParams({ limit: String(limit) })
  if (level) params.set('level', level)
  const res = await fetch(`${API_BASE}/logs?${params}`)
  return res.json()
}

// System
export async function getSystemStatus() {
  const res = await fetch(`${API_BASE}/system/status`)
  return res.json()
}

export async function getSystemConfig() {
  const res = await fetch(`${API_BASE}/system/config`)
  return res.json()
}

export async function updateSystemConfig(data: any) {
  const res = await fetch(`${API_BASE}/system/config`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return res.json()
}

// GoPhish Integration
export async function getGoPhishSummary() {
  const res = await fetch(`${API_BASE}/gophish/summary`)
  return res.json()
}

export async function getGoPhishCampaigns() {
  const res = await fetch(`${API_BASE}/gophish/campaigns`)
  return res.json()
}

export async function getGoPhishGroups() {
  const res = await fetch(`${API_BASE}/gophish/groups`)
  return res.json()
}

export async function getGoPhishTemplates() {
  const res = await fetch(`${API_BASE}/gophish/templates`)
  return res.json()
}

export async function getGoPhishPages() {
  const res = await fetch(`${API_BASE}/gophish/pages`)
  return res.json()
}

export async function getGoPhishProfiles() {
  const res = await fetch(`${API_BASE}/gophish/profiles`)
  return res.json()
}
