<template>
  <div class="config-form">
    <template v-if="form.useAdvanced">
      <div class="form-group">
        <label>{{ t('flow.advancedConfig') }}</label>
        <textarea
          v-model="form.advancedJson"
          rows="10"
          class="config-json"
          :placeholder="'{}'"
        />
      </div>
    </template>
    <template v-else>
      <!-- timestamp -->
      <template v-if="transformationType === 'timestamp'">
        <div class="form-group">
          <label>{{ t('flow.timestampFieldName') }}</label>
          <input
            v-model="form.fieldName"
            type="text"
            class="form-input"
            :placeholder="t('flow.timestampFieldNamePlaceholder')"
          />
        </div>
        <div class="form-group">
          <label>{{ t('flow.timestampFormat') }}</label>
          <select v-model="form.format" class="form-select">
            <option value="RFC3339">RFC3339</option>
            <option value="Unix">Unix</option>
            <option value="custom">Custom (specify in advanced)</option>
          </select>
        </div>
      </template>

      <!-- flatten -->
      <template v-else-if="transformationType === 'flatten'">
        <div class="form-group">
          <label>{{ t('flow.flattenField') }}</label>
          <input
            v-model="form.field"
            type="text"
            class="form-input"
            :placeholder="t('flow.flattenFieldPlaceholder')"
          />
        </div>
      </template>

      <!-- filter -->
      <template v-else-if="transformationType === 'filter'">
        <div class="form-group">
          <label>{{ t('flow.filterCondition') }}</label>
          <input
            v-model="form.condition"
            type="text"
            class="form-input"
            :placeholder="t('flow.filterConditionPlaceholder')"
          />
        </div>
      </template>

      <!-- select -->
      <template v-else-if="transformationType === 'select'">
        <div class="form-group">
          <label>{{ t('flow.selectFields') }}</label>
          <textarea
            v-model="form.fields"
            rows="4"
            class="form-input"
            :placeholder="t('flow.selectFieldsPlaceholder')"
          />
        </div>
      </template>

      <!-- remove -->
      <template v-else-if="transformationType === 'remove'">
        <div class="form-group">
          <label>{{ t('flow.removeFields') }}</label>
          <textarea
            v-model="form.fields"
            rows="4"
            class="form-input"
            :placeholder="t('flow.removeFieldsPlaceholder')"
          />
        </div>
      </template>

      <!-- mask -->
      <template v-else-if="transformationType === 'mask'">
        <div class="form-group">
          <label>{{ t('flow.maskFields') }}</label>
          <textarea
            v-model="form.fields"
            rows="4"
            class="form-input"
            :placeholder="t('flow.maskFieldsPlaceholder')"
          />
        </div>
        <div class="form-group form-row">
          <label class="checkbox-label">
            <input v-model="form.keepLength" type="checkbox" />
            {{ t('flow.maskKeepLength') }}
          </label>
        </div>
      </template>

      <!-- router -->
      <template v-else-if="transformationType === 'router'">
        <p class="form-hint">{{ t('flow.routerConditionHint') }}</p>
        <div class="routes-list">
          <div
            v-for="(route, index) in form.routes"
            :key="index"
            class="route-card"
          >
            <div class="route-header">
              <span class="route-title">{{ t('flow.routerRouteLabel', { index: index + 1 }) }}</span>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="form.routes.length <= 1"
                @click="removeRoute(index)"
              >
                {{ t('flow.routerRemoveRoute') }}
              </button>
            </div>
            <div class="form-group">
              <label>{{ t('flow.routerCondition') }}</label>
              <input
                v-model="route.condition"
                type="text"
                class="form-input"
                :placeholder="t('flow.routerConditionPlaceholder')"
              />
            </div>
            <div class="form-group">
              <label>{{ t('flow.routerSinkType') }}</label>
              <select v-model="route.sinkType" class="form-select">
                <option v-for="ct in connectorTypes" :key="ct" :value="ct">
                  {{ ct }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>{{ t('flow.routerSinkConfig') }}</label>
              <textarea
                v-model="route.sinkConfigJson"
                rows="6"
                class="config-json"
                :placeholder="'{}'"
              />
            </div>
          </div>
        </div>
        <button type="button" class="btn btn-secondary btn-sm route-add" @click="addRoute">
          {{ t('flow.routerAddRoute') }}
        </button>
      </template>

      <!-- snakeCase, camelCase -->
      <template v-else-if="transformationType === 'snakeCase' || transformationType === 'camelCase'">
        <div class="form-group form-row">
          <label class="checkbox-label">
            <input v-model="form.deep" type="checkbox" />
            {{ t('flow.transformDeep') }}
          </label>
        </div>
      </template>

      <div v-if="needsAdvancedOption" class="form-group form-row advanced-toggle">
        <label class="checkbox-label">
          <input v-model="showAdvanced" type="checkbox" />
          {{ t('flow.showAdvanced') }}
        </label>
      </div>
      <div v-if="showAdvanced && needsAdvancedOption" class="form-group">
        <label>{{ t('flow.advancedConfig') }}</label>
        <textarea
          v-model="form.advancedJson"
          rows="6"
          class="config-json"
          :placeholder="'{}'"
        />
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { CONNECTOR_TYPES } from '../../composables/useFlowManifest'

const props = defineProps({
  modelValue: { type: Object, required: true },
  transformationType: { type: String, required: true },
})

const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const form = ref({ ...props.modelValue })
const showAdvanced = ref(false)
const connectorTypes = CONNECTOR_TYPES

const needsAdvancedOption = computed(() =>
  ['timestamp', 'flatten', 'filter', 'select', 'remove', 'mask', 'snakeCase', 'camelCase', 'router'].includes(
    props.transformationType
  )
)

function ensureRouterRoutes() {
  if (!Array.isArray(form.value.routes) || form.value.routes.length === 0) {
    form.value.routes = [{ condition: '', sinkType: 'kafka', sinkConfigJson: '{}' }]
  }
}

function addRoute() {
  ensureRouterRoutes()
  form.value.routes.push({ condition: '', sinkType: 'kafka', sinkConfigJson: '{}' })
}

function removeRoute(index) {
  if (!Array.isArray(form.value.routes) || form.value.routes.length <= 1) return
  form.value.routes.splice(index, 1)
}

watch(
  () => props.modelValue,
  (v) => {
    form.value = { ...v, advancedJson: v?.advancedJson ?? '' }
    if (props.transformationType === 'router') {
      ensureRouterRoutes()
    }
    showAdvanced.value = !!(v?.advancedJson && v.advancedJson.trim())
  },
  { deep: true, immediate: true }
)

watch(
  () => props.transformationType,
  (type) => {
    if (type === 'router') {
      ensureRouterRoutes()
    }
  }
)

watch(
  form,
  () => emit('update:modelValue', { ...form.value }),
  { deep: true }
)
</script>

<style scoped>
.config-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.form-group.form-row {
  flex-direction: row;
  align-items: center;
}

.form-input {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.9rem;
  background: var(--bg-page);
  color: var(--text);
}

.form-select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 0.9rem;
  background: var(--bg-page);
  color: var(--text);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-weight: normal;
}

.config-json {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
  background: var(--bg-page);
  color: var(--text);
}

.form-hint {
  margin: 0;
  color: var(--text-muted);
  font-size: 0.9rem;
}

.advanced-toggle {
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border);
}

.routes-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.route-card {
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0.75rem;
  background: var(--bg-page);
}

.route-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.route-title {
  font-weight: 600;
  font-size: 0.9rem;
}

.route-add {
  align-self: flex-start;
}
</style>
