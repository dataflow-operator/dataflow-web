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

export async function listDataFlowCrons(namespace = 'default') {
  return request(`/dataflowcrons?namespace=${encodeURIComponent(namespace)}`)
}

export async function getDataFlowCron(namespace, name) {
  return request(
    `/dataflowcrons/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`
  )
}

export async function createDataFlowCron(namespace, body) {
  return request(`/dataflowcrons?namespace=${encodeURIComponent(namespace)}`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function updateDataFlowCron(namespace, name, body) {
  return request(
    `/dataflowcrons/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`,
    {
      method: 'PUT',
      body: JSON.stringify(body),
    }
  )
}

export async function deleteDataFlowCron(namespace, name) {
  const url = `${API_BASE}/dataflowcrons/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`
  const res = await fetch(url, { method: 'DELETE' })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
}

export async function triggerDataFlowCron(namespace, name) {
  return request(
    `/dataflowcrons/${encodeURIComponent(name)}/trigger?namespace=${encodeURIComponent(namespace)}`,
    { method: 'POST' }
  )
}

export async function suspendDataFlowCron(namespace, name, suspend) {
  return request(
    `/dataflowcrons/${encodeURIComponent(name)}/suspend?namespace=${encodeURIComponent(namespace)}`,
    {
      method: 'POST',
      body: JSON.stringify({ suspend }),
    }
  )
}

function logsQuery(namespace, name, { tailLines = 100, kind = '' } = {}) {
  const params = new URLSearchParams({
    namespace,
    name,
    tailLines: String(tailLines),
  })
  if (kind) params.set('kind', kind)
  return params.toString()
}

export async function getLogs(namespace, name, { tailLines = 100, kind = '' } = {}) {
  const path = `/logs?${logsQuery(namespace, name, { tailLines, kind })}&follow=false`
  return requestText(path)
}

export function createLogStream(namespace, name, { tailLines = 100, kind = '' } = {}, onMessage) {
  const path = `${API_BASE}/logs?${logsQuery(namespace, name, { tailLines, kind })}&follow=true`
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

export async function getRuntime(namespace, name) {
  return request(
    `/runtime?namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`
  )
}

export async function getPrometheusRange(namespace, name, panel, { start, end, step } = {}) {
  const params = new URLSearchParams({
    namespace,
    name,
    panel,
  })
  if (start != null) params.set('start', String(start))
  if (end != null) params.set('end', String(end))
  if (step != null) params.set('step', String(step))
  return request(`/prometheus/range?${params.toString()}`)
}

export async function getPrometheusInstant(namespace, name, panel) {
  const params = new URLSearchParams({ namespace, name, panel })
  return request(`/prometheus/instant?${params.toString()}`)
}

export async function getEvents(namespace, name = null) {
  const params = new URLSearchParams({ namespace })
  if (name) params.set('name', name)
  return request(`/events?${params.toString()}`)
}

export async function listSecrets(namespace = 'default') {
  return request(`/secrets?namespace=${encodeURIComponent(namespace)}`)
}

export async function getSecret(namespace, name) {
  return request(
    `/secrets/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`
  )
}

export async function createSecret(namespace, body) {
  return request(`/secrets?namespace=${encodeURIComponent(namespace)}`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function updateSecret(namespace, name, body) {
  return request(
    `/secrets/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`,
    {
      method: 'PUT',
      body: JSON.stringify(body),
    }
  )
}

export async function deleteSecret(namespace, name) {
  const url = `${API_BASE}/secrets/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`
  const res = await fetch(url, { method: 'DELETE' })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
}
