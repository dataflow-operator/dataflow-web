<template>
  <div class="flow-node flow-node-source" @click="$emit('config', $event)">
    <Handle type="source" :position="Position.Right" />
    <div class="flow-node-label">
      <span class="flow-node-icon">📥</span>
      {{ data?.label || 'Source' }}
    </div>
  </div>
</template>

<script setup>
import { Handle, Position } from '@vue-flow/core'

defineProps({
  data: { type: Object, default: () => ({}) },
})

defineEmits(['config'])
</script>

<style scoped>
.flow-node-source {
  padding: 0.75rem 1rem;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  min-width: 120px;
  text-align: center;
  border: 2px solid rgba(255, 255, 255, 0.3);
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
