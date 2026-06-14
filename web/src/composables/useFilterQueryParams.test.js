import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref, nextTick } from 'vue'
import { useFilterQueryParams } from './useFilterQueryParams'

const replace = vi.fn()
let routeQuery = {}

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: routeQuery }),
  useRouter: () => ({ replace }),
}))

describe('useFilterQueryParams', () => {
  beforeEach(() => {
    routeQuery = {}
    replace.mockClear()
  })

  it('reads namespace and dataflow from URL on init', () => {
    routeQuery = { namespace: 'prod', dataflow: 'sync-job' }
    const { namespace, dataflow } = useFilterQueryParams({ dataflow: true })
    expect(namespace.value).toBe('prod')
    expect(dataflow.value).toBe('sync-job')
  })

  it('defaults namespace to default when query is empty', () => {
    const { namespace, dataflow } = useFilterQueryParams({ dataflow: true })
    expect(namespace.value).toBe('default')
    expect(dataflow.value).toBe('')
  })

  it('writes namespace to URL when changed', async () => {
    const { namespace } = useFilterQueryParams()
    namespace.value = 'staging'
    await nextTick()
    expect(replace).toHaveBeenCalledWith({ query: { namespace: 'staging' } })
  })

  it('writes dataflow to URL when changed', async () => {
    const { namespace, dataflow } = useFilterQueryParams({ dataflow: true })
    namespace.value = 'prod'
    dataflow.value = 'orders-etl'
    await nextTick()
    expect(replace).toHaveBeenCalledWith({
      query: { namespace: 'prod', dataflow: 'orders-etl' },
    })
  })

  it('removes dataflow from URL when cleared', async () => {
    routeQuery = { namespace: 'prod', dataflow: 'orders-etl' }
    const { dataflow } = useFilterQueryParams({ dataflow: true })
    dataflow.value = ''
    await nextTick()
    expect(replace).toHaveBeenCalledWith({ query: { namespace: 'prod' } })
  })

  it('ignores dataflow param when dataflow sync is disabled', () => {
    routeQuery = { namespace: 'prod', dataflow: 'orders-etl' }
    const { dataflow } = useFilterQueryParams()
    expect(dataflow.value).toBe('')
  })
})
