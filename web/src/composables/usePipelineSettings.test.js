import { describe, it, expect } from 'vitest'
import {
  createDefaultPipelineSettings,
  specToPipelineSettings,
  applyPipelineSettingsToSpec,
  formatDurationForInput,
  isValidDuration,
  isReplicasSupported,
  ACK_GRANULARITY_BATCH,
  ACK_GRANULARITY_MESSAGE,
} from './usePipelineSettings'
import { graphToManifest, manifestToGraph, createEmptyGraph } from './useFlowManifest'

describe('usePipelineSettings', () => {
  it('createDefaultPipelineSettings returns expected defaults', () => {
    const defaults = createDefaultPipelineSettings()
    expect(defaults.checkpointPersistence).toBe(true)
    expect(defaults.checkpointSyncOnAck).toBe(false)
    expect(defaults.checkpointSaveInterval).toBe('')
    expect(defaults.checkpointReset).toBe(false)
    expect(defaults.strictIdempotency).toBe(false)
    expect(defaults.ackGranularity).toBe(ACK_GRANULARITY_BATCH)
    expect(defaults.replicas).toBe(1)
    expect(defaults.channelBufferSize).toBe('')
  })

  it('specToPipelineSettings reads spec fields', () => {
    const settings = specToPipelineSettings({
      checkpointPersistence: false,
      checkpointSyncOnAck: true,
      checkpointSaveInterval: '5s',
      checkpointReset: true,
      strictIdempotency: true,
      ackGranularity: ACK_GRANULARITY_MESSAGE,
      replicas: 3,
      channelBufferSize: 500,
    })
    expect(settings.checkpointPersistence).toBe(false)
    expect(settings.checkpointSyncOnAck).toBe(true)
    expect(settings.checkpointSaveInterval).toBe('5s')
    expect(settings.checkpointReset).toBe(true)
    expect(settings.strictIdempotency).toBe(true)
    expect(settings.ackGranularity).toBe(ACK_GRANULARITY_MESSAGE)
    expect(settings.replicas).toBe(3)
    expect(settings.channelBufferSize).toBe(500)
  })

  it('formatDurationForInput handles string and object forms', () => {
    expect(formatDurationForInput('30s')).toBe('30s')
    expect(formatDurationForInput({ duration: '1m' })).toBe('1m')
    expect(formatDurationForInput(null)).toBe('')
  })

  it('isValidDuration validates Kubernetes duration format', () => {
    expect(isValidDuration('30s')).toBe(true)
    expect(isValidDuration('5m')).toBe(true)
    expect(isValidDuration('1h')).toBe(true)
    expect(isValidDuration('500ms')).toBe(true)
    expect(isValidDuration('invalid')).toBe(false)
    expect(isValidDuration('')).toBe(false)
  })

  it('applyPipelineSettingsToSpec omits default values', () => {
    const spec = applyPipelineSettingsToSpec(
      { source: { type: 'kafka' }, sink: { type: 'postgresql' } },
      createDefaultPipelineSettings(),
      { sourceType: 'kafka' }
    )
    expect(spec.checkpointPersistence).toBeUndefined()
    expect(spec.checkpointSyncOnAck).toBeUndefined()
    expect(spec.ackGranularity).toBeUndefined()
    expect(spec.replicas).toBeUndefined()
  })

  it('applyPipelineSettingsToSpec writes non-default settings', () => {
    const settings = {
      ...createDefaultPipelineSettings(),
      checkpointPersistence: false,
      checkpointSyncOnAck: true,
      checkpointSaveInterval: '5s',
      checkpointReset: true,
      strictIdempotency: true,
      ackGranularity: ACK_GRANULARITY_MESSAGE,
      replicas: 2,
      channelBufferSize: 1000,
    }
    const spec = applyPipelineSettingsToSpec(
      { source: { type: 'kafka' }, sink: { type: 'postgresql' } },
      settings,
      { sourceType: 'kafka' }
    )
    expect(spec.checkpointPersistence).toBe(false)
    expect(spec.checkpointSyncOnAck).toBe(true)
    expect(spec.checkpointSaveInterval).toBe('5s')
    expect(spec.checkpointReset).toBe(true)
    expect(spec.strictIdempotency).toBe(true)
    expect(spec.ackGranularity).toBe(ACK_GRANULARITY_MESSAGE)
    expect(spec.replicas).toBe(2)
    expect(spec.channelBufferSize).toBe(1000)
  })

  it('applyPipelineSettingsToSpec removes replicas for non-kafka source', () => {
    const settings = { ...createDefaultPipelineSettings(), replicas: 3 }
    const spec = applyPipelineSettingsToSpec(
      { source: { type: 'postgresql' }, sink: { type: 'postgresql' }, replicas: 3 },
      settings,
      { sourceType: 'postgresql' }
    )
    expect(spec.replicas).toBeUndefined()
  })

  it('applyPipelineSettingsToSpec ignores invalid checkpointSaveInterval', () => {
    const settings = {
      ...createDefaultPipelineSettings(),
      checkpointSaveInterval: 'not-a-duration',
    }
    const spec = applyPipelineSettingsToSpec({}, settings)
    expect(spec.checkpointSaveInterval).toBeUndefined()
  })

  it('isReplicasSupported returns true only for kafka', () => {
    expect(isReplicasSupported('kafka')).toBe(true)
    expect(isReplicasSupported('postgresql')).toBe(false)
  })

  it('graphToManifest round-trips pipeline settings via baseManifest spec', () => {
    const baseSpec = applyPipelineSettingsToSpec(
      {
        source: { type: 'kafka', config: { topic: 'in' } },
        sink: { type: 'postgresql', config: { table: 'out' } },
      },
      {
        ...createDefaultPipelineSettings(),
        checkpointSyncOnAck: true,
        ackGranularity: ACK_GRANULARITY_MESSAGE,
        channelBufferSize: 500,
      },
      { sourceType: 'kafka' }
    )
    const manifest = {
      metadata: { name: 'test', namespace: 'default' },
      spec: baseSpec,
    }
    const { nodes, edges } = manifestToGraph(manifest)
    const result = graphToManifest(nodes, edges, manifest)
    expect(result.spec.checkpointSyncOnAck).toBe(true)
    expect(result.spec.ackGranularity).toBe(ACK_GRANULARITY_MESSAGE)
    expect(result.spec.channelBufferSize).toBe(500)
  })
})
