<template>
  <div class="policies-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>策略管理</span>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            新增策略
          </el-button>
        </div>
      </template>

      <!-- 搜索筛选 -->
      <div class="search-section">
        <el-form :model="searchForm" inline>
          <el-form-item label="策略名称">
            <el-input
              v-model="searchForm.name"
              placeholder="请输入策略名称"
              clearable
              style="width: 200px"
            />
          </el-form-item>
          <el-form-item label="域名">
            <el-input
              v-model="searchForm.domain"
              placeholder="请输入域名"
              clearable
              style="width: 200px"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select
              v-model="searchForm.enabled"
              placeholder="请选择状态"
              clearable
              style="width: 120px"
            >
              <el-option label="启用" :value="true" />
              <el-option label="禁用" :value="false" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadPolicies">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
            <el-button @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 批量操作 -->
      <div class="batch-actions" v-if="selectedPolicies.length > 0">
        <el-alert
          :title="`已选择 ${selectedPolicies.length} 项`"
          type="info"
          show-icon
          :closable="false"
        >
          <template #default>
            <el-button size="small" type="danger" @click="batchDelete">
              批量删除
            </el-button>
          </template>
        </el-alert>
      </div>

      <!-- 策略列表 -->
      <el-table
        :data="policies"
        v-loading="loading"
        @selection-change="handleSelectionChange"
        style="width: 100%"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="策略名称" min-width="150">
          <template #default="{ row }">
            <el-link type="primary" @click="showDetail(row)">
              {{ row.name }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="domain" label="域名" min-width="150" show-overflow-tooltip />
        <el-table-column label="关联规则" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ getRuleCount(row) }} 个</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="toggleStatus(row)"
              :loading="row.toggling"
            />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="showDetail(row)">详情</el-button>
            <el-button size="small" type="primary" @click="showEditDialog(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deletePolicy(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadPolicies"
          @current-change="loadPolicies"
        />
      </div>
    </el-card>

    <!-- 创建/编辑策略对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑策略' : '新增策略'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="策略名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入策略名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入策略描述"
          />
        </el-form-item>
        <el-form-item label="应用域名" prop="domain_id">
          <el-select
            v-model="form.domain_id"
            placeholder="请选择要应用策略的域名"
            style="width: 100%"
            :loading="domainsLoading"
            filterable
          >
            <el-option
              v-for="domain in availableDomains"
              :key="domain.id"
              :label="`${domain.domain} (${domain.protocol}://${domain.domain}:${domain.port})`"
              :value="domain.id"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ domain.domain }}</span>
                <div>
                  <el-tag size="small" :type="domain.protocol === 'https' ? 'success' : 'info'">
                    {{ domain.protocol.toUpperCase() }}
                  </el-tag>
                  <el-tag size="small" type="warning" style="margin-left: 4px;">
                    :{{ domain.port }}
                  </el-tag>
                </div>
              </div>
            </el-option>
          </el-select>
          <div class="form-tip">
            <el-text size="small" type="warning">
              <el-icon><Warning /></el-icon>
              必须选择域名，策略将应用到指定域名的所有请求
            </el-text>
          </div>
        </el-form-item>
        <el-form-item label="关联规则">
          <el-select
            v-model="form.rule_ids"
            multiple
            placeholder="请选择关联的规则（可选）"
            style="width: 100%"
            :loading="rulesLoading"
            collapse-tags
            collapse-tags-tooltip
            :max-collapse-tags="3"
          >
            <el-option
              v-for="rule in availableRules"
              :key="rule.id"
              :label="`${rule.name} (${getActionText(rule.action)})`"
              :value="rule.id"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ rule.name }}</span>
                <el-tag :type="getActionColor(rule.action)" size="small">
                  {{ getActionText(rule.action) }}
                </el-tag>
              </div>
            </el-option>
          </el-select>
          <div class="form-tip">
            <el-text size="small" type="info">
              可选，不选择规则时策略将允许所有请求通过
            </el-text>
          </div>
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 策略详情对话框 -->
    <el-dialog
      v-model="detailVisible"
      title="策略详情"
      width="800px"
    >
      <div v-if="currentPolicy">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="策略名称">
            {{ currentPolicy.name }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentPolicy.enabled ? 'success' : 'danger'">
              {{ currentPolicy.enabled ? '启用' : '禁用' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="域名">
            {{ currentPolicy.domain || '未指定' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(currentPolicy.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            {{ currentPolicy.description || '无描述' }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">关联规则</el-divider>
        
        <el-table :data="currentPolicyRules" style="width: 100%">
          <el-table-column prop="name" label="规则名称" min-width="150" />
          <el-table-column prop="match_type" label="匹配类型" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ getMatchTypeText(row.match_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="pattern" label="匹配模式" min-width="200" show-overflow-tooltip />
          <el-table-column prop="action" label="动作" width="100">
            <template #default="{ row }">
              <el-tag :type="getActionColor(row.action)" size="small">
                {{ getActionText(row.action) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="priority" label="优先级" width="80" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.enabled ? 'success' : 'danger'" size="small">
                {{ row.enabled ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus, Search, Warning } from '@element-plus/icons-vue'
import {
  policyApi,
  type PolicyConfig,
  type PolicyListRequest,
  type PolicyListResponse,
  type PolicyWithRules
} from '@/api/policies'
import { domainApi, type DomainConfig } from '@/api/domains'
import { formatTime } from '@/utils/format'

// 响应式数据
const loading = ref(false)
const policies = ref<PolicyConfig[]>([])
const selectedPolicies = ref<PolicyConfig[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const rulesLoading = ref(false)
const domainsLoading = ref(false)
const availableRules = ref<any[]>([])
const availableDomains = ref<DomainConfig[]>([])
const currentPolicy = ref<PolicyWithRules | null>(null)
const currentPolicyRules = ref<any[]>([])
const currentEditingPolicyId = ref<number | undefined>()

// 表单引用
const formRef = ref<FormInstance>()

// 搜索表单
const searchForm = reactive<PolicyListRequest>({
  name: '',
  domain: '',
  enabled: undefined
})

// 分页
const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

// 表单数据
const form = reactive({
  name: '',
  description: '',
  domain_id: undefined as number | undefined,
  rule_ids: [] as number[],
  enabled: true
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入策略名称', trigger: 'blur' }
  ],
  domain_id: [
    { required: true, message: '请选择应用域名', trigger: 'change' }
  ]
}

// 加载策略列表
const loadPolicies = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.size,
      ...searchForm
    }
    
    // 过滤掉undefined值
    Object.keys(params).forEach(key => {
      if (params[key as keyof typeof params] === undefined || params[key as keyof typeof params] === '') {
        delete params[key as keyof typeof params]
      }
    })

    const response = await policyApi.getPolicyList(params)
    policies.value = response.data.list
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('加载策略列表失败')
  } finally {
    loading.value = false
  }
}

// 加载可用规则
const loadAvailableRules = async () => {
  rulesLoading.value = true
  try {
    const response = await policyApi.getAvailableRules()
    availableRules.value = response.data
  } catch (error) {
    ElMessage.error('加载规则列表失败')
  } finally {
    rulesLoading.value = false
  }
}

// 加载可用域名
const loadDomains = async () => {
  domainsLoading.value = true
  try {
    const response = await domainApi.getDomains({ enabled: true })
    availableDomains.value = response.data.list
  } catch (error) {
    console.error('加载域名列表失败:', error)
    ElMessage.error('加载域名列表失败')
  } finally {
    domainsLoading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  searchForm.name = ''
  searchForm.domain = ''
  searchForm.enabled = undefined
  pagination.page = 1
  loadPolicies()
}

// 显示创建对话框
const showCreateDialog = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
  loadAvailableRules()
  loadDomains()
}

// 显示编辑对话框
const showEditDialog = (policy: PolicyConfig) => {
  isEdit.value = true
  // 保存当前编辑的策略ID
  currentEditingPolicyId.value = policy.id
  Object.assign(form, {
    name: policy.name,
    description: policy.description || '',
    domain_id: policy.domain_id,
    rule_ids: Array.isArray(policy.rule_ids) ? policy.rule_ids : [],
    enabled: policy.enabled
  })
  dialogVisible.value = true
  loadAvailableRules()
  loadDomains()
}

// 重置表单
const resetForm = () => {
  Object.assign(form, {
    name: '',
    description: '',
    domain_id: undefined,
    rule_ids: [],
    enabled: true
  })
  currentEditingPolicyId.value = undefined
  formRef.value?.clearValidate()
}

// 提交表单
const submitForm = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    console.log('提交的表单数据:', form)
    if (isEdit.value) {
      if (currentEditingPolicyId.value) {
        await policyApi.updatePolicy(currentEditingPolicyId.value, form)
        ElMessage.success('更新成功')
      }
    } else {
      await policyApi.createPolicy(form)
      ElMessage.success('创建成功')
    }
    
    dialogVisible.value = false
    loadPolicies()
  } catch (error) {
    ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

// 显示详情
const showDetail = async (policy: PolicyConfig) => {
  if (!policy.id) return
  
  try {
    const response = await policyApi.getPolicyWithRules(policy.id)
    currentPolicy.value = response.data
    currentPolicyRules.value = response.data.rules || []
    detailVisible.value = true
  } catch (error) {
    ElMessage.error('加载策略详情失败')
  }
}

// 切换状态
const toggleStatus = async (policy: PolicyConfig) => {
  if (!policy.id) return
  
  policy.toggling = true
  try {
    await policyApi.togglePolicy(policy.id)
    ElMessage.success('状态切换成功')
  } catch (error) {
    // 恢复原状态
    policy.enabled = !policy.enabled
    ElMessage.error('状态切换失败')
  } finally {
    policy.toggling = false
  }
}

// 删除策略
const deletePolicy = async (policy: PolicyConfig) => {
  if (!policy.id) return
  
  try {
    await ElMessageBox.confirm(
      `确定要删除策略 "${policy.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await policyApi.deletePolicy(policy.id)
    ElMessage.success('删除成功')
    loadPolicies()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 批量删除
const batchDelete = async () => {
  if (selectedPolicies.value.length === 0) return
  
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedPolicies.value.length} 个策略吗？`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const ids = selectedPolicies.value.map(p => p.id!).filter(id => id)
    await policyApi.batchDeletePolicies(ids)
    ElMessage.success('批量删除成功')
    selectedPolicies.value = []
    loadPolicies()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 处理选择变化
const handleSelectionChange = (selection: PolicyConfig[]) => {
  selectedPolicies.value = selection
}

// 获取规则数量
const getRuleCount = (policy: PolicyConfig) => {
  if (Array.isArray(policy.rule_ids)) {
    return policy.rule_ids.length
  }
  return 0
}

// 获取匹配类型文本
const getMatchTypeText = (type: string) => {
  const typeMap: Record<string, string> = {
    uri: 'URI',
    ip: 'IP',
    header: '请求头',
    body: '请求体',
    user_agent: 'User-Agent'
  }
  return typeMap[type] || type
}

// 获取动作文本
const getActionText = (action: string) => {
  const actionMap: Record<string, string> = {
    block: '拦截',
    allow: '允许',
    log: '记录'
  }
  return actionMap[action] || action
}

// 获取动作颜色
const getActionColor = (action: string) => {
  const colorMap: Record<string, string> = {
    block: 'danger',
    allow: 'success',
    log: 'warning'
  }
  return colorMap[action] || ''
}

// 初始化
onMounted(() => {
  loadPolicies()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-section {
  margin-bottom: 20px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.batch-actions {
  margin-bottom: 20px;
}

.pagination-wrapper {
  margin-top: 20px;
  text-align: right;
}

.form-tip {
  margin-top: 5px;
}

:deep(.el-table .el-table__cell) {
  padding: 8px 0;
}

:deep(.el-descriptions__label) {
  font-weight: bold;
}
</style> 