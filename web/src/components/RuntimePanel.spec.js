import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import RuntimePanel from './RuntimePanel.vue'
import * as client from '../api/client'

vi.mock('../api/client', () => ({
  getDataFlow: vi.fn(),
  updateDataFlow: vi.fn(),
}))

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: {
    en: {
      common: { enabled: 'Enabled', disabled: 'Disabled' },
      metrics: {
        runtimeTitle: 'Runtime and checkpoint',
        loadingRuntime: 'Loading runtime...',
        noRuntime: 'No runtime data',
        checkpointPersistence: 'Checkpoint persistence',
        checkpointConfigMap: 'Checkpoint ConfigMap',
        processorReady: 'Processor ready',
        checkpoints: 'Checkpoint (raw JSON)',
        processorPods: 'Processor pods',
        podPhase: 'Phase',
        podReady: 'Ready',
        podRestarts: 'Restarts',
        podNode: 'Node',
        conditions: 'Conditions',
        resetCheckpoint: 'Reset checkpoint',
        resetCheckpointConfirmTitle: 'Reset checkpoint?',
        resetCheckpointConfirmMessage: 'Confirm reset message',
        resetCheckpointSuccess: 'Checkpoint reset requested',
        resettingCheckpoint: 'Requesting checkpoint reset...',
      },
    },
  },
})

const runtimeFixture = {
  checkpointPersistence: true,
  checkpointConfigMap: 'df-test-checkpoint',
  processor: { readyReplicas: 1, replicas: 1, pods: [] },
}

describe('RuntimePanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows reset checkpoint button when namespace and dataflow are set', () => {
    const wrapper = mount(RuntimePanel, {
      props: {
        runtime: runtimeFixture,
        namespace: 'default',
        dataflowName: 'test-flow',
      },
      global: {
        plugins: [i18n],
        stubs: { LoadingSpinner: true, ConfirmModal: true },
      },
    })
    expect(wrapper.text()).toContain('Reset checkpoint')
  })

  it('hides reset checkpoint button without dataflow name', () => {
    const wrapper = mount(RuntimePanel, {
      props: {
        runtime: runtimeFixture,
        namespace: 'default',
        dataflowName: '',
      },
      global: {
        plugins: [i18n],
        stubs: { LoadingSpinner: true, ConfirmModal: true },
      },
    })
    expect(wrapper.text()).not.toContain('Reset checkpoint')
  })

  it('calls updateDataFlow with checkpointReset on confirm', async () => {
    const current = {
      metadata: { name: 'test-flow', namespace: 'default', resourceVersion: '99', uid: 'abc' },
      spec: { source: { type: 'kafka' }, sink: { type: 'postgresql' } },
    }
    vi.mocked(client.getDataFlow).mockResolvedValue(current)
    vi.mocked(client.updateDataFlow).mockResolvedValue({})

    const wrapper = mount(RuntimePanel, {
      props: {
        runtime: runtimeFixture,
        namespace: 'default',
        dataflowName: 'test-flow',
      },
      global: {
        plugins: [i18n],
        stubs: { LoadingSpinner: true, ConfirmModal: true },
      },
    })

    await wrapper.vm.onConfirmReset()
    await flushPromises()

    expect(client.getDataFlow).toHaveBeenCalledWith('default', 'test-flow')
    expect(client.updateDataFlow).toHaveBeenCalledWith(
      'default',
      'test-flow',
      expect.objectContaining({
        metadata: expect.objectContaining({ resourceVersion: '99', uid: 'abc' }),
        spec: expect.objectContaining({ checkpointReset: true }),
      })
    )
    expect(wrapper.emitted('checkpoint-reset')).toBeTruthy()
    expect(wrapper.text()).toContain('Checkpoint reset requested')
  })

  it('shows error when updateDataFlow fails', async () => {
    vi.mocked(client.getDataFlow).mockRejectedValue(new Error('API error'))

    const wrapper = mount(RuntimePanel, {
      props: {
        runtime: runtimeFixture,
        namespace: 'default',
        dataflowName: 'test-flow',
      },
      global: {
        plugins: [i18n],
        stubs: { LoadingSpinner: true, ConfirmModal: true },
      },
    })

    await wrapper.vm.onConfirmReset()
    await flushPromises()

    expect(wrapper.text()).toContain('API error')
    expect(wrapper.emitted('checkpoint-reset')).toBeFalsy()
  })
})
