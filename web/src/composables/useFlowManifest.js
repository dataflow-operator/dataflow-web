/**
 * Converts between Vue Flow graph (nodes, edges) and DataFlow manifest spec.
 */

const SOURCE_NODE_ID = 'source'
const SINK_NODE_ID = 'sink'
const TRANSFORMATION_PREFIX = 't-'

export const CONNECTOR_TYPES = ['kafka', 'postgresql', 'postgresql-cdc', 'trino', 'clickhouse', 'nessie', 'iceberg']
export const TRANSFORMATION_TYPES = [
  'timestamp',
  'flatten',
  'filter',
  'mask',
  'router',
  'select',
  'remove',
  'snakeCase',
  'camelCase',
  'debeziumUnwrap',
  'replaceField',
  'headersToPayload',
  'structFlatten',
  'extractField',
  'hoistField',
  'cast',
  'timezone',
]

/**
 * Normalizes config from manifest (handles RawExtension, nested objects, etc.)
 * @param {*} raw - config from API/manifest
 * @returns {Object}
 */
function normalizeConfig(raw) {
  if (raw == null) return {}
  if (typeof raw === 'object' && !Array.isArray(raw)) {
    // Kubernetes RawExtension may have Raw as base64 or string
    if (raw.Raw != null) {
      try {
        const decoded = typeof raw.Raw === 'string'
          ? (raw.Raw.startsWith('{') ? JSON.parse(raw.Raw) : JSON.parse(atob(raw.Raw)))
          : raw.Raw
        return typeof decoded === 'object' && decoded !== null ? decoded : {}
      } catch {
        return {}
      }
    }
    return JSON.parse(JSON.stringify(raw))
  }
  return {}
}

/**
 * @param {Object} manifest - Full DataFlow manifest
 * @returns {{ nodes: Array, edges: Array }}
 */
export function manifestToGraph(manifest) {
  const nodes = []
  const edges = []
  const spec = manifest?.spec || {}

  // Source node
  const source = spec.source || {}
  nodes.push({
    id: SOURCE_NODE_ID,
    type: 'source',
    position: { x: 50, y: 200 },
    data: {
      label: source.type || 'kafka',
      connectorType: source.type || 'kafka',
      config: normalizeConfig(source.config),
    },
  })

  // Transformation nodes
  const transformations = spec.transformations || []
  let prevId = SOURCE_NODE_ID
  transformations.forEach((t, i) => {
    const nodeId = `${TRANSFORMATION_PREFIX}${i}`
    nodes.push({
      id: nodeId,
      type: 'transformation',
      position: { x: 250 + i * 200, y: 200 },
      data: {
        label: t.type || 'transformation',
        transformationType: t.type || 'timestamp',
        config: normalizeConfig(t.config),
      },
    })
    edges.push({ id: `e-${prevId}-${nodeId}`, source: prevId, target: nodeId })
    prevId = nodeId
  })

  // Sink node
  const sink = spec.sink || {}
  nodes.push({
    id: SINK_NODE_ID,
    type: 'sink',
    position: { x: 250 + transformations.length * 200, y: 200 },
    data: {
      label: sink.type || 'postgresql',
      connectorType: sink.type || 'postgresql',
      config: normalizeConfig(sink.config),
    },
  })
  edges.push({ id: `e-${prevId}-sink`, source: prevId, target: SINK_NODE_ID })

  return { nodes, edges }
}

/**
 * @param {Array} nodes - Vue Flow nodes
 * @param {Array} edges - Vue Flow edges
 * @param {Object} baseManifest - Base manifest (metadata, etc.)
 * @returns {Object} DataFlow manifest
 */
export function graphToManifest(nodes, edges, baseManifest = {}) {
  const nodeMap = {}
  for (const n of nodes) {
    nodeMap[n.id] = n
  }

  const sourceNode = nodeMap[SOURCE_NODE_ID]
  const sinkNode = nodeMap[SINK_NODE_ID]
  if (!sourceNode || !sinkNode) {
    throw new Error('Source and Sink nodes are required')
  }

  // Build transformations in order (follow edges from source to sink)
  const transformations = []
  let currentId = SOURCE_NODE_ID
  const visited = new Set([SOURCE_NODE_ID])

  while (currentId !== SINK_NODE_ID) {
    const outgoing = edges.filter((e) => e.source === currentId)
    if (outgoing.length === 0) break
    const nextEdge = outgoing[0]
    const nextId = nextEdge.target
    if (visited.has(nextId)) break
    visited.add(nextId)

    const nextNode = nodeMap[nextId]
    if (nextNode?.type === 'transformation') {
      const cfg = nextNode.data?.config || {}
      const cleanConfig = Object.keys(cfg).length ? cfg : undefined
      transformations.push({
        type: nextNode.data?.transformationType || 'timestamp',
        config: cleanConfig,
      })
    }
    currentId = nextId
  }

  const sourceConfig = sourceNode.data?.config || {}
  const sinkConfig = sinkNode.data?.config || {}
  const cleanSourceConfig = Object.keys(sourceConfig).length ? sourceConfig : undefined
  const cleanSinkConfig = Object.keys(sinkConfig).length ? sinkConfig : undefined

  const spec = baseManifest?.spec
    ? JSON.parse(JSON.stringify(baseManifest.spec))
    : {}

  spec.source = {
    type: sourceNode.data?.connectorType || 'kafka',
    config: cleanSourceConfig,
  }
  spec.sink = {
    type: sinkNode.data?.connectorType || 'postgresql',
    config: cleanSinkConfig,
  }
  if (transformations.length > 0) {
    spec.transformations = transformations
  } else {
    delete spec.transformations
  }

  return {
    apiVersion: baseManifest.apiVersion || 'dataflow.dataflow.io/v1',
    kind: baseManifest.kind || 'DataFlow',
    metadata: baseManifest.metadata || { name: '', namespace: 'default' },
    spec,
  }
}

export function createEmptyGraph() {
  return manifestToGraph({
    spec: {
      source: { type: 'kafka', config: {} },
      sink: { type: 'postgresql', config: {} },
      transformations: [],
    },
  })
}
