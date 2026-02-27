<template>
  <div class="form-group namespace-selector">
    <label for="namespace-select">{{ t('namespaceSelect.label') }}</label>
    <select
      id="namespace-select"
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target).value)"
      :aria-label="t('namespaceSelect.aria')"
    >
      <option v-if="loading" value="">{{ t('common.loading') }}</option>
      <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
    </select>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
import { getNamespaces } from '../api/client'

const props = defineProps({
  modelValue: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])

const namespaces = ref([])
const loading = ref(true)

async function load() {
  loading.value = true
  try {
    namespaces.value = await getNamespaces()
  } catch (e) {
    namespaces.value = ['default']
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(namespaces, (list) => {
  if (list.length && props.modelValue && !list.includes(props.modelValue)) {
    emit('update:modelValue', list[0])
  }
}, { flush: 'post' })
</script>
