<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-overlay"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.self="close"
    >
      <div class="modal-content">
        <div class="modal-header">
          <h2 :id="titleId" class="modal-title">{{ title }}</h2>
          <button
            type="button"
            class="modal-close"
            :aria-label="t('modal.closeAria')"
            @click="close"
          >
            &times;
          </button>
        </div>
        <div class="modal-body">
          <div ref="editorWrap" class="yaml-editor-wrap"></div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="close">
            {{ t('yaml.close') }}
          </button>
          <button
            v-if="mode === 'edit'"
            type="button"
            class="btn btn-primary"
            @click="save"
          >
            {{ t('yaml.save') }}
          </button>
          <button
            v-if="mode === 'create'"
            type="button"
            class="btn btn-primary"
            @click="create"
          >
            {{ t('yaml.create') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import CodeMirror from 'codemirror'

const { t } = useI18n()
import 'codemirror/mode/yaml/yaml'
import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/monokai.css'
import yaml from 'js-yaml'

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  value: { type: [Object, String], default: null },
  mode: { type: String, default: 'edit' }, // 'view' | 'edit' | 'create'
})
const emit = defineEmits(['close', 'save', 'create'])

const titleId = computed(() => 'modal-title-' + Math.random().toString(36).slice(2))
const editorWrap = ref(null)
let editor = null

function getYamlText(val) {
  if (val == null) return ''
  if (typeof val === 'string') return val
  try {
    return yaml.dump(val, { indent: 2, lineWidth: -1, quotingType: '"', forceQuotes: false })
  } catch {
    return JSON.stringify(val, null, 2)
  }
}

function initEditor() {
  if (!editorWrap.value || editor) return
  editor = CodeMirror(editorWrap.value, {
    value: getYamlText(props.value),
    mode: 'yaml',
    theme: 'monokai',
    lineNumbers: true,
    indentUnit: 2,
    lineWrapping: true,
    readOnly: props.mode === 'view',
  })
}

function close() {
  emit('close')
}

function getParsed() {
  if (!editor) return null
  const text = editor.getValue()
  try {
    return yaml.load(text) || null
  } catch {
    try {
      return JSON.parse(text)
    } catch (e) {
      throw new Error(t('yaml.invalidFormat', { msg: e.message }))
    }
  }
}

function save() {
  try {
    const parsed = getParsed()
    if (!parsed) throw new Error(t('yaml.emptyManifest'))
    emit('save', parsed)
  } catch (e) {
    emit('save', null, e)
  }
}

function create() {
  try {
    const parsed = getParsed()
    if (!parsed) throw new Error(t('yaml.emptyManifest'))
    if (!parsed.metadata?.name) throw new Error(t('yaml.nameRequired'))
    emit('create', parsed)
  } catch (e) {
    emit('create', null, e)
  }
}

function destroyEditor() {
  if (editor && editor.getWrapperElement()) {
    editor.getWrapperElement().remove()
    editor = null
  }
}

watch(
  () => [props.open, props.value],
  async () => {
    if (!props.open) {
      destroyEditor()
      return
    }
    await nextTick()
    let attempts = 0
    const maxAttempts = 10
    const tryInit = () => {
      if (!props.open) return
      if (editorWrap.value) {
        if (!editor) {
          initEditor()
        } else {
          editor.setValue(getYamlText(props.value))
          editor.setOption('readOnly', props.mode === 'view')
        }
      } else if (attempts++ < maxAttempts) {
        requestAnimationFrame(tryInit)
      }
    }
    tryInit()
  },
  { immediate: true }
)

onMounted(async () => {
  if (props.open) {
    await nextTick()
    if (editorWrap.value) initEditor()
  }
})

onUnmounted(destroyEditor)
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.modal-content {
  background: var(--bg-card);
  border-radius: 8px;
  width: 100%;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

.modal-header {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, var(--gradient-start), var(--gradient-end));
  color: white;
  border-radius: 8px 8px 0 0;
}

.modal-title {
  margin: 0;
  font-size: 1.25rem;
}

.modal-close {
  background: none;
  border: none;
  color: white;
  font-size: 1.75rem;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  border-radius: 4px;
}

.modal-close:hover {
  opacity: 0.85;
}

.modal-body {
  padding: 1rem;
  overflow: auto;
  flex: 1;
  min-height: 300px;
}

.yaml-editor-wrap {
  min-height: 400px;
}

.yaml-editor-wrap :deep(.CodeMirror) {
  height: 450px;
  font-size: 0.9rem;
  border: 1px solid var(--border);
  border-radius: 6px;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
</style>
