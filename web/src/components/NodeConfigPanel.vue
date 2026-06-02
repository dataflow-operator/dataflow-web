<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-overlay"
      role="dialog"
      aria-modal="true"
      @click.self="$emit('close')"
    >
      <div class="modal-content node-config-modal">
        <div class="modal-header">
          <h2 class="modal-title">{{ panelTitle }}</h2>
          <button type="button" class="modal-close" @click="$emit('close')">&times;</button>
        </div>
        <div class="modal-body">
          <template v-if="nodeType === 'source' || nodeType === 'sink'">
            <div class="form-group">
              <label>{{ t('flow.connectorType') }}</label>
              <select v-model="form.connectorType" @change="onConnectorTypeChange">
                <option
                  v-for="ct in connectorTypes"
                  :key="ct"
                  :value="ct"
                >
                  {{ ct }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>{{ t('flow.useConnection') }}</label>
              <select v-model="form.connectionRef" @change="onConnectionSelect">
                <option value="">{{ t('flow.manualConfig') }}</option>
                <option
                  v-for="c in connectionOptions"
                  :key="c.metadata?.name"
                  :value="c.metadata?.name"
                >
                  {{ c.metadata?.name }} ({{ c.metadata?.labels?.[CONNECTION_TYPE_LABEL] || '—' }})
                </option>
              </select>
            </div>
            <template v-if="!form.connectionRef">
              <ConnectorConfigForm
                v-model="form.structuredConfig"
                :connector-type="form.connectorType"
                :role="nodeType"
              />
            </template>
            <div v-else class="form-group">
              <label>{{ t('flow.configJson') }}</label>
              <textarea
                v-model="form.configJson"
                rows="8"
                class="config-json"
                :placeholder="'{}'"
              />
            </div>
          </template>
          <template v-else-if="nodeType === 'transformation'">
            <div class="form-group">
              <label>{{ t('flow.transformationType') }}</label>
              <select v-model="form.transformationType" @change="onTransformationTypeChange">
                <option
                  v-for="tt in transformationTypes"
                  :key="tt"
                  :value="tt"
                >
                  {{ tt }}
                </option>
              </select>
            </div>
            <TransformationConfigForm
              v-model="form.structuredConfig"
              :transformation-type="form.transformationType"
            />
          </template>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="$emit('close')">
            {{ t('common.cancel') }}
          </button>
          <button type="button" class="btn btn-primary" @click="save">
            {{ t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  CONNECTOR_TYPES,
  TRANSFORMATION_TYPES,
} from '../composables/useFlowManifest'
import ConnectorConfigForm from './config-forms/ConnectorConfigForm.vue'
import TransformationConfigForm from './config-forms/TransformationConfigForm.vue'
import {
  configToForm,
  formToConfig,
  transformationConfigToForm,
  transformationFormToConfig,
} from './config-forms/useConfigForm'

const CONNECTION_TYPE_LABEL = 'dataflow.dataflow.io/connection-type'

const props = defineProps({
  node: { type: Object, default: null },
  nodeType: { type: String, default: '' },
  connections: { type: Array, default: () => [] },
  namespace: { type: String, default: 'default' },
})

const emit = defineEmits(['close', 'save'])

const { t } = useI18n()

const open = computed(() => !!props.node)

const form = ref({
  connectorType: 'kafka',
  transformationType: 'timestamp',
  connectionRef: '',
  configJson: '{}',
  structuredConfig: {},
})

const panelTitle = computed(() => {
  if (props.nodeType === 'source') return t('flow.configSource')
  if (props.nodeType === 'sink') return t('flow.configSink')
  return t('flow.configTransformation')
})

const connectorTypes = CONNECTOR_TYPES
const transformationTypes = TRANSFORMATION_TYPES

const connectionOptions = computed(() => {
  const type = form.value.connectorType
  return props.connections.filter(
    (c) => (c.metadata?.labels?.[CONNECTION_TYPE_LABEL] || '') === type
  )
})

function buildSecretRef(secretName, key) {
  return { name: secretName, key }
}

function onConnectorTypeChange() {
  form.value.connectionRef = ''
  form.value.structuredConfig = configToForm(
    {},
    form.value.connectorType,
    props.nodeType
  )
}

function onTransformationTypeChange() {
  form.value.structuredConfig = transformationConfigToForm(
    {},
    form.value.transformationType
  )
}

function onConnectionSelect() {
  const name = form.value.connectionRef
  if (!name) return
  const conn = props.connections.find((c) => c.metadata?.name === name)
  if (!conn?.keys) return
  const config = {}
  const saslKeys = ['username', 'password']
  for (const key of conn.keys) {
    const ref = buildSecretRef(name, key)
    if (saslKeys.includes(key.toLowerCase())) {
      if (!config.sasl) config.sasl = { mechanism: 'plain' }
      config.sasl[`${key}SecretRef`] = ref
    } else {
      const camelKey = key.charAt(0).toLowerCase() + key.slice(1)
      config[`${camelKey}SecretRef`] = ref
    }
  }
  form.value.configJson = JSON.stringify(config, null, 2)
}

function getPlainNodeData(node) {
  if (!node?.data) return {}
  try {
    return JSON.parse(JSON.stringify(node.data))
  } catch {
    return {}
  }
}

function parseAdvancedConfig(advancedJson) {
  if (!advancedJson?.trim()) return {}
  try {
    return JSON.parse(advancedJson) || {}
  } catch {
    return {}
  }
}

watch(
  () => [props.node, props.nodeType],
  () => {
    const n = props.node
    if (!n) return
    const data = getPlainNodeData(n)
    if (props.nodeType === 'source' || props.nodeType === 'sink') {
      const connectorType = data.connectorType || 'kafka'
      form.value.connectorType = connectorType
      form.value.connectionRef = ''
      form.value.structuredConfig = configToForm(
        data.config || {},
        connectorType,
        props.nodeType
      )
      form.value.configJson = '{}'
    } else {
      const transformationType = data.transformationType || 'timestamp'
      form.value.transformationType = transformationType
      form.value.structuredConfig = transformationConfigToForm(
        data.config || {},
        transformationType
      )
    }
  },
  { immediate: true }
)

function save() {
  if (props.nodeType === 'source' || props.nodeType === 'sink') {
    let config
    if (form.value.connectionRef) {
      try {
        config = form.value.configJson?.trim()
          ? JSON.parse(form.value.configJson)
          : {}
      } catch {
        config = {}
      }
    } else {
      const sc = form.value.structuredConfig || {}
      const advanced = parseAdvancedConfig(sc.advancedJson)
      config = formToConfig(
        sc,
        form.value.connectorType,
        props.nodeType,
        advanced
      )
    }
    emit('save', {
      ...props.node,
      data: {
        ...props.node.data,
        label: form.value.connectorType,
        connectorType: form.value.connectorType,
        config,
      },
    })
  } else {
    const sc = form.value.structuredConfig || {}
    const advanced = parseAdvancedConfig(sc.advancedJson)
    const config = transformationFormToConfig(
      sc,
      form.value.transformationType,
      advanced
    )
    emit('save', {
      ...props.node,
      data: {
        ...props.node.data,
        label: form.value.transformationType,
        transformationType: form.value.transformationType,
        config,
      },
    })
  }
  emit('close')
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1001;
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
  max-width: 560px;
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
}

.modal-body {
  padding: 1rem 1.5rem;
  overflow: auto;
  flex: 1;
}

.modal-body .form-group {
  margin-bottom: 0.75rem;
}

.modal-body .form-group:last-child {
  margin-bottom: 0;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.config-json {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  background: var(--bg-page);
  color: var(--text);
}
</style>
