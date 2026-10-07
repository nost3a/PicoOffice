// minimal hand-rolled i18n: reactive dict + global $t, default zh, persists in localStorage
import { reactive } from 'vue'
import zh from './zh'
import en from './en'

const dicts = { zh, en }

const state = reactive({
  lang: localStorage.getItem('pico_lang') || 'zh'
})

// lookup: supports {name} interpolation; missing key falls back to zh, then to key
export function t(key, vars) {
  let s = dicts[state.lang]?.[key] ?? dicts.zh[key] ?? key
  if (vars && typeof s === 'string') {
    for (const k in vars) {
      s = s.replace(new RegExp('\\{' + k + '\\}', 'g'), vars[k])
    }
  }
  return s
}

export function getLocale() {
  return state.lang
}

export function setLocale(lang) {
  state.lang = dicts[lang] ? lang : 'zh'
  localStorage.setItem('pico_lang', state.lang)
  // also update document dir/lang attributes
  document.documentElement.lang = state.lang
}

// in component: const { t, locale, setLocale } = useI18n()
export function useI18n() {
  return { t, state, locale: () => state.lang, setLocale }
}

// after setupI18n(app) in main.js, templates can use $t('key')
export function setupI18n(app) {
  app.config.globalProperties.$t = t
  document.documentElement.lang = state.lang
}
