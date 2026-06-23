/**
 * Pipeline-level spec settings (fault tolerance, scaling, throughput).
 */

export const ACK_GRANULARITY_BATCH = 'batch'
export const ACK_GRANULARITY_MESSAGE = 'message'

export const ACK_GRANULARITY_OPTIONS = [ACK_GRANULARITY_BATCH, ACK_GRANULARITY_MESSAGE]

export const ERROR_ACK_POLICIES = ['afterWrite', 'never', 'afterMainSinkSuccess']

const DURATION_REGEX = /^(\d+(?:\.\d+)?)(ms|s|m|h)$/

/**
 * @returns {import('./usePipelineSettings').ErrorSinkSettings}
 */
export function createDefaultErrorSinkSettings() {
  return {
    enabled: false,
    connectorType: 'kafka',
    config: {},
    ackPolicy: 'afterWrite',
  }
}

/**
 * @returns {import('./usePipelineSettings').PipelineSettings}
 */
export function createDefaultPipelineSettings() {
  return {
    checkpointPersistence: true,
    checkpointSyncOnAck: false,
    checkpointSaveInterval: '',
    checkpointReset: false,
    strictIdempotency: false,
    ackGranularity: ACK_GRANULARITY_BATCH,
    replicas: 1,
    channelBufferSize: '',
    errorSink: createDefaultErrorSinkSettings(),
    maintenance: createDefaultMaintenanceSettings(),
  }
}

export function createDefaultMaintenanceSettings() {
  return {
    startTime: '',
    duration: '',
    repeat: '',
    timezone: '',
    description: '',
  }
}

/**
 * @param {*} value - Duration from manifest (string or metav1 object)
 * @returns {string}
 */
export function formatDurationForInput(value) {
  if (value == null || value === '') return ''
  if (typeof value === 'string') return value.trim()
  if (typeof value === 'object' && value.duration != null) {
    return String(value.duration).trim()
  }
  return String(value).trim()
}

/**
 * @param {string} value
 * @returns {boolean}
 */
export function isValidDuration(value) {
  if (!value || typeof value !== 'string') return false
  return DURATION_REGEX.test(value.trim())
}

/**
 * @typedef {Object} PipelineSettings
 * @property {boolean} checkpointPersistence
 * @property {boolean} checkpointSyncOnAck
 * @property {string} checkpointSaveInterval
 * @property {boolean} checkpointReset
 * @property {boolean} strictIdempotency
 * @property {string} ackGranularity
 * @property {number} replicas
 * @property {number|string} channelBufferSize
 * @property {ErrorSinkSettings} errorSink
 * @property {MaintenanceSettings} maintenance
 */

/**
 * @typedef {Object} MaintenanceSettings
 * @property {string} startTime
 * @property {string} duration
 * @property {string} repeat
 * @property {string} timezone
 * @property {string} description
 */

/**
 * @typedef {Object} ErrorSinkSettings
 * @property {boolean} enabled
 * @property {string} connectorType
 * @property {Object} config
 * @property {string} ackPolicy
 */

/**
 * @param {Object} [errors]
 * @returns {ErrorSinkSettings}
 */
export function specToErrorSinkSettings(errors) {
  if (!errors?.type) {
    return createDefaultErrorSinkSettings()
  }
  return {
    enabled: true,
    connectorType: errors.type,
    config: errors.config ? JSON.parse(JSON.stringify(errors.config)) : {},
    ackPolicy: errors.ackPolicy || 'afterWrite',
  }
}

/**
 * @param {ErrorSinkSettings} errorSink
 * @returns {Object|undefined}
 */
export function errorSinkSettingsToSpec(errorSink) {
  if (!errorSink?.enabled || !errorSink.connectorType) {
    return undefined
  }
  const errors = {
    type: errorSink.connectorType,
  }
  const config = errorSink.config || {}
  if (Object.keys(config).length > 0) {
    errors.config = JSON.parse(JSON.stringify(config))
  }
  if (errorSink.ackPolicy && errorSink.ackPolicy !== 'afterWrite') {
    errors.ackPolicy = errorSink.ackPolicy
  }
  return errors
}

/**
 * @param {Object} [spec]
 * @returns {PipelineSettings}
 */
export function specToPipelineSettings(spec = {}) {
  const defaults = createDefaultPipelineSettings()
  return {
    checkpointPersistence: spec.checkpointPersistence ?? defaults.checkpointPersistence,
    checkpointSyncOnAck: spec.checkpointSyncOnAck ?? defaults.checkpointSyncOnAck,
    checkpointSaveInterval: formatDurationForInput(spec.checkpointSaveInterval),
    checkpointReset: spec.checkpointReset ?? false,
    strictIdempotency: spec.strictIdempotency ?? defaults.strictIdempotency,
    ackGranularity: spec.ackGranularity || defaults.ackGranularity,
    replicas: spec.replicas ?? defaults.replicas,
    channelBufferSize: spec.channelBufferSize ?? '',
    errorSink: specToErrorSinkSettings(spec.errors),
    maintenance: specToMaintenanceSettings(spec.maintenance),
  }
}

function specToMaintenanceSettings(maintenance) {
  const defaults = createDefaultMaintenanceSettings()
  if (!maintenance) return defaults
  return {
    startTime: maintenance.startTime || '',
    duration: maintenance.duration || '',
    repeat: maintenance.repeat || '',
    timezone: maintenance.timezone || '',
    description: maintenance.description || '',
  }
}

/**
 * @param {Object} spec
 * @param {PipelineSettings} settings
 * @param {{ sourceType?: string }} [options]
 * @returns {Object}
 */
export function applyPipelineSettingsToSpec(spec, settings, options = {}) {
  const next = { ...spec }
  const defaults = createDefaultPipelineSettings()
  const sourceType = options.sourceType || spec.source?.type || ''

  if (settings.checkpointPersistence !== defaults.checkpointPersistence) {
    next.checkpointPersistence = settings.checkpointPersistence
  } else {
    delete next.checkpointPersistence
  }

  if (settings.checkpointSyncOnAck) {
    next.checkpointSyncOnAck = true
  } else {
    delete next.checkpointSyncOnAck
  }

  if (settings.checkpointReset) {
    next.checkpointReset = true
  } else {
    delete next.checkpointReset
  }

  if (settings.strictIdempotency) {
    next.strictIdempotency = true
  } else {
    delete next.strictIdempotency
  }

  const interval = formatDurationForInput(settings.checkpointSaveInterval)
  if (interval) {
    if (isValidDuration(interval)) {
      next.checkpointSaveInterval = interval
    }
  } else {
    delete next.checkpointSaveInterval
  }

  if (settings.ackGranularity && settings.ackGranularity !== defaults.ackGranularity) {
    next.ackGranularity = settings.ackGranularity
  } else {
    delete next.ackGranularity
  }

  const isKafka = sourceType === 'kafka'
  if (isKafka) {
    const replicas = Number(settings.replicas)
    if (!Number.isNaN(replicas) && replicas >= 0 && replicas !== defaults.replicas) {
      next.replicas = replicas
    } else {
      delete next.replicas
    }
  } else {
    delete next.replicas
  }

  const buffer = Number(settings.channelBufferSize)
  if (!Number.isNaN(buffer) && buffer > 0) {
    next.channelBufferSize = buffer
  } else {
    delete next.channelBufferSize
  }

  const errors = errorSinkSettingsToSpec(settings.errorSink)
  if (errors) {
    next.errors = errors
  } else {
    delete next.errors
  }

  const maintenance = maintenanceSettingsToSpec(settings.maintenance)
  if (maintenance) {
    next.maintenance = maintenance
  } else {
    delete next.maintenance
  }

  return next
}

function maintenanceSettingsToSpec(maintenance) {
  if (!maintenance) return undefined
  const startTime = (maintenance.startTime || '').trim()
  const duration = (maintenance.duration || '').trim()
  const repeat = (maintenance.repeat || '').trim()
  const timezone = (maintenance.timezone || '').trim()
  const description = (maintenance.description || '').trim()

  if (!startTime && !duration && !repeat && !timezone && !description) {
    return undefined
  }

  const out = {}
  if (startTime) out.startTime = startTime
  if (duration) out.duration = duration
  if (repeat) out.repeat = repeat
  if (timezone) out.timezone = timezone
  if (description) out.description = description
  return out
}

/**
 * @param {string} sourceType
 * @returns {boolean}
 */
export function isReplicasSupported(sourceType) {
  return sourceType === 'kafka'
}
