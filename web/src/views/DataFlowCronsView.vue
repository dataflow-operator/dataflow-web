<template>
  <div class="crons">
    <div class="card">
      <div class="card-header">
        <h2>{{ t('crons.title') }}</h2>
        <div class="card-header-actions">
          <button class="btn btn-primary" @click="openCreate">{{ t('crons.createNew') }}</button>
        </div>
      </div>
      <NamespaceSelect v-model="namespace" @update:model-value="loadCrons" />
      <LoadingSpinner v-if="loading" :message="t('crons.loading')" />
      <div v-else-if="error" class="error-message">{{ error }}</div>
      <div v-else-if="filteredCrons.length === 0" class="empty-state">
        <p>{{ t('crons.empty') }}</p>
        <p class="empty-hint">{{ t('crons.emptyHint') }}</p>
      </div>
      <template v-else>
        <div class="table-toolbar">
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('crons.searchPlaceholder')"
            class="search-input"
            :aria-label="t('crons.searchAria')"
          />
        </div>
        <div class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>{{ t('crons.name') }}</th>
                <th>{{ t('common.namespace') }}</th>
                <th>{{ t('crons.schedule') }}</th>
                <th>{{ t('crons.status') }}</th>
                <th>{{ t('crons.lastRun') }}</th>
                <th>{{ t('crons.suspended') }}</th>
                <th>{{ t('crons.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="dfc in filteredCrons" :key="dfc.metadata?.name + dfc.metadata?.namespace">
                <td :data-label="t('crons.name')">{{ dfc.metadata?.name }}</td>
                <td :data-label="t('common.namespace')">{{ dfc.metadata?.namespace }}</td>
                <td :data-label="t('crons.schedule')"><code>{{ dfc.spec?.schedule || '—' }}</code></td>
                <td :data-label="t('crons.status')">
                  <span :class="['status-badge', statusClass(dfc.status?.phase)]">
                    {{ dfc.status?.phase || '—' }}
                  </span>
                </td>
                <td :data-label="t('crons.lastRun')">{{ formatLastRun(dfc) }}</td>
                <td :data-label="t('crons.suspended')">
                  {{ dfc.spec?.suspend ? t('common.enabled') : t('common.disabled') }}
                </td>
                <td :data-label="t('crons.actions')">
                  <button
                    type="button"
                    class="btn btn-secondary btn-sm"
                    @click="openEdit(dfc.metadata.namespace, dfc.metadata.name)"
                  >
                    {{ t('crons.viewEdit') }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary btn-sm"
                    :disabled="dfc.spec?.suspend"
                    :title="dfc.spec?.suspend ? t('crons.suspendedHint') : ''"
                    @click="confirmTrigger(dfc.metadata.namespace, dfc.metadata.name)"
                  >
                    {{ t('crons.runNow') }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-secondary btn-sm"
                    @click="toggleSuspend(dfc)"
                  >
                    {{ dfc.spec?.suspend ? t('crons.resume') : t('crons.suspend') }}
                  </button>
                  <router-link
                    :to="{
                      name: 'logs',
                      query: {
                        namespace: dfc.metadata.namespace,
                        dataflow: dfc.metadata.name,
                        kind: 'dataflowcron',
                      },
                    }"
                    class="btn btn-secondary btn-sm"
                  >
                    {{ t('crons.logs') }}
                  </router-link>
                  <button
                    type="button"
                    class="btn btn-danger btn-sm"
                    @click="confirmDelete(dfc.metadata.namespace, dfc.metadata.name)"
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
      :title="confirm.title"
      :message="confirm.message"
      :confirm-label="confirm.confirmLabel"
      @cancel="confirm.open = false"
      @confirm="onConfirm"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import NamespaceSelect from '../components/NamespaceSelect.vue'
import { useFilterQueryParams } from '../composables/useFilterQueryParams'
import YamlEditorModal from '../components/YamlEditorModal.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import { useToast } from '../composables/useToast'
import {
  listDataFlowCrons,
  getDataFlowCron,
  createDataFlowCron,
  updateDataFlowCron,
  deleteDataFlowCron,
  triggerDataFlowCron,
  suspendDataFlowCron,
} from '../api/client'
import { sanitizeManifestForDisplay, mergeManifestForUpdate } from '../utils/manifest'

const { t } = useI18n()
const { success, error: showError } = useToast()

const { namespace } = useFilterQueryParams()
const crons = ref([])
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
  action: '',
  namespace: '',
  name: '',
  title: '',
  message: '',
  confirmLabel: '',
})

const filteredCrons = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return crons.value
  return crons.value.filter((dfc) => (dfc.metadata?.name || '').toLowerCase().includes(q))
})

function statusClass(phase) {
  const p = (phase || '').toLowerCase()
  if (p === 'completed') return 'status-running'
  if (p === 'runningtriggers') return 'status-pending'
  if (p === 'failed') return 'status-error'
  return 'status-stopped'
}

function formatLastRun(dfc) {
  const ok = dfc.status?.lastSuccessfulTime
  const fail = dfc.status?.lastFailedTime
  if (ok) return new Date(ok).toLocaleString()
  if (fail) return new Date(fail).toLocaleString()
  return '—'
}

async function loadCrons() {
  if (!namespace.value) return
  loading.value = true
  error.value = ''
  try {
    crons.value = await listDataFlowCrons(namespace.value)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(namespace, (ns) => {
  if (ns) loadCrons()
}, { immediate: true })

async function openEdit(ns, name) {
  try {
    const dfc = await getDataFlowCron(ns, name)
    yamlModal.value = {
      open: true,
      title: t('crons.editTitle', { name, ns }),
      value: sanitizeManifestForDisplay(dfc),
      mode: 'edit',
      editingNamespace: ns,
      editingName: name,
      originalManifest: dfc,
    }
  } catch (e) {
    showError(e.message)
  }
}

function openCreate() {
  const ns = namespace.value || 'default'
  yamlModal.value = {
    open: true,
    title: t('crons.createTitle'),
    value: {
      apiVersion: 'dataflow.dataflow.io/v1',
      kind: 'DataFlowCron',
      metadata: { name: '', namespace: ns },
      spec: {
        schedule: '0 * * * *',
        concurrencyPolicy: 'Forbid',
        source: { type: 'postgresql', config: {} },
        sink: { type: 'postgresql', config: {} },
        transformations: [],
      },
    },
    mode: 'create',
    editingNamespace: ns,
    editingName: '',
    originalManifest: null,
  }
}

function onSave(parsed, err) {
  if (err) {
    showError(err.message)
    return
  }
  const { editingNamespace, editingName, originalManifest } = yamlModal.value
  const toUpdate = originalManifest ? mergeManifestForUpdate(parsed, originalManifest) : parsed
  updateDataFlowCron(editingNamespace, editingName, toUpdate)
    .then(() => {
      yamlModal.value.open = false
      loadCrons()
      success(t('crons.updated'))
    })
    .catch((e) => showError(e.message))
}

function onCreate(parsed, err) {
  if (err) {
    showError(err.message)
    return
  }
  const ns = parsed.metadata?.namespace || namespace.value || 'default'
  createDataFlowCron(ns, parsed)
    .then(() => {
      yamlModal.value.open = false
      loadCrons()
      success(t('crons.created'))
    })
    .catch((e) => showError(e.message))
}

function confirmDelete(ns, name) {
  confirm.value = {
    open: true,
    action: 'delete',
    namespace: ns,
    name,
    title: t('crons.deleteTitle'),
    message: t('crons.deleteConfirm', { name }),
    confirmLabel: t('common.delete'),
  }
}

function confirmTrigger(ns, name) {
  confirm.value = {
    open: true,
    action: 'trigger',
    namespace: ns,
    name,
    title: t('crons.triggerTitle'),
    message: t('crons.triggerConfirm', { name }),
    confirmLabel: t('crons.runNow'),
  }
}

function onConfirm() {
  const { action, namespace: ns, name } = confirm.value
  confirm.value.open = false
  if (action === 'delete') {
    deleteDataFlowCron(ns, name)
      .then(() => {
        loadCrons()
        success(t('crons.deleted'))
      })
      .catch((e) => showError(e.message))
    return
  }
  if (action === 'trigger') {
    triggerDataFlowCron(ns, name)
      .then((res) => {
        loadCrons()
        success(t('crons.triggered', { job: res.jobName }))
      })
      .catch((e) => showError(e.message))
  }
}

async function toggleSuspend(dfc) {
  const ns = dfc.metadata.namespace
  const name = dfc.metadata.name
  const suspend = !dfc.spec?.suspend
  try {
    await suspendDataFlowCron(ns, name, suspend)
    loadCrons()
    success(suspend ? t('crons.suspended') : t('crons.resumed'))
  } catch (e) {
    showError(e.message)
  }
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

code {
  font-size: 0.85rem;
}
</style>
