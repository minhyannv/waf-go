<template>
  <div class="domains-view">
    <!-- 搜索表单 -->
    <el-form class="search-form" inline>
      <el-form-item label="域名">
        <el-input
          v-model="searchForm.domain"
          placeholder="请输入域名"
          clearable
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </el-form-item>

      <el-form-item label="协议">
        <el-select
          v-model="searchForm.protocol"
          placeholder="请选择"
          clearable
          @change="handleSearch"
        >
          <el-option label="HTTP" value="http" />
          <el-option label="HTTPS" value="https" />
        </el-select>
      </el-form-item>

      <el-form-item label="状态">
        <el-select
          v-model="searchForm.enabled"
          placeholder="请选择"
          clearable
          @change="handleSearch"
        >
          <el-option label="启用" :value="true" />
          <el-option label="禁用" :value="false" />
        </el-select>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="handleReset">
          重置
        </el-button>
      </el-form-item>
    </el-form>

    <!-- 工具栏 -->
    <div class="button-group">
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        添加域名
      </el-button>

      <el-button 
        type="danger" 
        :disabled="selectedDomains.length === 0"
        @click="handleBatchDelete"
      >
        <el-icon><Delete /></el-icon>
        批量删除
      </el-button>
    </div>

    <!-- 域名列表 -->
    <div class="table-container">
      <el-table
        :data="domains"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
        @sort-change="handleSortChange"
      >
        <el-table-column type="selection" width="55" />

        <el-table-column prop="domain" label="域名" sortable min-width="200">
          <template #default="{ row }">
            <div class="domain-cell">
              <el-icon class="domain-icon"><Connection /></el-icon>
              <span class="domain-name">{{ row.domain }}</span>
              <el-tag v-if="row.protocol === 'https'" type="success" size="small">SSL</el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="protocol" label="协议" width="100">
          <template #default="{ row }">
            <el-tag :type="getProtocolTagType(row.protocol)" size="small">
              {{ row.protocol.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="port" label="端口" width="80" />

        <el-table-column prop="backend_url" label="后端地址" min-width="200">
          <template #default="{ row }">
            <span class="backend-url">{{ row.backend_url }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="toggleDomain(row)"
            />
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button 
              type="primary" 
              link
              @click="handleEdit(row)"
            >
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button 
              type="danger" 
              link
              @click="handleDelete(row)"
            >
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="val => { pagination.page_size = val; loadDomains(); }"
          @current-change="val => { pagination.page = val; loadDomains(); }"
        />
      </div>
    </div>

    <!-- 域名表单对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="650px"
    >
      <el-form
        ref="formRef"
        :model="domainForm"
        :rules="rules"
        label-width="100px"
      >
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本配置" name="basic">
            <el-form-item label="域名" prop="domain">
              <el-input v-model="domainForm.domain" placeholder="请输入域名，例如: example.com" />
            </el-form-item>

            <el-form-item label="协议" prop="protocol">
              <el-radio-group v-model="domainForm.protocol">
                <el-radio label="http">HTTP</el-radio>
                <el-radio label="https">HTTPS</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="端口" prop="port">
              <el-input-number 
                v-model="domainForm.port" 
                :min="1" 
                :max="65535"
                placeholder="端口号"
              />
              <span class="port-hint">
                默认端口：HTTP - 80，HTTPS - 443
              </span>
            </el-form-item>

            <el-form-item 
              v-if="domainForm.protocol === 'https'" 
              label="SSL证书" 
              prop="ssl_certificate"
            >
              <el-input 
                v-model="domainForm.ssl_certificate" 
                type="textarea" 
                rows="5"
                placeholder="请输入 SSL 证书内容（PEM 格式）" 
              />
            </el-form-item>

            <el-form-item 
              v-if="domainForm.protocol === 'https'" 
              label="SSL私钥" 
              prop="ssl_private_key"
            >
              <el-input 
                v-model="domainForm.ssl_private_key" 
                type="textarea" 
                rows="5"
                placeholder="请输入 SSL 私钥内容（PEM 格式）" 
              />
            </el-form-item>

            <el-form-item label="后端地址" prop="backend_url">
              <el-input 
                v-model="domainForm.backend_url" 
                placeholder="请输入后端服务地址，例如: http://localhost:3000" 
              />
              <span class="backend-hint">
                支持 HTTP/HTTPS 地址，需要包含协议前缀
              </span>
            </el-form-item>

            <el-form-item label="状态" prop="enabled">
              <el-switch v-model="domainForm.enabled" />
            </el-form-item>
          </el-tab-pane>
        </el-tabs>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Edit, Delete, Connection } from '@element-plus/icons-vue'
import { domainApi, type DomainConfig } from '@/api/domains'
import { formatTime } from '@/utils/format'

// 域名列表数据
const domains = ref<DomainConfig[]>([])
const total = ref(0)
const loading = ref(false)
const searchForm = reactive({
  domain: '',
  protocol: '',
  enabled: undefined as boolean | undefined
})

// 分页配置
const pagination = reactive({
  page: 1,
  page_size: 10
})

// 域名表单
const domainForm = ref<DomainConfig>({
  domain: '',
  protocol: 'http',
  port: 80,
  ssl_certificate: '',
  ssl_private_key: '',
  backend_url: '',
  enabled: true
})

const dialogVisible = ref(false)
const dialogTitle = ref('')
const activeTab = ref('basic')
const formRef = ref()

// 加载域名列表
const loadDomains = async () => {
  loading.value = true
  try {
    const { data } = await domainApi.list({
      ...pagination,
      ...searchForm
    })
    domains.value = data.list
    total.value = data.total
  } catch (error) {
    console.error('加载域名列表失败:', error)
    ElMessage.error('加载域名列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadDomains()
}

// 重置搜索
const handleReset = () => {
  searchForm.domain = ''
  searchForm.protocol = ''
  searchForm.enabled = undefined
  handleSearch()
}

// 切换域名状态
const toggleDomain = async (row: DomainConfig) => {
  try {
    await domainApi.toggle(row.id!)
    ElMessage.success('状态更新成功')
    loadDomains()
  } catch (error) {
    console.error('更新状态失败:', error)
    ElMessage.error('更新状态失败')
    row.enabled = !row.enabled
  }
}

// 打开创建域名对话框
const handleCreate = () => {
  dialogTitle.value = '创建域名'
  domainForm.value = {
    domain: '',
    protocol: 'http',
    port: 80,
    ssl_certificate: '',
    ssl_private_key: '',
    backend_url: '',
    enabled: true
  }
  dialogVisible.value = true
  activeTab.value = 'basic'
}

// 打开编辑域名对话框
const handleEdit = async (row: DomainConfig) => {
  dialogTitle.value = '编辑域名'
  const { data } = await domainApi.get(row.id!)
  domainForm.value = { ...data }
  dialogVisible.value = true
  activeTab.value = 'basic'
}

// 删除域名
const handleDelete = async (row: DomainConfig) => {
  try {
    await ElMessageBox.confirm('确定要删除该域名吗？', '提示', {
      type: 'warning'
    })
    await domainApi.delete(row.id!)
    ElMessage.success('删除成功')
    loadDomains()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 批量删除域名
const handleBatchDelete = async () => {
  const selectedIds = selectedDomains.value.map(domain => domain.id!)
  if (selectedIds.length === 0) {
    ElMessage.warning('请选择要删除的域名')
    return
  }

  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.length} 个域名吗？`, '提示', {
      type: 'warning'
    })
    await domainApi.batchDelete(selectedIds)
    ElMessage.success('批量删除成功')
    loadDomains()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    
    if (domainForm.value.id) {
      await domainApi.update(domainForm.value.id, domainForm.value)
      ElMessage.success('更新成功')
    } else {
      await domainApi.create(domainForm.value)
      ElMessage.success('创建成功')
    }
    
    dialogVisible.value = false
    loadDomains()
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('保存失败')
  }
}

// 表单验证规则
const rules = {
  domain: [
    { required: true, message: '请输入域名', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9][-a-zA-Z0-9]*(\.[a-zA-Z0-9][-a-zA-Z0-9]*)*$/, message: '请输入有效的域名', trigger: 'blur' }
  ],
  protocol: [
    { required: true, message: '请选择协议', trigger: 'change' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口号范围为 1-65535', trigger: 'blur' }
  ],
  ssl_certificate: [
    { required: true, message: '请输入 SSL 证书', trigger: 'blur', when: (form: any) => form.protocol === 'https' }
  ],
  ssl_private_key: [
    { required: true, message: '请输入 SSL 私钥', trigger: 'blur', when: (form: any) => form.protocol === 'https' }
  ],
  backend_url: [
    { required: true, message: '请输入后端地址', trigger: 'blur' },
    { pattern: /^https?:\/\/.+/, message: '请输入有效的 HTTP/HTTPS 地址', trigger: 'blur' },
    {
      validator: (rule: any, value: string, callback: Function) => {
        if (!value) {
          callback()
          return
        }
        const backendProtocol = value.startsWith('https://') ? 'https' : 'http'
        if (domainForm.value.protocol === 'https' && backendProtocol !== 'https') {
          callback(new Error('HTTPS 域名必须使用 HTTPS 后端地址'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 监听协议变化，自动设置默认端口和检查后端地址
watch(() => domainForm.value.protocol, (protocol) => {
  // 设置默认端口
  if (!domainForm.value.port || domainForm.value.port === 80 || domainForm.value.port === 443) {
    domainForm.value.port = protocol === 'https' ? 443 : 80
  }

  // 检查后端地址协议
  if (domainForm.value.backend_url) {
    const backendProtocol = domainForm.value.backend_url.startsWith('https://') ? 'https' : 'http'
    if (protocol === 'https' && backendProtocol !== 'https') {
      ElMessage.warning('HTTPS 域名建议使用 HTTPS 后端地址')
    }
  }
})

// 表格多选
const selectedDomains = ref<DomainConfig[]>([])
const handleSelectionChange = (selection: DomainConfig[]) => {
  selectedDomains.value = selection
}

// 表格排序
const handleSortChange = ({ prop, order }: { prop?: string, order?: string }) => {
  // 暂不支持后端排序，仅前端排序
  if (!prop || !order) return
  
  domains.value.sort((a: any, b: any) => {
    const aValue = a[prop]
    const bValue = b[prop]
    if (order === 'ascending') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })
}

// 获取协议标签类型
const getProtocolTagType = (protocol: string) => {
  return protocol === 'https' ? 'success' : ''
}

onMounted(() => {
  loadDomains()
})
</script>

<style scoped>
.domains-view {
  padding: 20px;
}

.search-form {
  margin-bottom: 20px;
  padding: 24px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.search-form :deep(.el-form-item) {
  margin-bottom: 0;
  margin-right: 16px;
}

.search-form :deep(.el-input),
.search-form :deep(.el-select) {
  width: 200px;
}

.search-form :deep(.el-form-item__label) {
  font-weight: 500;
}

.button-group {
  margin-bottom: 20px;
}

.table-container {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.domain-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.domain-icon {
  color: #409EFF;
}

.domain-name {
  flex: 1;
}

.port-hint {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
  margin-top: 4px;
  line-height: 1.4;
}

.backend-hint {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
  margin-top: 4px;
  line-height: 1.4;
}

.backend-url {
  color: #606266;
  font-family: monospace;
}

.pagination-container {
  padding: 20px;
  display: flex;
  justify-content: center;
}
</style> 