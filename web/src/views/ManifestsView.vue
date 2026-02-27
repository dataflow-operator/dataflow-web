<template>
  <div class="manifests">
    <div class="card">
      <div class="card-header">
        <h2>{{ t('manifests.title') }}</h2>
        <button class="btn btn-primary" @click="openCreate">{{ t('manifests.createNew') }}</button>
      </div>
      <NamespaceSelect v-model="namespace" @update:model-value="loadDataFlows" />
      <LoadingSpinner v-if="loading" :message="t('manifests.loading')" />
      <div v-else-if="error" class="error-message">{{ error }}</div>
      <div v-else-if="filteredFlows.length === 0" class="empty-state">
        <p>{{ t('manifests.empty') }}</p>
        <p class="empty-hint">{{ t('manifests.emptyHint') }}</p>
      </div>
      <template v-else>
        <div class="table-toolbar">
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('manifests.searchPlaceholder')"
            class="search-input"
            :aria-label="t('manifests.searchAria')"
          />
        </div>
        <div class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>{{ t('manifests.name') }}</th>
                <th>{{ t('common.namespace') }}</th>
                <th>{{ t('manifests.status') }}</th>
                <th>{{ t('manifests.processed') }}</th>
                <th>{{ t('manifests.errors') }}</th>
                <th>{{ t('manifests.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="df in filteredFlows" :key="df.metadata?.name + df.metadata?.namespace">
                <td :data-label="t('manifests.name')">{{ df.metadata?.name }}</td>
                <td :data-label="t('common.namespace')">{{ df.metadata?.namespace }}</td>
                <td :data-label="t('manifests.status')">
                  <span :class="['status-badge', statusClass(df.status?.phase)]">
                    {{ df.status?.phase || 'Unknown' }}
                  </span>
                </td>
                <td :data-label="t('manifests.processed')">{{ df.status?.processedCount ?? 0 }}</td>
                <td :data-label="t('manifests.errors')">{{ df.status?.errorCount ?? 0 }}</td>
                <td :data-label="t('manifests.actions')">
                  <button
                    type="button"
                    class="btn btn-secondary btn-sm"
                    @click="openEdit(df.metadata.namespace, df.metadata.name)"
                  >
                    {{ t('manifests.viewEdit') }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-danger btn-sm"
                    @click="confirmDelete(df.metadata.namespace, df.metadata.name)"
                  >
                    {{ t('common.delete') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <YamlEditorModal
      :open="yamlModal.open"
      :title="yamlModal.title"
      :value="yamlModal.value"
      :mode="yamlModal.mode"
      @close="yamlModal.open = false"
      @save="onSave"
      @create="onCreate"
    />

    <ConfirmModal
      :open="confirm.open"
      :title="t('manifests.deleteTitle')"
      :message="confirm.message"
      :confirm-label="t('common.delete')"
      @cancel="confirm.open = false"
      @confirm="doDelete"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import NamespaceSelect from '../components/NamespaceSelect.vue'
import YamlEditorModal from '../components/YamlEditorModal.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import { useToast } from '../composables/useToast'
import {
  listDataFlows,
  getDataFlow,
  createDataFlow,
  updateDataFlow,
  deleteDataFlow,
} from '../api/client'

const { t } = useI18n()
const { success, error: showError } = useToast()

const route = useRoute()
const namespace = ref(route.query.namespace || 'default')
const dataflows = ref([])
const loading = ref(false)
const error = ref('')
const searchQuery = ref('')

const yamlModal = ref({
  open: false,
  title: '',
  value: null,
  mode: 'edit',
  editingNamespace: '',
  editingName: '',
})

const confirm = ref({
  open: false,
  namespace: '',
  name: '',
  message: '',
})

const filteredFlows = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return dataflows.value
  return dataflows.value.filter(
    (df) => (df.metadata?.name || '').toLowerCase().includes(q)
  )
})

function statusClass(phase) {
  const p = (phase || '').toLowerCase()
  if (p === 'running') return 'status-running'
  if (p === 'pending') return 'status-pending'
  if (p === 'error') return 'status-error'
  return 'status-stopped'
}

async function loadDataFlows() {
  if (!namespace.value) return
  loading.value = true
  error.value = ''
  try {
    dataflows.value = await listDataFlows(namespace.value)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(namespace, (ns) => {
  if (ns) loadDataFlows()
})

onMounted(() => {
  if (route.query.namespace) namespace.value = route.query.namespace
})

async function openEdit(ns, name) {
  try {
    const df = await getDataFlow(ns, name)
    yamlModal.value = {
      open: true,
      title: t('manifests.editTitle', { name, ns }),
      value: df,
      mode: 'edit',
      editingNamespace: ns,
      editingName: name,
    }
  } catch (e) {
    showError(e.message)
  }
}

function openCreate() {
  const ns = namespace.value || 'default'
  yamlModal.value = {
    open: true,
    title: t('manifests.createTitle'),
    value: {
      apiVersion: 'dataflow.dataflow.io/v1',
      kind: 'DataFlow',
      metadata: { name: '', namespace: ns },
      spec: {
        source: { type: 'kafka', kafka: { brokers: [], topic: '' } },
        sink: { type: 'postgresql', postgresql: { connectionString: '', table: '' } },
        transformations: [],
      },
    },
    mode: 'create',
    editingNamespace: ns,
    editingName: '',
  }
}

function onSave(parsed, err) {
  if (err) {
    showError(err.message)
    return
  }
  const { editingNamespace, editingName } = yamlModal.value
  updateDataFlow(editingNamespace, editingName, parsed)
    .then(() => {
      yamlModal.value.open = false
      loadDataFlows()
      success(t('manifests.updated'))
    })
    .catch((e) => showError(e.message))
}

function onCreate(parsed, err) {
  if (err) {
    showError(err.message)
    return
  }
  const ns = parsed.metadata?.namespace || namespace.value || 'default'
  createDataFlow(ns, parsed)
    .then(() => {
      yamlModal.value.open = false
      loadDataFlows()
      success(t('manifests.created'))
    })
    .catch((e) => showError(e.message))
}

function confirmDelete(ns, name) {
  confirm.value = {
    open: true,
    namespace: ns,
    name,
    message: t('manifests.deleteConfirm', { name }),
  }
}

function doDelete() {
  const { namespace: ns, name } = confirm.value
  confirm.value.open = false
  deleteDataFlow(ns, name)
    .then(() => {
      loadDataFlows()
      success(t('manifests.deleted'))
    })
    .catch((e) => showError(e.message))
}
</script>

<style scoped>
.table-toolbar {
  margin-bottom: 1rem;
}

.search-input {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.95rem;
  max-width: 280px;
  width: 100%;
}

.table-wrap {
  overflow-x: auto;
}

.btn-sm {
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  margin-right: 0.5rem;
  margin-bottom: 0.25rem;
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--text-muted);
}

.empty-hint {
  margin-top: 0.5rem;
  font-size: 0.9rem;
}
</style>
