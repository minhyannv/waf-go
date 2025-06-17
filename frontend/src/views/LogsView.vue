<template>
  <div class="logs-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>攻击日志</span>
          <div class="header-actions">
            <el-button type="primary" @click="exportLogs" :disabled="selectedLogs.length === 0">
              <el-icon><Download /></el-icon>
              导出选中
            </el-button>
            <el-button type="warning" @click="showCleanDialog">
              <el-icon><Delete /></el-icon>
              清理日志
            </el-button>
          </div>
        </div>
      </template>
      
      <!-- 搜索区域 -->
      <div class="search-area">
        <el-form :model="searchForm" inline>
          <el-form-item label="客户端IP">
            <el-input
              v-model="searchForm.client_ip"
              placeholder="请输入IP地址"
              clearable
              style="width: 160px"
            />
          </el-form-item>
          <el-form-item label="请求URI">
            <el-input
              v-model="searchForm.request_uri"
              placeholder="请输入URI"
              clearable
              style="width: 200px"
            />
          </el-form-item>
          <el-form-item label="规则名称">
            <el-input
              v-model="searchForm.rule_name"
              placeholder="请输入规则名称"
              clearable
              style="width: 160px"
            />
          </el-form-item>
          <el-form-item label="动作">
            <el-select v-model="searchForm.action" placeholder="请选择动作" clearable style="width: 120px">
              <el-option label="拦截" value="block" />
              <el-option label="允许" value="allow" />
              <el-option label="记录" value="log" />
            </el-select>
          </el-form-item>
          <el-form-item label="时间范围">
            <el-date-picker
              v-model="timeRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
              @change="handleTimeRangeChange"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadLogs">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
            <el-button @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>
      
      <!-- 批量操作 -->
      <div v-if="selectedLogs.length > 0" class="batch-actions">
        <span>已选择 {{ selectedLogs.length }} 条记录</span>
        <el-button type="danger" size="small" @click="batchDelete">批量删除</el-button>
      </div>
      
      <!-- 日志表格 -->
      <el-table
        v-loading="loading"
        :data="logs"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="client_ip" label="客户端IP" width="140">
          <template #default="{ row }">
            <el-tag type="info">{{ row.client_ip }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="request_method" label="方法" width="80">
          <template #default="{ row }">
            <el-tag :type="getMethodColor(row.request_method)">
              {{ row.request_method }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="request_uri" label="请求URI" min-width="250" show-overflow-tooltip />
        <el-table-column prop="rule_name" label="触发规则" width="150" show-overflow-tooltip />
        <el-table-column prop="action" label="动作" width="80">
          <template #default="{ row }">
            <el-tag :type="getActionColor(row.action)">
              {{ getActionText(row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="response_code" label="响应码" width="80">
          <template #default="{ row }">
            <el-tag :type="getResponseCodeColor(row.response_code)">
              {{ row.response_code }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="showDetail(row)">详情</el-button>
            <el-button type="danger" link @click="deleteLog(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadLogs"
          @current-change="loadLogs"
        />
      </div>
    </el-card>

    <!-- 日志详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="攻击日志详情" width="800px">
      <el-descriptions :column="2" border v-if="currentLog">
        <el-descriptions-item label="日志ID">{{ currentLog.id }}</el-descriptions-item>
        <el-descriptions-item label="请求ID">{{ currentLog.request_id }}</el-descriptions-item>
        <el-descriptions-item label="客户端IP">
          <el-tag type="info">{{ currentLog.client_ip }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="请求方法">
          <el-tag :type="getMethodColor(currentLog.request_method)">
            {{ currentLog.request_method }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="请求URI" :span="2">{{ currentLog?.request_uri }}</el-descriptions-item>
        <el-descriptions-item label="User-Agent" :span="2">{{ currentLog?.user_agent || '无' }}</el-descriptions-item>
        <el-descriptions-item label="触发规则">{{ currentLog.rule_name }}</el-descriptions-item>
        <el-descriptions-item label="匹配字段">{{ currentLog.match_field }}</el-descriptions-item>
        <el-descriptions-item label="匹配值" :span="2">{{ currentLog.match_value }}</el-descriptions-item>
        <el-descriptions-item label="执行动作">
          <el-tag :type="getActionColor(currentLog.action)">
            {{ getActionText(currentLog.action) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="响应码">
          <el-tag :type="getResponseCodeColor(currentLog.response_code)">
            {{ currentLog.response_code }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间" :span="2">{{ formatDate(currentLog.created_at) }}</el-descriptions-item>
      </el-descriptions>
      
      <!-- 请求详情 -->
      <el-divider>请求详情</el-divider>
      <el-tabs>
        <el-tab-pane label="请求头" name="headers">
          <el-input
            v-model="currentLog!.request_headers"
            type="textarea"
            :rows="8"
            readonly
            placeholder="无请求头信息"
          />
        </el-tab-pane>
        <el-tab-pane label="请求体" name="body">
          <el-input
            v-model="currentLog!.request_body"
            type="textarea"
            :rows="8"
            readonly
            placeholder="无请求体信息"
          />
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <!-- 清理日志对话框 -->
    <el-dialog v-model="cleanDialogVisible" title="清理旧日志" width="400px">
      <el-form :model="cleanForm" label-width="100px">
        <el-form-item label="保留天数">
          <el-input-number
            v-model="cleanForm.days"
            :min="1"
            :max="365"
            placeholder="保留最近几天的日志"
            style="width: 100%"
          />
          <div class="form-tip">
            <el-text size="small" type="info">
              将删除 {{ cleanForm.days }} 天前的所有攻击日志
            </el-text>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cleanDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmClean" :loading="cleaning">确定清理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Download, Delete } from '@element-plus/icons-vue'
import { 
  getAttackLogList, 
  getAttackLogDetail, 
  deleteAttackLog, 
  batchDeleteAttackLogs,
  cleanOldLogs,
  exportAttackLogs,
  type AttackLog 
} from '@/api/logs'

// 响应式数据
const loading = ref(false)
const logs = ref<AttackLog[]>([])
const selectedLogs = ref<AttackLog[]>([])
const timeRange = ref<[string, string] | null>(null)
const searchForm = reactive({
  client_ip: '',
  request_uri: '',
  rule_name: '',
  action: '',
  start_time: '',
  end_time: '',
  domain: ''
})
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 对话框相关
const detailDialogVisible = ref(false)
const cleanDialogVisible = ref(false)
const currentLog = ref<AttackLog | null>(null)
const cleaning = ref(false)
const cleanForm = reactive({
  days: 30
})

// 工具函数
const getMethodColor = (method: string) => {
  const map: Record<string, string> = {
    GET: 'success',
    POST: 'primary',
    PUT: 'warning',
    DELETE: 'danger',
    PATCH: 'info'
  }
  return map[method] || 'info'
}

const getActionColor = (action: string) => {
  const map: Record<string, string> = {
    block: 'danger',
    allow: 'success',
    log: 'warning'
  }
  return map[action] || 'info'
}

const getActionText = (action: string) => {
  const map: Record<string, string> = {
    block: '拦截',
    allow: '允许',
    log: '记录'
  }
  return map[action] || action
}

const getResponseCodeColor = (code: number) => {
  if (code >= 200 && code < 300) return 'success'
  if (code >= 300 && code < 400) return 'info'
  if (code >= 400 && code < 500) return 'warning'
  if (code >= 500) return 'danger'
  return 'info'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

// 数据加载
const loadLogs = async () => {
  loading.value = true
  try {
    const response = await getAttackLogList({
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm
    })
    logs.value = response.data.list || []
    pagination.total = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载攻击日志失败')
    logs.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

const handleTimeRangeChange = (value: [string, string] | null) => {
  if (value) {
    searchForm.start_time = value[0]
    searchForm.end_time = value[1]
  } else {
    searchForm.start_time = ''
    searchForm.end_time = ''
  }
}

const resetSearch = () => {
  Object.assign(searchForm, {
    client_ip: '',
    request_uri: '',
    rule_name: '',
    action: '',
    start_time: '',
    end_time: ''
  })
  timeRange.value = null
  pagination.page = 1
  loadLogs()
}

const handleSelectionChange = (selection: AttackLog[]) => {
  selectedLogs.value = selection
}

// 详情操作
const showDetail = async (row: AttackLog) => {
  try {
    const response = await getAttackLogDetail(row.id)
    currentLog.value = response.data
    detailDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取日志详情失败')
  }
}

// 删除操作
const deleteLog = async (row: AttackLog) => {
  try {
    await ElMessageBox.confirm(`确定要删除ID为 ${row.id} 的攻击日志吗？`, '提示', {
      type: 'warning'
    })
    
    await deleteAttackLog(row.id)
    ElMessage.success('删除成功')
    loadLogs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedLogs.value.length} 条攻击日志吗？`, '提示', {
      type: 'warning'
    })
    
    const ids = selectedLogs.value.map(log => log.id)
    await batchDeleteAttackLogs(ids)
    ElMessage.success('批量删除成功')
    loadLogs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 清理日志
const showCleanDialog = () => {
  cleanDialogVisible.value = true
}

const confirmClean = async () => {
  try {
    cleaning.value = true
    const response = await cleanOldLogs(cleanForm.days)
    ElMessage.success(`清理完成，删除了 ${response.data} 条旧日志`)
    cleanDialogVisible.value = false
    loadLogs()
  } catch (error) {
    ElMessage.error('清理日志失败')
  } finally {
    cleaning.value = false
  }
}

// 导出日志
const exportLogs = async () => {
  try {
    const ids = selectedLogs.value.map(log => log.id)
    const response = await exportAttackLogs(ids)
    
    // 创建下载链接
    const blob = new Blob([response.data])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `attack_logs_${new Date().getTime()}.json`
    link.click()
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.search-area {
  margin-bottom: 20px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 6px;
}

.batch-actions {
  margin-bottom: 15px;
  padding: 10px 15px;
  background: #e3f2fd;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.form-tip {
  margin-top: 5px;
}
</style> 