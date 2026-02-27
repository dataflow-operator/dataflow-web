<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-overlay"
      role="alertdialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      @click.self="$emit('cancel')"
    >
      <div class="modal-content confirm-modal">
        <div class="modal-header">
          <h2 :id="titleId" class="modal-title">{{ title || t('confirm.defaultTitle') }}</h2>
          <button
            type="button"
            class="modal-close"
            :aria-label="t('modal.closeAria')"
            @click="$emit('cancel')"
          >
            &times;
          </button>
        </div>
        <div class="modal-body">
          <p>{{ message }}</p>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="$emit('cancel')">
            {{ t('confirm.cancel') }}
          </button>
          <button type="button" class="btn btn-danger" @click="$emit('confirm')">
            {{ confirmLabel ?? t('common.delete') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  message: { type: String, default: '' },
  confirmLabel: { type: String, default: null },
})
defineEmits(['cancel', 'confirm'])

const titleId = computed(() => 'confirm-title-' + Math.random().toString(36).slice(2))
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1001;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.modal-content.confirm-modal {
  max-width: 420px;
}

.modal-header {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-card);
  color: var(--text);
  border-radius: 8px 8px 0 0;
}

.modal-title {
  margin: 0;
  font-size: 1.15rem;
}

.modal-close {
  background: none;
  border: none;
  color: var(--text);
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.modal-body {
  padding: 1.25rem 1.5rem;
}

.modal-body p {
  margin: 0;
  color: var(--text-muted);
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
</style>
