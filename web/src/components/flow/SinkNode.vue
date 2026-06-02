<template>
  <div class="flow-node flow-node-sink" @click="$emit('config', $event)">
    <Handle type="target" :position="Position.Left" />
    <div class="flow-node-label">
      <span class="flow-node-icon">📤</span>
      {{ data?.label || 'Sink' }}
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
.flow-node-sink {
  padding: 0.75rem 1rem;
  border-radius: 8px;
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
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
