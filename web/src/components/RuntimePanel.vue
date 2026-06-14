<template>
  <div class="card runtime">
    <div class="card-header runtime-header">
      <h3>{{ t('metrics.runtimeTitle') }}</h3>
      <button
        v-if="canResetCheckpoint"
        type="button"
        class="btn btn-secondary btn-sm"
        :disabled="resetting"
        @click="confirmOpen = true"
      >
        {{ resetting ? t('metrics.resettingCheckpoint') : t('metrics.resetCheckpoint') }}
      </button>
    </div>

    <LoadingSpinner v-if="loading" :message="t('metrics.loadingRuntime')" />
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else-if="!runtime" class="empty-state">
      {{ t('metrics.noRuntime') }}
    </div>
    <div v-else class="runtime-body">
      <div v-if="resetSuccess" class="success-message">{{ resetSuccess }}</div>
      <div v-if="resetError" class="error-message">{{ resetError }}</div>

      <div class="runtime-grid">
        <div class="metric-card">
          <h4>{{ t('metrics.checkpointPersistence') }}</h4>
          <div class="value">{{ runtime.checkpointPersistence ? t('common.enabled') : t('common.disabled') }}</div>
        </div>
        <div class="metric-card">
          <h4>{{ t('metrics.checkpointConfigMap') }}</h4>
          <div class="value value-sm">{{ runtime.checkpointConfigMap }}</div>
        </div>
        <div class="metric-card">
          <h4>{{ t('metrics.processorReady') }}</h4>
          <div class="value">
            {{ runtime.processor?.readyReplicas ?? 0 }} / {{ runtime.processor?.replicas ?? 0 }}
          </div>
        </div>
      </div>

      <div v-if="runtime.checkpointMessage" class="hint">
        {{ runtime.checkpointMessage }}
      </div>

      <details class="section" v-if="runtime.checkpoints && Object.keys(runtime.checkpoints).length">
        <summary>{{ t('metrics.checkpoints') }}</summary>
        <div class="checkpoints">
          <div v-for="(v, k) in runtime.checkpoints" :key="k" class="checkpoint">
            <div class="checkpoint-title">{{ k }}</div>
            <pre class="checkpoint-json">{{ pretty(v) }}</pre>
          </div>
        </div>
      </details>

      <details class="section" v-if="runtime.processor?.pods?.length">
        <summary>{{ t('metrics.processorPods') }}</summary>
        <div class="pods">
          <div v-for="p in runtime.processor.pods" :key="p.name" class="pod">
            <div class="pod-name">{{ p.name }}</div>
            <div class="pod-meta">
              {{ t('metrics.podPhase') }}: {{ p.phase }} ·
              {{ t('metrics.podReady') }}: {{ p.ready }} ·
              {{ t('metrics.podRestarts') }}: {{ p.restarts }} ·
              {{ t('metrics.podNode') }}: {{ p.node || '—' }}
            </div>
          </div>
        </div>
      </details>

      <details class="section" v-if="runtime.conditions?.length">
        <summary>{{ t('metrics.conditions') }}</summary>
        <pre class="checkpoint-json">{{ pretty(runtime.conditions) }}</pre>
      </details>
    </div>

    <ConfirmModal
      :open="confirmOpen"
      :title="t('metrics.resetCheckpointConfirmTitle')"
      :message="t('metrics.resetCheckpointConfirmMessage')"
      :confirm-label="t('metrics.resetCheckpoint')"
      @cancel="confirmOpen = false"
      @confirm="onConfirmReset"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from './LoadingSpinner.vue'
import ConfirmModal from './ConfirmModal.vue'
import { getDataFlow, updateDataFlow } from '../api/client'
import { mergeManifestForUpdate } from '../utils/manifest'

const props = defineProps({
  runtime: { type: Object, default: null },
  loading: { type: Boolean, default: false },
  error: { type: String, default: '' },
  namespace: { type: String, default: '' },
  dataflowName: { type: String, default: '' },
})

const emit = defineEmits(['checkpoint-reset'])

const { t } = useI18n()

const confirmOpen = ref(false)
const resetting = ref(false)
const resetError = ref('')
const resetSuccess = ref('')

const canResetCheckpoint = computed(
  () => !!props.namespace && !!props.dataflowName && !props.loading
)

watch(
  () => [props.namespace, props.dataflowName],
  () => {
    resetError.value = ''
    resetSuccess.value = ''
  }
)

async function onConfirmReset() {
  confirmOpen.value = false
  if (!props.namespace || !props.dataflowName) return

  resetting.value = true
  resetError.value = ''
  resetSuccess.value = ''

  try {
    const current = await getDataFlow(props.namespace, props.dataflowName)
    const updated = mergeManifestForUpdate(
      {
        ...current,
        spec: {
          ...current.spec,
          checkpointReset: true,
        },
      },
      current
    )
    await updateDataFlow(props.namespace, props.dataflowName, updated)
    resetSuccess.value = t('metrics.resetCheckpointSuccess')
    emit('checkpoint-reset')
  } catch (e) {
    resetError.value = e.message
  } finally {
    resetting.value = false
  }
}

function pretty(obj) {
  try {
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(obj)
  }
}
</script>

<style scoped>
.runtime-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
}

.runtime-body {
  padding: 0.5rem 1rem 1rem;
}

.runtime-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.hint {
  margin-top: 0.75rem;
  color: var(--text-muted);
}

.success-message {
  margin-bottom: 0.75rem;
  color: var(--success, #15803d);
}

.error-message {
  margin-bottom: 0.75rem;
}

.section {
  margin-top: 1rem;
}

.checkpoints {
  margin-top: 0.5rem;
  display: grid;
  gap: 0.75rem;
}

.checkpoint-title {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.checkpoint-json {
  font-size: 0.85rem;
  overflow: auto;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  padding: 0.5rem;
  border-radius: 6px;
}

.pods {
  margin-top: 0.5rem;
  display: grid;
  gap: 0.5rem;
}

.pod-name {
  font-weight: 600;
}

.pod-meta {
  color: var(--text-muted);
  font-size: 0.9rem;
}
</style>
