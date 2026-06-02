<template>
  <div class="manifests">
    <div class="card">
      <div class="card-header">
        <h2>{{ t('manifests.title') }}</h2>
        <div class="card-header-actions">
          <button class="btn btn-secondary" @click="openCreateInConstructor">
            {{ t('manifests.createInConstructor') }}
          </button>
          <button class="btn btn-primary" @click="openCreate">{{ t('manifests.createNew') }}</button>
        </div>
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
                    @click="openEditInConstructor(df.metadata.namespace, df.metadata.name)"
                  >
                    {{ t('manifests.openInConstructor') }}
                  </button>
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

    <div v-if="constructorModal.open" class="constructor-modal-overlay" @click.self="constructorModal.open = false">
      <div class="constructor-modal">
        <div class="constructor-modal-header">
          <h2>{{ constructorModal.title }}</h2>
          <button type="button" class="modal-close" @click="constructorModal.open = false">&times;</button>
        </div>
        <div class="constructor-modal-body">
          <FlowCanvas
            :initial-manifest="constructorModal.manifest"
            :connections="connections"
            :namespace="constructorModal.namespace"
            :mode="constructorModal.mode"
            @save="onConstructorSave"
          />
        </div>
      </div>
    </div>
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
import FlowCanvas from '../components/FlowCanvas.vue'
import { useToast } from '../composables/useToast'
import {
  listDataFlows,
  getDataFlow,
  createDataFlow,
  updateDataFlow,
  deleteDataFlow,
  listSecrets,
} from '../api/client'
import { sanitizeManifestForDisplay, mergeManifestForUpdate } from '../utils/manifest'

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
  originalManifest: null,
})

const confirm = ref({
  open: false,
  namespace: '',
  name: '',
  message: '',
})

const constructorModal = ref({
  open: false,
  title: '',
  manifest: null,
  namespace: 'default',
  mode: 'create',
  editingNamespace: '',
  editingName: '',
})

const connections = ref([])

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
      value: sanitizeManifestForDisplay(df),
      mode: 'edit',
      editingNamespace: ns,
      editingName: name,
      originalManifest: df,
    }
  } catch (e) {
    showError(e.message)
  }
}

async function openCreateInConstructor() {
  const ns = namespace.value || 'default'
  try {
    connections.value = await listSecrets(ns)
  } catch {
    connections.value = []
  }
  constructorModal.value = {
    open: true,
    title: t('manifests.createTitle'),
    manifest: null,
    namespace: ns,
    mode: 'create',
    editingNamespace: ns,
    editingName: '',
  }
}

async function openEditInConstructor(ns, name) {
  try {
    const [df, conns] = await Promise.all([getDataFlow(ns, name), listSecrets(ns)])
    connections.value = conns
    constructorModal.value = {
      open: true,
      title: t('manifests.editTitle', { name, ns }),
      manifest: df,
      namespace: ns,
      mode: 'edit',
      editingNamespace: ns,
      editingName: name,
    }
  } catch (e) {
    showError(e.message)
  }
}

function onConstructorSave(manifest, err) {
  if (err) {
    showError(err?.message || 'Invalid flow')
    return
  }
  const { editingNamespace, editingName, mode } = constructorModal.value
  if (mode === 'edit') {
    updateDataFlow(editingNamespace, editingName, manifest)
      .then(() => {
        constructorModal.value.open = false
        loadDataFlows()
        success(t('manifests.updated'))
      })
      .catch((e) => showError(e.message))
  } else {
    const ns = manifest.metadata?.namespace || namespace.value || 'default'
    createDataFlow(ns, manifest)
      .then(() => {
        constructorModal.value.open = false
        loadDataFlows()
        success(t('manifests.created'))
      })
      .catch((e) => showError(e.message))
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
        source: { type: 'kafka', config: {} },
        sink: { type: 'postgresql', config: {} },
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
  const { editingNamespace, editingName, originalManifest } = yamlModal.value
  const toUpdate = originalManifest ? mergeManifestForUpdate(parsed, originalManifest) : parsed
  updateDataFlow(editingNamespace, editingName, toUpdate)
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

.card-header-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.constructor-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.constructor-modal {
  background: var(--bg-card);
  border-radius: 8px;
  width: 100%;
  max-width: 1100px;
  max-height: 95vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

.constructor-modal-header {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, var(--gradient-start), var(--gradient-end));
  color: white;
  border-radius: 8px 8px 0 0;
}

.constructor-modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
}

.constructor-modal-header .modal-close {
  background: none;
  border: none;
  color: white;
  font-size: 1.75rem;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.constructor-modal-body {
  padding: 1rem;
  overflow: auto;
  flex: 1;
}
</style>
