import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

function readQueryValue(value) {
  if (value == null) return ''
  return Array.isArray(value) ? String(value[0] || '') : String(value)
}

function queriesEqual(a, b) {
  const keys = new Set([...Object.keys(a || {}), ...Object.keys(b || {})])
  for (const key of keys) {
    if (readQueryValue(a?.[key]) !== readQueryValue(b?.[key])) return false
  }
  return true
}

/**
 * Keeps namespace / dataflow filters in sync with URL query params
 * (?namespace=&dataflow=) for reload and shareable links.
 */
export function useFilterQueryParams(options = {}) {
  const { dataflow: includeDataflow = false } = options
  const route = useRoute()
  const router = useRouter()

  const namespace = ref(readQueryValue(route.query.namespace) || 'default')
  const dataflow = ref(includeDataflow ? readQueryValue(route.query.dataflow) : '')

  let syncingFromRoute = false

  function buildQuery() {
    const query = { ...route.query }
    const ns = namespace.value.trim()
    if (ns) {
      query.namespace = ns
    } else {
      delete query.namespace
    }
    if (includeDataflow) {
      const df = dataflow.value.trim()
      if (df) {
        query.dataflow = df
      } else {
        delete query.dataflow
      }
    }
    return query
  }

  function syncToRoute() {
    if (syncingFromRoute) return
    const query = buildQuery()
    if (queriesEqual(query, route.query)) return
    router.replace({ query })
  }

  watch(namespace, syncToRoute)
  if (includeDataflow) {
    watch(dataflow, syncToRoute)
  }

  watch(
    () => route.query,
    (query) => {
      syncingFromRoute = true
      const ns = readQueryValue(query.namespace)
      if (ns) {
        namespace.value = ns
      } else if (!readQueryValue(query.namespace)) {
        namespace.value = 'default'
      }
      if (includeDataflow) {
        dataflow.value = readQueryValue(query.dataflow)
      }
      syncingFromRoute = false
    }
  )

  return { namespace, dataflow }
}
