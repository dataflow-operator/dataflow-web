/**
 * Composable for converting between structured form state and connector/transformation config.
 * Handles configToForm and formToConfig for all connector and transformation types.
 */

const CONNECTOR_KNOWN_KEYS = {
  kafka: ['brokers', 'topic', 'consumerGroup', 'format'],
  postgresql: ['connectionString', 'table', 'query', 'autoCreateTable', 'batchSize', 'rawMode', 'orderByColumn', 'changeTrackingColumn', 'readBatchSize', 'pollInterval', 'upsertMode', 'conflictKey'],
  trino: ['serverURL', 'catalog', 'schema', 'table', 'query', 'orderByColumn', 'changeTrackingColumn', 'readBatchSize', 'pollInterval', 'upsertMode', 'conflictKey'],
  clickhouse: ['connectionString', 'table', 'query', 'autoCreateTable', 'batchSize', 'rawMode', 'orderByColumn', 'changeTrackingColumn', 'readBatchSize', 'pollInterval', 'upsertMode', 'conflictKey'],
  nessie: ['baseURL', 'namespace', 'table', 'branch'],
  iceberg: [
    'catalogURI', 'prefix', 'warehouse', 'namespace', 'table', 'query', 'pollInterval',
    'authenticationType', 'bearerToken', 'basicAuthUsername', 'basicAuthPassword',
    'oauth2ServerURI', 'oauth2ClientID', 'oauth2ClientSecret', 'oauth2Scope',
    'incrementalBySnapshot', 'startSnapshotID', 'batchSize', 'batchFlushIntervalSeconds',
    'autoCreateTable', 'rawMode', 's3Endpoint', 's3Region', 'upsertMode', 'conflictKey',
  ],
}

export const SINK_IDEMPOTENCY_CONNECTORS = ['postgresql', 'trino', 'clickhouse', 'iceberg']

export const POLLING_SOURCE_TYPES = ['postgresql', 'clickhouse', 'trino', 'nessie', 'iceberg']

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

function sinkIdempotencyFromConfig(c) {
  return {
    upsertMode: !!c.upsertMode,
    conflictKey: c.conflictKey || '',
  }
}

function sinkIdempotencyToConfig(form, role) {
  if (role !== 'sink') return {}
  return {
    ...(form?.upsertMode ? { upsertMode: true } : {}),
    ...(form?.conflictKey ? { conflictKey: form.conflictKey } : {}),
  }
}

function basicAuthFromConfig(c) {
  return {
    basicAuthUsername: c.basicAuth?.username || '',
    basicAuthPassword: c.basicAuth?.password || '',
  }
}

function basicAuthToConfig(form) {
  const username = form?.basicAuthUsername?.trim()
  const password = form?.basicAuthPassword
  if (!username && !password) return {}
  return {
    basicAuth: {
      ...(username ? { username } : {}),
      ...(password ? { password } : {}),
    },
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
        ...(role === 'sink' ? sinkIdempotencyFromConfig(c) : {}),
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
        ...(role === 'sink' ? sinkIdempotencyFromConfig(c) : {}),
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
        ...(role === 'sink' ? sinkIdempotencyFromConfig(c) : {}),
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
    case 'iceberg':
      return {
        catalogURI: c.catalogURI || '',
        prefix: c.prefix || '',
        warehouse: c.warehouse || '',
        namespace: c.namespace || '',
        table: c.table || '',
        query: c.query || '',
        pollInterval: c.pollInterval ?? '',
        authenticationType: c.authenticationType || 'AUTO',
        bearerToken: c.bearerToken || '',
        ...basicAuthFromConfig(c),
        oauth2ServerURI: c.oauth2ServerURI || '',
        oauth2ClientID: c.oauth2ClientID || '',
        oauth2ClientSecret: c.oauth2ClientSecret || '',
        oauth2Scope: c.oauth2Scope || '',
        incrementalBySnapshot: !!c.incrementalBySnapshot,
        startSnapshotID: c.startSnapshotID || '',
        batchSize: c.batchSize ?? '',
        batchFlushIntervalSeconds: c.batchFlushIntervalSeconds ?? '',
        autoCreateTable: !!c.autoCreateTable,
        rawMode: !!c.rawMode,
        s3Endpoint: c.s3Endpoint || '',
        s3Region: c.s3Region || '',
        ...(role === 'sink' ? sinkIdempotencyFromConfig(c) : {}),
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
          ...sinkIdempotencyToConfig(form, role),
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
          ...sinkIdempotencyToConfig(form, role),
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
          ...sinkIdempotencyToConfig(form, role),
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
      case 'iceberg':
        base = {
          ...(form?.catalogURI ? { catalogURI: form.catalogURI } : {}),
          ...(form?.prefix ? { prefix: form.prefix } : {}),
          ...(form?.warehouse ? { warehouse: form.warehouse } : {}),
          ...(form?.namespace ? { namespace: form.namespace } : {}),
          ...(form?.table ? { table: form.table } : {}),
          ...(form?.query ? { query: form.query } : {}),
          ...(form?.pollInterval != null && form.pollInterval !== '' && Number(form.pollInterval) > 0
            ? { pollInterval: Number(form.pollInterval) }
            : {}),
          ...(form?.authenticationType && form.authenticationType !== 'AUTO'
            ? { authenticationType: form.authenticationType }
            : {}),
          ...(form?.bearerToken ? { bearerToken: form.bearerToken } : {}),
          ...basicAuthToConfig(form),
          ...(form?.oauth2ServerURI ? { oauth2ServerURI: form.oauth2ServerURI } : {}),
          ...(form?.oauth2ClientID ? { oauth2ClientID: form.oauth2ClientID } : {}),
          ...(form?.oauth2ClientSecret ? { oauth2ClientSecret: form.oauth2ClientSecret } : {}),
          ...(form?.oauth2Scope ? { oauth2Scope: form.oauth2Scope } : {}),
          ...(form?.incrementalBySnapshot ? { incrementalBySnapshot: true } : {}),
          ...(form?.startSnapshotID ? { startSnapshotID: form.startSnapshotID } : {}),
          ...(form?.batchSize != null && form.batchSize !== '' ? { batchSize: Number(form.batchSize) } : {}),
          ...(form?.batchFlushIntervalSeconds != null && form.batchFlushIntervalSeconds !== ''
            ? { batchFlushIntervalSeconds: Number(form.batchFlushIntervalSeconds) }
            : {}),
          ...(form?.autoCreateTable ? { autoCreateTable: true } : {}),
          ...(form?.rawMode ? { rawMode: true } : {}),
          ...(form?.s3Endpoint ? { s3Endpoint: form.s3Endpoint } : {}),
          ...(form?.s3Region ? { s3Region: form.s3Region } : {}),
          ...sinkIdempotencyToConfig(form, role),
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
  router: ['routes'],
}

function emptyRouterRoute() {
  return {
    condition: '',
    sinkType: 'kafka',
    sinkConfigJson: '{}',
  }
}

function routerConfigHasSecretRefs(config) {
  if (!config?.routes?.length) return false
  return config.routes.some(
    (route) => hasSecretRefs(route?.sink?.config) || hasSecretRefs(route?.sink)
  )
}

function routeConfigToFormRoute(route) {
  const sink = route?.sink || {}
  const sinkConfig = sink.config || {}
  return {
    condition: route?.condition || '',
    sinkType: sink.type || 'kafka',
    sinkConfigJson: JSON.stringify(sinkConfig, null, 2),
  }
}

function formRouteToRouteConfig(formRoute) {
  let sinkConfig = {}
  if (formRoute?.sinkConfigJson?.trim()) {
    try {
      sinkConfig = JSON.parse(formRoute.sinkConfigJson) || {}
    } catch {
      sinkConfig = {}
    }
  }
  return {
    condition: (formRoute?.condition || '').trim(),
    sink: {
      type: formRoute?.sinkType || 'kafka',
      ...(Object.keys(sinkConfig).length > 0 ? { config: sinkConfig } : {}),
    },
  }
}

/**
 * @param {Object} config - transformation config
 * @param {string} transformationType
 * @returns {Object} form state
 */
export function transformationConfigToForm(config, transformationType) {
  const c = config || {}
  if (hasSecretRefs(c)) {
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
    case 'router':
      if (routerConfigHasSecretRefs(c)) {
        return { useAdvanced: true, advancedJson: JSON.stringify(c, null, 2) }
      }
      return {
        routes:
          Array.isArray(c.routes) && c.routes.length > 0
            ? c.routes.map(routeConfigToFormRoute)
            : [emptyRouterRoute()],
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
      case 'router':
        base = {
          routes: (form?.routes || [])
            .map(formRouteToRouteConfig)
            .filter((route) => route.condition),
        }
        break
      default:
        base = {}
    }
  }

  return { ...base, ...advancedConfig }
}
