// ECharts按需加载工具
// 使用数据驱动的方式，一个函数处理所有图表类型

let echartsInstance = null

/**
 * 加载ECharts及指定的图表类型
 * @param {Array<string>} chartTypes - 需要的图表类型，如 ['Line', 'Pie', 'Heatmap']
 * @returns {Object} { init } - ECharts的init函数
 */
export const loadECharts = async (chartTypes = []) => {
  if (echartsInstance) return echartsInstance

  const [{ init, use }, chartsModule, componentsModule, { CanvasRenderer }] = await Promise.all([
    import('echarts/core'),
    import('echarts/charts'),
    import('echarts/components'),
    import('echarts/renderers')
  ])

  // 根据需要的图表类型动态注册
  const charts = chartTypes
    .map((type) => chartsModule[`${type}Chart`])
    .filter(Boolean)

  // 注册常用组件
  const components = [
    componentsModule.TitleComponent,
    componentsModule.TooltipComponent,
    componentsModule.LegendComponent,
    componentsModule.GridComponent,
    componentsModule.CalendarComponent,
    componentsModule.VisualMapComponent
  ].filter(Boolean)

  use([...charts, ...components, CanvasRenderer])

  echartsInstance = { init }
  return echartsInstance
}

/**
 * 图表resize防抖工具
 * @param {Object} chart - ECharts实例
 * @returns {Function} cleanup函数
 */
export const useChartResize = (chart) => {
  let timer = null
  const resizeHandler = () => {
    clearTimeout(timer)
    timer = setTimeout(() => chart?.resize(), 200)
  }

  window.addEventListener('resize', resizeHandler)
  return () => {
    window.removeEventListener('resize', resizeHandler)
    clearTimeout(timer)
  }
}
