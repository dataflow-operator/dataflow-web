<template>
  <div class="connections">
    <div class="card">
      <div class="card-header">
        <h2>{{ t('connections.title') }}</h2>
        <button class="btn btn-primary" @click="openCreate">
          {{ t('connections.createNew') }}
        </button>
      </div>
      <NamespaceSelect v-model="namespace" @update:model-value="loadSecrets" />
      <LoadingSpinner v-if="loading" :message="t('connections.loading')" />
      <div v-else-if="error" class="error-message">{{ error }}</div>
      <div v-else-if="filteredSecrets.length === 0" class="empty-state">
        <p>{{ t('connections.empty') }}</p>
        <p class="empty-hint">{{ t('connections.emptyHint') }}</p>
      </div>
      <template v-else>
        <div class="table-toolbar">
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('connections.searchPlaceholder')"
            class="search-input"
            :aria-label="t('connections.searchPlaceholder')"
          />
        </div>
        <div class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>{{ t('connections.name') }}</th>
                <th>{{ t('common.namespace') }}</th>
                <th>{{ t('connections.type') }}</th>
                <th>{{ t('connections.keys') }}</th>
                <th>{{ t('connections.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="s in filteredSecrets" :key="s.metadata?.name + (s.metadata?.namespace || namespace)">
                <td :data-label="t('connections.name')">{{ s.metadata?.name }}</td>
                <td :data-label="t('common.namespace')">{{ s.metadata?.namespace || namespace }}</td>
                <td :data-label="t('connections.type')">
                  {{ connectionTypeLabel(getConnectionType(s)) }}
                </td>
                <td :data-label="t('connections.keys')">
                  <span class="keys-list">{{ (s.keys || []).join(', ') || '—' }}</span>
                </td>
                <td :data-label="t('connections.actions')">
                  <button
                    type="button"
                    class="btn btn-secondary btn-sm"
                    @click="openEdit(s.metadata?.namespace || namespace, s.metadata?.name)"
                  >
                    {{ t('connections.edit') }}
                  </button>
                  <button
                    type="button"
                    class="btn btn-danger btn-sm"
                    @click="confirmDelete(s.metadata?.namespace || namespace, s.metadata?.name)"
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

    <ConnectionFormModal
      :open="formModal.open"
      :title="formModal.title"
      :secret="formModal.secret"
      :mode="formModal.mode"
      :namespace="formModal.namespace"
      @close="formModal.open = false"
      @save="onSave"
      @create="onCreate"
    />

    <ConfirmModal
      :open="confirm.open"
      :title="t('connections.deleteTitle')"
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
import ConnectionFormModal from '../components/ConnectionFormModal.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import { useToast } from '../composables/useToast'
import {
  listSecrets,
  getSecret,
  createSecret,
  updateSecret,
  deleteSecret,
} from '../api/client'

const CONNECTION_TYPE_LABEL = 'dataflow.dataflow.io/connection-type'

const { t } = useI18n()
const { success, error: showError } = useToast()

const route = useRoute()
const namespace = ref(route.query.namespace || 'default')
const secrets = ref([])
const loading = ref(false)
const error = ref('')
const searchQuery = ref('')

const formModal = ref({
  open: false,
  title: '',
  secret: null,
  mode: 'edit',
  namespace: 'default',
})

const confirm = ref({
  open: false,
  namespace: '',
  name: '',
  message: '',
})

function getConnectionType(item) {
  return item?.metadata?.labels?.[CONNECTION_TYPE_LABEL] || ''
}

function connectionTypeLabel(type) {
  if (!type) return '—'
  const key = `connections.connectionTypes.${type}`
  const translated = t(key)
  return translated !== key ? translated : type
}

const filteredSecrets = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return secrets.value
  return secrets.value.filter(
    (s) => (s.metadata?.name || '').toLowerCase().includes(q)
  )
})

async function loadSecrets() {
  if (!namespace.value) return
  loading.value = true
  error.value = ''
  try {
    secrets.value = await listSecrets(namespace.value)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(namespace, (ns) => {
  if (ns) loadSecrets()
})

onMounted(() => {
  if (route.query.namespace) namespace.value = route.query.namespace
  loadSecrets()
})

async function openEdit(ns, name) {
  try {
    const secret = await getSecret(ns, name)
    formModal.value = {
      open: true,
      title: t('connections.editTitle', { name }),
      secret: { ...secret, metadata: { ...secret.metadata, namespace: ns } },
      mode: 'edit',
      namespace: ns,
    }
  } catch (e) {
    showError(e.message)
  }
}

function openCreate() {
  const ns = namespace.value || 'default'
  formModal.value = {
    open: true,
    title: t('connections.createTitle'),
    secret: {
      metadata: { name: '', namespace: ns },
      type: 'Opaque',
      stringData: {},
    },
    mode: 'create',
    namespace: ns,
  }
}

function onSave(secret) {
  const ns = formModal.value.namespace
  const name = secret.metadata?.name
  if (!name) {
    showError('Name required')
    return
  }
  updateSecret(ns, name, secret)
    .then(() => {
      formModal.value.open = false
      loadSecrets()
      success(t('connections.updated'))
    })
    .catch((e) => showError(e.message))
}

function onCreate(secret) {
  const ns = secret.metadata?.namespace || namespace.value || 'default'
  const name = secret.metadata?.name
  if (!name) {
    showError('Name required')
    return
  }
  createSecret(ns, secret)
    .then(() => {
      formModal.value.open = false
      loadSecrets()
      success(t('connections.created'))
    })
    .catch((e) => showError(e.message))
}

function confirmDelete(ns, name) {
  confirm.value = {
    open: true,
    namespace: ns,
    name,
    message: t('connections.deleteConfirm', { name }),
  }
}

function doDelete() {
  const { namespace: ns, name } = confirm.value
  confirm.value.open = false
  deleteSecret(ns, name)
    .then(() => {
      loadSecrets()
      success(t('connections.deleted'))
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

.keys-list {
  font-size: 0.9rem;
  color: var(--text-muted);
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
