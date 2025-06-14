<template>
  <div class="rules-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>规则管理</span>
          <div class="header-actions">
            <el-button @click="showImportDialog">
              <el-icon><Upload /></el-icon>
              导入规则
            </el-button>
            <el-button @click="exportRules" :disabled="!selectedRules.length">
              <el-icon><Download /></el-icon>
              导出规则
            </el-button>
            <el-button type="primary" @click="showCreateDialog">
              <el-icon><Plus /></el-icon>
              新增规则
            </el-button>
          </div>
        </div>
      </template>
      
      <!-- 搜索区域 -->
      <div class="search-area">
        <el-form :model="searchForm" inline>
          <el-form-item label="规则名称">
            <el-input
              v-model="searchForm.name"
              placeholder="请输入规则名称"
              clearable
              style="width: 200px"
            />
          </el-form-item>
          <el-form-item label="匹配类型">
            <el-select v-model="searchForm.match_type" placeholder="请选择匹配类型" clearable style="width: 150px">
              <el-option label="URI" value="uri" />
              <el-option label="IP" value="ip" />
              <el-option label="请求头" value="header" />
              <el-option label="请求体" value="body" />
              <el-option label="User-Agent" value="user_agent" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="searchForm.enabled" placeholder="请选择状态" clearable style="width: 120px">
              <el-option label="启用" :value="true" />
              <el-option label="禁用" :value="false" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadRules">搜索</el-button>
            <el-button @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>
      
      <!-- 批量操作 -->
      <div class="batch-actions" v-if="selectedRules.length > 0">
        <span>已选择 {{ selectedRules.length }} 项</span>
        <el-button type="success" size="small" @click="batchEnable">批量启用</el-button>
        <el-button type="warning" size="small" @click="batchDisable">批量禁用</el-button>
        <el-button type="danger" size="small" @click="batchDelete">批量删除</el-button>
      </div>
      
      <!-- 规则表格 -->
      <el-table
        v-loading="loading"
        :data="rules"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="规则名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column prop="match_type" label="匹配类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getMatchTypeColor(row.match_type)">
              {{ getMatchTypeText(row.match_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="match_mode" label="匹配方式" width="100">
          <template #default="{ row }">
            <el-tag :type="getMatchModeColor(row.match_mode)">
              {{ getMatchModeText(row.match_mode) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="pattern" label="匹配模式" min-width="200" show-overflow-tooltip />
        <el-table-column prop="action" label="动作" width="80">
          <template #default="{ row }">
            <el-tag :type="getActionColor(row.action)">
              {{ getActionText(row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" sortable />
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="toggleRule(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="showEditDialog(row)">编辑</el-button>
            <el-button type="info" link @click="showDetailDialog(row)">详情</el-button>
            <el-button type="danger" link @click="deleteRule(row)">删除</el-button>
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
          @size-change="loadRules"
          @current-change="loadRules"
        />
      </div>
    </el-card>

    <!-- 创建/编辑规则对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="2"
            placeholder="请输入规则描述"
          />
        </el-form-item>
        <el-form-item label="匹配类型" prop="match_type">
          <el-select v-model="form.match_type" placeholder="请选择匹配类型" style="width: 100%">
            <el-option label="URI 路径" value="uri" />
            <el-option label="IP 地址" value="ip" />
            <el-option label="请求头" value="header" />
            <el-option label="请求体" value="body" />
            <el-option label="User-Agent" value="user_agent" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配方式" prop="match_mode">
          <el-select v-model="form.match_mode" placeholder="请选择匹配方式" style="width: 100%">
            <el-option label="精确匹配" value="exact" />
            <el-option label="正则匹配" value="regex" />
            <el-option label="包含匹配" value="contains" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配模式" prop="pattern">
          <el-input
            v-model="form.pattern"
            placeholder="请输入匹配模式"
          />
          <div class="form-tip">
            <el-text size="small" type="info">
              <span v-if="form.match_mode === 'exact'">精确匹配：完全相等才匹配，如：/admin</span>
              <span v-else-if="form.match_mode === 'regex'">正则匹配：支持正则表达式，如：^/admin.*、.*\.php$</span>
              <span v-else-if="form.match_mode === 'contains'">包含匹配：包含指定字符串即匹配，如：admin</span>
              <span v-else>请先选择匹配方式</span>
            </el-text>
          </div>
        </el-form-item>
        <el-form-item label="执行动作" prop="action">
          <el-radio-group v-model="form.action">
            <el-radio value="block">拦截</el-radio>
            <el-radio value="allow">允许</el-radio>
            <el-radio value="log">仅记录</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-input-number
            v-model="form.priority"
            :min="1"
            :max="1000"
            placeholder="数值越大优先级越高"
            style="width: 100%"
          />
          <div class="form-tip">
            <el-text size="small" type="info">
              优先级范围 1-1000，数值越大优先级越高
            </el-text>
          </div>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 规则详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="规则详情" width="500px">
      <el-descriptions :column="1" border v-if="currentRule">
        <el-descriptions-item label="ID">{{ currentRule.id }}</el-descriptions-item>
        <el-descriptions-item label="规则名称">{{ currentRule.name }}</el-descriptions-item>
        <el-descriptions-item label="描述">{{ currentRule.description || '无' }}</el-descriptions-item>
        <el-descriptions-item label="匹配类型">
          <el-tag :type="getMatchTypeColor(currentRule.match_type)">
            {{ getMatchTypeText(currentRule.match_type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="匹配方式">
          <el-tag :type="getMatchModeColor(currentRule.match_mode)">
            {{ getMatchModeText(currentRule.match_mode) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="匹配模式">{{ currentRule.pattern }}</el-descriptions-item>
        <el-descriptions-item label="执行动作">
          <el-tag :type="getActionColor(currentRule.action)">
            {{ getActionText(currentRule.action) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="优先级">{{ currentRule.priority }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentRule.enabled ? 'success' : 'danger'">
            {{ currentRule.enabled ? '启用' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(currentRule.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(currentRule.updated_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 导入规则对话框 -->
    <el-dialog v-model="importDialogVisible" title="导入规则" width="400px">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :limit="1"
        accept=".json,.csv"
        drag
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            支持 JSON 和 CSV 格式文件
          </div>
        </template>
      </el-upload>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleImport">确定导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Upload, Download, UploadFilled } from '@element-plus/icons-vue'
import { 
  getRuleList, 
  createRule, 
  updateRule, 
  deleteRule as deleteRuleApi, 
  toggleRuleStatus,
  batchUpdateRules,
  exportRules as exportRulesApi,
  importRules,
  type Rule 
} from '@/api/rules'

// 响应式数据
const loading = ref(false)
const rules = ref<Rule[]>([])
const selectedRules = ref<Rule[]>([])
const searchForm = reactive({
  name: '',
  enabled: undefined as boolean | undefined,
  match_type: ''
})
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 对话框相关
const dialogVisible = ref(false)
const detailDialogVisible = ref(false)
const importDialogVisible = ref(false)
const dialogTitle = ref('')
const isEdit = ref(false)
const submitting = ref(false)
const currentRule = ref<Rule | null>(null)

// 表单相关
const formRef = ref<FormInstance>()
const uploadRef = ref()
const form = reactive<Partial<Rule>>({
  name: '',
  description: '',
  match_type: 'uri',
  match_mode: 'regex',
  pattern: '',
  action: 'block',
  priority: 100,
  enabled: true
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' },
    { min: 2, max: 50, message: '规则名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  match_type: [
    { required: true, message: '请选择匹配类型', trigger: 'change' }
  ],
  match_mode: [
    { required: true, message: '请选择匹配方式', trigger: 'change' }
  ],
  pattern: [
    { required: true, message: '请输入匹配模式', trigger: 'blur' }
  ],
  action: [
    { required: true, message: '请选择执行动作', trigger: 'change' }
  ],
  priority: [
    { required: true, message: '请输入优先级', trigger: 'blur' },
    { type: 'number', min: 1, max: 1000, message: '优先级范围 1-1000', trigger: 'blur' }
  ]
}

// 工具函数
const getMatchTypeText = (type: string) => {
  const map: Record<string, string> = {
    uri: 'URI',
    ip: 'IP',
    header: '请求头',
    body: '请求体',
    user_agent: 'User-Agent'
  }
  return map[type] || type
}

const getMatchTypeColor = (type: string) => {
  const map: Record<string, string> = {
    uri: 'primary',
    ip: 'success',
    header: 'warning',
    body: 'danger',
    user_agent: 'info'
  }
  return map[type] || 'primary'
}

const getActionText = (action: string) => {
  const map: Record<string, string> = {
    block: '拦截',
    allow: '允许',
    log: '记录'
  }
  return map[action] || action
}

const getActionColor = (action: string) => {
  const map: Record<string, string> = {
    block: 'danger',
    allow: 'success',
    log: 'warning'
  }
  return map[action] || 'primary'
}

const getMatchModeText = (mode: string) => {
  const map: Record<string, string> = {
    exact: '精确匹配',
    regex: '正则匹配',
    contains: '包含匹配'
  }
  return map[mode] || mode
}

const getMatchModeColor = (mode: string) => {
  const map: Record<string, string> = {
    exact: 'success',
    regex: 'warning',
    contains: 'info'
  }
  return map[mode] || 'primary'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

// 数据加载
const loadRules = async () => {
  loading.value = true
  try {
    const response = await getRuleList({
      page: pagination.page,
      page_size: pagination.pageSize,
      ...searchForm
    })
    rules.value = response.data.list
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('加载规则列表失败')
    console.error('Load rules error:', error)
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  searchForm.name = ''
  searchForm.enabled = undefined
  searchForm.match_type = ''
  pagination.page = 1
  loadRules()
}

// 表格操作
const handleSelectionChange = (selection: Rule[]) => {
  selectedRules.value = selection
}

const toggleRule = async (row: Rule) => {
  try {
    await toggleRuleStatus(row.id!, row.enabled)
    ElMessage.success('状态切换成功')
  } catch (error) {
    ElMessage.error('状态切换失败')
    row.enabled = !row.enabled // 回滚状态
  }
}

// 对话框操作
const showCreateDialog = () => {
  dialogTitle.value = '新增规则'
  isEdit.value = false
  dialogVisible.value = true
  resetForm()
}

const showEditDialog = (row: Rule) => {
  dialogTitle.value = '编辑规则'
  isEdit.value = true
  currentRule.value = row
  Object.assign(form, row)
  dialogVisible.value = true
}

const showDetailDialog = (row: Rule) => {
  currentRule.value = row
  detailDialogVisible.value = true
}

const resetForm = () => {
  Object.assign(form, {
    name: '',
    description: '',
    match_type: 'uri',
    match_mode: 'regex',
    pattern: '',
    action: 'block',
    priority: 100,
    enabled: true
  })
  formRef.value?.clearValidate()
}

const submitForm = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (isEdit.value && currentRule.value) {
          await updateRule(currentRule.value.id!, form)
          ElMessage.success('更新成功')
        } else {
          await createRule(form as Omit<Rule, 'id' | 'created_at' | 'updated_at'>)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        loadRules()
      } catch (error) {
        ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

// 删除操作
const deleteRule = async (row: Rule) => {
  try {
    await ElMessageBox.confirm(`确定要删除规则 "${row.name}" 吗？`, '提示', {
      type: 'warning'
    })
    
    await deleteRuleApi(row.id!)
    ElMessage.success('删除成功')
    loadRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 批量操作
const batchEnable = async () => {
  try {
    const ids = selectedRules.value.map(rule => rule.id!)
    await batchUpdateRules(ids, { enabled: true })
    ElMessage.success('批量启用成功')
    loadRules()
  } catch (error) {
    ElMessage.error('批量启用失败')
  }
}

const batchDisable = async () => {
  try {
    const ids = selectedRules.value.map(rule => rule.id!)
    await batchUpdateRules(ids, { enabled: false })
    ElMessage.success('批量禁用成功')
    loadRules()
  } catch (error) {
    ElMessage.error('批量禁用失败')
  }
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedRules.value.length} 条规则吗？`, '提示', {
      type: 'warning'
    })
    
    const deletePromises = selectedRules.value.map(rule => deleteRuleApi(rule.id!))
    await Promise.all(deletePromises)
    ElMessage.success('批量删除成功')
    loadRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 导入导出
const showImportDialog = () => {
  importDialogVisible.value = true
}

const handleImport = async () => {
  const files = uploadRef.value?.uploadFiles
  if (!files || files.length === 0) {
    ElMessage.warning('请选择要导入的文件')
    return
  }
  
  try {
    await importRules(files[0].raw)
    ElMessage.success('导入成功')
    importDialogVisible.value = false
    loadRules()
  } catch (error) {
    ElMessage.error('导入失败')
  }
}

const exportRules = async () => {
  try {
    const ids = selectedRules.value.map(rule => rule.id!)
    const response = await exportRulesApi({ ids })
    
    // 创建下载链接
    const blob = new Blob([response.data])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `rules_${new Date().getTime()}.json`
    link.click()
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

onMounted(() => {
  loadRules()
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