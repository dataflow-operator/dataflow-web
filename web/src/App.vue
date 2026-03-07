<template>
  <div class="app">
    <header class="app-header" role="banner">
      <h1>{{ t('app.title') }}</h1>
      <div class="header-actions">
        <nav class="app-nav" :aria-label="t('app.navAria')">
          <router-link to="/">{{ t('nav.dashboard') }}</router-link>
          <router-link to="/manifests">{{ t('nav.manifests') }}</router-link>
          <router-link to="/logs">{{ t('nav.logs') }}</router-link>
          <router-link to="/metrics">{{ t('nav.metrics') }}</router-link>
          <router-link to="/events">{{ t('nav.events') }}</router-link>
        </nav>
        <div class="header-controls">
          <select
            :value="locale"
            class="locale-select"
            :aria-label="t('locale.label')"
            @change="onLocaleChange"
          >
            <option value="ru">{{ t('locale.ru') }}</option>
            <option value="en">{{ t('locale.en') }}</option>
          </select>
          <button
            type="button"
            class="theme-toggle"
            :aria-label="isDark ? t('theme.switchLight') : t('theme.switchDark')"
            @click="toggleTheme"
          >
            {{ isDark ? t('theme.light') : t('theme.dark') }}
          </button>
        </div>
      </div>
    </header>
    <main class="container" role="main">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
    <ToastContainer />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, i18n } from './i18n'
import ToastContainer from './components/ToastContainer.vue'

const { t } = useI18n()
const locale = i18n.global.locale
const isDark = ref(false)

function onLocaleChange(e) {
  const value = e.target.value
  setLocale(value)
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.setAttribute('data-theme', isDark.value ? 'dark' : '')
  try {
    localStorage.setItem('dataflow-theme', isDark.value ? 'dark' : 'light')
  } catch (_) {}
}

onMounted(() => {
  const saved = localStorage.getItem('dataflow-theme')
  if (saved === 'dark') {
    isDark.value = true
    document.documentElement.setAttribute('data-theme', 'dark')
  }
})
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.locale-select {
  padding: 0.4rem 0.6rem;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  background: rgba(255, 255, 255, 0.15);
  color: white;
  font-size: 0.9rem;
  cursor: pointer;
}

.locale-select:hover {
  background: rgba(255, 255, 255, 0.25);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
