<template>
  <div class="card runtime">
    <div class="card-header">
      <h3>{{ t('metrics.runtimeTitle') }}</h3>
    </div>

    <LoadingSpinner v-if="loading" :message="t('metrics.loadingRuntime')" />
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else-if="!runtime" class="empty-state">
      {{ t('metrics.noRuntime') }}
    </div>
    <div v-else class="runtime-body">
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
  </div>
</template>

<script setup>
import LoadingSpinner from './LoadingSpinner.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  runtime: { type: Object, default: null },
  loading: { type: Boolean, default: false },
  error: { type: String, default: '' },
})

function pretty(obj) {
  try {
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(obj)
  }
}
</script>

<style scoped>
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

