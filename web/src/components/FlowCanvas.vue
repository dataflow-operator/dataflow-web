<template>
  <div class="flow-canvas-wrap">
    <div class="flow-toolbar">
      <input
        v-model="manifestName"
        type="text"
        :placeholder="t('flow.manifestNamePlaceholder')"
        class="flow-name-input"
      />
      <button type="button" class="btn btn-secondary btn-sm" @click="addTransformation">
        {{ t('flow.addTransformation') }}
      </button>
      <button type="button" class="btn btn-secondary btn-sm" @click="applyLayout">
        {{ t('flow.applyLayout') }}
      </button>
      <button
        type="button"
        class="btn btn-secondary btn-sm"
        :disabled="!previewManifest"
        @click="openPreview"
      >
        {{ t('flow.preview') }}
      </button>
      <button
        type="button"
        class="btn btn-secondary btn-sm"
        :disabled="!previewManifest"
        @click="exportManifest"
      >
        {{ t('flow.export') }}
      </button>
      <button
        type="button"
        class="btn btn-secondary btn-sm"
        :disabled="!previewManifest"
        @click="copyManifest"
      >
        {{ t('flow.copy') }}
      </button>
      <button type="button" class="btn btn-primary" @click="emitSave">
        {{ t('flow.save') }}
      </button>
    </div>
    <div class="flow-canvas">
      <VueFlow
        v-model:nodes="nodes"
        v-model:edges="edges"
        :default-viewport="{ x: 0, y: 0, zoom: 1 }"
        fit-view-on-init
        @node-click="onNodeClick"
      >
        <Background />
        <Controls />
        <template #node-source="props">
          <SourceNode v-bind="props" @config="openConfig($event, 'source', props.node)" />
        </template>
        <template #node-sink="props">
          <SinkNode v-bind="props" @config="openConfig($event, 'sink', props.node)" />
        </template>
        <template #node-transformation="props">
          <TransformationNode
            v-bind="props"
            :can-move-up="canMoveTransformationUp(props.node?.id ?? props.id)"
            :can-move-down="canMoveTransformationDown(props.node?.id ?? props.id)"
            @config="openConfig($event, 'transformation', props.node)"
            @remove="removeTransformation($event)"
            @move-up="moveTransformationUp($event)"
            @move-down="moveTransformationDown($event)"
          />
        </template>
      </VueFlow>
    </div>

    <NodeConfigPanel
      v-if="configPanel.open"
      :key="`config-${configPanel.nodeId}`"
      :node="configPanel.node"
      :node-type="configPanel.nodeType"
      :connections="connections"
      :namespace="namespace"
      @close="configPanel.open = false"
      @save="onConfigSave"
    />

    <YamlEditorModal
      :open="previewOpen"
      :title="t('flow.previewTitle')"
      :value="previewManifest"
      mode="view"
      @close="previewOpen = false"
    />
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { VueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import { useI18n } from 'vue-i18n'
import SourceNode from './flow/SourceNode.vue'
import SinkNode from './flow/SinkNode.vue'
import TransformationNode from './flow/TransformationNode.vue'
import NodeConfigPanel from './NodeConfigPanel.vue'
import YamlEditorModal from './YamlEditorModal.vue'
import { useToast } from '../composables/useToast'
import yaml from 'js-yaml'
import {
  manifestToGraph,
  graphToManifest,
  createEmptyGraph,
  CONNECTOR_TYPES,
  TRANSFORMATION_TYPES,
} from '../composables/useFlowManifest'

const props = defineProps({
  initialManifest: { type: Object, default: null },
  connections: { type: Array, default: () => [] },
  namespace: { type: String, default: 'default' },
  mode: { type: String, default: 'create' },
})

const emit = defineEmits(['save', 'close'])

const { t } = useI18n()
const { success } = useToast()

const nodes = ref([])
const edges = ref([])
const manifestName = ref('')

const configPanel = ref({
  open: false,
  nodeId: '',
  node: null,
  nodeType: '',
})

const previewOpen = ref(false)

const LAYOUT_SPACING = 220

function getNodeOrder() {
  const order = []
  let currentId = 'source'
  const visited = new Set()
  while (currentId && currentId !== 'sink') {
    if (visited.has(currentId)) break
    visited.add(currentId)
    order.push(currentId)
    const out = edges.value.find((e) => e.source === currentId)
    currentId = out?.target || null
  }
  order.push('sink')
  return order
}

function applyLayout() {
  const order = getNodeOrder()
  const next = nodes.value.map((n) => {
    const idx = order.indexOf(n.id)
    if (idx < 0) return n
    return {
      ...n,
      position: { x: 50 + idx * LAYOUT_SPACING, y: 200 },
    }
  })
  nodes.value = next
}

function initFromManifest(manifest) {
  if (manifest) {
    const { nodes: n, edges: e } = manifestToGraph(manifest)
    nodes.value = n
    edges.value = e
    manifestName.value = manifest.metadata?.name || ''
  } else {
    const { nodes: n, edges: e } = createEmptyGraph()
    nodes.value = n
    edges.value = e
    manifestName.value = ''
  }
  applyLayout()
}

watch(
  () => props.initialManifest,
  (m) => initFromManifest(m),
  { immediate: true }
)

function addTransformation() {
  const n = nodes.value.filter((x) => x.id.startsWith('t-'))
  const nextIdx = n.length
  const tId = `t-${nextIdx}`

  const sinkNode = nodes.value.find((x) => x.id === 'sink')
  const prevId = nextIdx === 0 ? 'source' : `t-${nextIdx - 1}`

  const newTransformation = {
    id: tId,
    type: 'transformation',
    position: { x: 250 + nextIdx * 200, y: 200 },
    data: {
      label: 'timestamp',
      transformationType: 'timestamp',
      config: {},
    },
  }

  nodes.value = [...nodes.value, newTransformation]

  const oldEdge = edges.value.find((e) => e.target === 'sink')
  if (oldEdge) {
    edges.value = edges.value.filter((e) => e.id !== oldEdge.id)
    edges.value.push({ id: `e-${prevId}-${tId}`, source: prevId, target: tId })
    edges.value.push({ id: `e-${tId}-sink`, source: tId, target: 'sink' })
  } else {
    edges.value.push({ id: `e-${prevId}-${tId}`, source: prevId, target: tId })
    edges.value.push({ id: `e-${tId}-sink`, source: tId, target: 'sink' })
  }
  applyLayout()
}

function removeTransformation(nodeId) {
  const node = nodes.value.find((n) => n.id === nodeId)
  if (!node || node.type !== 'transformation') return
  const inEdge = edges.value.find((e) => e.target === nodeId)
  const outEdge = edges.value.find((e) => e.source === nodeId)
  if (!inEdge || !outEdge) return
  const prevId = inEdge.source
  const nextId = outEdge.target
  nodes.value = nodes.value.filter((n) => n.id !== nodeId)
  edges.value = edges.value.filter(
    (e) => e.source !== nodeId && e.target !== nodeId
  )
  edges.value.push({ id: `e-${prevId}-${nextId}`, source: prevId, target: nextId })
  applyLayout()
}

function getTransformationOrder() {
  const order = getNodeOrder()
  return order.filter((id) => id.startsWith('t-'))
}

function moveTransformationUp(nodeId) {
  const order = getTransformationOrder()
  const idx = order.indexOf(nodeId)
  if (idx <= 0) return
  const prevId = order[idx - 1]
  swapTransformationData(nodeId, prevId)
}

function moveTransformationDown(nodeId) {
  const order = getTransformationOrder()
  const idx = order.indexOf(nodeId)
  if (idx < 0 || idx >= order.length - 1) return
  const nextId = order[idx + 1]
  swapTransformationData(nodeId, nextId)
}

function canMoveTransformationUp(nodeId) {
  const order = getTransformationOrder()
  return order.indexOf(nodeId) > 0
}

function canMoveTransformationDown(nodeId) {
  const order = getTransformationOrder()
  const idx = order.indexOf(nodeId)
  return idx >= 0 && idx < order.length - 1
}

function swapTransformationData(idA, idB) {
  const next = nodes.value.map((n) => {
    if (n.id === idA) {
      const nodeB = nodes.value.find((x) => x.id === idB)
      return nodeB ? { ...n, data: { ...nodeB.data } } : n
    }
    if (n.id === idB) {
      const nodeA = nodes.value.find((x) => x.id === idA)
      return nodeA ? { ...n, data: { ...nodeA.data } } : n
    }
    return n
  })
  nodes.value = next
}

function onNodeClick({ node }) {
  const type = node.type
  if (type === 'source' || type === 'sink' || type === 'transformation') {
    openConfig(null, type, node)
  }
}

function openConfig(ev, nodeType, node) {
  const nodeId = node?.id || ''
  const fromGraph = nodes.value.find((n) => n.id === nodeId)
  const source = fromGraph || node
  const nodeData = source?.data
    ? JSON.parse(JSON.stringify(source.data))
    : {}
  configPanel.value = {
    open: true,
    nodeId,
    node: { ...source, id: nodeId, data: nodeData },
    nodeType,
  }
}

function onConfigSave(updatedNode) {
  const idx = nodes.value.findIndex((n) => n.id === updatedNode.id)
  if (idx >= 0) {
    const next = nodes.value.map((n) =>
      n.id === updatedNode.id
        ? { ...n, data: { ...n.data, ...updatedNode.data } }
        : n
    )
    nodes.value = next
  }
  configPanel.value.open = false
}

function buildManifest() {
  try {
    const nodesPlain = JSON.parse(JSON.stringify(nodes.value))
    const edgesPlain = JSON.parse(JSON.stringify(edges.value))
    const name = manifestName.value?.trim() || props.initialManifest?.metadata?.name || 'dataflow'
    const ns = props.namespace || props.initialManifest?.metadata?.namespace || 'default'
    const manifest = graphToManifest(nodesPlain, edgesPlain, {
      ...(props.initialManifest || {}),
      metadata: {
        ...(props.initialManifest?.metadata || {}),
        name,
        namespace: ns,
      },
    })
    return manifest
  } catch {
    return null
  }
}

const previewManifest = computed(() => buildManifest())

function openPreview() {
  if (previewManifest.value) previewOpen.value = true
}

function emitSave() {
  const manifest = buildManifest()
  if (manifest) {
    emit('save', manifest)
  } else {
    emit('save', null, new Error('Invalid flow'))
  }
}

function exportManifest() {
  const manifest = buildManifest()
  if (!manifest) return
  const yamlText = yaml.dump(manifest, {
    indent: 2,
    lineWidth: -1,
    quotingType: '"',
    forceQuotes: false,
  })
  const blob = new Blob([yamlText], { type: 'application/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `dataflow-${manifest.metadata?.name || 'manifest'}.yaml`
  a.click()
  URL.revokeObjectURL(url)
}

function copyManifest() {
  const manifest = buildManifest()
  if (!manifest) return
  const yamlText = yaml.dump(manifest, {
    indent: 2,
    lineWidth: -1,
    quotingType: '"',
    forceQuotes: false,
  })
  navigator.clipboard.writeText(yamlText).then(() => {
    success(t('flow.exported'))
  })
}
</script>

<style scoped>
.flow-canvas-wrap {
  display: flex;
  flex-direction: column;
  height: 500px;
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.flow-toolbar {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
}

.flow-name-input {
  flex: 1;
  max-width: 320px;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.95rem;
}

.flow-canvas {
  flex: 1;
  min-height: 400px;
}

.flow-canvas :deep(.vue-flow) {
  background: var(--bg-page);
}
</style>
