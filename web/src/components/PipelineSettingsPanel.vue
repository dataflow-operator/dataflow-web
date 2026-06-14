<template>
  <aside class="pipeline-settings-panel" role="complementary">
    <div class="panel-header">
      <h3 class="panel-title">{{ t('flow.pipelineSettingsTitle') }}</h3>
      <button
        type="button"
        class="panel-close"
        :aria-label="t('modal.closeAria')"
        @click="$emit('close')"
      >
        &times;
      </button>
    </div>
    <div class="panel-body">
      <section class="settings-section">
        <h4 class="section-title">{{ t('flow.pipelineSettingsCheckpoint') }}</h4>
        <div class="form-group checkbox-group">
          <label class="checkbox-label">
            <input v-model="local.checkpointPersistence" type="checkbox" />
            {{ t('flow.pipelineSettingsCheckpointPersistence') }}
          </label>
          <p class="field-hint">{{ t('flow.pipelineSettingsCheckpointPersistenceHint') }}</p>
        </div>
        <div class="form-group checkbox-group">
          <label class="checkbox-label">
            <input v-model="local.checkpointSyncOnAck" type="checkbox" />
            {{ t('flow.pipelineSettingsCheckpointSyncOnAck') }}
          </label>
          <p class="field-hint">{{ t('flow.pipelineSettingsCheckpointSyncOnAckHint') }}</p>
        </div>
        <div class="form-group">
          <label for="checkpoint-save-interval">{{ t('flow.pipelineSettingsCheckpointSaveInterval') }}</label>
          <input
            id="checkpoint-save-interval"
            v-model="local.checkpointSaveInterval"
            type="text"
            :placeholder="t('flow.pipelineSettingsCheckpointSaveIntervalPlaceholder')"
            :class="{ 'input-invalid': intervalInvalid }"
          />
          <p v-if="intervalInvalid" class="field-error">
            {{ t('flow.pipelineSettingsInvalidDuration') }}
          </p>
          <p v-else class="field-hint">{{ t('flow.pipelineSettingsCheckpointSaveIntervalHint') }}</p>
        </div>
        <div class="form-group checkbox-group">
          <label class="checkbox-label">
            <input v-model="local.checkpointReset" type="checkbox" />
            {{ t('flow.pipelineSettingsCheckpointReset') }}
          </label>
          <p class="field-hint field-warning">{{ t('flow.pipelineSettingsCheckpointResetHint') }}</p>
        </div>
      </section>

      <section class="settings-section">
        <h4 class="section-title">{{ t('flow.pipelineSettingsProcessing') }}</h4>
        <div class="form-group checkbox-group">
          <label class="checkbox-label">
            <input v-model="local.strictIdempotency" type="checkbox" />
            {{ t('flow.pipelineSettingsStrictIdempotency') }}
          </label>
          <p class="field-hint">{{ t('flow.pipelineSettingsStrictIdempotencyHint') }}</p>
        </div>
        <div class="form-group">
          <label for="ack-granularity">{{ t('flow.pipelineSettingsAckGranularity') }}</label>
          <select id="ack-granularity" v-model="local.ackGranularity">
            <option v-for="opt in ackGranularityOptions" :key="opt" :value="opt">
              {{ t(`flow.pipelineSettingsAckGranularity_${opt}`) }}
            </option>
          </select>
          <p class="field-hint">{{ t('flow.pipelineSettingsAckGranularityHint') }}</p>
        </div>
        <div class="form-group">
          <label for="channel-buffer-size">{{ t('flow.pipelineSettingsChannelBufferSize') }}</label>
          <input
            id="channel-buffer-size"
            v-model.number="local.channelBufferSize"
            type="number"
            min="1"
            :placeholder="t('flow.pipelineSettingsChannelBufferSizePlaceholder')"
          />
          <p class="field-hint">{{ t('flow.pipelineSettingsChannelBufferSizeHint') }}</p>
        </div>
      </section>

      <section class="settings-section">
        <h4 class="section-title">{{ t('flow.pipelineSettingsScaling') }}</h4>
        <div class="form-group">
          <label for="replicas">{{ t('flow.pipelineSettingsReplicas') }}</label>
          <input
            id="replicas"
            v-model.number="local.replicas"
            type="number"
            min="0"
            :disabled="!replicasEnabled"
          />
          <p v-if="!replicasEnabled" class="field-hint">
            {{ t('flow.pipelineSettingsReplicasKafkaOnly') }}
          </p>
          <p v-else class="field-hint">{{ t('flow.pipelineSettingsReplicasHint') }}</p>
        </div>
      </section>
    </div>
  </aside>
</template>

<script setup>
import { computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ACK_GRANULARITY_OPTIONS,
  createDefaultPipelineSettings,
  isReplicasSupported,
  isValidDuration,
  formatDurationForInput,
} from '../composables/usePipelineSettings'

const props = defineProps({
  modelValue: { type: Object, default: () => createDefaultPipelineSettings() },
  sourceType: { type: String, default: 'kafka' },
})

const emit = defineEmits(['update:modelValue', 'close'])

const { t } = useI18n()

const ackGranularityOptions = ACK_GRANULARITY_OPTIONS

const local = reactive(createDefaultPipelineSettings())
let syncingFromProps = false

function syncFromProps(value) {
  syncingFromProps = true
  const next = value || createDefaultPipelineSettings()
  Object.assign(local, {
    ...createDefaultPipelineSettings(),
    ...next,
    checkpointSaveInterval: formatDurationForInput(next.checkpointSaveInterval),
  })
  syncingFromProps = false
}

watch(
  () => props.modelValue,
  (value) => syncFromProps(value),
  { immediate: true, deep: true }
)

watch(
  local,
  () => {
    if (syncingFromProps) return
    emit('update:modelValue', { ...local })
  },
  { deep: true }
)

const replicasEnabled = computed(() => isReplicasSupported(props.sourceType))

const intervalInvalid = computed(() => {
  const value = formatDurationForInput(local.checkpointSaveInterval)
  return value !== '' && !isValidDuration(value)
})

watch(replicasEnabled, (enabled) => {
  if (!enabled) {
    local.replicas = 1
  }
})
</script>

<style scoped>
.pipeline-settings-panel {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-left: 1px solid var(--border);
  background: var(--bg-card);
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border);
  background: linear-gradient(135deg, var(--gradient-start), var(--gradient-end));
  color: white;
}

.panel-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
}

.panel-close {
  background: none;
  border: none;
  color: white;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.panel-body {
  flex: 1;
  overflow: auto;
  padding: 1rem;
}

.settings-section {
  margin-bottom: 1.25rem;
}

.settings-section:last-child {
  margin-bottom: 0;
}

.section-title {
  margin: 0 0 0.75rem;
  font-size: 0.85rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-muted);
}

.checkbox-group {
  margin-bottom: 0.75rem;
}

.checkbox-label {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-weight: 500;
  color: var(--text);
  cursor: pointer;
}

.checkbox-label input[type='checkbox'] {
  width: auto;
  max-width: none;
  margin-top: 0.15rem;
  flex-shrink: 0;
}

.panel-body .form-group input,
.panel-body .form-group select {
  max-width: none;
}

.field-hint {
  margin: 0.35rem 0 0;
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.field-warning {
  color: var(--warning-text, #b45309);
}

.field-error {
  margin: 0.35rem 0 0;
  font-size: 0.8rem;
  color: var(--error-text);
}

.input-invalid {
  border-color: var(--error-text) !important;
}
</style>
