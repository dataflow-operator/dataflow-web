<template>
  <div class="metrics">
    <div class="card">
      <div class="card-header">
        <h2>{{ t('metrics.title') }}</h2>
      </div>
      <div class="form-group">
        <label>{{ t('common.namespace') }}</label>
        <select v-model="namespace" class="form-select" @change="loadDataFlowList">
          <option value="">{{ t('common.loading') }}</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>
      <div class="form-group">
        <label>{{ t('common.dataflow') }}</label>
        <select v-model="selectedName" class="form-select" @change="loadMetrics">
          <option value="">{{ t('metrics.selectDataflow') }}</option>
          <option v-for="df in dataflowList" :key="df.metadata.name" :value="df.metadata.name">
            {{ df.metadata.name }}
          </option>
        </select>
      </div>
      <div v-if="!selectedName" class="empty-state">
        <p>{{ t('metrics.selectHint') }}</p>
      </div>
      <LoadingSpinner v-else-if="loading" :message="t('metrics.loading')" />
      <div v-else-if="error" class="error-message">{{ error }}</div>
      <template v-else-if="status">
        <div class="metrics-grid">
          <div class="metric-card">
            <h3>{{ t('metrics.status') }}</h3>
            <div class="value">{{ status.phase || 'Unknown' }}</div>
          </div>
          <div class="metric-card">
            <h3>{{ t('metrics.processedMessages') }}</h3>
            <div class="value">{{ status.processedCount ?? 0 }}</div>
          </div>
          <div class="metric-card">
            <h3>{{ t('metrics.errors') }}</h3>
            <div class="value">{{ status.errorCount ?? 0 }}</div>
          </div>
          <div class="metric-card">
            <h3>{{ t('metrics.lastProcessed') }}</h3>
            <div class="value value-sm">
              {{ lastProcessedFormatted }}
            </div>
          </div>
        </div>
        <div class="card status-message">
          <h3>{{ t('metrics.statusMessage') }}</h3>
          <p>{{ status.message || t('metrics.noMessage') }}</p>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const { t, locale } = useI18n()
import { getNamespaces } from '../api/client'
import { listDataFlows, getStatus } from '../api/client'

const namespace = ref('default')
const namespaces = ref([])
const dataflowList = ref([])
const selectedName = ref('')
const status = ref(null)
const loading = ref(false)
const error = ref('')

const lastProcessedFormatted = computed(() => {
  const ts = status.value?.lastProcessedTime
  if (!ts) return 'N/A'
  try {
    const localeTag = locale.value === 'ru' ? 'ru-RU' : 'en-US'
    return new Date(ts).toLocaleString(localeTag)
  } catch {
    return String(ts)
  }
})

async function loadNamespaces() {
  try {
    namespaces.value = await getNamespaces()
    if (namespaces.value.length && !namespaces.value.includes(namespace.value)) {
      namespace.value = namespaces.value[0]
    }
  } catch {
    namespaces.value = ['default']
  }
}

async function loadDataFlowList() {
  if (!namespace.value) return
  try {
    dataflowList.value = await listDataFlows(namespace.value)
    if (!dataflowList.value.some((df) => df.metadata.name === selectedName.value)) {
      selectedName.value = ''
      status.value = null
    }
  } catch {
    dataflowList.value = []
    selectedName.value = ''
    status.value = null
  }
}

async function loadMetrics() {
  if (!selectedName.value) {
    status.value = null
    return
  }
  loading.value = true
  error.value = ''
  try {
    status.value = await getStatus(namespace.value, selectedName.value)
  } catch (e) {
    error.value = e.message
    status.value = null
  } finally {
    loading.value = false
  }
}

watch(namespace, loadDataFlowList)
onMounted(() => {
  loadNamespaces().then(() => loadDataFlowList())
})
</script>

<style scoped>
.form-select {
  max-width: 400px;
}

.empty-state {
  padding: 2rem;
  color: var(--text-muted);
  text-align: center;
}

.value-sm {
  font-size: 1rem !important;
}

.status-message {
  margin-top: 0;
}

.status-message h3 {
  margin-bottom: 0.5rem;
  font-size: 1rem;
}

.status-message p {
  margin: 0;
  color: var(--text-muted);
}
</style>
