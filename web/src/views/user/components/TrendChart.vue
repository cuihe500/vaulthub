<template>
  <div ref="chartRef" class="chart-container"></div>
</template>

<script>
import { loadECharts, useChartResize } from '@/utils/echarts'
import dayjs from 'dayjs'

export default {
  name: 'TrendChart',
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
    chartData() {
      // 按日期升序排序
      const sorted = [...this.data].sort((a, b) => new Date(a.stat_date) - new Date(b.stat_date))

      return {
        dates: sorted.map((item) => dayjs(item.stat_date).format('MM-DD')),
        total: sorted.map((item) => item.total_secrets),
        apiKey: sorted.map((item) => item.api_key_count),
        password: sorted.map((item) => item.password_count),
        certificate: sorted.map((item) => item.certificate_count),
        sshKey: sorted.map((item) => item.ssh_key_count),
        database: sorted.map((item) => item.database_count),
        token: sorted.map((item) => item.token_count)
      }
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

      const { init } = await loadECharts(['Line'])
      this.chart = init(this.$refs.chartRef)

      const option = {
        tooltip: {
          trigger: 'axis'
        },
        legend: {
          data: ['总计', 'API密钥', '密码', '证书', 'SSH密钥', '数据库凭证', 'Token']
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true
        },
        xAxis: {
          type: 'category',
          boundaryGap: false,
          data: this.chartData.dates
        },
        yAxis: {
          type: 'value'
        },
        series: [
          {
            name: '总计',
            type: 'line',
            data: this.chartData.total,
            smooth: true
          },
          {
            name: 'API密钥',
            type: 'line',
            data: this.chartData.apiKey,
            smooth: true
          },
          {
            name: '密码',
            type: 'line',
            data: this.chartData.password,
            smooth: true
          },
          {
            name: '证书',
            type: 'line',
            data: this.chartData.certificate,
            smooth: true
          },
          {
            name: 'SSH密钥',
            type: 'line',
            data: this.chartData.sshKey,
            smooth: true
          },
          {
            name: '数据库凭证',
            type: 'line',
            data: this.chartData.database,
            smooth: true
          },
          {
            name: 'Token',
            type: 'line',
            data: this.chartData.token,
            smooth: true
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
  height: 400px;
}
</style>
