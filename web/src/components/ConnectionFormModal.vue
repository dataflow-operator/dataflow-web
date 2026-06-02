<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-overlay"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.self="close"
    >
      <div class="modal-content connection-form-modal">
        <div class="modal-header">
          <h2 :id="titleId" class="modal-title">{{ title }}</h2>
          <button
            type="button"
            class="modal-close"
            :aria-label="t('modal.closeAria')"
            @click="close"
          >
            &times;
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label for="conn-name">{{ t('connections.name') }}</label>
            <input
              id="conn-name"
              v-model="formData.name"
              type="text"
              :disabled="mode === 'edit'"
              placeholder="e.g. kafka-credentials"
            />
          </div>
          <div class="form-group">
            <label for="conn-type">{{ t('connections.type') }}</label>
            <select id="conn-type" v-model="formData.connectionType" @change="onTypeChange">
              <option value="">{{ t('connections.typePlaceholder') }}</option>
              <option
                v-for="(label, key) in connectionTypes"
                :key="key"
                :value="key"
              >
                {{ label }}
              </option>
            </select>
          </div>
          <template v-if="formData.connectionType">
            <div
              v-for="field in currentFields"
              :key="field.key"
              class="form-group"
            >
              <label :for="'field-' + field.key">{{ field.label }}</label>
              <input
                :id="'field-' + field.key"
                v-model="formData.stringData[field.key]"
                :type="field.sensitive ? 'password' : 'text'"
                :placeholder="field.placeholder"
              />
            </div>
          </template>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="close">
            {{ t('common.cancel') }}
          </button>
          <button
            v-if="mode === 'edit'"
            type="button"
            class="btn btn-primary"
            @click="save"
          >
            {{ t('common.save') }}
          </button>
          <button
            v-if="mode === 'create'"
            type="button"
            class="btn btn-primary"
            @click="create"
          >
            {{ t('common.create') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const CONNECTION_TYPE_LABEL = 'dataflow.dataflow.io/connection-type'

const CONNECTION_FIELDS = {
  kafka: [
    { key: 'brokers', label: 'Brokers (comma-separated)', placeholder: 'localhost:9092', sensitive: false },
    { key: 'topic', label: 'Topic', placeholder: 'input-topic', sensitive: false },
    { key: 'consumerGroup', label: 'Consumer Group', placeholder: 'dataflow-group', sensitive: false },
    { key: 'username', label: 'SASL Username', placeholder: '', sensitive: false },
    { key: 'password', label: 'SASL Password', placeholder: '', sensitive: true },
  ],
  postgresql: [
    { key: 'connectionString', label: 'Connection String', placeholder: 'postgres://user:pass@host:5432/db', sensitive: true },
    { key: 'table', label: 'Table', placeholder: 'output_table', sensitive: false },
  ],
  trino: [
    { key: 'serverURL', label: 'Server URL', placeholder: 'http://trino:8080', sensitive: false },
    { key: 'catalog', label: 'Catalog', placeholder: 'hive', sensitive: false },
    { key: 'schema', label: 'Schema', placeholder: 'default', sensitive: false },
    { key: 'table', label: 'Table', placeholder: 'table_name', sensitive: false },
  ],
  clickhouse: [
    { key: 'connectionString', label: 'Connection String', placeholder: 'clickhouse://host:9000/default', sensitive: true },
    { key: 'table', label: 'Table', placeholder: 'table_name', sensitive: false },
  ],
  nessie: [
    { key: 'baseURL', label: 'Base URL', placeholder: 'http://nessie:19120/api/v1', sensitive: false },
    { key: 'token', label: 'Bearer Token', placeholder: '', sensitive: true },
    { key: 'namespace', label: 'Namespace', placeholder: 'default', sensitive: false },
    { key: 'table', label: 'Table', placeholder: 'table_name', sensitive: false },
  ],
}

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  secret: { type: Object, default: null },
  mode: { type: String, default: 'edit' },
  namespace: { type: String, default: 'default' },
})

const emit = defineEmits(['close', 'save', 'create'])

const { t } = useI18n()

const titleId = computed(() => 'modal-title-' + Math.random().toString(36).slice(2))

const formData = ref({
  name: '',
  connectionType: '',
  stringData: {},
})

const connectionTypes = computed(() => ({
  kafka: t('connections.connectionTypes.kafka'),
  postgresql: t('connections.connectionTypes.postgresql'),
  trino: t('connections.connectionTypes.trino'),
  clickhouse: t('connections.connectionTypes.clickhouse'),
  nessie: t('connections.connectionTypes.nessie'),
}))

const currentFields = computed(() => {
  const type = formData.value.connectionType
  return CONNECTION_FIELDS[type] || []
})

function onTypeChange() {
  const type = formData.value.connectionType
  const existing = formData.value.stringData
  const newData = {}
  const fields = CONNECTION_FIELDS[type] || []
  for (const f of fields) {
    newData[f.key] = existing[f.key] ?? ''
  }
  formData.value.stringData = newData
}

function close() {
  emit('close')
}

function buildSecret() {
  const ns = props.namespace || 'default'
  const name = formData.value.name?.trim()
  if (!name) return null
  const stringData = { ...formData.value.stringData }
  for (const k of Object.keys(stringData)) {
    const v = stringData[k]
    if (v === '') delete stringData[k]
    else if (v === '****' && props.mode === 'create') delete stringData[k]
  }
  // In edit mode, "****" is sent so backend preserves existing sensitive values
  return {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name,
      namespace: ns,
      labels: {
        [CONNECTION_TYPE_LABEL]: formData.value.connectionType,
      },
    },
    type: 'Opaque',
    stringData,
  }
}

function save() {
  const secret = buildSecret()
  if (!secret) {
    emit('save', null)
    return
  }
  emit('save', secret)
}

function create() {
  const secret = buildSecret()
  if (!secret) {
    emit('create', null)
    return
  }
  emit('create', secret)
}

watch(
  () => [props.open, props.secret],
  () => {
    if (!props.open) return
    const s = props.secret
    if (s) {
      formData.value = {
        name: s.metadata?.name || '',
        connectionType: s.metadata?.labels?.[CONNECTION_TYPE_LABEL] || '',
        stringData: { ...(s.stringData || {}) },
      }
      if (!formData.value.connectionType && Object.keys(formData.value.stringData || {}).length > 0) {
        formData.value.connectionType = 'kafka'
        onTypeChange()
      }
    } else {
      formData.value = {
        name: '',
        connectionType: '',
        stringData: {},
      }
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.modal-content {
  background: var(--bg-card);
  border-radius: 8px;
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

.modal-header {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, var(--gradient-start), var(--gradient-end));
  color: white;
  border-radius: 8px 8px 0 0;
}

.modal-title {
  margin: 0;
  font-size: 1.25rem;
}

.modal-close {
  background: none;
  border: none;
  color: white;
  font-size: 1.75rem;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  border-radius: 4px;
}

.modal-close:hover {
  opacity: 0.85;
}

.modal-body {
  padding: 1rem 1.5rem;
  overflow: auto;
  flex: 1;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.35rem;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.95rem;
  background: var(--bg-card);
  color: var(--text);
}

.form-group input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>
