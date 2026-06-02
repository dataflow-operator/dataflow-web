import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  getNamespaces,
  listDataFlows,
  getDataFlow,
  createDataFlow,
  updateDataFlow,
  deleteDataFlow,
  getLogs,
  getStatus,
  getRuntime,
  getPrometheusRange,
  getPrometheusInstant,
  getEvents,
  listSecrets,
  getSecret,
  createSecret,
  updateSecret,
  deleteSecret,
} from './client'

describe('API client', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  it('getNamespaces calls fetch with /api/namespaces', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(['default', 'kube-system']) })
    const result = await getNamespaces()
    expect(fetch).toHaveBeenCalledWith(expect.stringContaining('/namespaces'), expect.any(Object))
    expect(result).toEqual(['default', 'kube-system'])
  })

  it('listDataFlows adds namespace query', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) })
    await listDataFlows('myns')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\?namespace=myns/),
      expect.any(Object)
    )
  })

  it('getDataFlow builds path with name and namespace', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ metadata: { name: 'df1' } }) })
    const result = await getDataFlow('default', 'df1')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/dataflows/df1'),
      expect.any(Object)
    )
    expect(result.metadata.name).toBe('df1')
  })

  it('createDataFlow sends POST with body', async () => {
    const body = { metadata: { name: 'new' }, spec: {} }
    fetch.mockResolvedValueOnce({ ok: true, status: 201, json: () => Promise.resolve(body) })
    await createDataFlow('default', body)
    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify(body),
      })
    )
  })

  it('updateDataFlow sends PUT with body', async () => {
    const body = { spec: {} }
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) })
    await updateDataFlow('default', 'df1', body)
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/dataflows/df1'),
      expect.objectContaining({ method: 'PUT' })
    )
  })

  it('deleteDataFlow sends DELETE', async () => {
    fetch.mockResolvedValueOnce({ ok: true, status: 204 })
    await deleteDataFlow('default', 'df1')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/dataflows/df1'),
      expect.objectContaining({ method: 'DELETE' })
    )
  })

  it('getLogs returns text', async () => {
    fetch.mockResolvedValueOnce({ ok: true, text: () => Promise.resolve('line1\nline2') })
    const result = await getLogs('default', 'df1', { tailLines: 50 })
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(/tailLines=50/),
      expect.any(Object)
    )
    expect(result).toBe('line1\nline2')
  })

  it('getStatus returns status object', async () => {
    fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ phase: 'Running', processedCount: 10 }),
    })
    const result = await getStatus('default', 'df1')
    expect(result.phase).toBe('Running')
    expect(result.processedCount).toBe(10)
  })

  it('getRuntime calls /api/runtime', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ checkpointConfigMap: 'df-x-checkpoint' }) })
    const result = await getRuntime('default', 'df1')
    expect(fetch).toHaveBeenCalledWith(expect.stringContaining('/runtime?'), expect.any(Object))
    expect(result.checkpointConfigMap).toBe('df-x-checkpoint')
  })

  it('getPrometheusRange calls /api/prometheus/range', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ mode: 'instant', series: [] }) })
    await getPrometheusRange('default', 'df1', 'throughput', { start: 1, end: 2, step: 15 })
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/prometheus\/range\?.*panel=throughput/),
      expect.any(Object)
    )
  })

  it('getPrometheusInstant calls /api/prometheus/instant', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ mode: 'instant', series: [] }) })
    await getPrometheusInstant('default', 'df1', 'queue_size')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/prometheus\/instant\?.*panel=queue_size/),
      expect.any(Object)
    )
  })

  it('getEvents returns all events when name is null', async () => {
    const events = [{ type: 'Normal', reason: 'ConfigMapCreated', message: 'Created ConfigMap' }]
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(events) })
    const result = await getEvents('default')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\?namespace=default$/),
      expect.any(Object)
    )
    expect(result).toEqual(events)
  })

  it('getEvents adds name param when filtering by manifest', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) })
    await getEvents('default', 'my-dataflow')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(/name=my-dataflow/),
      expect.any(Object)
    )
  })

  it('throws on non-ok response', async () => {
    fetch.mockResolvedValueOnce({ ok: false, text: () => Promise.resolve('Not found') })
    await expect(getNamespaces()).rejects.toThrow()
  })

  it('listSecrets adds namespace query', async () => {
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) })
    await listSecrets('myns')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\?namespace=myns/),
      expect.any(Object)
    )
  })

  it('getSecret builds path with name and namespace', async () => {
    fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ metadata: { name: 'kafka-creds' }, stringData: {} }),
    })
    const result = await getSecret('default', 'kafka-creds')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/secrets/kafka-creds'),
      expect.any(Object)
    )
    expect(result.metadata.name).toBe('kafka-creds')
  })

  it('createSecret sends POST with body', async () => {
    const body = { metadata: { name: 'new-secret' }, type: 'Opaque', stringData: { key: 'val' } }
    fetch.mockResolvedValueOnce({ ok: true, status: 201, json: () => Promise.resolve(body) })
    await createSecret('default', body)
    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify(body),
      })
    )
  })

  it('updateSecret sends PUT with body', async () => {
    const body = { stringData: { key: 'updated' } }
    fetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(body) })
    await updateSecret('default', 'my-secret', body)
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/secrets/my-secret'),
      expect.objectContaining({ method: 'PUT' })
    )
  })

  it('deleteSecret sends DELETE', async () => {
    fetch.mockResolvedValueOnce({ ok: true, status: 204 })
    await deleteSecret('default', 'my-secret')
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/secrets/my-secret'),
      expect.objectContaining({ method: 'DELETE' })
    )
  })
})
