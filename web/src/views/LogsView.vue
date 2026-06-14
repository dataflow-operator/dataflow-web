<template>
  <div class="logs">
    <div class="card">
      <div class="card-header">
        <h2>{{ t('logs.title') }}</h2>
      </div>
      <div class="form-group">
        <label>{{ t('common.namespace') }}</label>
        <select
          v-model="namespace"
          class="form-select"
          @change="loadResourceList"
        >
          <option value="">{{ t('common.loading') }}</option>
          <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
        </select>
      </div>
      <div class="form-group">
        <label>{{ t('logs.resourceKind') }}</label>
        <select v-model="resourceKind" class="form-select" @change="onKindChange">
          <option value="dataflow">{{ t('logs.kindDataflow') }}</option>
          <option value="dataflowcron">{{ t('logs.kindDataflowCron') }}</option>
        </select>
      </div>
      <div class="form-group">
        <label>{{ resourceKind === 'dataflowcron' ? t('logs.kindDataflowCron') : t('common.dataflow') }}</label>
        <select
          v-model="selectedName"
          class="form-select"
        >
          <option value="">{{ selectPlaceholder }}</option>
          <option v-for="item in resourceList" :key="item.metadata.name" :value="item.metadata.name">
            {{ item.metadata.name }}
          </option>
        </select>
      </div>
      <div class="form-group">
        <label>{{ t('logs.linesCount') }}</label>
        <input v-model.number="tailLines" type="number" min="1" max="10000" class="form-input" />
      </div>
      <div class="button-row">
        <button type="button" class="btn btn-primary" :disabled="!selectedName || loading" @click="loadLogs">
          {{ t('logs.load') }}
        </button>
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="!selectedName"
          @click="toggleFollow"
        >
          {{ follow ? t('logs.stopFollow') : t('logs.follow') }}
        </button>
      </div>
      <div v-show="logsVisible" class="logs-section">
        <button
          v-if="logsText"
          type="button"
          class="btn btn-secondary btn-sm copy-logs-btn"
          @click="copyLogs"
        >
          {{ t('logs.copy') }}
        </button>
        <div
          ref="logsContainer"
          class="logs-container"
          role="log"
          aria-live="polite"
        >
          <template v-if="loading && !follow">{{ t('logs.loading') }}</template>
          <template v-else-if="follow && connecting">{{ t('logs.connecting') }}</template>
          <template v-else>{{ logsText || t('logs.empty') }}</template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  getNamespaces,
  listDataFlows,
  listDataFlowCrons,
  getLogs,
  createLogStream,
} from '../api/client'
import { useToast } from '../composables/useToast'
import { useFilterQueryParams } from '../composables/useFilterQueryParams'

const { t } = useI18n()
const { success } = useToast()
const route = useRoute()

const { namespace, dataflow: selectedName } = useFilterQueryParams({ dataflow: true })
const namespaces = ref([])
const resourceList = ref([])
const resourceKind = ref(route.query.kind === 'dataflowcron' ? 'dataflowcron' : 'dataflow')
const tailLines = ref(100)
const logsText = ref('')
const logsVisible = ref(false)
const loading = ref(false)
const follow = ref(false)
const connecting = ref(false)
let stopStream = null
const logsContainer = ref(null)

const selectPlaceholder = computed(() =>
  resourceKind.value === 'dataflowcron' ? t('logs.selectDataflowCron') : t('logs.selectDataflow')
)

const logsKindParam = computed(() =>
  resourceKind.value === 'dataflowcron' ? 'dataflowcron' : ''
)

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

async function loadResourceList() {
  if (!namespace.value) return
  try {
    resourceList.value =
      resourceKind.value === 'dataflowcron'
        ? await listDataFlowCrons(namespace.value)
        : await listDataFlows(namespace.value)
    if (!resourceList.value.some((item) => item.metadata.name === selectedName.value)) {
      selectedName.value = ''
    }
  } catch {
    resourceList.value = []
    selectedName.value = ''
  }
}

function onKindChange() {
  selectedName.value = ''
  loadResourceList()
}

watch(namespace, loadResourceList)
watch(
  () => route.query.kind,
  (kind) => {
    if (kind === 'dataflowcron' || kind === 'dataflow') {
      resourceKind.value = kind
      loadResourceList()
    }
  }
)

onMounted(() => {
  loadNamespaces().then(() => loadResourceList())
})

function scrollToBottom() {
  if (logsContainer.value) {
    logsContainer.value.scrollTop = logsContainer.value.scrollHeight
  }
}

async function loadLogs() {
  if (!selectedName.value) return
  logsVisible.value = true
  loading.value = true
  logsText.value = ''
  try {
    logsText.value = await getLogs(namespace.value, selectedName.value, {
      tailLines: tailLines.value,
      kind: logsKindParam.value,
    })
    nextTick(() => scrollToBottom())
  } catch (e) {
    logsText.value = t('common.error') + ': ' + e.message
  } finally {
    loading.value = false
  }
}

function toggleFollow() {
  if (follow.value) {
    if (stopStream) stopStream()
    stopStream = null
    follow.value = false
    connecting.value = false
    return
  }
  if (!selectedName.value) return
  follow.value = true
  connecting.value = true
  logsVisible.value = true
  logsText.value = t('logs.connecting') + '\n'
  stopStream = createLogStream(
    namespace.value,
    selectedName.value,
    { tailLines: tailLines.value, kind: logsKindParam.value },
    (line) => {
      connecting.value = false
      logsText.value += line + '\n'
      nextTick(() => scrollToBottom())
    }
  )
}

function copyLogs() {
  if (!logsText.value) return
  navigator.clipboard.writeText(logsText.value).then(
    () => success(t('logs.copied')),
    () => {}
  )
}

onUnmounted(() => {
  if (stopStream) stopStream()
})
</script>

<style scoped>
.form-select,
.form-input {
  max-width: 400px;
}

.button-row {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.logs-section {
  margin-top: 1rem;
  position: relative;
}

.copy-logs-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  z-index: 1;
}
</style>
