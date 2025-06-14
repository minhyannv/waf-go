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

    <!-- 攻击趋势分析 - 单独一行 -->
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
              :key="item.ip"
              class="top-item"
              @click="showIPDetail(item.ip)"
            >
              <div class="rank" :class="getRankClass(index)">{{ index + 1 }}</div>
              <div class="item-content">
                <div class="item-value">{{ item.ip }}</div>
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
              :key="item.uri"
              class="top-item"
              @click="showURIDetail(item.uri)"
            >
              <div class="rank" :class="getRankClass(index)">{{ index + 1 }}</div>
              <div class="item-content">
                <div class="item-value" :title="item.uri">{{ item.uri }}</div>
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

    <!-- 实时监控 -->
    <el-row :gutter="20" class="realtime-section">
      <el-col :span="24">
        <el-card class="realtime-card">
          <template #header>
            <div class="card-header">
              <h3>实时攻击监控</h3>
              <div class="realtime-status">
                <el-tag :type="realtimeStatus.type" effect="dark">
                  {{ realtimeStatus.text }}
                </el-tag>
                <span class="last-update">最后更新: {{ lastUpdateTime }}</span>
              </div>
            </div>
          </template>
          <div ref="realtimeChart" class="chart-container realtime-chart"></div>
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
            <el-table-column prop="ip" label="IP地址" width="200" />
            <el-table-column prop="count" label="攻击次数" width="120" />
            <el-table-column label="占比" width="120">
              <template #default="scope">
                {{ getPercentage(scope.row.count, stats.total_requests) }}%
              </template>
            </el-table-column>

            <el-table-column label="操作">
              <template #default="scope">
                <el-button size="small" @click="addToBlacklist(scope.row.ip)">
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
            <el-table-column prop="uri" label="URI路径" min-width="300" show-overflow-tooltip />
            <el-table-column prop="count" label="攻击次数" width="120" />
            <el-table-column label="占比" width="120">
              <template #default="scope">
                {{ getPercentage(scope.row.count, stats.total_requests) }}%
              </template>
            </el-table-column>
            <el-table-column label="操作">
              <template #default="scope">
                <el-button size="small" @click="addURIToBlacklist(scope.row.uri)">
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
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
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
import { getDashboardStats, getRealtimeStats } from '@/api/dashboard'
import { createBlackList } from '@/api/blacklists'
import type { DashboardStats } from '@/api/dashboard'

// 响应式数据
const stats = ref<DashboardStats>({
  total_requests: 0,
  blocked_requests: 0,
  allowed_requests: 0,
  top_attack_ips: [],
  top_attack_uris: [],
  top_attack_rules: [],
  top_attack_user_agents: [],
  hourly_stats: [],
  daily_stats: [],
  active_rules: 0,
  active_policies: 0
})

const loading = ref(false)
const autoRefresh = ref(false)
const refreshInterval = ref<number>()
const selectedDays = ref(7)
const chartType = ref('daily')

// 图表引用
const trendChart = ref<HTMLElement>()
const typeChart = ref<HTMLElement>()
const realtimeChart = ref<HTMLElement>()

// ECharts实例
let trendChartInstance: echarts.ECharts | null = null
let typeChartInstance: echarts.ECharts | null = null
let realtimeChartInstance: echarts.ECharts | null = null

// 趋势数据 - 已移除，改为显示静态信息

// 详情弹窗
const detailVisible = ref(false)
const detailTitle = ref('')
const detailType = ref('')

// 计算属性
const blockRate = computed(() => {
  if (stats.value.total_requests === 0) return 0
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
    const response = await getDashboardStats(selectedDays.value)
    const newStats = response.data
    
    // 计算趋势
    calculateTrends(newStats)
    
    stats.value = newStats
    lastUpdateTime.value = new Date().toLocaleTimeString()
    
    // 更新图表
    await nextTick()
    renderCharts()
  } catch (error) {
    ElMessage.error('加载统计数据失败')
  } finally {
    loading.value = false
  }
}

const calculateTrends = (newStats: DashboardStats) => {
  // 趋势计算已移除，改为显示静态信息
}

const renderCharts = () => {
  renderTrendChart()
  renderTypeChart()
  renderRealtimeChart()
}

const renderTrendChart = () => {
  if (!trendChart.value) return
  
  if (!trendChartInstance) {
    trendChartInstance = echarts.init(trendChart.value)
  }
  
  const data = chartType.value === 'daily' ? (stats.value.daily_stats || []) : (stats.value.hourly_stats || [])
  const xAxisData = data.map(item => 'date' in item ? item.date : item.hour)
  const seriesData = data.map(item => item.count)
  
  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: xAxisData,
      axisLine: {
        lineStyle: {
          color: '#e0e6ed'
        }
      }
    },
    yAxis: {
      type: 'value',
      axisLine: {
        lineStyle: {
          color: '#e0e6ed'
        }
      }
    },
    series: [
      {
        name: '攻击次数',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: {
          width: 3,
          color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
            { offset: 0, color: '#667eea' },
            { offset: 1, color: '#764ba2' }
          ])
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(102, 126, 234, 0.3)' },
            { offset: 1, color: 'rgba(118, 75, 162, 0.1)' }
          ])
        },
        data: seriesData
      }
    ]
  }
  
  trendChartInstance.setOption(option)
}

const renderTypeChart = () => {
  if (!typeChart.value) return
  
  if (!typeChartInstance) {
    typeChartInstance = echarts.init(typeChart.value)
  }
  
  // 直接使用攻击规则数据作为攻击类型分布
  const typeData: Array<{ name: string; value: number }> = []
  
  if (stats.value.top_attack_rules && stats.value.top_attack_rules.length > 0) {
    // 直接使用规则名称和对应的攻击次数
    stats.value.top_attack_rules.forEach(rule => {
      typeData.push({ 
        name: rule.rule_name, 
        value: rule.count 
      })
    })
  } else {
    // 如果没有数据，显示空状态
    typeData.push({ name: '暂无数据', value: 1 })
  }
  
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      textStyle: {
        fontSize: 12
      }
    },
    series: [
      {
        name: '攻击类型',
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['60%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: '18',
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: typeData
      }
    ]
  }
  
  typeChartInstance.setOption(option)
}

const renderRealtimeChart = async () => {
  if (!realtimeChart.value) return
  
  if (!realtimeChartInstance) {
    realtimeChartInstance = echarts.init(realtimeChart.value)
  }
  
  let realtimeData = []
  
  try {
    // 获取真实的实时攻击数据
    const response = await getRealtimeStats()
    realtimeData = response.data.map((item: { minute: string; count: number }) => ({
      time: new Date(item.minute).toLocaleTimeString('zh-CN', { 
        hour: '2-digit', 
        minute: '2-digit' 
      }),
      value: item.count
    }))
  } catch (error) {
    console.error('获取实时数据失败:', error)
    // 如果获取失败，使用空数据
    const now = new Date()
    for (let i = 59; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60000)
      realtimeData.push({
        time: time.toLocaleTimeString('zh-CN', { 
          hour: '2-digit', 
          minute: '2-digit' 
        }),
        value: 0
      })
    }
  }
  
  const option = {
    tooltip: {
      trigger: 'axis',
      formatter: function(params: any) {
        return `${params[0].name}<br/>攻击次数: ${params[0].value}`
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
              data: realtimeData.map((item: { time: string; value: number }) => item.time),
      axisLabel: {
        interval: 9 // 每10个点显示一个标签
      }
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '实时攻击',
        type: 'bar',
        barWidth: '60%',
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#f093fb' },
            { offset: 1, color: '#f5576c' }
          ])
        },
        data: realtimeData.map((item: { time: string; value: number }) => item.value)
      }
    ]
  }
  
  realtimeChartInstance.setOption(option)
}

const updateTrendChart = () => {
  renderTrendChart()
}

const refreshData = () => {
  loadStats()
}

const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
  
  if (autoRefresh.value) {
    refreshInterval.value = setInterval(() => {
      loadStats()
      renderRealtimeChart() // 同时更新实时图表
    }, 30000) // 30秒刷新一次
    ElMessage.success('已开启自动刷新')
  } else {
    if (refreshInterval.value) {
      clearInterval(refreshInterval.value)
    }
    ElMessage.info('已停止自动刷新')
  }
}

const exportData = async () => {
  try {
    await ElMessageBox.confirm('确定要导出当前统计报告吗？', '导出确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'info'
    })
    
    // 这里实现导出逻辑
    ElMessage.success('报告导出成功')
  } catch {
    // 用户取消
  }
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

const addToBlacklist = async (ip: string) => {
  try {
    await ElMessageBox.confirm(`确定要将IP ${ip} 加入黑名单吗？`, '确认操作', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 调用API将IP加入黑名单
    await createBlackList({
      type: 'ip',
      value: ip,
      comment: `从攻击IP统计自动添加 - ${new Date().toLocaleString()}`,
      enabled: true
    })
    
    ElMessage.success(`IP ${ip} 已成功加入黑名单`)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('加入黑名单失败:', error)
      ElMessage.error('加入黑名单失败，请稍后重试')
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

const showIPDetail = (ip: string) => {
  const ipData = (stats.value.top_attack_ips || []).find(item => item.ip === ip)
  if (ipData) {
    ElMessageBox.alert(
      `IP地址: ${ip}\n攻击次数: ${ipData.count}\n占比: ${getPercentage(ipData.count, stats.value.total_requests)}%`,
      'IP详情',
      {
        confirmButtonText: '确定',
        type: 'info'
      }
    )
  }
}

const showURIDetail = (uri: string) => {
  const uriData = (stats.value.top_attack_uris || []).find(item => item.uri === uri)
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

// 生命周期
onMounted(() => {
  loadStats()
  
  // 监听窗口大小变化
  window.addEventListener('resize', () => {
    trendChartInstance?.resize()
    typeChartInstance?.resize()
    realtimeChartInstance?.resize()
  })
})

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }
  
  // 销毁图表实例
  trendChartInstance?.dispose()
  typeChartInstance?.dispose()
  realtimeChartInstance?.dispose()
  
  window.removeEventListener('resize', () => {})
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-left h2 {
  margin: 0 0 8px 0;
  color: #1f2937;
  font-size: 24px;
  font-weight: 600;
}

.header-desc {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.stats-row {
  margin-bottom: 24px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
}

.stat-card.total-requests::before {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-card.blocked-requests::before {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-card.allowed-requests::before {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-card.active-rules::before {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-card .stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: white;
  margin-bottom: 16px;
}

.total-requests .stat-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.blocked-requests .stat-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.allowed-requests .stat-icon {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.active-rules .stat-icon {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: #1f2937;
  line-height: 1;
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
  padding-right: 8px; /* 给滚动条留出空间 */
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
  margin-right: 8px; /* 与滚动条保持距离 */
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
</style> 