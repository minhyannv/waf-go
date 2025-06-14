<template>
  <div class="settings-container">
    <el-tabs v-model="activeTab" class="settings-tabs">
      <!-- 系统配置 -->
      <el-tab-pane label="系统配置" name="config">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>系统配置</span>
              <div>
                <el-button @click="resetConfig" :loading="resetting">重置默认</el-button>
                <el-button type="primary" @click="saveConfig" :loading="saving">保存配置</el-button>
              </div>
            </div>
          </template>

          <el-form :model="configForm" label-width="150px" class="config-form">
            <!-- WAF引擎配置 -->
            <el-divider content-position="left">
              <el-text size="large" tag="b">WAF引擎配置</el-text>
            </el-divider>
            
            <el-form-item label="启用速率限制">
              <el-switch v-model="configForm.waf.enable_rate_limit" />
              <el-text size="small" type="info" style="margin-left: 10px">
                开启后将对请求进行速率限制
              </el-text>
            </el-form-item>

            <el-form-item label="时间窗口(秒)" v-if="configForm.waf.enable_rate_limit">
              <el-input-number 
                v-model="configForm.waf.rate_limit_window" 
                :min="1" 
                :max="3600" 
                style="width: 200px"
              />
              <el-text size="small" type="info" style="margin-left: 10px">
                统计请求数的时间窗口
              </el-text>
            </el-form-item>

            <el-form-item label="最大请求数" v-if="configForm.waf.enable_rate_limit">
              <el-input-number 
                v-model="configForm.waf.max_requests" 
                :min="1" 
                :max="10000" 
                style="width: 200px"
              />
              <el-text size="small" type="info" style="margin-left: 10px">
                时间窗口内允许的最大请求数
              </el-text>
            </el-form-item>

            <el-form-item label="启用黑名单">
              <el-switch v-model="configForm.waf.enable_blacklist" />
              <el-text size="small" type="info" style="margin-left: 10px">
                开启后将检查黑名单规则
              </el-text>
            </el-form-item>

            <el-form-item label="启用白名单">
              <el-switch v-model="configForm.waf.enable_whitelist" />
              <el-text size="small" type="info" style="margin-left: 10px">
                开启后将检查白名单规则
              </el-text>
            </el-form-item>

            <!-- 服务器配置 -->
            <el-divider content-position="left">
              <el-text size="large" tag="b">服务器配置</el-text>
            </el-divider>

            <el-form-item label="运行模式">
              <el-select v-model="configForm.server.mode" style="width: 200px">
                <el-option label="调试模式" value="debug" />
                <el-option label="发布模式" value="release" />
                <el-option label="测试模式" value="test" />
              </el-select>
              <el-text size="small" type="info" style="margin-left: 10px">
                生产环境建议使用发布模式
              </el-text>
            </el-form-item>

            <!-- 日志配置 -->
            <el-divider content-position="left">
              <el-text size="large" tag="b">日志配置</el-text>
            </el-divider>

            <el-form-item label="日志级别">
              <el-select v-model="configForm.log.level" style="width: 200px">
                <el-option label="调试" value="debug" />
                <el-option label="信息" value="info" />
                <el-option label="警告" value="warn" />
                <el-option label="错误" value="error" />
              </el-select>
              <el-text size="small" type="info" style="margin-left: 10px">
                日志记录的详细程度
              </el-text>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>

      <!-- 白名单管理 -->
      <el-tab-pane label="白名单管理" name="whitelist">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>白名单管理</span>
              <div>
                <el-button 
                  type="danger" 
                  :disabled="whitelistSelection.length === 0"
                  @click="batchDeleteWhitelist"
                >
                  批量删除 ({{ whitelistSelection.length }})
                </el-button>
                <el-button type="primary" @click="showWhitelistDialog('create')">添加白名单</el-button>
              </div>
            </div>
          </template>

          <!-- 搜索筛选 -->
          <div class="search-bar">
            <el-form :model="whitelistQuery" inline>
              <el-form-item label="类型">
                <el-select v-model="whitelistQuery.type" clearable placeholder="全部类型" style="width: 120px">
                  <el-option label="IP地址" value="ip" />
                  <el-option label="URI路径" value="uri" />
                  <el-option label="User-Agent" value="user_agent" />
                </el-select>
              </el-form-item>
              <el-form-item label="值">
                <el-input v-model="whitelistQuery.value" placeholder="搜索值" clearable style="width: 200px" />
              </el-form-item>
              <el-form-item label="状态">
                <el-select v-model="whitelistQuery.enabled" clearable placeholder="全部状态" style="width: 120px">
                  <el-option label="启用" :value="true" />
                  <el-option label="禁用" :value="false" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="loadWhitelistData">搜索</el-button>
                <el-button @click="resetWhitelistQuery">重置</el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- 白名单表格 -->
          <el-table
            :data="whitelistData"
            v-loading="whitelistLoading"
            @selection-change="handleWhitelistSelectionChange"
            style="width: 100%"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="type" label="类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getTypeTagType(row.type)">
                  {{ getTypeLabel(row.type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="value" label="值" min-width="200" />
            <el-table-column prop="comment" label="备注" min-width="150" />
            <el-table-column prop="enabled" label="状态" width="100">
              <template #default="{ row }">
                <el-switch
                  v-model="row.enabled"
                  @change="toggleWhitelistStatus(row)"
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
                <el-button size="small" @click="showWhitelistDialog('edit', row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteWhitelistItem(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="pagination-wrapper">
            <el-pagination
              v-model:current-page="whitelistQuery.page"
              v-model:page-size="whitelistQuery.page_size"
              :page-sizes="[10, 20, 50, 100]"
              :total="whitelistTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="loadWhitelistData"
              @current-change="loadWhitelistData"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 黑名单管理 -->
      <el-tab-pane label="黑名单管理" name="blacklist">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>黑名单管理</span>
              <div>
                <el-button 
                  type="danger" 
                  :disabled="blacklistSelection.length === 0"
                  @click="batchDeleteBlacklist"
                >
                  批量删除 ({{ blacklistSelection.length }})
                </el-button>
                <el-button type="primary" @click="showBlacklistDialog('create')">添加黑名单</el-button>
              </div>
            </div>
          </template>

          <!-- 搜索筛选 -->
          <div class="search-bar">
            <el-form :model="blacklistQuery" inline>
              <el-form-item label="类型">
                <el-select v-model="blacklistQuery.type" clearable placeholder="全部类型" style="width: 120px">
                  <el-option label="IP地址" value="ip" />
                  <el-option label="URI路径" value="uri" />
                  <el-option label="User-Agent" value="user_agent" />
                </el-select>
              </el-form-item>
              <el-form-item label="值">
                <el-input v-model="blacklistQuery.value" placeholder="搜索值" clearable style="width: 200px" />
              </el-form-item>
              <el-form-item label="状态">
                <el-select v-model="blacklistQuery.enabled" clearable placeholder="全部状态" style="width: 120px">
                  <el-option label="启用" :value="true" />
                  <el-option label="禁用" :value="false" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="loadBlacklistData">搜索</el-button>
                <el-button @click="resetBlacklistQuery">重置</el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- 黑名单表格 -->
          <el-table
            :data="blacklistData"
            v-loading="blacklistLoading"
            @selection-change="handleBlacklistSelectionChange"
            style="width: 100%"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="type" label="类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getTypeTagType(row.type)">
                  {{ getTypeLabel(row.type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="value" label="值" min-width="200" />
            <el-table-column prop="comment" label="备注" min-width="150" />
            <el-table-column prop="enabled" label="状态" width="100">
              <template #default="{ row }">
                <el-switch
                  v-model="row.enabled"
                  @change="toggleBlacklistStatus(row)"
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
                <el-button size="small" @click="showBlacklistDialog('edit', row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteBlacklistItem(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <div class="pagination-wrapper">
            <el-pagination
              v-model:current-page="blacklistQuery.page"
              v-model:page-size="blacklistQuery.page_size"
              :page-sizes="[10, 20, 50, 100]"
              :total="blacklistTotal"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="loadBlacklistData"
              @current-change="loadBlacklistData"
            />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 白名单编辑对话框 -->
    <el-dialog
      v-model="whitelistDialogVisible"
      :title="whitelistDialogMode === 'create' ? '添加白名单' : '编辑白名单'"
      width="500px"
    >
      <el-form :model="whitelistForm" :rules="whitelistRules" ref="whitelistFormRef" label-width="100px">
        <el-form-item label="类型" prop="type">
          <el-select v-model="whitelistForm.type" placeholder="请选择类型" style="width: 100%">
            <el-option label="IP地址" value="ip" />
            <el-option label="URI路径" value="uri" />
            <el-option label="User-Agent" value="user_agent" />
          </el-select>
        </el-form-item>
        <el-form-item label="值" prop="value">
          <el-input v-model="whitelistForm.value" placeholder="请输入值" />
          <div class="form-tip">
            <el-text size="small" type="info">
              <span v-if="whitelistForm.type === 'ip'">支持单个IP或CIDR格式，如：192.168.1.1 或 192.168.1.0/24</span>
              <span v-else-if="whitelistForm.type === 'uri'">URI路径，如：/api/public</span>
              <span v-else-if="whitelistForm.type === 'user_agent'">User-Agent字符串</span>
            </el-text>
          </div>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="whitelistForm.comment" placeholder="请输入备注" />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="whitelistForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="whitelistDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitWhitelistForm" :loading="whitelistSubmitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 黑名单编辑对话框 -->
    <el-dialog
      v-model="blacklistDialogVisible"
      :title="blacklistDialogMode === 'create' ? '添加黑名单' : '编辑黑名单'"
      width="500px"
    >
      <el-form :model="blacklistForm" :rules="blacklistRules" ref="blacklistFormRef" label-width="100px">
        <el-form-item label="类型" prop="type">
          <el-select v-model="blacklistForm.type" placeholder="请选择类型" style="width: 100%">
            <el-option label="IP地址" value="ip" />
            <el-option label="URI路径" value="uri" />
            <el-option label="User-Agent" value="user_agent" />
          </el-select>
        </el-form-item>
        <el-form-item label="值" prop="value">
          <el-input v-model="blacklistForm.value" placeholder="请输入值" />
          <div class="form-tip">
            <el-text size="small" type="info">
              <span v-if="blacklistForm.type === 'ip'">支持单个IP或CIDR格式，如：192.168.1.1 或 192.168.1.0/24</span>
              <span v-else-if="blacklistForm.type === 'uri'">URI路径，如：/admin</span>
              <span v-else-if="blacklistForm.type === 'user_agent'">User-Agent字符串</span>
            </el-text>
          </div>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="blacklistForm.comment" placeholder="请输入备注" />
        </el-form-item>
        <el-form-item label="启用状态">
          <el-switch v-model="blacklistForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="blacklistDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitBlacklistForm" :loading="blacklistSubmitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { getSystemConfig, updateSystemConfig, resetSystemConfig } from '@/api/config'
import { 
  getWhiteListList, 
  createWhiteList, 
  updateWhiteList, 
  deleteWhiteList, 
  batchDeleteWhiteList, 
  toggleWhiteListStatus 
} from '@/api/whitelists'
import { 
  getBlackListList, 
  createBlackList, 
  updateBlackList, 
  deleteBlackList, 
  batchDeleteBlackList, 
  toggleBlackListStatus 
} from '@/api/blacklists'
import type { SystemConfig } from '@/api/config'
import type { WhiteList } from '@/api/whitelists'
import type { BlackList } from '@/api/blacklists'

// 当前激活的标签页
const activeTab = ref('config')

// 系统配置相关
const configForm = ref<SystemConfig>({
  waf: {
    rate_limit_window: 60,
    max_requests: 100,
    enable_rate_limit: false,
    enable_blacklist: true,
    enable_whitelist: true
  },
  server: {
    mode: 'debug'
  },
  log: {
    level: 'info'
  }
})

const saving = ref(false)
const resetting = ref(false)

// 白名单相关
const whitelistData = ref<WhiteList[]>([])
const whitelistLoading = ref(false)
const whitelistTotal = ref(0)
const whitelistSelection = ref<WhiteList[]>([])

const whitelistQuery = reactive({
  page: 1,
  page_size: 10,
  type: '',
  value: '',
  enabled: undefined as boolean | undefined
})

const whitelistDialogVisible = ref(false)
const whitelistDialogMode = ref<'create' | 'edit'>('create')
const whitelistSubmitting = ref(false)
const whitelistFormRef = ref<FormInstance>()

const whitelistForm = reactive({
  id: 0,
  type: '' as 'ip' | 'uri' | 'user_agent' | '',
  value: '',
  comment: '',
  enabled: true
})

const whitelistRules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  value: [{ required: true, message: '请输入值', trigger: 'blur' }]
}

// 黑名单相关
const blacklistData = ref<BlackList[]>([])
const blacklistLoading = ref(false)
const blacklistTotal = ref(0)
const blacklistSelection = ref<BlackList[]>([])

const blacklistQuery = reactive({
  page: 1,
  page_size: 10,
  type: '',
  value: '',
  enabled: undefined as boolean | undefined
})

const blacklistDialogVisible = ref(false)
const blacklistDialogMode = ref<'create' | 'edit'>('create')
const blacklistSubmitting = ref(false)
const blacklistFormRef = ref<FormInstance>()

const blacklistForm = reactive({
  id: 0,
  type: '' as 'ip' | 'uri' | 'user_agent' | '',
  value: '',
  comment: '',
  enabled: true
})

const blacklistRules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  value: [{ required: true, message: '请输入值', trigger: 'blur' }]
}

// 系统配置方法
const loadConfig = async () => {
  try {
    const response = await getSystemConfig()
    configForm.value = response.data
  } catch (error) {
    console.error('加载配置失败:', error)
    ElMessage.error('加载配置失败')
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    await updateSystemConfig(configForm.value)
    ElMessage.success('配置保存成功')
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存配置失败')
  } finally {
    saving.value = false
  }
}

const resetConfig = async () => {
  try {
    await ElMessageBox.confirm('确定要重置为默认配置吗？', '确认重置', {
      type: 'warning'
    })
    
    resetting.value = true
    await resetSystemConfig()
    await loadConfig()
    ElMessage.success('配置重置成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重置配置失败:', error)
      ElMessage.error('重置配置失败')
    }
  } finally {
    resetting.value = false
  }
}

// 白名单方法
const loadWhitelistData = async () => {
  whitelistLoading.value = true
  try {
    const response = await getWhiteListList(whitelistQuery)
    whitelistData.value = response.data.list
    whitelistTotal.value = response.data.total
  } catch (error) {
    console.error('加载白名单失败:', error)
    ElMessage.error('加载白名单失败')
  } finally {
    whitelistLoading.value = false
  }
}

const resetWhitelistQuery = () => {
  whitelistQuery.page = 1
  whitelistQuery.page_size = 10
  whitelistQuery.type = ''
  whitelistQuery.value = ''
  whitelistQuery.enabled = undefined
  loadWhitelistData()
}

const handleWhitelistSelectionChange = (selection: WhiteList[]) => {
  whitelistSelection.value = selection
}

const showWhitelistDialog = (mode: 'create' | 'edit', row?: WhiteList) => {
  whitelistDialogMode.value = mode
  if (mode === 'create') {
    Object.assign(whitelistForm, {
      id: 0,
      type: '',
      value: '',
      comment: '',
      enabled: true
    })
  } else if (row) {
    Object.assign(whitelistForm, { ...row })
  }
  whitelistDialogVisible.value = true
}

const submitWhitelistForm = async () => {
  if (!whitelistFormRef.value) return
  
  try {
    await whitelistFormRef.value.validate()
    whitelistSubmitting.value = true
    
    if (whitelistDialogMode.value === 'create') {
      const { id, ...createData } = whitelistForm
      await createWhiteList(createData as Omit<WhiteList, 'id' | 'created_at' | 'updated_at'>)
      ElMessage.success('添加成功')
    } else {
      await updateWhiteList(whitelistForm.id, whitelistForm as Partial<WhiteList>)
      ElMessage.success('更新成功')
    }
    
    whitelistDialogVisible.value = false
    loadWhitelistData()
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败')
  } finally {
    whitelistSubmitting.value = false
  }
}

const deleteWhitelistItem = async (row: WhiteList) => {
  try {
    await ElMessageBox.confirm(`确定要删除白名单 "${row.value}" 吗？`, '确认删除', {
      type: 'warning'
    })
    
    await deleteWhiteList(row.id!)
    ElMessage.success('删除成功')
    loadWhitelistData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const batchDeleteWhitelist = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${whitelistSelection.value.length} 个白名单吗？`, '确认批量删除', {
      type: 'warning'
    })
    
    const ids = whitelistSelection.value.map(item => item.id!)
    await batchDeleteWhiteList(ids)
    ElMessage.success('批量删除成功')
    loadWhitelistData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  }
}

const toggleWhitelistStatus = async (row: WhiteList) => {
  try {
    await toggleWhiteListStatus(row.id!)
    ElMessage.success('状态更新成功')
  } catch (error) {
    console.error('状态更新失败:', error)
    ElMessage.error('状态更新失败')
    // 恢复原状态
    row.enabled = !row.enabled
  }
}

// 黑名单方法
const loadBlacklistData = async () => {
  blacklistLoading.value = true
  try {
    const response = await getBlackListList(blacklistQuery)
    blacklistData.value = response.data.list
    blacklistTotal.value = response.data.total
  } catch (error) {
    console.error('加载黑名单失败:', error)
    ElMessage.error('加载黑名单失败')
  } finally {
    blacklistLoading.value = false
  }
}

const resetBlacklistQuery = () => {
  blacklistQuery.page = 1
  blacklistQuery.page_size = 10
  blacklistQuery.type = ''
  blacklistQuery.value = ''
  blacklistQuery.enabled = undefined
  loadBlacklistData()
}

const handleBlacklistSelectionChange = (selection: BlackList[]) => {
  blacklistSelection.value = selection
}

const showBlacklistDialog = (mode: 'create' | 'edit', row?: BlackList) => {
  blacklistDialogMode.value = mode
  if (mode === 'create') {
    Object.assign(blacklistForm, {
      id: 0,
      type: '',
      value: '',
      comment: '',
      enabled: true
    })
  } else if (row) {
    Object.assign(blacklistForm, { ...row })
  }
  blacklistDialogVisible.value = true
}

const submitBlacklistForm = async () => {
  if (!blacklistFormRef.value) return
  
  try {
    await blacklistFormRef.value.validate()
    blacklistSubmitting.value = true
    
    if (blacklistDialogMode.value === 'create') {
      const { id, ...createData } = blacklistForm
      await createBlackList(createData as Omit<BlackList, 'id' | 'created_at' | 'updated_at'>)
      ElMessage.success('添加成功')
    } else {
      await updateBlackList(blacklistForm.id, blacklistForm as Partial<BlackList>)
      ElMessage.success('更新成功')
    }
    
    blacklistDialogVisible.value = false
    loadBlacklistData()
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败')
  } finally {
    blacklistSubmitting.value = false
  }
}

const deleteBlacklistItem = async (row: BlackList) => {
  try {
    await ElMessageBox.confirm(`确定要删除黑名单 "${row.value}" 吗？`, '确认删除', {
      type: 'warning'
    })
    
    await deleteBlackList(row.id!)
    ElMessage.success('删除成功')
    loadBlacklistData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const batchDeleteBlacklist = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${blacklistSelection.value.length} 个黑名单吗？`, '确认批量删除', {
      type: 'warning'
    })
    
    const ids = blacklistSelection.value.map(item => item.id!)
    await batchDeleteBlackList(ids)
    ElMessage.success('批量删除成功')
    loadBlacklistData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  }
}

const toggleBlacklistStatus = async (row: BlackList) => {
  try {
    await toggleBlackListStatus(row.id!)
    ElMessage.success('状态更新成功')
  } catch (error) {
    console.error('状态更新失败:', error)
    ElMessage.error('状态更新失败')
    // 恢复原状态
    row.enabled = !row.enabled
  }
}

// 工具方法
const getTypeTagType = (type: string) => {
  switch (type) {
    case 'ip': return 'success'
    case 'uri': return 'warning'
    case 'user_agent': return 'info'
    default: return ''
  }
}

const getTypeLabel = (type: string) => {
  switch (type) {
    case 'ip': return 'IP地址'
    case 'uri': return 'URI路径'
    case 'user_agent': return 'User-Agent'
    default: return type
  }
}

const formatTime = (time: string) => {
  return new Date(time).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  loadConfig()
  loadWhitelistData()
  loadBlacklistData()
})
</script>

<style scoped>
.settings-container {
  padding: 20px;
}

.settings-tabs {
  background: white;
  border-radius: 8px;
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.config-form {
  max-width: 800px;
}

.search-bar {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.form-tip {
  margin-top: 5px;
}

:deep(.el-divider__text) {
  background-color: white;
}
</style> 