import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { i18n } from '../i18n'
import NamespaceSelect from './NamespaceSelect.vue'

vi.mock('../api/client', () => ({
  getNamespaces: vi.fn(),
}))

describe('NamespaceSelect', () => {
  beforeEach(async () => {
    const { getNamespaces } = await import('../api/client')
    getNamespaces.mockResolvedValue(['default', 'kube-system'])
  })

  it('renders select and loads namespaces', async () => {
    const wrapper = mount(NamespaceSelect, {
      props: { modelValue: 'default' },
      global: { plugins: [i18n] },
    })
    await flushPromises()
    const select = wrapper.find('select')
    expect(select.exists()).toBe(true)
    const options = wrapper.findAll('option')
    expect(options.length).toBeGreaterThanOrEqual(2)
  })

  it('emits update:modelValue on change', async () => {
    const wrapper = mount(NamespaceSelect, {
      props: { modelValue: 'default' },
      global: { plugins: [i18n] },
    })
    await flushPromises()
    await wrapper.find('select').setValue('kube-system')
    expect(wrapper.emitted('update:modelValue')).toEqual([['kube-system']])
  })
})
