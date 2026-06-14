import { describe, it, expect } from 'vitest'
import {
  configToForm,
  formToConfig,
  transformationConfigToForm,
  transformationFormToConfig,
} from './useConfigForm'

describe('useConfigForm', () => {
  describe('configToForm / formToConfig', () => {
    it('Kafka source round-trip', () => {
      const config = {
        brokers: ['localhost:9092', 'kafka2:9092'],
        topic: 'input-topic',
        consumerGroup: 'dataflow-group',
        format: 'json',
      }
      const form = configToForm(config, 'kafka', 'source')
      expect(form.brokers).toBe('localhost:9092, kafka2:9092')
      expect(form.topic).toBe('input-topic')
      expect(form.consumerGroup).toBe('dataflow-group')
      expect(form.format).toBe('json')

      const result = formToConfig(form, 'kafka', 'source', {})
      expect(result.brokers).toEqual(['localhost:9092', 'kafka2:9092'])
      expect(result.topic).toBe('input-topic')
      expect(result.consumerGroup).toBe('dataflow-group')
    })

    it('Kafka sink omits consumerGroup', () => {
      const config = { brokers: ['kafka:9092'], topic: 'out' }
      const form = configToForm(config, 'kafka', 'sink')
      expect(form.consumerGroup).toBe('')

      const result = formToConfig(form, 'kafka', 'sink', {})
      expect(result.consumerGroup).toBeUndefined()
    })

    it('PostgreSQL round-trip', () => {
      const config = {
        connectionString: 'postgres://user:pass@host:5432/db',
        table: 'output_table',
        autoCreateTable: true,
        batchSize: 1000,
        rawMode: true,
      }
      const form = configToForm(config, 'postgresql', 'sink')
      expect(form.connectionString).toBe(config.connectionString)
      expect(form.table).toBe(config.table)
      expect(form.autoCreateTable).toBe(true)
      expect(form.batchSize).toBe(1000)
      expect(form.rawMode).toBe(true)

      const result = formToConfig(form, 'postgresql', 'sink', {})
      expect(result.connectionString).toBe(config.connectionString)
      expect(result.table).toBe(config.table)
      expect(result.autoCreateTable).toBe(true)
      expect(result.batchSize).toBe(1000)
      expect(result.rawMode).toBe(true)
    })

    it('Trino round-trip', () => {
      const config = {
        serverURL: 'http://trino:8080',
        catalog: 'hive',
        schema: 'default',
        table: 'events',
      }
      const form = configToForm(config, 'trino', 'sink')
      expect(form.serverURL).toBe(config.serverURL)
      expect(form.catalog).toBe(config.catalog)
      expect(form.schema).toBe(config.schema)
      expect(form.table).toBe(config.table)

      const result = formToConfig(form, 'trino', 'sink', {})
      expect(result).toEqual(config)
    })

    it('ClickHouse round-trip', () => {
      const config = {
        connectionString: 'clickhouse://default@ch:9000/default',
        table: 'output',
        autoCreateTable: true,
      }
      const form = configToForm(config, 'clickhouse', 'sink')
      expect(form.connectionString).toBe(config.connectionString)
      expect(form.table).toBe(config.table)
      expect(form.autoCreateTable).toBe(true)

      const result = formToConfig(form, 'clickhouse', 'sink', {})
      expect(result.connectionString).toBe(config.connectionString)
      expect(result.table).toBe(config.table)
      expect(result.autoCreateTable).toBe(true)
    })

    it('PostgreSQL source incremental round-trip', () => {
      const config = {
        connectionString: 'postgres://localhost/db',
        table: 'events',
        query: 'SELECT * FROM events',
        changeTrackingColumn: 'updated_at',
        orderByColumn: 'id',
        readBatchSize: 500,
        pollInterval: 10,
      }
      const form = configToForm(config, 'postgresql', 'source')
      expect(form.changeTrackingColumn).toBe('updated_at')
      expect(form.orderByColumn).toBe('id')
      expect(form.readBatchSize).toBe(500)
      expect(form.pollInterval).toBe(10)

      const result = formToConfig(form, 'postgresql', 'source', {})
      expect(result.changeTrackingColumn).toBe('updated_at')
      expect(result.orderByColumn).toBe('id')
      expect(result.readBatchSize).toBe(500)
      expect(result.pollInterval).toBe(10)
    })

    it('Trino source incremental round-trip', () => {
      const config = {
        serverURL: 'http://trino:8080',
        catalog: 'hive',
        schema: 'default',
        table: 'events',
        changeTrackingColumn: 'updated_at',
        readBatchSize: 1000,
      }
      const form = configToForm(config, 'trino', 'source')
      expect(form.changeTrackingColumn).toBe('updated_at')
      expect(form.readBatchSize).toBe(1000)

      const result = formToConfig(form, 'trino', 'source', {})
      expect(result.changeTrackingColumn).toBe('updated_at')
      expect(result.readBatchSize).toBe(1000)
    })

    it('Nessie round-trip', () => {
      const config = {
        baseURL: 'http://nessie:19120',
        namespace: 'my_schema',
        table: 'my_table',
        branch: 'main',
      }
      const form = configToForm(config, 'nessie', 'sink')
      expect(form.baseURL).toBe(config.baseURL)
      expect(form.namespace).toBe(config.namespace)
      expect(form.table).toBe(config.table)
      expect(form.branch).toBe(config.branch)

      const result = formToConfig(form, 'nessie', 'sink', {})
      expect(result).toEqual(config)
    })

    it('Iceberg source round-trip', () => {
      const config = {
        catalogURI: 'https://catalog:8181',
        namespace: 'analytics',
        table: 'events',
        pollInterval: 30,
        incrementalBySnapshot: true,
        startSnapshotID: '12345',
        authenticationType: 'BEARER',
        bearerToken: 'secret',
      }
      const form = configToForm(config, 'iceberg', 'source')
      expect(form.catalogURI).toBe(config.catalogURI)
      expect(form.namespace).toBe(config.namespace)
      expect(form.table).toBe(config.table)
      expect(form.pollInterval).toBe(30)
      expect(form.incrementalBySnapshot).toBe(true)
      expect(form.startSnapshotID).toBe('12345')
      expect(form.authenticationType).toBe('BEARER')
      expect(form.bearerToken).toBe('secret')

      const result = formToConfig(form, 'iceberg', 'source', {})
      expect(result.catalogURI).toBe(config.catalogURI)
      expect(result.pollInterval).toBe(30)
      expect(result.incrementalBySnapshot).toBe(true)
      expect(result.startSnapshotID).toBe('12345')
      expect(result.authenticationType).toBe('BEARER')
      expect(result.bearerToken).toBe('secret')
    })

    it('Iceberg sink round-trip with upsertMode', () => {
      const config = {
        catalogURI: 'https://catalog:8181',
        namespace: 'analytics',
        table: 'events',
        autoCreateTable: true,
        batchSize: 500,
        upsertMode: true,
        conflictKey: 'id',
        basicAuth: { username: 'user', password: 'pass' },
      }
      const form = configToForm(config, 'iceberg', 'sink')
      expect(form.upsertMode).toBe(true)
      expect(form.conflictKey).toBe('id')
      expect(form.basicAuthUsername).toBe('user')
      expect(form.basicAuthPassword).toBe('pass')

      const result = formToConfig(form, 'iceberg', 'sink', {})
      expect(result.upsertMode).toBe(true)
      expect(result.conflictKey).toBe('id')
      expect(result.basicAuth).toEqual({ username: 'user', password: 'pass' })
      expect(result.autoCreateTable).toBe(true)
      expect(result.batchSize).toBe(500)
    })

    it('PostgreSQL sink upsertMode round-trip', () => {
      const config = {
        connectionString: 'postgres://localhost/db',
        table: 'output',
        upsertMode: true,
        conflictKey: 'id',
      }
      const form = configToForm(config, 'postgresql', 'sink')
      expect(form.upsertMode).toBe(true)
      expect(form.conflictKey).toBe('id')

      const result = formToConfig(form, 'postgresql', 'sink', {})
      expect(result.upsertMode).toBe(true)
      expect(result.conflictKey).toBe('id')
    })

    it('Trino sink upsertMode round-trip', () => {
      const config = {
        serverURL: 'http://trino:8080',
        catalog: 'iceberg',
        schema: 'default',
        table: 'events',
        upsertMode: true,
        conflictKey: 'event_id',
      }
      const form = configToForm(config, 'trino', 'sink')
      const result = formToConfig(form, 'trino', 'sink', {})
      expect(result.upsertMode).toBe(true)
      expect(result.conflictKey).toBe('event_id')
    })

    it('ClickHouse sink upsertMode round-trip', () => {
      const config = {
        connectionString: 'clickhouse://localhost/default',
        table: 'output',
        upsertMode: true,
        conflictKey: 'id',
      }
      const form = configToForm(config, 'clickhouse', 'sink')
      const result = formToConfig(form, 'clickhouse', 'sink', {})
      expect(result.upsertMode).toBe(true)
      expect(result.conflictKey).toBe('id')
    })

    it('config with SecretRefs returns useAdvanced', () => {
      const config = {
        brokersSecretRef: { name: 'kafka-secret', key: 'brokers' },
        topicSecretRef: { name: 'kafka-secret', key: 'topic' },
      }
      const form = configToForm(config, 'kafka', 'source')
      expect(form.useAdvanced).toBe(true)
      expect(form.advancedJson).toContain('brokersSecretRef')
    })

    it('extra config keys go to advancedJson', () => {
      const config = {
        brokers: ['kafka:9092'],
        topic: 't',
        sasl: { mechanism: 'plain', username: 'u', password: 'p' },
      }
      const form = configToForm(config, 'kafka', 'source')
      expect(form.brokers).toBe('kafka:9092')
      expect(form.topic).toBe('t')
      expect(form.advancedJson).toContain('sasl')
      expect(JSON.parse(form.advancedJson)).toEqual({ sasl: config.sasl })
    })

    it('formToConfig merges advancedConfig', () => {
      const form = {
        brokers: 'kafka:9092',
        topic: 't',
        consumerGroup: 'cg',
        format: 'json',
        advancedJson: '',
      }
      const advanced = { sasl: { mechanism: 'plain' } }
      const result = formToConfig(form, 'kafka', 'source', advanced)
      expect(result.brokers).toEqual(['kafka:9092'])
      expect(result.topic).toBe('t')
      expect(result.sasl).toEqual({ mechanism: 'plain' })
    })
  })

  describe('transformationConfigToForm / transformationFormToConfig', () => {
    it('timestamp round-trip', () => {
      const config = { fieldName: 'created_at', format: 'RFC3339' }
      const form = transformationConfigToForm(config, 'timestamp')
      expect(form.fieldName).toBe('created_at')
      expect(form.format).toBe('RFC3339')

      const result = transformationFormToConfig(form, 'timestamp', {})
      expect(result.fieldName).toBe('created_at')
      if (result.format !== undefined) {
        expect(result.format).toBe('RFC3339')
      }
    })

    it('flatten round-trip', () => {
      const config = { field: '$.items' }
      const form = transformationConfigToForm(config, 'flatten')
      expect(form.field).toBe('$.items')

      const result = transformationFormToConfig(form, 'flatten', {})
      expect(result).toEqual(config)
    })

    it('filter round-trip', () => {
      const config = { condition: '$.enabled == true' }
      const form = transformationConfigToForm(config, 'filter')
      expect(form.condition).toBe('$.enabled == true')

      const result = transformationFormToConfig(form, 'filter', {})
      expect(result).toEqual(config)
    })

    it('select round-trip', () => {
      const config = { fields: ['$.id', '$.name'] }
      const form = transformationConfigToForm(config, 'select')
      expect(form.fields).toBe('$.id\n$.name')

      const result = transformationFormToConfig(form, 'select', {})
      expect(result.fields).toEqual(['$.id', '$.name'])
    })

    it('remove round-trip', () => {
      const config = { fields: ['$.password', '$.secret'] }
      const form = transformationConfigToForm(config, 'remove')
      expect(form.fields).toBe('$.password\n$.secret')

      const result = transformationFormToConfig(form, 'remove', {})
      expect(result.fields).toEqual(['$.password', '$.secret'])
    })

    it('mask round-trip', () => {
      const config = { fields: ['password', 'email'], keepLength: true }
      const form = transformationConfigToForm(config, 'mask')
      expect(form.fields).toBe('password\nemail')
      expect(form.keepLength).toBe(true)

      const result = transformationFormToConfig(form, 'mask', {})
      expect(result.fields).toEqual(['password', 'email'])
      expect(result.keepLength).toBe(true)
    })

    it('router round-trip with structured routes', () => {
      const config = {
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
      }
      const form = transformationConfigToForm(config, 'router')
      expect(form.useAdvanced).toBeUndefined()
      expect(form.routes).toHaveLength(2)
      expect(form.routes[0].condition).toBe("$.type == 'order'")
      expect(form.routes[0].sinkType).toBe('kafka')
      expect(JSON.parse(form.routes[0].sinkConfigJson).topic).toBe('orders')

      const result = transformationFormToConfig(form, 'router', {})
      expect(result.routes).toHaveLength(2)
      expect(result.routes[0].condition).toBe("$.type == 'order'")
      expect(result.routes[0].sink.type).toBe('kafka')
      expect(result.routes[0].sink.config.topic).toBe('orders')
      expect(result.routes[1].sink.config.topic).toBe('users')
    })

    it('router with SecretRefs in nested sink uses advanced mode', () => {
      const config = {
        routes: [
          {
            condition: 'true',
            sink: {
              type: 'kafka',
              config: {
                brokersSecretRef: { name: 'kafka-secret', key: 'brokers' },
                topic: 'errors',
              },
            },
          },
        ],
      }
      const form = transformationConfigToForm(config, 'router')
      expect(form.useAdvanced).toBe(true)
      expect(form.advancedJson).toContain('brokersSecretRef')
    })

    it('snakeCase/camelCase round-trip', () => {
      const config = { deep: false }
      const form = transformationConfigToForm(config, 'snakeCase')
      expect(form.deep).toBe(false)
      expect(form.advancedJson).toBe('')

      const result = transformationFormToConfig(form, 'snakeCase', {})
      expect(result).toEqual({ deep: false })

      const configDeep = { deep: true }
      const formDeep = transformationConfigToForm(configDeep, 'camelCase')
      expect(formDeep.deep).toBe(true)
      const resultDeep = transformationFormToConfig(formDeep, 'camelCase', {})
      expect(resultDeep).toEqual({ deep: true })
    })
  })
})
