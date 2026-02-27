import { createI18n } from 'vue-i18n'
import ru from './locales/ru'
import en from './locales/en'

const savedLocale = (typeof localStorage !== 'undefined' && localStorage.getItem('dataflow-locale')) || 'ru'

export const i18n = createI18n({
  legacy: false,
  locale: savedLocale,
  fallbackLocale: 'en',
  messages: { ru, en },
})

if (typeof document !== 'undefined' && document.documentElement) {
  document.documentElement.lang = savedLocale === 'ru' ? 'ru' : 'en'
}

export function setLocale(locale) {
  i18n.global.locale.value = locale
  if (typeof document !== 'undefined' && document.documentElement) {
    document.documentElement.lang = locale === 'ru' ? 'ru' : 'en'
  }
  try {
    localStorage.setItem('dataflow-locale', locale)
  } catch (_) {}
}
