<template>
  <div class="metric-chart card">
    <div class="card-header">
      <h3>{{ title }}</h3>
      <div v-if="subtitle" class="subtitle">{{ subtitle }}</div>
    </div>

    <div v-if="loading" class="chart-loading">
      <LoadingSpinner :message="t('metrics.loadingCharts')" />
    </div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else class="chart-body">
      <VChart class="chart" :option="option" :autoresize="true" />
      <div v-if="message" class="chart-message">{{ message }}</div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import LoadingSpinner from './LoadingSpinner.vue'

const { t } = useI18n()

use([CanvasRenderer, LineChart, GridComponent, LegendComponent, TooltipComponent])

const props = defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  series: { type: Array, default: () => [] }, // [{ metric, values }]
  loading: { type: Boolean, default: false },
  error: { type: String, default: '' },
  message: { type: String, default: '' },
})

function formatLegend(metric = {}) {
  const parts = []
  for (const [k, v] of Object.entries(metric)) {
    if (k === '__name__' || k === 'namespace' || k === 'name') continue
    parts.push(`${k}=${v}`)
  }
  return parts.join(', ') || metric.__name__ || 'series'
}

const option = computed(() => {
  const series = (props.series || []).map((s) => ({
    name: formatLegend(s.metric),
    type: 'line',
    showSymbol: false,
    smooth: true,
    data: (s.values || [])
      .map(([ts, value]) => [Number(ts) * 1000, Number(value)])
      .filter((p) => Number.isFinite(p[0]) && Number.isFinite(p[1])),
  }))
  return {
    tooltip: { trigger: 'axis' },
    legend: { type: 'scroll' },
    grid: { left: 40, right: 20, top: 30, bottom: 30 },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', scale: true },
    series,
  }
})
</script>

<style scoped>
.subtitle {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.chart-body {
  padding: 0.5rem 1rem 1rem;
}

.chart {
  height: 260px;
}

.chart-message {
  margin-top: 0.5rem;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.chart-loading {
  padding: 1rem;
}
</style>

