<template>
  <div ref="chartRef" class="chart-container"></div>
</template>

<script>
import { loadECharts, useChartResize } from '@/utils/echarts'
import dayjs from 'dayjs'

export default {
  name: 'ActivityHeatmap',
  props: {
    data: {
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
  computed: {
    heatmapData() {
      return this.data.map((item) => [dayjs(item.stat_date).format('YYYY-MM-DD'), item.total_operations])
    },
    maxOperations() {
      if (this.data.length === 0) return 0
      return Math.max(...this.data.map((item) => item.total_operations))
    },
    calendarRange() {
      if (this.data.length === 0) return [dayjs().format('YYYY-MM-DD'), dayjs().format('YYYY-MM-DD')]
      const dates = this.data.map((item) => dayjs(item.stat_date))
      const minDate = dayjs.min(dates)
      const maxDate = dayjs.max(dates)
      return [minDate.format('YYYY-MM-DD'), maxDate.format('YYYY-MM-DD')]
    }
  },
  watch: {
    data: {
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
      if (!this.data || this.data.length === 0) return

      const { init } = await loadECharts(['Heatmap'])
      this.chart = init(this.$refs.chartRef)

      const option = {
        tooltip: {
          position: 'top',
          formatter: (params) => {
            return `${params.value[0]}<br/>操作次数: ${params.value[1]}`
          }
        },
        visualMap: {
          min: 0,
          max: this.maxOperations,
          calculable: true,
          orient: 'horizontal',
          left: 'center',
          bottom: '5%',
          inRange: {
            color: ['#ebedf0', '#c6e48b', '#7bc96f', '#239a3b', '#196127']
          }
        },
        calendar: {
          range: this.calendarRange,
          cellSize: ['auto', 15],
          left: 50,
          right: 30,
          top: 30,
          bottom: 80,
          yearLabel: { show: false }
        },
        series: [
          {
            type: 'heatmap',
            coordinateSystem: 'calendar',
            data: this.heatmapData
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
  height: 200px;
}
</style>
