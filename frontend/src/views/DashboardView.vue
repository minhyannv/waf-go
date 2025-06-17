<template>
  <div class="dashboard">
    <!-- 页面头部 -->
    <div class="dashboard-header">
      <div class="header-left">
        <h2>安全仪表盘</h2>
        <p class="header-desc">实时监控WAF防护状态和攻击趋势</p>
      </div>
      <div class="header-right">
        <el-button-group>
          <el-button 
            :type="autoRefresh ? 'primary' : 'default'" 
            @click="toggleAutoRefresh"
            :icon="autoRefresh ? 'VideoPause' : 'VideoPlay'"
          >
            {{ autoRefresh ? '停止刷新' : '自动刷新' }}
          </el-button>
          <el-button @click="refreshData" :loading="loading" icon="Refresh">
            刷新数据
          </el-button>
          <el-button @click="exportData" icon="Download">
            导出报告
          </el-button>
        </el-button-group>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="6">
        <div class="stat-card total-requests" @click="showDetailModal('total')">
          <div class="stat-icon">
            <el-icon><DataAnalysis /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(stats.total_requests) }}</div>
            <div class="stat-label">总请求数</div>
            <div class="stat-extra">今日总计</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="12" :sm="6">
        <div class="stat-card blocked-requests" @click="showDetailModal('blocked')">
          <div class="stat-icon">
            <el-icon><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(stats.blocked_requests) }}</div>
            <div class="stat-label">拦截请求</div>
            <div class="stat-extra">{{ blockRate }}% 拦截率</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="12" :sm="6">
        <div class="stat-card allowed-requests" @click="showDetailModal('allowed')">
          <div class="stat-icon">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatNumber(stats.allowed_requests) }}</div>
            <div class="stat-label">允许请求</div>
            <div class="stat-extra">正常通过</div>
          </div>
        </div>
      </el-col>
      
      <el-col :xs="12" :sm="6">
        <div class="stat-card active-rules" @click="showDetailModal('rules')">
          <div class="stat-icon">
            <el-icon><Lock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.active_rules }}</div>
            <div class="stat-label">活跃规则</div>
            <div class="stat-extra">{{ stats.active_policies }} 个策略</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 攻击趋势分析 -->
    <el-row :gutter="20" class="trend-section">
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <h3>攻击趋势分析</h3>
                <span class="header-desc">最近{{ selectedDays }}天的攻击趋势</span>
              </div>
              <div class="header-right">
                <el-radio-group v-model="chartType" @change="updateTrendChart">
                  <el-radio-button label="daily">按天</el-radio-button>
                  <el-radio-button label="hourly">按小时</el-radio-button>
                </el-radio-group>
                <el-select v-model="selectedDays" @change="loadStats" size="small" style="margin-left: 10px;">
                  <el-option label="最近7天" :value="7" />
                  <el-option label="最近15天" :value="15" />
                  <el-option label="最近30天" :value="30" />
                </el-select>
              </div>
            </div>
          </template>
          <div ref="trendChart" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 攻击规则分布和Top攻击规则 -->
    <el-row :gutter="20" class="rule-charts">
      <el-col :xs="24" :lg="12">
        <el-card class="top-list-card">
          <template #header>
            <div class="card-header">
              <h3>攻击规则分布</h3>
              <span class="header-desc">各规则触发占比</span>
            </div>
          </template>
          <div ref="typeChart" class="chart-container rule-chart"></div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :lg="12">
        <el-card class="top-list-card">
          <template #header>
            <div class="card-header">
              <h3>Top 攻击规则</h3>
              <el-button text @click="showAllRules">查看全部</el-button>
            </div>
          </template>
          <div class="top-list">
            <div
              v-for="(item, index) in (stats.top_attack_rules || [])"
              :key="item.rule_name"
              class="top-item"
              @click="showRuleDetail(item.rule_name)"
            >
              <div class="rank" :class="getRankClass(index)">{{ index + 1 }}</div>
              <div class="item-content">
                <div class="item-value" :title="item.rule_name">{{ item.rule_name }}</div>
                <div class="item-progress">
                  <el-progress 
                    :percentage="getPercentage(item.count, (stats.top_attack_rules || [])[0]?.count || 1)" 
                    :show-text="false" 
                    :stroke-width="4"
                  />
                </div>
              </div>
              <div class="item-count">{{ formatNumber(item.count) }}</div>
            </div>
            <div v-if="(stats.top_attack_rules || []).length === 0" class="empty-state">
              <el-empty description="暂无规则触发数据" :image-size="80" />
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Top攻击统计 - IP、URI、User-Agent -->
    <el-row :gutter="20" class="attack-stats">
      <el-col :xs="24" :md="8">
        <el-card class="top-list-card">
          <template #header>
            <div class="card-header">
              <h3>Top 攻击 IP</h3>
              <el-button text @click="showAllIPs">查看全部</el-button>
            </div>
          </template>
          <div class="top-list">
            <div
              v-for="(item, index) in (stats.top_attack_ips || [])"
              :key="item.client_ip"
              class="top-item"
              @click="showIPDetail(item.client_ip)"
            >
              <div class="rank" :class="getRankClass(index)">{{ index + 1 }}</div>
              <div class="item-content">
                <div class="item-value">{{ item.client_ip }}</div>
                <div class="item-progress">
                  <el-progress 
                    :percentage="getPercentage(item.count, (stats.top_attack_ips || [])[0]?.count || 1)" 
                    :show-text="false" 
                    :stroke-width="4"
                  />
                </div>
              </div>
              <div class="item-count">{{ formatNumber(item.count) }}</div>
            </div>
            <div v-if="(stats.top_attack_ips || []).length === 0" class="empty-state">
              <el-empty description="暂无攻击IP数据" :image-size="80" />
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :md="8">
        <el-card class="top-list-card">
          <template #header>
            <div class="card-header">
              <h3>Top 攻击 URI</h3>
              <el-button text @click="showAllURIs">查看全部</el-button>
            </div>
          </template>
          <div class="top-list">
            <div
              v-for="(item, index) in (stats.top_attack_uris || [])"
              :key="item.request_uri"
              class="top-item"
              @click="showURIDetail(item.request_uri)"
            >
              <div class="rank" :class="getRankClass(index)">{{ index + 1 }}</div>
              <div class="item-content">
                <div class="item-value" :title="item.request_uri">{{ item.request_uri }}</div>
                <div class="item-progress">
                  <el-progress 
                    :percentage="getPercentage(item.count, (stats.top_attack_uris || [])[0]?.count || 1)" 
                    :show-text="false" 
                    :stroke-width="4"
                  />
                </div>
              </div>
              <div class="item-count">{{ formatNumber(item.count) }}</div>
            </div>
            <div v-if="(stats.top_attack_uris || []).length === 0" class="empty-state">
              <el-empty description="暂无攻击URI数据" :image-size="80" />
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :md="8">
        <el-card class="top-list-card">
          <template #header>
            <div class="card-header">
              <h3>Top 攻击 User-Agent</h3>
              <el-button text @click="showAllUserAgents">查看全部</el-button>
            </div>
          </template>
          <div class="top-list">
            <div
              v-for="(item, index) in (stats.top_attack_user_agents || [])"
              :key="item.user_agent"
              class="top-item"
              @click="showUserAgentDetail(item.user_agent)"
            >
              <div class="rank" :class="getRankClass(index)">{{ index + 1 }}</div>
              <div class="item-content">
                <div class="item-value" :title="item.user_agent">{{ truncateUserAgent(item.user_agent) }}</div>
                <div class="item-progress">
                  <el-progress 
                    :percentage="getPercentage(item.count, (stats.top_attack_user_agents || [])[0]?.count || 1)" 
                    :show-text="false" 
                    :stroke-width="4"
                  />
                </div>
              </div>
              <div class="item-count">{{ formatNumber(item.count) }}</div>
            </div>
            <div v-if="(stats.top_attack_user_agents || []).length === 0" class="empty-state">
              <el-empty description="暂无User-Agent数据" :image-size="80" />
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" :title="detailTitle" width="70%" :before-close="closeDetailModal">
      <div class="detail-content">
        <div v-if="detailType === 'ips'" class="detail-list">
          <div class="detail-header">
            <h4>Top 攻击 IP 详情</h4>
            <span class="detail-desc">显示攻击次数最多的IP地址</span>
          </div>
          <el-table :data="stats.top_attack_ips || []" style="width: 100%">
            <el-table-column prop="client_ip" label="IP地址" width="200" />
            <el-table-column prop="count" label="攻击次数" width="120" />
            <el-table-column label="占比" width="120">
              <template #default="scope">
                {{ getPercentage(scope.row.count, stats.total_requests) }}%
              </template>
            </el-table-column>

            <el-table-column label="操作">
              <template #default="scope">
                <el-button size="small" @click="addToBlacklist(scope.row.client_ip)">
                  加入黑名单
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
        
        <div v-else-if="detailType === 'uris'" class="detail-list">
          <div class="detail-header">
            <h4>Top 攻击 URI 详情</h4>
            <span class="detail-desc">显示被攻击最多的URI路径</span>
          </div>
          <el-table :data="stats.top_attack_uris || []" style="width: 100%">
            <el-table-column prop="request_uri" label="URI路径" min-width="300" show-overflow-tooltip />
            <el-table-column prop="count" label="攻击次数" width="120" />
            <el-table-column label="占比" width="120">
              <template #default="scope">
                {{ getPercentage(scope.row.count, stats.total_requests) }}%
              </template>
            </el-table-column>
            <el-table-column label="操作">
              <template #default="scope">
                <el-button size="small" @click="addURIToBlacklist(scope.row.request_uri)">
                  加入黑名单
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
        
        <div v-else-if="detailType === 'rules'" class="detail-list">
          <div class="detail-header">
            <h4>Top 攻击规则详情</h4>
            <span class="detail-desc">显示触发次数最多的WAF规则</span>
          </div>
          <el-table :data="stats.top_attack_rules || []" style="width: 100%">
            <el-table-column prop="rule_name" label="规则名称" min-width="200" />
            <el-table-column prop="count" label="触发次数" width="120" />
            <el-table-column label="占比" width="120">
              <template #default="scope">
                {{ getPercentage(scope.row.count, stats.total_requests) }}%
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div v-else-if="detailType === 'user_agents'" class="detail-list">
          <div class="detail-header">
            <h4>Top 攻击 User-Agent 详情</h4>
            <span class="detail-desc">显示攻击次数最多的User-Agent</span>
          </div>
          <el-table :data="stats.top_attack_user_agents || []" style="width: 100%">
            <el-table-column prop="user_agent" label="User-Agent" min-width="400" show-overflow-tooltip />
            <el-table-column prop="count" label="攻击次数" width="120" />
            <el-table-column label="占比" width="120">
              <template #default="scope">
                {{ getPercentage(scope.row.count, stats.total_requests) }}%
              </template>
            </el-table-column>
            <el-table-column label="操作">
              <template #default="scope">
                <el-button size="small" @click="addUserAgentToBlacklist(scope.row.user_agent)">
                  加入黑名单
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
        
        <div v-else class="detail-summary">
          <el-descriptions title="统计概览" :column="2" border>
            <el-descriptions-item label="总请求数">{{ formatNumber(stats.total_requests) }}</el-descriptions-item>
            <el-descriptions-item label="拦截请求">{{ formatNumber(stats.blocked_requests) }}</el-descriptions-item>
            <el-descriptions-item label="允许请求">{{ formatNumber(stats.allowed_requests) }}</el-descriptions-item>
            <el-descriptions-item label="拦截率">{{ blockRate }}%</el-descriptions-item>
            <el-descriptions-item label="活跃规则">{{ stats.active_rules }}</el-descriptions-item>
            <el-descriptions-item label="活跃策略">{{ stats.active_policies }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  DataAnalysis, 
  CircleClose, 
  CircleCheck, 
  Lock, 
  ArrowUp, 
  ArrowDown,
  Refresh,
  Download,
  VideoPlay,
  VideoPause
} from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import type { EChartsType } from 'echarts'
import { 
  getDashboardOverview,
  getAttackTrend,
  getTopRules,
  getTopIPs,
  getTopURIs,
  getTopUserAgents,
  type DashboardOverview,
  type AttackTrend,
  type TopRule,
  type TopIP,
  type TopURI
} from '@/api/dashboard'
import { createBlackList } from '@/api/blacklist'

interface DashboardStats {
  total_requests: number
  blocked_requests: number
  allowed_requests: number
  active_rules: number
  active_policies: number
  top_attack_ips: TopIP[]
  top_attack_uris: TopURI[]
  top_attack_rules: TopRule[]
  top_attack_user_agents: Array<{ user_agent: string; count: number }>
  hourly_stats: Array<{ hour: string; count: number }>
  daily_stats: AttackTrend[]
}

// 数据加载和初始化
const loading = ref(false)
const stats = ref<DashboardStats>({
  total_requests: 0,
  blocked_requests: 0,
  allowed_requests: 0,
  active_rules: 0,
  active_policies: 0,
  top_attack_ips: [],
  top_attack_uris: [],
  top_attack_rules: [],
  top_attack_user_agents: [],
  hourly_stats: [],
  daily_stats: []
})

const autoRefresh = ref(false)
const refreshInterval = ref<number>()
const selectedDays = ref(7)
const chartType = ref('daily')

// 图表实例引用
const trendChart = ref<HTMLElement | null>(null)
const typeChart = ref<HTMLElement | null>(null)
let trendChartInstance: echarts.ECharts | null = null
let typeChartInstance: echarts.ECharts | null = null

// 详情弹窗
const detailVisible = ref(false)
const detailTitle = ref('')
const detailType = ref('')

// 计算属性
const blockRate = computed(() => {
  if (!stats.value.total_requests) return 0
  return Math.round((stats.value.blocked_requests / stats.value.total_requests) * 100)
})

const realtimeStatus = computed(() => {
  const rate = blockRate.value
  if (rate > 50) return { type: 'danger', text: '高风险' }
  if (rate > 20) return { type: 'warning', text: '中风险' }
  return { type: 'success', text: '安全' }
})

const lastUpdateTime = ref('')

// 方法
const loadStats = async () => {
  loading.value = true
  try {
    // 并行加载所有数据
    const [overview, trend, rules, ips, uris, userAgents] = await Promise.all([
      getDashboardOverview(),
      getAttackTrend(selectedDays.value),
      getTopRules(),
      getTopIPs(),
      getTopURIs(),
      getTopUserAgents()
    ])

    // 更新统计数据
    stats.value = {
      ...stats.value,
      total_requests: overview.data.total_attack_logs || 0,
      blocked_requests: overview.data.blocked_requests || 0,
      allowed_requests: overview.data.passed_requests || 0,
      active_rules: overview.data.total_rules || 0,
      active_policies: overview.data.total_policies || 0,
      top_attack_ips: ips.data || [],
      top_attack_uris: uris.data || [],
      top_attack_rules: rules.data || [],
      top_attack_user_agents: userAgents.data || [],
      daily_stats: trend.data || []
    }

    // 更新图表
    await updateCharts()
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

const calculateTrends = (newStats: DashboardStats) => {
  // 趋势计算已移除，改为显示静态信息
}

const renderCharts = () => {
  if (!stats.value) return
  renderTrendChart()
  renderTypeChart()
}

const renderTrendChart = () => {
  if (!trendChart.value || !stats.value?.daily_stats) return

  if (!trendChartInstance && trendChart.value) {
    trendChartInstance = echarts.init(trendChart.value)
  }

  const chartData = chartType.value === 'hourly' 
    ? stats.value.hourly_stats
    : stats.value.daily_stats

  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: chartData.map(item => {
        if ('hour' in item) {
          // 按小时显示时，只显示小时部分
          return item.hour.split(' ')[1].substring(0, 5)
        }
        // 按天显示时，只显示日期部分
        return item.time.split(' ')[0]
      }),
      axisLine: { lineStyle: { color: '#e0e6ed' } }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#e0e6ed' } }
    },
    series: [{
      name: '攻击次数',
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      data: chartData.map(item => item.count),
      areaStyle: {
        opacity: 0.1
      },
      lineStyle: {
        width: 2
      }
    }]
  }

  trendChartInstance?.setOption(option)
}

const renderTypeChart = () => {
  if (!typeChart.value || !stats.value?.top_attack_rules) return

  if (!typeChartInstance && typeChart.value) {
    typeChartInstance = echarts.init(typeChart.value)
  }

  const typeData = stats.value.top_attack_rules.map(rule => ({
    name: rule.rule_name,
    value: rule.count
  }))

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      textStyle: { fontSize: 12 }
    },
    series: [{
      name: '攻击类型',
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['60%', '50%'],
      data: typeData
    }]
  }

  typeChartInstance?.setOption(option)
}

const renderRealtimeChart = () => {}

const updateTrendChart = async () => {
  try {
    loading.value = true
    // 根据选择的类型加载不同的数据
    const response = await getAttackTrend(selectedDays.value, chartType.value as 'hourly' | 'daily')
    
    // 更新统计数据
    if (chartType.value === 'hourly') {
      stats.value.hourly_stats = response.data.map(item => ({
        hour: item.time,
        count: item.count
      }))
    } else {
      stats.value.daily_stats = response.data
    }
    
    // 重新渲染图表
    await nextTick()
    renderTrendChart()
  } catch (error) {
    console.error('更新趋势图表失败:', error)
    ElMessage.error('更新趋势图表失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

// 监听图表类型变化
watch(chartType, () => {
  updateTrendChart()
})

// 监听天数变化
watch(selectedDays, () => {
  updateTrendChart()
})

const refreshData = () => {
  loadStats()
}

const updateCharts = async () => {
  await nextTick()
  renderCharts()
}

// 自动刷新数据
const startAutoRefresh = () => {
  if (refreshInterval.value) return
  refreshInterval.value = setInterval(async () => {
    await loadStats()
  }, 30000) // 每30秒刷新一次
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
    refreshInterval.value = undefined
  }
}

// 切换自动刷新状态
const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

const exportData = () => {
  ElMessage.info('导出功能开发中')
}

// 工具方法
const formatNumber = (num: number) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

// getTrendClass方法已移除，不再需要趋势样式

const getRankClass = (index: number) => {
  if (index === 0) return 'rank-first'
  if (index === 1) return 'rank-second'
  if (index === 2) return 'rank-third'
  return 'rank-normal'
}

const getPercentage = (value: number, max: number) => {
  return Math.round((value / max) * 100)
}

// 详情相关方法
const showDetailModal = (type: string) => {
  detailVisible.value = true
  detailTitle.value = getDetailTitle(type)
  detailType.value = type
}

const closeDetailModal = () => {
  detailVisible.value = false
  detailType.value = ''
}

const getDetailTitle = (type: string) => {
  const titles: Record<string, string> = {
    total: '总请求详情',
    blocked: '拦截请求详情',
    allowed: '允许请求详情',
    ips: 'Top 攻击 IP',
    uris: 'Top 攻击 URI',
    rules: 'Top 攻击规则',
    user_agents: 'Top 攻击 User-Agent'
  }
  return titles[type] || '详情'
}

const showAllIPs = () => {
  showDetailModal('ips')
}

const showAllURIs = () => {
  showDetailModal('uris')
}

const showAllRules = () => {
  showDetailModal('rules')
}

const showAllUserAgents = () => {
  showDetailModal('user_agents')
}

// 新增的辅助方法
const getRiskLevel = (count: number) => {
  if (count >= 10) return { type: 'danger', text: '高风险' }
  if (count >= 5) return { type: 'warning', text: '中风险' }
  return { type: 'success', text: '低风险' }
}

// 已移除 getAttackTypeFromURI 函数，现在直接使用后端返回的 attack_type

const addToBlacklist = async (clientIp: string) => {
  try {
    await ElMessageBox.confirm(`确定要将IP ${clientIp} 加入黑名单吗？`, '确认操作', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 调用API将IP加入黑名单
    await createBlackList({
      type: 'ip',
      value: clientIp,
      comment: `从攻击IP统计自动添加 - ${new Date().toLocaleString()}`
    })
    
    ElMessage.success('已成功添加到黑名单')
    await loadStats() // 重新加载数据
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('添加黑名单失败')
    }
  }
}

const viewRuleDetail = async (ruleName: string) => {
  try {
    // 显示规则详情弹窗
    const ruleData = (stats.value.top_attack_rules || []).find(item => item.rule_name === ruleName)
    if (ruleData) {
      await ElMessageBox.alert(
        `规则名称: ${ruleName}\n触发次数: ${ruleData.count}\n占比: ${getPercentage(ruleData.count, stats.value.total_requests)}%\n状态: 活跃`,
        '规则详情',
        {
          confirmButtonText: '确定',
          type: 'info'
        }
      )
    } else {
      ElMessage.warning('未找到规则详情')
    }
  } catch (error) {
    console.error('查看规则详情失败:', error)
  }
}

const showIPDetail = (clientIp: string) => {
  const ipData = (stats.value.top_attack_ips || []).find(item => item.client_ip === clientIp)
  if (ipData) {
    ElMessageBox.alert(
      `IP地址: ${clientIp}\n攻击次数: ${ipData.count}\n占比: ${getPercentage(ipData.count, stats.value.total_requests)}%`,
      'IP详情',
      {
        confirmButtonText: '确定',
        type: 'info'
      }
    )
  }
}

const showURIDetail = (uri: string) => {
  const uriData = (stats.value.top_attack_uris || []).find(item => item.request_uri === uri)
  if (uriData) {
    ElMessageBox.alert(
      `URI路径: ${uri}\n攻击次数: ${uriData.count}\n占比: ${getPercentage(uriData.count, stats.value.total_requests)}%`,
      'URI详情',
      {
        confirmButtonText: '确定',
        type: 'info'
      }
    )
  }
}

const showRuleDetail = (ruleName: string) => {
  const ruleData = (stats.value.top_attack_rules || []).find(item => item.rule_name === ruleName)
  if (ruleData) {
    ElMessageBox.alert(
      `规则名称: ${ruleName}\n触发次数: ${ruleData.count}\n占比: ${getPercentage(ruleData.count, stats.value.total_requests)}%`,
      '规则详情',
      {
        confirmButtonText: '确定',
        type: 'info'
      }
    )
  }
}

const showUserAgentDetail = (userAgent: string) => {
  const userAgentData = (stats.value.top_attack_user_agents || []).find(item => item.user_agent === userAgent)
  if (userAgentData) {
    ElMessageBox.alert(
      `User-Agent: ${userAgent}\n攻击次数: ${userAgentData.count}\n占比: ${getPercentage(userAgentData.count, stats.value.total_requests)}%`,
      'User-Agent详情',
      {
        confirmButtonText: '确定',
        type: 'info'
      }
    )
  }
}

const truncateUserAgent = (userAgent: string) => {
  if (userAgent.length > 50) {
    return userAgent.substring(0, 50) + '...'
  }
  return userAgent
}

const addURIToBlacklist = async (uri: string) => {
  try {
    await ElMessageBox.confirm(`确定要将URI ${uri} 加入黑名单吗？`, '确认操作', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 调用API将URI加入黑名单
    await createBlackList({
      type: 'uri',
      value: uri,
      comment: `从攻击URI统计自动添加 - ${new Date().toLocaleString()}`,
      enabled: true
    })
    
    ElMessage.success(`URI ${uri} 已成功加入黑名单`)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('加入黑名单失败:', error)
      ElMessage.error('加入黑名单失败，请稍后重试')
    }
  }
}

const addUserAgentToBlacklist = async (userAgent: string) => {
  try {
    await ElMessageBox.confirm(`确定要将User-Agent加入黑名单吗？\n\n${userAgent}`, '确认操作', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 调用API将User-Agent加入黑名单
    await createBlackList({
      type: 'user_agent',
      value: userAgent,
      comment: `从攻击User-Agent统计自动添加 - ${new Date().toLocaleString()}`,
      enabled: true
    })
    
    ElMessage.success('User-Agent已成功加入黑名单')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('加入黑名单失败:', error)
      ElMessage.error('加入黑名单失败，请稍后重试')
    }
  }
}

// 生命周期钩子
onMounted(async () => {
  await loadStats()
  if (autoRefresh.value) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  stopAutoRefresh()
  if (trendChartInstance) {
    trendChartInstance.dispose()
    trendChartInstance = null
  }
  if (typeChartInstance) {
    typeChartInstance.dispose()
    typeChartInstance = null
  }
})
</script>

<style scoped>
.dashboard {
  height: 100%;
  width: 100%;
  padding: 24px;
  overflow-y: auto;
  overflow-x: hidden;
  background-color: #f5f7fa;
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left {
  h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: #1f2937;
  }

  .header-desc {
    margin-top: 4px;
    color: #6b7280;
    font-size: 14px;
  }
}

.header-right {
  display: flex;
  gap: 12px;
}

.stats-row {
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  height: 100%;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  flex-shrink: 0;
}

.stat-icon .el-icon {
  font-size: 24px;
  color: white;
}

.total-requests .stat-icon {
  background: linear-gradient(135deg, #60a5fa, #3b82f6);
}

.blocked-requests .stat-icon {
  background: linear-gradient(135deg, #f87171, #ef4444);
}

.allowed-requests .stat-icon {
  background: linear-gradient(135deg, #34d399, #10b981);
}

.active-rules .stat-icon {
  background: linear-gradient(135deg, #a78bfa, #8b5cf6);
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 8px;
}

.stat-trend {
  display: flex;
  align-items: center;
  font-size: 12px;
  font-weight: 600;
}

.stat-trend.trend-up {
  color: #ef4444;
}

.stat-trend.trend-down {
  color: #10b981;
}

.stat-rate, .stat-extra {
  font-size: 12px;
  color: #6b7280;
}

.trend-section {
  margin-bottom: 24px;
}

.rule-charts {
  margin-bottom: 24px;
}

.attack-stats {
  margin-bottom: 24px;
}

.chart-card, .top-list-card, .realtime-card {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
}

.card-header .header-desc {
  color: #6b7280;
  font-size: 12px;
  margin-left: 8px;
}

.chart-container {
  height: 300px;
  width: 100%;
}

.rule-chart {
  height: 400px;
}

.realtime-chart {
  height: 200px;
}

.top-list {
  max-height: 400px;
  overflow-y: auto;
  padding-right: 8px;
}

.top-item {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
  transition: background-color 0.2s;
}

.top-item:hover {
  background-color: #f9fafb;
}

.top-item:last-child {
  border-bottom: none;
}

.rank {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: white;
  margin-right: 12px;
  flex-shrink: 0;
}

.rank-first {
  background: #ffd700;
}

.rank-second {
  background: #c0c0c0;
}

.rank-third {
  background: #cd7f32;
}

.rank-normal {
  background: #6b7280;
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-value {
  font-size: 14px;
  color: #1f2937;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.item-progress {
  width: 100%;
}

.item-count {
  color: #6b7280;
  font-size: 12px;
  font-weight: 600;
  margin-left: 12px;
  margin-right: 8px;
  flex-shrink: 0;
}

.empty-state {
  padding: 40px 0;
  text-align: center;
}

.realtime-section {
  margin-bottom: 24px;
}

.realtime-status {
  display: flex;
  align-items: center;
  gap: 12px;
}

.last-update {
  font-size: 12px;
  color: #6b7280;
}

.detail-content {
  min-height: 200px;
}

.detail-header {
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.detail-header h4 {
  margin: 0 0 4px 0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

.detail-desc {
  color: #6b7280;
  font-size: 14px;
}

.detail-list {
  padding: 0;
}

.detail-summary {
  padding: 20px 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .dashboard {
    padding: 12px;
  }
  
  .dashboard-header {
    flex-direction: column;
    gap: 16px;
    text-align: center;
  }
  
  .stat-card {
    padding: 16px;
  }
  
  .stat-value {
    font-size: 24px;
  }
  
  .chart-container {
    height: 250px;
  }
}

.dashboard-container {
  padding: 20px;
}

.overview-cards {
  margin-bottom: 20px;
}

.overview-card {
  height: 100%;
  transition: all 0.3s;
}

.overview-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.card-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.header-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.chart-container {
  height: 300px;
  width: 100%;
}

.rule-chart {
  height: 300px;
}

.top-list-card {
  height: 400px;  /* 设置固定高度 */
  display: flex;
  flex-direction: column;
}

.top-list-card :deep(.el-card__body) {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.top-list {
  flex: 1;
  overflow-y: auto;
  padding-right: 10px;
}

.top-list::-webkit-scrollbar {
  width: 6px;
}

.top-list::-webkit-scrollbar-thumb {
  background-color: var(--el-border-color);
  border-radius: 3px;
}

.top-item {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
  transition: all 0.3s;
}

.top-item:hover {
  background-color: var(--el-fill-color-light);
}

.rank {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
  margin-right: 12px;
  flex-shrink: 0;
}

.rank-1 {
  background-color: #f56c6c;
  color: white;
}

.rank-2 {
  background-color: #e6a23c;
  color: white;
}

.rank-3 {
  background-color: #67c23a;
  color: white;
}

.rank-other {
  background-color: var(--el-fill-color);
  color: var(--el-text-color-regular);
}

.item-content {
  flex: 1;
  min-width: 0;
  margin-right: 12px;
}

.item-value {
  font-size: 14px;
  color: var(--el-text-color-primary);
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-progress {
  width: 100%;
}

.item-count {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-left: 12px;
  flex-shrink: 0;
}

.empty-state {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.rule-charts {
  margin-bottom: 20px;
}

.attack-stats {
  margin-bottom: 20px;
}

/* 添加响应式布局调整 */
@media screen and (max-width: 768px) {
  .top-list-card {
    height: 350px;  /* 在移动端稍微降低高度 */
  }
  
  .chart-container,
  .rule-chart {
    height: 250px;  /* 在移动端降低图表高度 */
  }
}
</style> 