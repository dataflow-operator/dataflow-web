<template>
  <div class="flow-node flow-node-transformation" @click="$emit('config', $event)">
    <Handle type="target" :position="Position.Left" />
    <Handle type="source" :position="Position.Right" />
    <div class="flow-node-actions">
      <button
        type="button"
        class="flow-node-action"
        :aria-label="t('flow.moveUp')"
        :disabled="!canMoveUp"
        :title="t('flow.moveUp')"
        @click.stop="canMoveUp && $emit('move-up', node?.id || id)"
      >
        ↑
      </button>
      <button
        type="button"
        class="flow-node-action"
        :aria-label="t('flow.moveDown')"
        :disabled="!canMoveDown"
        :title="t('flow.moveDown')"
        @click.stop="canMoveDown && $emit('move-down', node?.id || id)"
      >
        ↓
      </button>
      <button
        type="button"
        class="flow-node-action flow-node-remove"
        :aria-label="t('flow.removeTransformation')"
        @click.stop="$emit('remove', node?.id || id)"
      >
        ×
      </button>
    </div>
    <div class="flow-node-label">
      <span class="flow-node-icon">⚙</span>
      {{ data?.label || 'Transform' }}
    </div>
  </div>
</template>

<script setup>
import { Handle, Position } from '@vue-flow/core'
import { useI18n } from 'vue-i18n'

defineProps({
  id: { type: String, default: '' },
  node: { type: Object, default: null },
  data: { type: Object, default: () => ({}) },
  canMoveUp: { type: Boolean, default: false },
  canMoveDown: { type: Boolean, default: false },
})

defineEmits(['config', 'remove', 'move-up', 'move-down'])

const { t } = useI18n()
</script>

<style scoped>
.flow-node-transformation {
  position: relative;
  padding: 0.75rem 1rem;
  padding-right: 4rem;
  border-radius: 8px;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
  min-width: 120px;
  text-align: center;
  border: 2px solid rgba(255, 255, 255, 0.3);
}

.flow-node-actions {
  position: absolute;
  top: 4px;
  right: 4px;
  display: flex;
  gap: 2px;
}

.flow-node-action {
  width: 22px;
  height: 22px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.3);
  color: white;
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
}

.flow-node-action:hover:not(:disabled) {
  background: rgba(0, 0, 0, 0.5);
}

.flow-node-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.flow-node-label {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.flow-node-icon {
  font-size: 1.2rem;
}
</style>
