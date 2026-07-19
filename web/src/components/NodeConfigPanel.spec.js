import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import NodeConfigPanel from './NodeConfigPanel.vue'

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: {
    en: {
      flow: {
        configSource: 'Configure source',
        configSink: 'Configure sink',
        configTransformation: 'Configure transformation',
        connectorType: 'Connector type',
        useConnection: 'Use connection',
        manualConfig: 'Manual config',
        configJson: 'Config (JSON)',
        transformationType: 'Transformation type',
        advancedConfig: 'Advanced config',
        showAdvanced: 'Show advanced',
        noConfigNeeded: 'No config',
        kafkaBrokers: 'Brokers',
        kafkaBrokersPlaceholder: 'Brokers',
        kafkaTopic: 'Topic',
        kafkaTopicPlaceholder: 'Topic',
        kafkaConsumerGroup: 'Consumer group',
        kafkaConsumerGroupPlaceholder: 'CG',
        kafkaFormat: 'Format',
        pgConnectionString: 'Connection string',
        pgConnectionStringPlaceholder: 'postgres://...',
        pgTable: 'Table',
        pgTablePlaceholder: 'table',
        pgQuery: 'Query',
        pgQueryPlaceholder: 'SELECT',
        pgAutoCreateTable: 'Auto-create',
        pgBatchSize: 'Batch size',
        pgBatchSizePlaceholder: '100',
        pgRawMode: 'Raw mode',
        sinkUpsertMode: 'Upsert mode',
        sinkUpsertModeHint: 'Enable idempotent writes.',
        sinkConflictKey: 'Conflict key',
        sinkConflictKeyPlaceholder: 'id',
        sinkConflictKeyHint: 'Deduplication column.',
        sinkUpsertPollingHint: 'Polling source warning.',
        sourceIncrementalHint: 'Incremental hint',
        sourceChangeTrackingColumn: 'Change tracking column',
        sourceChangeTrackingColumnPlaceholderPg: 'updated_at',
        sourceOrderByColumn: 'Order by column',
        sourceOrderByColumnPlaceholder: 'id',
        sourceReadBatchSize: 'Read batch size',
        sourceReadBatchSizePlaceholder: '0',
        sourcePollInterval: 'Poll interval',
        sourcePollIntervalPlaceholder: '5',
        trinoServerURL: 'Server URL',
        trinoServerURLPlaceholder: 'http://trino:8080',
        trinoCatalog: 'Catalog',
        trinoCatalogPlaceholder: 'hive',
        trinoSchema: 'Schema',
        trinoSchemaPlaceholder: 'default',
        trinoTable: 'Table',
        trinoTablePlaceholder: 'table',
        trinoQuery: 'Query',
        trinoQueryPlaceholder: 'SELECT',
        chConnectionString: 'Connection string',
        chConnectionStringPlaceholder: 'clickhouse://...',
        chTable: 'Table',
        chTablePlaceholder: 'table',
        chQuery: 'Query',
        chQueryPlaceholder: 'SELECT',
        chAutoCreateTable: 'Auto-create',
        chBatchSize: 'Batch size',
        chBatchSizePlaceholder: '100',
        chRawMode: 'Raw mode',
        nessieBaseURL: 'Base URL',
        nessieBaseURLPlaceholder: 'http://nessie:19120',
        nessieNamespace: 'Namespace',
        nessieNamespacePlaceholder: 'schema',
        nessieTable: 'Table',
        nessieTablePlaceholder: 'table',
        nessieBranch: 'Branch',
        nessieBranchPlaceholder: 'main',
        timestampFieldName: 'Field name',
        timestampFieldNamePlaceholder: 'created_at',
        timestampFormat: 'Format',
        flattenField: 'Field',
        flattenFieldPlaceholder: '$.items',
        filterCondition: 'Condition',
        filterConditionPlaceholder: '$.x',
        selectFields: 'Fields',
        selectFieldsPlaceholder: '$.id',
        removeFields: 'Fields',
        removeFieldsPlaceholder: '$.p',
        maskFields: 'Fields',
        maskFieldsPlaceholder: 'p',
        maskKeepLength: 'Keep length',
        transformDeep: 'Deep',
        debeziumInferDeleteFromTombstone: 'Infer tombstone',
        debeziumIncludeSourceInMetadata: 'Source metadata',
        debeziumSnapshotOperation: 'Snapshot op',
        debeziumAddOperationFields: 'Add op fields',
        debeziumAddSourceFields: 'Source fields',
        debeziumAddSourceFieldsPlaceholder: 'table',
        debeziumUnwrapHint: 'Hint',
        replaceFieldRenames: 'Renames',
        replaceFieldRenamesPlaceholder: 'a:b',
        replaceFieldInclude: 'Include',
        replaceFieldIncludePlaceholder: 'id',
        replaceFieldExclude: 'Exclude',
        replaceFieldExcludePlaceholder: 'secret',
        replaceFieldHint: 'Hint',
        headersToPayloadMappings: 'Mappings',
        headersToPayloadMappingsPlaceholder: 'X-Id:id',
        headersToPayloadHint: 'Hint',
        structFlattenDelimiter: 'Delimiter',
        structFlattenDelimiterPlaceholder: '.',
        structFlattenHint: 'Hint',
        extractFieldField: 'Field',
        extractFieldFieldPlaceholder: 'payload.after',
        extractFieldHint: 'Hint',
        hoistFieldField: 'Wrapper key',
        hoistFieldFieldPlaceholder: 'record',
        hoistFieldHint: 'Hint',
        castSpec: 'Casts',
        castSpecPlaceholder: 'id:int64',
        castHint: 'Hint',
        timezoneTimezone: 'Target timezone',
        timezoneTimezonePlaceholder: 'Europe/Moscow',
        timezoneFields: 'Fields',
        timezoneFieldsPlaceholder: 'created_at',
        timezoneSourceTimezone: 'Source timezone',
        timezoneSourceTimezonePlaceholder: 'UTC',
        timezoneFormat: 'Format',
        timezoneFormatPlaceholder: 'RFC3339Nano',
        timezoneHint: 'Hint',
      },
      common: { cancel: 'Cancel', save: 'Save' },
    },
  },
})

describe('NodeConfigPanel', () => {
  function mountPanel(props) {
    return mount(NodeConfigPanel, {
      props,
      global: {
        plugins: [i18n],
        stubs: {
          teleport: true,
        },
      },
    })
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders source config panel', () => {
    const wrapper = mountPanel({
        node: {
          id: 'source',
          type: 'source',
          data: {
            connectorType: 'kafka',
            config: { brokers: ['localhost:9092'], topic: 'input', consumerGroup: 'group' },
          },
        },
        nodeType: 'source',
        connections: [],
    })
    expect(wrapper.text()).toContain('Configure source')
    expect(wrapper.text()).toContain('Connector type')
    expect(wrapper.text()).toContain('Use connection')
    expect(wrapper.find('select').element.value).toBe('kafka')
  })

  it('renders sink config panel', () => {
    const wrapper = mountPanel({
        node: {
          id: 'sink',
          type: 'sink',
          data: {
            connectorType: 'postgresql',
            config: { connectionString: 'postgres://...', table: 'out' },
          },
        },
        nodeType: 'sink',
        connections: [],
    })
    expect(wrapper.text()).toContain('Configure sink')
    expect(wrapper.find('select').element.value).toBe('postgresql')
  })

  it('renders transformation config panel', () => {
    const wrapper = mountPanel({
        node: {
          id: 't-0',
          type: 'transformation',
          data: {
            transformationType: 'timestamp',
            config: { fieldName: 'created_at', format: 'RFC3339' },
          },
        },
        nodeType: 'transformation',
        connections: [],
    })
    expect(wrapper.text()).toContain('Configure transformation')
    expect(wrapper.text()).toContain('Transformation type')
    expect(wrapper.find('select').element.value).toBe('timestamp')
  })

  it('emits save with structured config for Kafka source', async () => {
    const wrapper = mountPanel({
        node: {
          id: 'source',
          type: 'source',
          data: {
            connectorType: 'kafka',
            config: { brokers: ['kafka:9092'], topic: 't', consumerGroup: 'cg' },
          },
        },
        nodeType: 'source',
        connections: [],
    })

    await wrapper.find('button.btn.btn-primary').trigger('click')

    expect(wrapper.emitted('save')).toHaveLength(1)
    const [saved] = wrapper.emitted('save')[0]
    expect(saved.data.connectorType).toBe('kafka')
    expect(saved.data.config.brokers).toEqual(['kafka:9092'])
    expect(saved.data.config.topic).toBe('t')
    expect(saved.data.config.consumerGroup).toBe('cg')
  })

  it('emits save with structured config for timestamp transformation', async () => {
    const wrapper = mountPanel({
        node: {
          id: 't-0',
          type: 'transformation',
          data: {
            transformationType: 'timestamp',
            config: { fieldName: 'processed_at', format: 'RFC3339' },
          },
        },
        nodeType: 'transformation',
        connections: [],
    })

    await wrapper.find('button.btn.btn-primary').trigger('click')

    expect(wrapper.emitted('save')).toHaveLength(1)
    const [saved] = wrapper.emitted('save')[0]
    expect(saved.data.transformationType).toBe('timestamp')
    expect(saved.data.config.fieldName).toBe('processed_at')
  })

  it('emits close when cancel clicked', async () => {
    const wrapper = mountPanel({
        node: {
          id: 'source',
          type: 'source',
          data: { connectorType: 'kafka', config: {} },
        },
        nodeType: 'source',
        connections: [],
    })

    await wrapper.find('button.btn.btn-secondary').trigger('click')

    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('shows connection options when connector type matches', () => {
    const wrapper = mountPanel({
        node: {
          id: 'source',
          type: 'source',
          data: { connectorType: 'kafka', config: {} },
        },
        nodeType: 'source',
        connections: [
          {
            metadata: {
              name: 'kafka-creds',
              labels: { 'dataflow.dataflow.io/connection-type': 'kafka' },
            },
          },
        ],
    })

    const selects = wrapper.findAll('select')
    expect(selects.length).toBeGreaterThan(1)
    const connectionSelect = selects[1]
    expect(connectionSelect.text()).toContain('kafka-creds')
  })
})
