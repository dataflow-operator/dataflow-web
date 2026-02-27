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

  it('throws on non-ok response', async () => {
    fetch.mockResolvedValueOnce({ ok: false, text: () => Promise.resolve('Not found') })
    await expect(getNamespaces()).rejects.toThrow()
  })
})
