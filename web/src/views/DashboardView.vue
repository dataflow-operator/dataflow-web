<template>
  <div class="dashboard">
    <div class="card">
      <div class="card-header">
        <h2>{{ t('dashboard.title') }}</h2>
      </div>
      <LoadingSpinner v-if="loading" :message="t('dashboard.loading')" />
      <div v-else-if="error" class="error-message">{{ error }}</div>
      <template v-else>
        <div class="metrics-grid dashboard-summary">
          <div class="metric-card">
            <h3>{{ t('dashboard.totalDataflows') }}</h3>
            <div class="value">{{ totalCount }}</div>
          </div>
          <div class="metric-card">
            <h3>{{ t('dashboard.running') }}</h3>
            <div class="value">{{ byPhase.Running ?? 0 }}</div>
          </div>
          <div class="metric-card">
            <h3>{{ t('dashboard.pending') }}</h3>
            <div class="value">{{ byPhase.Pending ?? 0 }}</div>
          </div>
          <div class="metric-card">
            <h3>{{ t('dashboard.error') }}</h3>
            <div class="value">{{ byPhase.Error ?? 0 }}</div>
          </div>
          <div class="metric-card">
            <h3>{{ t('dashboard.stopped') }}</h3>
            <div class="value">{{ byPhase.Stopped ?? 0 }}</div>
          </div>
        </div>
        <div v-if="namespaces.length" class="namespace-summary">
          <h3>{{ t('dashboard.byNamespace') }}</h3>
          <div class="ns-list">
            <div
              v-for="ns in namespaces"
              :key="ns"
              class="ns-card"
            >
              <div class="ns-name">{{ ns }}</div>
              <div class="ns-count">{{ t('dashboard.dataflowCount', { count: (byNamespace[ns] || []).length }) }}</div>
              <router-link :to="{ name: 'manifests', query: { namespace: ns } }" class="ns-link">
                {{ t('dashboard.openManifests') }} →
              </router-link>
            </div>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>{{ t('dashboard.empty') }}</p>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { getNamespaces } from '../api/client'
import { listDataFlows } from '../api/client'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const { t } = useI18n()
const namespaces = ref([])
const byNamespace = ref({})
const loading = ref(true)
const error = ref('')

const totalCount = computed(() => {
  return Object.values(byNamespace.value).reduce((sum, list) => sum + list.length, 0)
})

const byPhase = computed(() => {
  const acc = { Running: 0, Pending: 0, Error: 0, Stopped: 0, Unknown: 0 }
  for (const list of Object.values(byNamespace.value)) {
    for (const df of list) {
      const p = (df.status?.phase || 'Unknown').trim() || 'Unknown'
      acc[p] = (acc[p] ?? 0) + 1
    }
  }
  return acc
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    namespaces.value = await getNamespaces()
    const data = {}
    for (const ns of namespaces.value) {
      try {
        data[ns] = await listDataFlows(ns)
      } catch {
        data[ns] = []
      }
    }
    byNamespace.value = data
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.dashboard-summary {
  margin-bottom: 2rem;
}

.namespace-summary h3 {
  margin-bottom: 1rem;
  font-size: 1.1rem;
  color: var(--text-muted);
}

.ns-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 1rem;
}

.ns-card {
  padding: 1rem;
  background: var(--bg-page);
  border-radius: 8px;
  border: 1px solid var(--border);
}

.ns-name {
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.ns-count {
  font-size: 0.9rem;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
}

.ns-link {
  font-size: 0.9rem;
  color: var(--primary);
  text-decoration: none;
}

.ns-link:hover {
  text-decoration: underline;
}
</style>
