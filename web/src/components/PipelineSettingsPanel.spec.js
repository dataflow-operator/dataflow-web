import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import PipelineSettingsPanel from './PipelineSettingsPanel.vue'
import { createDefaultPipelineSettings } from '../composables/usePipelineSettings'

function createI18nInstance() {
  return createI18n({
    legacy: false,
    locale: 'en',
    messages: {
      en: {
        modal: { closeAria: 'Close' },
        flow: {
          pipelineSettingsTitle: 'Pipeline settings',
          pipelineSettingsCheckpoint: 'Checkpoint',
          pipelineSettingsCheckpointPersistence: 'Checkpoint persistence',
          pipelineSettingsCheckpointPersistenceHint: 'Hint',
          pipelineSettingsCheckpointSyncOnAck: 'Sync checkpoint on ack',
          pipelineSettingsCheckpointSyncOnAckHint: 'Hint',
          pipelineSettingsCheckpointSaveInterval: 'Checkpoint save interval',
          pipelineSettingsCheckpointSaveIntervalPlaceholder: '30s',
          pipelineSettingsCheckpointSaveIntervalHint: 'Hint',
          pipelineSettingsInvalidDuration: 'Invalid duration',
          pipelineSettingsCheckpointReset: 'Reset checkpoint',
          pipelineSettingsCheckpointResetHint: 'Warning',
          pipelineSettingsProcessing: 'Processing',
          pipelineSettingsStrictIdempotency: 'Strict idempotency',
          pipelineSettingsStrictIdempotencyHint: 'Hint',
          pipelineSettingsAckGranularity: 'Ack granularity',
          pipelineSettingsAckGranularity_batch: 'Batch',
          pipelineSettingsAckGranularity_message: 'Message',
          pipelineSettingsAckGranularityHint: 'Hint',
          pipelineSettingsChannelBufferSize: 'Channel buffer size',
          pipelineSettingsChannelBufferSizePlaceholder: '100',
          pipelineSettingsChannelBufferSizeHint: 'Hint',
          pipelineSettingsScaling: 'Scaling',
          pipelineSettingsReplicas: 'Replicas',
          pipelineSettingsReplicasHint: 'Hint',
          pipelineSettingsReplicasKafkaOnly: 'Kafka only',
          pipelineSettingsErrorSink: 'Error sink',
          pipelineSettingsErrorSinkEnabled: 'Enable error sink',
          pipelineSettingsErrorSinkEnabledHint: 'Route failed messages to a dedicated sink.',
          pipelineSettingsErrorSinkAckPolicy: 'Ack policy',
          pipelineSettingsErrorSinkAckPolicy_afterWrite: 'After write',
          pipelineSettingsErrorSinkAckPolicy_never: 'Never',
          pipelineSettingsErrorSinkAckPolicy_afterMainSinkSuccess: 'After main sink success',
          pipelineSettingsErrorSinkAckPolicyHint: 'Ack policy hint',
          connectorType: 'Connector type',
        },
      },
    },
  })
}

describe('PipelineSettingsPanel', () => {
  let i18n

  beforeEach(() => {
    i18n = createI18nInstance()
  })

  it('renders checkpoint and processing fields', () => {
    const wrapper = mount(PipelineSettingsPanel, {
      props: {
        modelValue: createDefaultPipelineSettings(),
        sourceType: 'kafka',
      },
      global: { plugins: [i18n] },
    })
    expect(wrapper.find('#checkpoint-save-interval').exists()).toBe(true)
    expect(wrapper.find('#ack-granularity').exists()).toBe(true)
    expect(wrapper.find('#replicas').exists()).toBe(true)
    expect(wrapper.find('#channel-buffer-size').exists()).toBe(true)
  })

  it('emits update:modelValue when checkpointSyncOnAck is toggled', async () => {
    const wrapper = mount(PipelineSettingsPanel, {
      props: {
        modelValue: createDefaultPipelineSettings(),
        sourceType: 'kafka',
      },
      global: { plugins: [i18n] },
    })
    const checkboxes = wrapper.findAll('input[type="checkbox"]')
    const syncCheckbox = checkboxes[1]
    await syncCheckbox.setValue(true)
    const emitted = wrapper.emitted('update:modelValue')
    expect(emitted).toBeTruthy()
    const last = emitted[emitted.length - 1][0]
    expect(last.checkpointSyncOnAck).toBe(true)
  })

  it('disables replicas input for non-kafka source', () => {
    const wrapper = mount(PipelineSettingsPanel, {
      props: {
        modelValue: createDefaultPipelineSettings(),
        sourceType: 'postgresql',
      },
      global: { plugins: [i18n] },
    })
    expect(wrapper.find('#replicas').attributes('disabled')).toBeDefined()
  })

  it('shows duration validation error for invalid interval', async () => {
    const wrapper = mount(PipelineSettingsPanel, {
      props: {
        modelValue: { ...createDefaultPipelineSettings(), checkpointSaveInterval: 'bad' },
        sourceType: 'kafka',
      },
      global: { plugins: [i18n] },
    })
    expect(wrapper.find('.field-error').exists()).toBe(true)
  })

  it('emits close when close button clicked', async () => {
    const wrapper = mount(PipelineSettingsPanel, {
      props: {
        modelValue: createDefaultPipelineSettings(),
        sourceType: 'kafka',
      },
      global: { plugins: [i18n] },
    })
    await wrapper.find('.panel-close').trigger('click')
    expect(wrapper.emitted('close')).toBeTruthy()
  })
})
