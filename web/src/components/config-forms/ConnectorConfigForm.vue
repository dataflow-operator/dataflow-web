<template>
  <div class="config-form">
    <template v-if="form.useAdvanced">
      <div class="form-group">
        <label>{{ t('flow.advancedConfig') }}</label>
        <textarea
          v-model="form.advancedJson"
          rows="10"
          class="config-json"
          :placeholder="'{}'"
        />
      </div>
    </template>
    <template v-else>
      <!-- Kafka -->
      <template v-if="connectorType === 'kafka'">
        <div class="form-group">
          <label>{{ t('flow.kafkaBrokers') }}</label>
          <input
            v-model="form.brokers"
            type="text"
            class="form-input"
            :placeholder="t('flow.kafkaBrokersPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.kafkaTopic') }}</label>
          <input
            v-model="form.topic"
            type="text"
            class="form-input"
            :placeholder="t('flow.kafkaTopicPlaceholder')"
          />
        </div>
        <div v-if="role === 'source'" class="form-group">
          <label>{{ t('flow.kafkaConsumerGroup') }}</label>
          <input
            v-model="form.consumerGroup"
            type="text"
            class="form-input"
            :placeholder="t('flow.kafkaConsumerGroupPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.kafkaFormat') }}</label>
          <select v-model="form.format" class="form-select">
            <option value="json">JSON</option>
            <option value="avro">Avro</option>
          </select>
        </div>
      </template>

      <!-- PostgreSQL -->
      <template v-else-if="connectorType === 'postgresql'">
        <div class="form-group">
          <label>{{ t('flow.pgConnectionString') }}</label>
          <input
            v-model="form.connectionString"
            type="password"
            class="form-input"
            autocomplete="off"
            :placeholder="t('flow.pgConnectionStringPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgTable') }}</label>
          <input
            v-model="form.table"
            type="text"
            class="form-input"
            :placeholder="t('flow.pgTablePlaceholder')"
          />
        </div>
        <div v-if="role === 'source'" class="form-group">
          <label>{{ t('flow.pgQuery') }}</label>
          <textarea
            v-model="form.query"
            rows="3"
            class="form-input"
            :placeholder="t('flow.pgQueryPlaceholder')"
          />
        </div>
        <template v-if="role === 'source'">
          <p class="form-hint">{{ t('flow.sourceIncrementalHint') }}</p>
          <div class="form-group">
            <label>{{ t('flow.sourceChangeTrackingColumn') }}</label>
            <input
              v-model="form.changeTrackingColumn"
              type="text"
              class="form-input"
              :placeholder="t('flow.sourceChangeTrackingColumnPlaceholderPg')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourceOrderByColumn') }}</label>
            <input
              v-model="form.orderByColumn"
              type="text"
              class="form-input"
              :placeholder="t('flow.sourceOrderByColumnPlaceholder')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourceReadBatchSize') }}</label>
            <input
              v-model.number="form.readBatchSize"
              type="number"
              class="form-input form-input-sm"
              min="0"
              :placeholder="t('flow.sourceReadBatchSizePlaceholder')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourcePollInterval') }}</label>
            <input
              v-model.number="form.pollInterval"
              type="number"
              class="form-input form-input-sm"
              min="1"
              :placeholder="t('flow.sourcePollIntervalPlaceholder')"
            />
          </div>
        </template>
        <div class="form-group form-row">
          <label class="checkbox-label">
            <input v-model="form.autoCreateTable" type="checkbox" />
            {{ t('flow.pgAutoCreateTable') }}
          </label>
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgBatchSize') }}</label>
          <input
            v-model.number="form.batchSize"
            type="number"
            class="form-input form-input-sm"
            min="1"
            :placeholder="t('flow.pgBatchSizePlaceholder')"
          />
        </div>
        <div v-if="role === 'sink'" class="form-group form-row">
          <label class="checkbox-label">
            <input v-model="form.rawMode" type="checkbox" />
            {{ t('flow.pgRawMode') }}
          </label>
        </div>
      </template>

      <!-- PostgreSQL CDC (source only) -->
      <template v-else-if="connectorType === 'postgresql-cdc'">
        <div class="form-group">
          <label>{{ t('flow.pgConnectionString') }}</label>
          <input
            v-model="form.connectionString"
            type="password"
            class="form-input"
            autocomplete="off"
            :placeholder="t('flow.pgConnectionStringPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcSlotName') }}</label>
          <input
            v-model="form.slotName"
            type="text"
            class="form-input"
            :placeholder="t('flow.pgCdcSlotNamePlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcPublicationName') }}</label>
          <input
            v-model="form.publicationName"
            type="text"
            class="form-input"
            :placeholder="t('flow.pgCdcPublicationNamePlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcTables') }}</label>
          <textarea
            v-model="form.tables"
            rows="3"
            class="form-input"
            :placeholder="t('flow.pgCdcTablesPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcSnapshotMode') }}</label>
          <select v-model="form.snapshotMode" class="form-select">
            <option value="initial">initial</option>
            <option value="never">never</option>
            <option value="always">always</option>
          </select>
        </div>
        <div class="form-group form-row">
          <label class="checkbox-label">
            <input v-model="form.createSlotIfNotExists" type="checkbox" />
            {{ t('flow.pgCdcCreateSlot') }}
          </label>
        </div>
        <div class="form-group form-row">
          <label class="checkbox-label">
            <input v-model="form.createPublicationIfNotExists" type="checkbox" />
            {{ t('flow.pgCdcCreatePublication') }}
          </label>
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcHeartbeatInterval') }}</label>
          <input v-model.number="form.heartbeatIntervalSeconds" type="number" class="form-input form-input-sm" min="0" />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcPrimaryKeyColumn') }}</label>
          <input
            v-model="form.primaryKeyColumn"
            type="text"
            class="form-input"
            :placeholder="t('flow.pgCdcPrimaryKeyColumnPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcIncludeColumns') }}</label>
          <textarea v-model="form.includeColumns" rows="2" class="form-input" />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcExcludeColumns') }}</label>
          <textarea v-model="form.excludeColumns" rows="2" class="form-input" />
        </div>
        <div class="form-group">
          <label>{{ t('flow.pgCdcEnvelopeFormat') }}</label>
          <select v-model="form.envelopeFormat" class="form-select">
            <option value="row">row</option>
            <option value="debezium">debezium</option>
          </select>
        </div>
      </template>

      <!-- Trino -->
      <template v-else-if="connectorType === 'trino'">
        <div class="form-group">
          <label>{{ t('flow.trinoServerURL') }}</label>
          <input
            v-model="form.serverURL"
            type="text"
            class="form-input"
            :placeholder="t('flow.trinoServerURLPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.trinoCatalog') }}</label>
          <input
            v-model="form.catalog"
            type="text"
            class="form-input"
            :placeholder="t('flow.trinoCatalogPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.trinoSchema') }}</label>
          <input
            v-model="form.schema"
            type="text"
            class="form-input"
            :placeholder="t('flow.trinoSchemaPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.trinoTable') }}</label>
          <input
            v-model="form.table"
            type="text"
            class="form-input"
            :placeholder="t('flow.trinoTablePlaceholder')"
          />
        </div>
        <div v-if="role === 'source'" class="form-group">
          <label>{{ t('flow.trinoQuery') }}</label>
          <textarea
            v-model="form.query"
            rows="3"
            class="form-input"
            :placeholder="t('flow.trinoQueryPlaceholder')"
          />
        </div>
        <template v-if="role === 'source'">
          <p class="form-hint">{{ t('flow.sourceIncrementalHint') }}</p>
          <div class="form-group">
            <label>{{ t('flow.sourceChangeTrackingColumn') }}</label>
            <input
              v-model="form.changeTrackingColumn"
              type="text"
              class="form-input"
              :placeholder="t('flow.sourceChangeTrackingColumnPlaceholderTrino')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourceOrderByColumn') }}</label>
            <input
              v-model="form.orderByColumn"
              type="text"
              class="form-input"
              :placeholder="t('flow.sourceOrderByColumnPlaceholder')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourceReadBatchSize') }}</label>
            <input
              v-model.number="form.readBatchSize"
              type="number"
              class="form-input form-input-sm"
              min="0"
              :placeholder="t('flow.sourceReadBatchSizePlaceholder')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourcePollInterval') }}</label>
            <input
              v-model.number="form.pollInterval"
              type="number"
              class="form-input form-input-sm"
              min="1"
              :placeholder="t('flow.sourcePollIntervalPlaceholder')"
            />
          </div>
        </template>
      </template>

      <!-- ClickHouse -->
      <template v-else-if="connectorType === 'clickhouse'">
        <div class="form-group">
          <label>{{ t('flow.chConnectionString') }}</label>
          <input
            v-model="form.connectionString"
            type="text"
            class="form-input"
            :placeholder="t('flow.chConnectionStringPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.chTable') }}</label>
          <input
            v-model="form.table"
            type="text"
            class="form-input"
            :placeholder="t('flow.chTablePlaceholder')"
          />
        </div>
        <div v-if="role === 'source'" class="form-group">
          <label>{{ t('flow.chQuery') }}</label>
          <textarea
            v-model="form.query"
            rows="3"
            class="form-input"
            :placeholder="t('flow.chQueryPlaceholder')"
          />
        </div>
        <template v-if="role === 'source'">
          <p class="form-hint">{{ t('flow.sourceIncrementalHint') }}</p>
          <div class="form-group">
            <label>{{ t('flow.sourceChangeTrackingColumn') }}</label>
            <input
              v-model="form.changeTrackingColumn"
              type="text"
              class="form-input"
              :placeholder="t('flow.sourceChangeTrackingColumnPlaceholderCh')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourceOrderByColumn') }}</label>
            <input
              v-model="form.orderByColumn"
              type="text"
              class="form-input"
              :placeholder="t('flow.sourceOrderByColumnPlaceholder')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourceReadBatchSize') }}</label>
            <input
              v-model.number="form.readBatchSize"
              type="number"
              class="form-input form-input-sm"
              min="0"
              :placeholder="t('flow.sourceReadBatchSizePlaceholder')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.sourcePollInterval') }}</label>
            <input
              v-model.number="form.pollInterval"
              type="number"
              class="form-input form-input-sm"
              min="1"
              :placeholder="t('flow.sourcePollIntervalPlaceholder')"
            />
          </div>
        </template>
        <div class="form-group form-row">
          <label class="checkbox-label">
            <input v-model="form.autoCreateTable" type="checkbox" />
            {{ t('flow.chAutoCreateTable') }}
          </label>
        </div>
        <div class="form-group">
          <label>{{ t('flow.chBatchSize') }}</label>
          <input
            v-model.number="form.batchSize"
            type="number"
            class="form-input form-input-sm"
            min="1"
            :placeholder="t('flow.chBatchSizePlaceholder')"
          />
        </div>
        <div v-if="role === 'sink'" class="form-group form-row">
          <label class="checkbox-label">
            <input v-model="form.rawMode" type="checkbox" />
            {{ t('flow.chRawMode') }}
          </label>
        </div>
      </template>

      <!-- Nessie -->
      <template v-else-if="connectorType === 'nessie'">
        <div class="form-group">
          <label>{{ t('flow.nessieBaseURL') }}</label>
          <input
            v-model="form.baseURL"
            type="text"
            class="form-input"
            :placeholder="t('flow.nessieBaseURLPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.nessieNamespace') }}</label>
          <input
            v-model="form.namespace"
            type="text"
            class="form-input"
            :placeholder="t('flow.nessieNamespacePlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.nessieTable') }}</label>
          <input
            v-model="form.table"
            type="text"
            class="form-input"
            :placeholder="t('flow.nessieTablePlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.nessieBranch') }}</label>
          <input
            v-model="form.branch"
            type="text"
            class="form-input"
            :placeholder="t('flow.nessieBranchPlaceholder')"
          />
        </div>
      </template>

      <!-- Iceberg -->
      <template v-else-if="connectorType === 'iceberg'">
        <div class="form-group">
          <label>{{ t('flow.icebergCatalogURI') }}</label>
          <input
            v-model="form.catalogURI"
            type="text"
            class="form-input"
            :placeholder="t('flow.icebergCatalogURIPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.icebergNamespace') }}</label>
          <input
            v-model="form.namespace"
            type="text"
            class="form-input"
            :placeholder="t('flow.icebergNamespacePlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.icebergTable') }}</label>
          <input
            v-model="form.table"
            type="text"
            class="form-input"
            :placeholder="t('flow.icebergTablePlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.icebergPrefix') }}</label>
          <input
            v-model="form.prefix"
            type="text"
            class="form-input"
            :placeholder="t('flow.icebergPrefixPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.icebergWarehouse') }}</label>
          <input
            v-model="form.warehouse"
            type="text"
            class="form-input"
            :placeholder="t('flow.icebergWarehousePlaceholder')"
          />
        </div>
        <div v-if="role === 'source'" class="form-group">
          <label>{{ t('flow.icebergQuery') }}</label>
          <textarea
            v-model="form.query"
            rows="3"
            class="form-input"
            :placeholder="t('flow.icebergQueryPlaceholder')"
          />
        </div>
        <template v-if="role === 'source'">
          <div class="form-group">
            <label>{{ t('flow.sourcePollInterval') }}</label>
            <input
              v-model.number="form.pollInterval"
              type="number"
              class="form-input form-input-sm"
              min="1"
              :placeholder="t('flow.sourcePollIntervalPlaceholder')"
            />
          </div>
          <div class="form-group form-row">
            <label class="checkbox-label">
              <input v-model="form.incrementalBySnapshot" type="checkbox" />
              {{ t('flow.icebergIncrementalBySnapshot') }}
            </label>
          </div>
          <p class="form-hint">{{ t('flow.icebergIncrementalBySnapshotHint') }}</p>
          <div class="form-group">
            <label>{{ t('flow.icebergStartSnapshotID') }}</label>
            <input
              v-model="form.startSnapshotID"
              type="text"
              class="form-input"
              :placeholder="t('flow.icebergStartSnapshotIDPlaceholder')"
            />
          </div>
        </template>
        <template v-if="role === 'sink'">
          <div class="form-group form-row">
            <label class="checkbox-label">
              <input v-model="form.autoCreateTable" type="checkbox" />
              {{ t('flow.icebergAutoCreateTable') }}
            </label>
          </div>
          <div class="form-group">
            <label>{{ t('flow.icebergBatchSize') }}</label>
            <input
              v-model.number="form.batchSize"
              type="number"
              class="form-input form-input-sm"
              min="1"
              :placeholder="t('flow.icebergBatchSizePlaceholder')"
            />
          </div>
          <div class="form-group form-row">
            <label class="checkbox-label">
              <input v-model="form.rawMode" type="checkbox" />
              {{ t('flow.icebergRawMode') }}
            </label>
          </div>
          <div class="form-group">
            <label>{{ t('flow.icebergS3Endpoint') }}</label>
            <input
              v-model="form.s3Endpoint"
              type="text"
              class="form-input"
              :placeholder="t('flow.icebergS3EndpointPlaceholder')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.icebergS3Region') }}</label>
            <input
              v-model="form.s3Region"
              type="text"
              class="form-input"
              :placeholder="t('flow.icebergS3RegionPlaceholder')"
            />
          </div>
        </template>
        <div class="form-group">
          <label>{{ t('flow.icebergAuthenticationType') }}</label>
          <select v-model="form.authenticationType" class="form-select">
            <option value="AUTO">{{ t('flow.icebergAuthAuto') }}</option>
            <option value="BEARER">{{ t('flow.icebergAuthBearer') }}</option>
            <option value="BASIC">{{ t('flow.icebergAuthBasic') }}</option>
            <option value="NONE">{{ t('flow.icebergAuthNone') }}</option>
          </select>
        </div>
        <div v-if="form.authenticationType === 'BEARER' || form.authenticationType === 'AUTO'" class="form-group">
          <label>{{ t('flow.icebergBearerToken') }}</label>
          <input
            v-model="form.bearerToken"
            type="password"
            class="form-input"
            autocomplete="off"
            :placeholder="t('flow.icebergBearerTokenPlaceholder')"
          />
        </div>
        <template v-if="form.authenticationType === 'BASIC' || form.authenticationType === 'AUTO'">
          <div class="form-group">
            <label>{{ t('flow.icebergBasicAuthUsername') }}</label>
            <input
              v-model="form.basicAuthUsername"
              type="text"
              class="form-input"
              :placeholder="t('flow.icebergBasicAuthUsernamePlaceholder')"
            />
          </div>
          <div class="form-group">
            <label>{{ t('flow.icebergBasicAuthPassword') }}</label>
            <input
              v-model="form.basicAuthPassword"
              type="password"
              class="form-input"
              autocomplete="off"
              :placeholder="t('flow.icebergBasicAuthPasswordPlaceholder')"
            />
          </div>
        </template>
        <div class="form-group">
          <label>{{ t('flow.icebergOAuth2ServerURI') }}</label>
          <input
            v-model="form.oauth2ServerURI"
            type="text"
            class="form-input"
            :placeholder="t('flow.icebergOAuth2ServerURIPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.icebergOAuth2ClientID') }}</label>
          <input
            v-model="form.oauth2ClientID"
            type="text"
            class="form-input"
            :placeholder="t('flow.icebergOAuth2ClientIDPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.icebergOAuth2ClientSecret') }}</label>
          <input
            v-model="form.oauth2ClientSecret"
            type="password"
            class="form-input"
            autocomplete="off"
            :placeholder="t('flow.icebergOAuth2ClientSecretPlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.icebergOAuth2Scope') }}</label>
          <input
            v-model="form.oauth2Scope"
            type="text"
            class="form-input"
            :placeholder="t('flow.icebergOAuth2ScopePlaceholder')"
          />
        </div>
      </template>

      <template v-if="showSinkIdempotency">
        <div class="idempotency-section">
          <p v-if="showPollingHint" class="form-hint form-warning">
            {{ t('flow.sinkUpsertPollingHint') }}
          </p>
          <div class="form-group form-row">
            <label class="checkbox-label">
              <input v-model="form.upsertMode" type="checkbox" />
              {{ t('flow.sinkUpsertMode') }}
            </label>
          </div>
          <p class="form-hint">{{ t('flow.sinkUpsertModeHint') }}</p>
          <div v-if="form.upsertMode" class="form-group">
            <label>{{ t('flow.sinkConflictKey') }}</label>
            <input
              v-model="form.conflictKey"
              type="text"
              class="form-input"
              :placeholder="t('flow.sinkConflictKeyPlaceholder')"
            />
            <p class="form-hint">{{ t('flow.sinkConflictKeyHint') }}</p>
          </div>
        </div>
      </template>

      <div class="form-group form-row advanced-toggle">
        <label class="checkbox-label">
          <input v-model="showAdvanced" type="checkbox" />
          {{ t('flow.showAdvanced') }}
        </label>
      </div>
      <div v-if="showAdvanced" class="form-group">
        <label>{{ t('flow.advancedConfig') }}</label>
        <textarea
          v-model="form.advancedJson"
          rows="6"
          class="config-json"
          :placeholder="'{}'"
        />
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { SINK_IDEMPOTENCY_CONNECTORS, POLLING_SOURCE_TYPES } from './useConfigForm'

const props = defineProps({
  modelValue: { type: Object, required: true },
  connectorType: { type: String, required: true },
  role: { type: String, default: 'sink' },
  sourceConnectorType: { type: String, default: '' },
})

const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const form = ref({ ...props.modelValue })
const showAdvanced = ref(false)

const showSinkIdempotency = computed(
  () => props.role === 'sink' && SINK_IDEMPOTENCY_CONNECTORS.includes(props.connectorType)
)

const showPollingHint = computed(
  () => showSinkIdempotency.value
    && POLLING_SOURCE_TYPES.includes(props.sourceConnectorType)
    && !form.value.upsertMode
)

watch(
  () => props.modelValue,
  (v) => {
    form.value = { ...v, advancedJson: v?.advancedJson ?? '' }
    showAdvanced.value = !!(v?.advancedJson && v.advancedJson.trim())
  },
  { deep: true, immediate: true }
)

watch(
  form,
  () => emit('update:modelValue', { ...form.value }),
  { deep: true }
)
</script>

<style scoped>
.config-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-group.form-row {
  flex-direction: row;
  align-items: center;
}

.form-input {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.9rem;
  background: var(--bg-page);
  color: var(--text);
}

.form-input-sm {
  max-width: 120px;
}

.form-select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.9rem;
  background: var(--bg-page);
  color: var(--text);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-weight: normal;
}

.config-json {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
  background: var(--bg-page);
  color: var(--text);
}

.advanced-toggle {
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border);
}

.idempotency-section {
  margin-top: 0.5rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--border);
}

.form-hint {
  margin: 0.25rem 0 0.5rem;
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.form-warning {
  color: var(--warning-text, #b45309);
}
</style>
