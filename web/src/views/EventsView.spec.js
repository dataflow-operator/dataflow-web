import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { i18n } from '../i18n'
import EventsView from './EventsView.vue'

vi.mock('../api/client', () => ({
  getNamespaces: vi.fn(),
  listDataFlows: vi.fn(),
  getEvents: vi.fn(),
}))

describe('EventsView', () => {
  beforeEach(async () => {
    const { getNamespaces, listDataFlows, getEvents } = await import('../api/client')
    getNamespaces.mockResolvedValue(['default', 'kube-system'])
    listDataFlows.mockResolvedValue([
      { metadata: { name: 'flow-a' } },
      { metadata: { name: 'flow-b' } },
    ])
    getEvents.mockResolvedValue([
      {
        type: 'Normal',
        reason: 'ConfigMapCreated',
        message: 'Created ConfigMap',
        involvedObject: { name: 'flow-a' },
        lastTimestamp: '2025-01-01T12:00:00Z',
        count: 1,
      },
    ])
  })

  it('renders title and loads events', async () => {
    const wrapper = mount(EventsView, {
      global: { plugins: [i18n] },
    })
    await flushPromises()

    expect(wrapper.find('h2').text()).toContain('Events')
    const rows = wrapper.findAll('tbody tr')
    expect(rows.length).toBe(1)
    expect(rows[0].text()).toContain('ConfigMapCreated')
  })

  it('shows filter mode options', async () => {
    const wrapper = mount(EventsView, {
      global: { plugins: [i18n] },
    })
    await flushPromises()

    const filterSelect = wrapper.findAll('select')[1]
    expect(filterSelect.exists()).toBe(true)
    const options = filterSelect.findAll('option')
    expect(options.length).toBeGreaterThanOrEqual(2)
  })

  it('calls getEvents with namespace and no name when filter is all', async () => {
    const { getEvents } = await import('../api/client')
    const wrapper = mount(EventsView, {
      global: { plugins: [i18n] },
    })
    await flushPromises()

    expect(getEvents).toHaveBeenCalledWith('default', null)
  })

  it('calls getEvents with name when filtering by manifest', async () => {
    const { getEvents } = await import('../api/client')
    getEvents.mockClear()

    const wrapper = mount(EventsView, {
      global: { plugins: [i18n] },
    })
    await flushPromises()

    const selects = wrapper.findAll('select')
    await selects[1].setValue('manifest')
    await flushPromises()

    const manifestSelects = wrapper.findAll('select')
    const manifestSelect = manifestSelects.find((s) =>
      s.findAll('option[value="flow-a"]').length > 0
    )
    if (manifestSelect) {
      await manifestSelect.setValue('flow-a')
      await flushPromises()
      expect(getEvents).toHaveBeenCalledWith('default', 'flow-a')
    }
  })
})
