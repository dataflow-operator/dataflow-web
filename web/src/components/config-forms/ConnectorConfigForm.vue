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
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  modelValue: { type: Object, required: true },
  connectorType: { type: String, required: true },
  role: { type: String, default: 'sink' },
})

const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const form = ref({ ...props.modelValue })
const showAdvanced = ref(false)

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
</style>
