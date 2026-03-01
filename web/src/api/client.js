const API_BASE = import.meta.env.VITE_API_BASE || '/api'

async function request(path, options = {}) {
  const url = path.startsWith('http') ? path : `${API_BASE}${path}`
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
  if (res.status === 204) return null
  return res.json()
}

async function requestText(path, options = {}) {
  const url = path.startsWith('http') ? path : `${API_BASE}${path}`
  const res = await fetch(url, options)
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
  return res.text()
}

export async function getNamespaces() {
  return request('/namespaces')
}

export async function listDataFlows(namespace = 'default') {
  return request(`/dataflows?namespace=${encodeURIComponent(namespace)}`)
}

export async function getDataFlow(namespace, name) {
  return request(
    `/dataflows/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`
  )
}

export async function createDataFlow(namespace, body) {
  return request(`/dataflows?namespace=${encodeURIComponent(namespace)}`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function updateDataFlow(namespace, name, body) {
  return request(
    `/dataflows/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`,
    {
      method: 'PUT',
      body: JSON.stringify(body),
    }
  )
}

export async function deleteDataFlow(namespace, name) {
  const url = `${API_BASE}/dataflows/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`
  const res = await fetch(url, { method: 'DELETE' })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
}

export async function getLogs(namespace, name, { tailLines = 100 } = {}) {
  const path = `/logs?namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}&tailLines=${tailLines}&follow=false`
  return requestText(path)
}

export function createLogStream(namespace, name, { tailLines = 100 } = {}, onMessage) {
  const path = `${API_BASE}/logs?namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}&tailLines=${tailLines}&follow=true`
  const es = new EventSource(path)
  es.onmessage = (e) => onMessage(e.data)
  return () => es.close()
}

export async function getStatus(namespace, name) {
  return request(
    `/status?namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`
  )
}

export async function getMetrics(namespace, name) {
  return requestText(
    `/metrics?namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`
  )
}
