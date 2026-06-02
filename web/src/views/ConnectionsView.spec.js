import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import ConnectionsView from './ConnectionsView.vue'
import * as client from '../api/client'

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
}))

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: {
    en: {
      connections: {
        title: 'Connections',
        createNew: 'Create connection',
        loading: 'Loading...',
        empty: 'No connections',
        emptyHint: 'Create a new connection or select another namespace.',
        searchPlaceholder: 'Search...',
        name: 'Name',
        type: 'Type',
        keys: 'Keys',
        actions: 'Actions',
        edit: 'Edit',
        editTitle: 'Edit: {name}',
        createTitle: 'Create connection',
        connectionTypes: { kafka: 'Kafka', postgresql: 'PostgreSQL' },
        deleteTitle: 'Delete',
        deleteConfirm: 'Delete "{name}"?',
      },
      common: { namespace: 'Namespace', delete: 'Delete' },
    },
  },
})

describe('ConnectionsView', () => {
  beforeEach(() => {
    vi.spyOn(client, 'listSecrets').mockResolvedValue([])
  })

  it('renders and loads connections', async () => {
    vi.mocked(client.listSecrets).mockResolvedValue([
      {
        metadata: { name: 'kafka-creds', namespace: 'default' },
        keys: ['brokers', 'topic'],
      },
    ])
    const wrapper = mount(ConnectionsView, {
      global: {
        plugins: [i18n],
        stubs: {
          NamespaceSelect: true,
          ConnectionFormModal: true,
          ConfirmModal: true,
          LoadingSpinner: true,
        },
      },
    })
    await flushPromises()
    expect(client.listSecrets).toHaveBeenCalledWith('default')
    expect(wrapper.text()).toContain('Connections')
  })

  it('shows create button', () => {
    const wrapper = mount(ConnectionsView, {
      global: {
        plugins: [i18n],
        stubs: {
          NamespaceSelect: true,
          ConnectionFormModal: true,
          ConfirmModal: true,
          LoadingSpinner: true,
        },
      },
    })
    expect(wrapper.text()).toContain('Create connection')
  })
})
