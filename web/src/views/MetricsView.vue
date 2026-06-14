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

        <div class="charts">
          <MetricsChart
            :title="t('metrics.throughputChart')"
            :series="charts.throughput.series"
            :loading="charts.throughput.loading"
            :error="charts.throughput.error"
            :message="charts.throughput.message"
          />
          <MetricsChart
            :title="t('metrics.successRateChart')"
            :series="charts.success_rate.series"
            :loading="charts.success_rate.loading"
            :error="charts.success_rate.error"
            :message="charts.success_rate.message"
          />
          <MetricsChart
            :title="t('metrics.queueSizeChart')"
            :series="charts.queue_size.series"
            :loading="charts.queue_size.loading"
            :error="charts.queue_size.error"
            :message="charts.queue_size.message"
          />
          <MetricsChart
            :title="t('metrics.connectorErrorsChart')"
            :series="charts.connector_errors.series"
            :loading="charts.connector_errors.loading"
            :error="charts.connector_errors.error"
            :message="charts.connector_errors.message"
          />
        </div>

        <RuntimePanel
          :runtime="runtime"
          :loading="runtimeLoading"
          :error="runtimeError"
          :namespace="namespace"
          :dataflow-name="selectedName"
          @checkpoint-reset="loadRuntime"
        />
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import MetricsChart from '../components/MetricsChart.vue'
import RuntimePanel from '../components/RuntimePanel.vue'
import { useFilterQueryParams } from '../composables/useFilterQueryParams'

const { t, locale } = useI18n()
import { getNamespaces } from '../api/client'
import { listDataFlows, getStatus, getRuntime, getPrometheusRange } from '../api/client'

const { namespace, dataflow: selectedName } = useFilterQueryParams({ dataflow: true })
const namespaces = ref([])
const dataflowList = ref([])
const status = ref(null)
const loading = ref(false)
const error = ref('')
const runtime = ref(null)
const runtimeLoading = ref(false)
const runtimeError = ref('')

const charts = ref({
  throughput: initChartState(),
  success_rate: initChartState(),
  queue_size: initChartState(),
  connector_errors: initChartState(),
})

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

function initChartState() {
  return { loading: false, error: '', message: '', series: [] }
}

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
    runtime.value = null
    return
  }
  loading.value = true
  error.value = ''
  try {
    status.value = await getStatus(namespace.value, selectedName.value)
    await Promise.all([loadRuntime(), loadCharts()])
  } catch (e) {
    error.value = e.message
    status.value = null
    runtime.value = null
  } finally {
    loading.value = false
  }
}

async function loadRuntime() {
  runtimeLoading.value = true
  runtimeError.value = ''
  try {
    runtime.value = await getRuntime(namespace.value, selectedName.value)
  } catch (e) {
    runtimeError.value = e.message
    runtime.value = null
  } finally {
    runtimeLoading.value = false
  }
}

async function loadCharts() {
  const now = Math.floor(Date.now() / 1000)
  const start = now - 3600
  const end = now
  const step = 30
  await Promise.all([
    loadChart('throughput', { start, end, step }),
    loadChart('success_rate', { start, end, step }),
    loadChart('queue_size', { start, end, step }),
    loadChart('connector_errors', { start, end, step }),
  ])
}

async function loadChart(panel, range) {
  const state = charts.value[panel]
  if (!state) return
  state.loading = true
  state.error = ''
  state.message = ''
  state.series = []
  try {
    const resp = await getPrometheusRange(namespace.value, selectedName.value, panel, range)
    state.series = resp.series || []
    if (resp.message) state.message = resp.message
  } catch (e) {
    state.error = e.message
  } finally {
    state.loading = false
  }
}

watch(namespace, loadDataFlowList)
onMounted(() => {
  loadNamespaces().then(() =>
    loadDataFlowList().then(() => {
      if (selectedName.value) loadMetrics()
    })
  )
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

.charts {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1rem;
}
</style>
