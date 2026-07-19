import { describe, it, expect } from 'vitest'
import {
  manifestToGraph,
  graphToManifest,
  createEmptyGraph,
  CONNECTOR_TYPES,
  TRANSFORMATION_TYPES,
} from './useFlowManifest'

describe('useFlowManifest', () => {
  it('manifestToGraph creates source and sink nodes', () => {
    const manifest = {
      spec: {
        source: { type: 'kafka', config: { brokers: ['localhost:9092'], topic: 'input' } },
        sink: { type: 'postgresql', config: { connectionString: 'postgres://...', table: 'out' } },
        transformations: [],
      },
    }
    const { nodes, edges } = manifestToGraph(manifest)
    expect(nodes).toHaveLength(2)
    expect(nodes[0].id).toBe('source')
    expect(nodes[0].data.connectorType).toBe('kafka')
    expect(nodes[0].data.config).toEqual({ brokers: ['localhost:9092'], topic: 'input' })
    expect(nodes[1].id).toBe('sink')
    expect(nodes[1].data.connectorType).toBe('postgresql')
    expect(nodes[1].data.config).toEqual({ connectionString: 'postgres://...', table: 'out' })
    expect(edges).toHaveLength(1)
    expect(edges[0].source).toBe('source')
    expect(edges[0].target).toBe('sink')
  })

  it('manifestToGraph preserves full source/sink config (sasl, keycloak, etc.)', () => {
    const manifest = {
      metadata: { name: 'com-dp-supply-status-io-prod', namespace: 'dataflow' },
      spec: {
        source: {
          type: 'kafka',
          config: {
            brokers: ['broker:9092'],
            topic: 'input',
            consumerGroup: 'cg',
            format: 'json',
            sasl: {
              mechanism: 'plain',
              usernameSecretRef: { name: 'kafka-secret', key: 'username', namespace: 'dataflow' },
              passwordSecretRef: { name: 'kafka-secret', key: 'password', namespace: 'dataflow' },
            },
            tls: { insecureSkipVerify: true },
          },
        },
        sink: {
          type: 'trino',
          config: {
            serverURL: 'https://trino:443',
            catalog: 'iceberg',
            schema: 'public',
            table: 'events',
            rawMode: true,
            batchSize: 1000,
            autoCreateTable: true,
            keycloak: {
              serverURL: 'https://keycloak/auth',
              realm: 'prod',
              clientID: 'trino',
              tokenSecretRef: { name: 'trino-token', key: 'token', namespace: 'dataflow' },
            },
          },
        },
        transformations: [],
      },
    }
    const { nodes, edges } = manifestToGraph(manifest)
    const result = graphToManifest(nodes, edges, { metadata: manifest.metadata })
    expect(result.spec.source.type).toBe('kafka')
    expect(result.spec.source.config.brokers).toEqual(['broker:9092'])
    expect(result.spec.source.config.sasl.mechanism).toBe('plain')
    expect(result.spec.source.config.sasl.passwordSecretRef.name).toBe('kafka-secret')
    expect(result.spec.sink.type).toBe('trino')
    expect(result.spec.sink.config.keycloak.realm).toBe('prod')
    expect(result.spec.sink.config.table).toBe('events')
  })

  it('manifestToGraph includes transformation nodes', () => {
    const manifest = {
      spec: {
        source: { type: 'kafka', config: {} },
        sink: { type: 'postgresql', config: {} },
        transformations: [
          { type: 'timestamp', config: { fieldName: 'created_at' } },
          { type: 'flatten', config: { field: '$.items' } },
        ],
      },
    }
    const { nodes, edges } = manifestToGraph(manifest)
    expect(nodes).toHaveLength(4)
    expect(nodes.find((n) => n.id === 't-0')).toBeDefined()
    expect(nodes.find((n) => n.id === 't-1')).toBeDefined()
    expect(edges).toHaveLength(3)
  })

  it('graphToManifest round-trips simple flow', () => {
    const manifest = {
      metadata: { name: 'test-flow', namespace: 'default' },
      spec: {
        source: { type: 'kafka', config: { topic: 'in' } },
        sink: { type: 'postgresql', config: { table: 'out' } },
        transformations: [],
      },
    }
    const { nodes, edges } = manifestToGraph(manifest)
    const result = graphToManifest(nodes, edges, { metadata: manifest.metadata })
    expect(result.spec.source.type).toBe('kafka')
    expect(result.spec.source.config.topic).toBe('in')
    expect(result.spec.sink.type).toBe('postgresql')
    expect(result.spec.sink.config.table).toBe('out')
    expect(result.spec.transformations).toBeUndefined()
  })

  it('graphToManifest round-trips with transformations', () => {
    const manifest = {
      metadata: { name: 'test-flow', namespace: 'default' },
      spec: {
        source: { type: 'kafka', config: {} },
        sink: { type: 'postgresql', config: {} },
        transformations: [{ type: 'timestamp', config: { fieldName: 'ts' } }],
      },
    }
    const { nodes, edges } = manifestToGraph(manifest)
    const result = graphToManifest(nodes, edges, { metadata: manifest.metadata })
    expect(result.spec.transformations).toHaveLength(1)
    expect(result.spec.transformations[0].type).toBe('timestamp')
    expect(result.spec.transformations[0].config.fieldName).toBe('ts')
  })

  it('createEmptyGraph returns valid graph', () => {
    const { nodes, edges } = createEmptyGraph()
    expect(nodes.length).toBeGreaterThanOrEqual(2)
    expect(edges.length).toBe(1)
    expect(nodes[0].id).toBe('source')
    expect(nodes[1].id).toBe('sink')
  })

  it('graphToManifest throws when source or sink missing', () => {
    const nodes = [{ id: 'source', type: 'source', position: { x: 0, y: 0 }, data: {} }]
    const edges = []
    expect(() => graphToManifest(nodes, edges)).toThrow('Source and Sink')
  })

  it('CONNECTOR_TYPES and TRANSFORMATION_TYPES are defined', () => {
    expect(CONNECTOR_TYPES).toContain('kafka')
    expect(CONNECTOR_TYPES).toContain('postgresql')
    expect(CONNECTOR_TYPES).toContain('iceberg')
    expect(TRANSFORMATION_TYPES).toContain('timestamp')
    expect(TRANSFORMATION_TYPES).toContain('flatten')
    expect(TRANSFORMATION_TYPES).toContain('replaceField')
    expect(TRANSFORMATION_TYPES).toContain('headersToPayload')
    expect(TRANSFORMATION_TYPES).toContain('structFlatten')
    expect(TRANSFORMATION_TYPES).toContain('debeziumUnwrap')
    expect(TRANSFORMATION_TYPES).toContain('extractField')
    expect(TRANSFORMATION_TYPES).toContain('hoistField')
    expect(TRANSFORMATION_TYPES).toContain('cast')
    expect(TRANSFORMATION_TYPES).toContain('timezone')
  })

  it('manifestToGraph round-trips iceberg source and sink', () => {
    const manifest = {
      metadata: { name: 'iceberg-flow', namespace: 'default' },
      spec: {
        source: {
          type: 'iceberg',
          config: {
            catalogURI: 'https://catalog:8181',
            namespace: 'analytics',
            table: 'events',
            pollInterval: 10,
            incrementalBySnapshot: true,
          },
        },
        sink: {
          type: 'iceberg',
          config: {
            catalogURI: 'https://catalog:8181',
            namespace: 'analytics',
            table: 'events_out',
            upsertMode: true,
            conflictKey: 'id',
          },
        },
        transformations: [],
      },
    }
    const { nodes, edges } = manifestToGraph(manifest)
    expect(nodes[0].data.connectorType).toBe('iceberg')
    expect(nodes[0].data.config.catalogURI).toBe('https://catalog:8181')
    expect(nodes[1].data.connectorType).toBe('iceberg')
    expect(nodes[1].data.config.upsertMode).toBe(true)

    const result = graphToManifest(nodes, edges, manifest)
    expect(result.spec.source.type).toBe('iceberg')
    expect(result.spec.source.config.incrementalBySnapshot).toBe(true)
    expect(result.spec.sink.config.upsertMode).toBe(true)
    expect(result.spec.sink.config.conflictKey).toBe('id')
  })

  it('graphToManifest uses baseManifest metadata', () => {
    const { nodes, edges } = createEmptyGraph()
    const result = graphToManifest(nodes, edges, {
      metadata: { name: 'my-flow', namespace: 'prod' },
    })
    expect(result.metadata.name).toBe('my-flow')
    expect(result.metadata.namespace).toBe('prod')
  })

  it('graphToManifest preserves non-graph spec fields from baseManifest', () => {
    const manifest = {
      metadata: { name: 'test-flow', namespace: 'default' },
      spec: {
        source: { type: 'kafka', config: { topic: 'in' } },
        sink: { type: 'postgresql', config: { table: 'out' } },
        transformations: [],
        checkpointPersistence: true,
        checkpointReset: true,
        strictIdempotency: true,
        replicas: 2,
        errors: {
          type: 'kafka',
          config: { topic: 'errors' },
          ackPolicy: 'afterWrite',
        },
      },
    }
    const { nodes, edges } = manifestToGraph(manifest)
    const result = graphToManifest(nodes, edges, manifest)
    expect(result.spec.checkpointPersistence).toBe(true)
    expect(result.spec.checkpointReset).toBe(true)
    expect(result.spec.strictIdempotency).toBe(true)
    expect(result.spec.replicas).toBe(2)
    expect(result.spec.errors).toEqual(manifest.spec.errors)
    expect(result.spec.source.config.topic).toBe('in')
    expect(result.spec.sink.config.table).toBe('out')
  })

  it('graphToManifest updates graph fields while preserving other spec fields', () => {
    const baseManifest = {
      metadata: { name: 'test-flow', namespace: 'default' },
      spec: {
        source: { type: 'kafka', config: { topic: 'old-in' } },
        sink: { type: 'postgresql', config: { table: 'old-out' } },
        transformations: [{ type: 'timestamp', config: { fieldName: 'ts' } }],
        checkpointSyncOnAck: true,
        ackGranularity: 'message',
      },
    }
    const { nodes, edges } = manifestToGraph(baseManifest)
    nodes.find((n) => n.id === 'source').data.config = { topic: 'new-in' }
    nodes.find((n) => n.id === 'sink').data.config = { table: 'new-out' }
    const result = graphToManifest(nodes, edges, baseManifest)
    expect(result.spec.source.config.topic).toBe('new-in')
    expect(result.spec.sink.config.table).toBe('new-out')
    expect(result.spec.checkpointSyncOnAck).toBe(true)
    expect(result.spec.ackGranularity).toBe('message')
    expect(result.spec.transformations).toHaveLength(1)
  })

  it('graphToManifest round-trips router transformation', () => {
    const manifest = {
      metadata: { name: 'router-flow', namespace: 'default' },
      spec: {
        source: { type: 'postgresql', config: { table: 'events' } },
        sink: { type: 'kafka', config: { topic: 'default' } },
        transformations: [
          {
            type: 'router',
            config: {
              routes: [
                {
                  condition: "$.type == 'order'",
                  sink: {
                    type: 'kafka',
                    config: { brokers: ['localhost:9092'], topic: 'orders' },
                  },
                },
                {
                  condition: "$.type == 'user'",
                  sink: {
                    type: 'kafka',
                    config: { brokers: ['localhost:9092'], topic: 'users' },
                  },
                },
              ],
            },
          },
        ],
      },
    }
    const { nodes, edges } = manifestToGraph(manifest)
    const routerNode = nodes.find((n) => n.data.transformationType === 'router')
    expect(routerNode).toBeDefined()
    expect(routerNode.data.config.routes).toHaveLength(2)

    const result = graphToManifest(nodes, edges, manifest)
    expect(result.spec.transformations).toHaveLength(1)
    expect(result.spec.transformations[0].type).toBe('router')
    expect(result.spec.transformations[0].config.routes[0].condition).toBe("$.type == 'order'")
    expect(result.spec.transformations[0].config.routes[0].sink.config.topic).toBe('orders')
  })

  it('parity regression: constructor save preserves iceberg, errors, fault-tolerance and upsertMode', () => {
    const manifest = {
      metadata: { name: 'parity-flow', namespace: 'prod', resourceVersion: '12345' },
      spec: {
        source: {
          type: 'iceberg',
          config: {
            catalogURI: 'https://catalog:8181',
            namespace: 'analytics',
            table: 'events_in',
            pollInterval: 15,
            incrementalBySnapshot: true,
          },
        },
        sink: {
          type: 'iceberg',
          config: {
            catalogURI: 'https://catalog:8181',
            namespace: 'analytics',
            table: 'events_out',
            upsertMode: true,
            conflictKey: 'event_id',
          },
        },
        transformations: [{ type: 'timestamp', config: { fieldName: 'processed_at' } }],
        checkpointPersistence: true,
        checkpointSyncOnAck: true,
        checkpointSaveInterval: '30s',
        strictIdempotency: true,
        ackGranularity: 'message',
        replicas: 2,
        channelBufferSize: 1000,
        errors: {
          type: 'kafka',
          config: { brokers: ['kafka:9092'], topic: 'dlq' },
          ackPolicy: 'afterWrite',
        },
      },
    }

    const { nodes, edges } = manifestToGraph(manifest)
    nodes.find((n) => n.id === 'source').data.config.pollInterval = 20

    const result = graphToManifest(nodes, edges, manifest)

    expect(result.spec.source.type).toBe('iceberg')
    expect(result.spec.source.config.pollInterval).toBe(20)
    expect(result.spec.source.config.incrementalBySnapshot).toBe(true)
    expect(result.spec.sink.config.upsertMode).toBe(true)
    expect(result.spec.sink.config.conflictKey).toBe('event_id')
    expect(result.spec.checkpointPersistence).toBe(true)
    expect(result.spec.checkpointSyncOnAck).toBe(true)
    expect(result.spec.checkpointSaveInterval).toBe('30s')
    expect(result.spec.strictIdempotency).toBe(true)
    expect(result.spec.ackGranularity).toBe('message')
    expect(result.spec.replicas).toBe(2)
    expect(result.spec.channelBufferSize).toBe(1000)
    expect(result.spec.errors).toEqual(manifest.spec.errors)
    expect(result.metadata.name).toBe('parity-flow')
    expect(result.metadata.namespace).toBe('prod')
  })
})
