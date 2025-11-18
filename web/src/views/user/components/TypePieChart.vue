<template>
  <div ref="chartRef" class="chart-container"></div>
</template>

<script>
import { loadECharts, useChartResize } from '@/utils/echarts'

export default {
  name: 'TypePieChart',
  props: {
    distribution: {
      type: Array,
      required: true
    }
  },
  data() {
    return {
      chart: null,
      cleanupResize: null
    }
  },
  watch: {
    distribution: {
      handler() {
        this.renderChart()
      },
      deep: true
    }
  },
  mounted() {
    this.renderChart()
  },
  beforeUnmount() {
    if (this.cleanupResize) {
      this.cleanupResize()
    }
    if (this.chart) {
      this.chart.dispose()
    }
  },
  methods: {
    async renderChart() {
      if (!this.distribution || this.distribution.length === 0) return

      const { init } = await loadECharts(['Pie'])
      this.chart = init(this.$refs.chartRef)

      const option = {
        tooltip: {
          trigger: 'item',
          formatter: '{b}: {c} ({d}%)'
        },
        legend: {
          orient: 'vertical',
          right: 'right'
        },
        series: [
          {
            type: 'pie',
            radius: '70%',
            data: this.distribution,
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.5)'
              }
            }
          }
        ]
      }

      this.chart.setOption(option)
      this.cleanupResize = useChartResize(this.chart)
    }
  }
}
</script>

<style scoped>
.chart-container {
  width: 100%;
  height: 300px;
}
</style>
