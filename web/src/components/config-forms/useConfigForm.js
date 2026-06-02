/**
 * Composable for converting between structured form state and connector/transformation config.
 * Handles configToForm and formToConfig for all connector and transformation types.
 */

const CONNECTOR_KNOWN_KEYS = {
  kafka: ['brokers', 'topic', 'consumerGroup', 'format'],
  postgresql: ['connectionString', 'table', 'query', 'autoCreateTable', 'batchSize', 'rawMode', 'orderByColumn', 'changeTrackingColumn', 'readBatchSize', 'pollInterval'],
  trino: ['serverURL', 'catalog', 'schema', 'table', 'query', 'orderByColumn', 'changeTrackingColumn', 'readBatchSize', 'pollInterval'],
  clickhouse: ['connectionString', 'table', 'query', 'autoCreateTable', 'batchSize', 'rawMode', 'orderByColumn', 'changeTrackingColumn', 'readBatchSize', 'pollInterval'],
  nessie: ['baseURL', 'namespace', 'table', 'branch'],
}

/**
 * Check if config uses SecretRef (from Connection selection) - cannot fully populate structured form
 */
function hasSecretRefs(config) {
  if (!config || typeof config !== 'object') return false
  const str = JSON.stringify(config)
  return str.includes('SecretRef') || str.includes('secretRef')
}

/**
 * Extract keys not in known set into a separate object
 */
function extractExtraConfig(config, knownKeys) {
  if (!config || typeof config !== 'object') return {}
  const extra = {}
  for (const k of Object.keys(config)) {
    if (!knownKeys.includes(k)) {
      extra[k] = config[k]
    }
  }
  return extra
}

/**
 * Extract brokers from config - supports array or comma-separated string
 */
function brokersFromConfig(config) {
  const v = config?.brokers
  if (Array.isArray(v)) return v.join(', ')
  if (typeof v === 'string') return v
  return ''
}

/**
 * Build brokers for config - always as array
 */
function brokersToConfig(val) {
  if (!val?.trim()) return undefined
  return val.split(',').map((s) => s.trim()).filter(Boolean)
}

function sourceIncrementalFromConfig(c) {
  return {
    orderByColumn: c.orderByColumn || '',
    changeTrackingColumn: c.changeTrackingColumn || '',
    readBatchSize: c.readBatchSize ?? '',
    pollInterval: c.pollInterval ?? '',
  }
}

function sourceIncrementalToConfig(form, role) {
  if (role !== 'source') return {}
  return {
    ...(form?.orderByColumn ? { orderByColumn: form.orderByColumn } : {}),
    ...(form?.changeTrackingColumn ? { changeTrackingColumn: form.changeTrackingColumn } : {}),
    ...(form?.readBatchSize != null && form.readBatchSize !== '' && Number(form.readBatchSize) > 0
      ? { readBatchSize: Number(form.readBatchSize) }
      : {}),
    ...(form?.pollInterval != null && form.pollInterval !== '' && Number(form.pollInterval) > 0
      ? { pollInterval: Number(form.pollInterval) }
      : {}),
  }
}

/**
 * @param {Object} config - connector config from manifest
 * @param {string} connectorType - kafka, postgresql, trino, clickhouse, nessie
 * @param {string} role - 'source' | 'sink'
 * @returns {Object} form state
 */
export function configToForm(config, connectorType, role = 'sink') {
  const c = config || {}
  if (hasSecretRefs(c)) {
    return { useAdvanced: true, advancedJson: JSON.stringify(c, null, 2) }
  }

  const knownKeys = CONNECTOR_KNOWN_KEYS[connectorType] || []
  const extra = extractExtraConfig(c, knownKeys)
  const advancedJson = Object.keys(extra).length > 0 ? JSON.stringify(extra, null, 2) : ''

  switch (connectorType) {
    case 'kafka':
      return {
        brokers: brokersFromConfig(c),
        topic: c.topic || '',
        consumerGroup: role === 'source' ? (c.consumerGroup || '') : '',
        format: c.format || 'json',
        advancedJson,
      }
    case 'postgresql':
      return {
        connectionString: c.connectionString || '',
        table: c.table || '',
        query: c.query || '',
        autoCreateTable: !!c.autoCreateTable,
        batchSize: c.batchSize ?? '',
        rawMode: !!c.rawMode,
        ...(role === 'source' ? sourceIncrementalFromConfig(c) : {}),
        advancedJson,
      }
    case 'trino':
      return {
        serverURL: c.serverURL || '',
        catalog: c.catalog || '',
        schema: c.schema || '',
        table: c.table || '',
        query: c.query || '',
        ...(role === 'source' ? sourceIncrementalFromConfig(c) : {}),
        advancedJson,
      }
    case 'clickhouse':
      return {
        connectionString: c.connectionString || '',
        table: c.table || '',
        query: c.query || '',
        autoCreateTable: !!c.autoCreateTable,
        batchSize: c.batchSize ?? '',
        rawMode: !!c.rawMode,
        ...(role === 'source' ? sourceIncrementalFromConfig(c) : {}),
        advancedJson,
      }
    case 'nessie':
      return {
        baseURL: c.baseURL || '',
        namespace: c.namespace || '',
        table: c.table || '',
        branch: c.branch || '',
        advancedJson,
      }
    default:
      return { useAdvanced: true, advancedJson: JSON.stringify(c, null, 2) }
  }
}

/**
 * @param {Object} form - form state
 * @param {string} connectorType
 * @param {string} role - 'source' | 'sink'
 * @param {Object} advancedConfig - merged advanced JSON (overrides)
 * @returns {Object} config for manifest
 */
export function formToConfig(form, connectorType, role = 'sink', advancedConfig = {}) {
  let base = {}

  if (form?.useAdvanced && form?.advancedJson) {
    try {
      base = JSON.parse(form.advancedJson) || {}
    } catch {
      return advancedConfig
    }
  } else {
    switch (connectorType) {
      case 'kafka': {
        const brokers = brokersToConfig(form?.brokers)
        base = {
          ...(brokers?.length ? { brokers } : {}),
          ...(form?.topic ? { topic: form.topic } : {}),
          ...(role === 'source' && form?.consumerGroup ? { consumerGroup: form.consumerGroup } : {}),
          ...(form?.format && form.format !== 'json' ? { format: form.format } : {}),
        }
        break
      }
      case 'postgresql':
        base = {
          ...(form?.connectionString ? { connectionString: form.connectionString } : {}),
          ...(form?.table ? { table: form.table } : {}),
          ...(form?.query ? { query: form.query } : {}),
          ...(form?.autoCreateTable ? { autoCreateTable: true } : {}),
          ...(form?.batchSize != null && form.batchSize !== '' ? { batchSize: Number(form.batchSize) } : {}),
          ...(form?.rawMode ? { rawMode: true } : {}),
          ...sourceIncrementalToConfig(form, role),
        }
        break
      case 'trino':
        base = {
          ...(form?.serverURL ? { serverURL: form.serverURL } : {}),
          ...(form?.catalog ? { catalog: form.catalog } : {}),
          ...(form?.schema ? { schema: form.schema } : {}),
          ...(form?.table ? { table: form.table } : {}),
          ...(form?.query ? { query: form.query } : {}),
          ...sourceIncrementalToConfig(form, role),
        }
        break
      case 'clickhouse':
        base = {
          ...(form?.connectionString ? { connectionString: form.connectionString } : {}),
          ...(form?.table ? { table: form.table } : {}),
          ...(form?.query ? { query: form.query } : {}),
          ...(form?.autoCreateTable ? { autoCreateTable: true } : {}),
          ...(form?.batchSize != null && form.batchSize !== '' ? { batchSize: Number(form.batchSize) } : {}),
          ...(form?.rawMode ? { rawMode: true } : {}),
          ...sourceIncrementalToConfig(form, role),
        }
        break
      case 'nessie':
        base = {
          ...(form?.baseURL ? { baseURL: form.baseURL } : {}),
          ...(form?.namespace ? { namespace: form.namespace } : {}),
          ...(form?.table ? { table: form.table } : {}),
          ...(form?.branch ? { branch: form.branch } : {}),
        }
        break
      default:
        base = {}
    }
  }

  return { ...base, ...advancedConfig }
}

const TRANSFORMATION_KNOWN_KEYS = {
  timestamp: ['fieldName', 'format'],
  flatten: ['field'],
  filter: ['condition'],
  select: ['fields'],
  remove: ['fields'],
  mask: ['fields', 'keepLength'],
  snakeCase: ['deep'],
  camelCase: ['deep'],
}

/**
 * @param {Object} config - transformation config
 * @param {string} transformationType
 * @returns {Object} form state
 */
export function transformationConfigToForm(config, transformationType) {
  const c = config || {}
  if (hasSecretRefs(c) || transformationType === 'router') {
    return { useAdvanced: true, advancedJson: JSON.stringify(c, null, 2) }
  }

  const knownKeys = TRANSFORMATION_KNOWN_KEYS[transformationType] || []
  const extra = extractExtraConfig(c, knownKeys)
  const advancedJson = Object.keys(extra).length > 0 ? JSON.stringify(extra, null, 2) : ''

  switch (transformationType) {
    case 'timestamp':
      return {
        fieldName: c.fieldName || 'created_at',
        format: c.format || 'RFC3339',
        advancedJson,
      }
    case 'flatten':
      return { field: c.field || '', advancedJson }
    case 'filter':
      return { condition: c.condition || '', advancedJson }
    case 'select':
      return {
        fields: Array.isArray(c.fields) ? c.fields.join('\n') : '',
        advancedJson,
      }
    case 'remove':
      return {
        fields: Array.isArray(c.fields) ? c.fields.join('\n') : '',
        advancedJson,
      }
    case 'mask':
      return {
        fields: Array.isArray(c.fields) ? c.fields.join('\n') : '',
        keepLength: !!c.keepLength,
        advancedJson,
      }
    case 'snakeCase':
    case 'camelCase':
      return {
        deep: !!c.deep,
        advancedJson,
      }
    default:
      return { useAdvanced: true, advancedJson: JSON.stringify(c, null, 2) }
  }
}

/**
 * @param {Object} form - form state
 * @param {string} transformationType
 * @param {Object} advancedConfig - merged advanced JSON
 * @returns {Object} config for manifest
 */
export function transformationFormToConfig(form, transformationType, advancedConfig = {}) {
  let base = {}

  if (form?.useAdvanced && form?.advancedJson) {
    try {
      base = JSON.parse(form.advancedJson) || {}
    } catch {
      return advancedConfig
    }
  } else {
    switch (transformationType) {
      case 'timestamp':
        base = {
          ...(form?.fieldName ? { fieldName: form.fieldName } : {}),
          ...(form?.format && form.format !== 'RFC3339' ? { format: form.format } : {}),
        }
        break
      case 'flatten':
        base = form?.field ? { field: form.field } : {}
        break
      case 'filter':
        base = form?.condition ? { condition: form.condition } : {}
        break
      case 'select':
        base = {
          fields: (form?.fields || '')
            .split('\n')
            .map((s) => s.trim())
            .filter(Boolean),
        }
        break
      case 'remove':
        base = {
          fields: (form?.fields || '')
            .split('\n')
            .map((s) => s.trim())
            .filter(Boolean),
        }
        break
      case 'mask':
        base = {
          fields: (form?.fields || '')
            .split('\n')
            .map((s) => s.trim())
            .filter(Boolean),
          ...(form?.keepLength ? { keepLength: true } : {}),
        }
        break
      case 'snakeCase':
      case 'camelCase':
        base = form && 'deep' in form ? { deep: !!form.deep } : {}
        break
      default:
        base = {}
    }
  }

  return { ...base, ...advancedConfig }
}
