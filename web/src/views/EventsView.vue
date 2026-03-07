<template>
  <div class="events">
    <div class="card">
      <div class="card-header">
        <h2>{{ t('events.title') }}</h2>
      </div>
      <div class="form-group">
        <label>{{ t('common.namespace') }}</label>
        <select
          v-model="namespace"
          class="form-select"
          @change="loadDataFlowList"
        >
          <option value="">{{ t('common.loading') }}</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>
      <div class="form-group">
        <label>{{ t('events.filterByManifest') }}</label>
        <div class="filter-row">
          <select
            v-model="filterMode"
            class="form-select filter-mode"
            @change="onFilterModeChange"
          >
            <option value="all">{{ t('events.filterAll') }}</option>
            <option value="manifest">{{ t('events.filterByManifest') }}</option>
          </select>
          <select
            v-if="filterMode === 'manifest'"
            v-model="selectedName"
            class="form-select"
            @change="loadEvents"
          >
            <option value="">{{ t('events.selectManifest') }}</option>
            <option v-for="df in dataflowList" :key="df.metadata.name" :value="df.metadata.name">
              {{ df.metadata.name }}
            </option>
          </select>
        </div>
      </div>
      <div class="button-row">
        <button
          type="button"
          class="btn btn-primary"
          :disabled="loading"
          @click="loadEvents"
        >
          {{ t('events.refresh') }}
        </button>
      </div>
      <LoadingSpinner v-if="loading" :message="t('events.loading')" />
      <div v-else-if="error" class="error-message">{{ error }}</div>
      <div v-else-if="events.length === 0" class="empty-state">
        <p>{{ filterMode === 'manifest' && selectedName ? t('events.emptyFiltered') : t('events.empty') }}</p>
      </div>
      <div v-else class="events-table-wrap">
        <table class="data-table events-table">
          <thead>
            <tr>
              <th>{{ t('events.type') }}</th>
              <th>{{ t('events.reason') }}</th>
              <th>{{ t('events.message') }}</th>
              <th>{{ t('events.object') }}</th>
              <th>{{ t('events.time') }}</th>
              <th>{{ t('events.count') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(ev, idx) in events"
              :key="ev.metadata?.uid || idx"
              :class="{ 'event-warning': ev.type === 'Warning' }"
            >
              <td>
                <span
                  class="event-type-badge"
                  :class="ev.type === 'Warning' ? 'event-type-warning' : 'event-type-normal'"
                >
                  {{ ev.type || 'Normal' }}
                </span>
              </td>
              <td>{{ ev.reason || '—' }}</td>
              <td class="event-message">{{ ev.message || '—' }}</td>
              <td>{{ ev.involvedObject?.name || '—' }}</td>
              <td class="event-time">{{ formatTime(ev.lastTimestamp || ev.eventTime) }}</td>
              <td>{{ ev.count ?? 1 }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import { getNamespaces } from '../api/client'
import { listDataFlows, getEvents } from '../api/client'

const { t, locale } = useI18n()

const namespace = ref('default')
const namespaces = ref([])
const dataflowList = ref([])
const filterMode = ref('all')
const selectedName = ref('')
const events = ref([])
const loading = ref(false)
const error = ref('')

function formatTime(ts) {
  if (!ts) return '—'
  try {
    const date = typeof ts === 'string' ? new Date(ts) : ts
    const localeTag = locale.value === 'ru' ? 'ru-RU' : 'en-US'
    return date.toLocaleString(localeTag)
  } catch {
    return String(ts)
  }
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
    if (filterMode.value === 'manifest' && !dataflowList.value.some((df) => df.metadata.name === selectedName.value)) {
      selectedName.value = ''
    }
  } catch {
    dataflowList.value = []
    selectedName.value = ''
  }
}

function onFilterModeChange() {
  if (filterMode.value === 'all') {
    selectedName.value = ''
  }
  loadEvents()
}

async function loadEvents() {
  if (!namespace.value) return
  if (filterMode.value === 'manifest' && !selectedName.value) {
    events.value = []
    return
  }
  loading.value = true
  error.value = ''
  try {
    const name = filterMode.value === 'manifest' ? selectedName.value : null
    events.value = await getEvents(namespace.value, name)
  } catch (e) {
    error.value = e.message
    events.value = []
  } finally {
    loading.value = false
  }
}

watch(namespace, () => {
  loadDataFlowList().then(loadEvents)
})

onMounted(() => {
  loadNamespaces().then(() => {
    loadDataFlowList().then(loadEvents)
  })
})
</script>

<style scoped>
.form-select {
  max-width: 400px;
}

.filter-row {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  align-items: center;
}

.filter-mode {
  max-width: 220px;
}

.button-row {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.empty-state {
  padding: 2rem;
  color: var(--text-muted);
  text-align: center;
}

.events-table-wrap {
  overflow-x: auto;
}

.events-table .event-message {
  max-width: 400px;
  word-break: break-word;
}

.events-table .event-time {
  white-space: nowrap;
}

.event-type-badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.event-type-normal {
  background: var(--success-bg);
  color: var(--success-text);
}

.event-type-warning {
  background: var(--warning-bg);
  color: var(--warning-text);
}

.event-warning {
  background: rgba(255, 193, 7, 0.08);
}
</style>
