<template>
  <div class="folder-tree">
    <div class="ft-head">
      <span class="ft-title">文件夹</span>
      <el-button size="small" text @click="createFolder(null)" title="新建根目录">
        <el-icon><Plus /></el-icon>
      </el-button>
    </div>

    <div
      class="ft-item"
      :class="{ active: modelValue === null }"
      @click="selectRoot"
    >
      <el-icon><FolderOpened /></el-icon>
      <span class="ft-name">全部文件</span>
    </div>

    <el-tree
      v-loading="loading"
      :data="tree"
      node-key="id"
      default-expand-all
      :expand-on-click-node="false"
      :props="{ label: 'name', children: 'children' }"
      @node-click="onNodeClick"
    >
      <template #default="{ node, data }">
        <span class="ft-node">
          <el-icon><Folder /></el-icon>
          <span class="ft-name" :title="data.name">{{ data.name }}</span>
          <span class="ft-ops">
            <el-icon title="新建子目录" @click.stop="createFolder(data.id)"><Plus /></el-icon>
            <el-icon title="重命名" @click.stop="renameFolder(data)"><Edit /></el-icon>
            <el-icon title="删除" @click.stop="removeFolder(data)"><Delete /></el-icon>
          </span>
        </span>
      </template>
    </el-tree>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  apiListFolders,
  apiCreateFolder,
  apiUpdateFolder,
  apiDeleteFolder
} from '@/api/extras'

const props = defineProps({
  modelValue: { default: null } // current folder id, null = all
})
const emit = defineEmits(['update:modelValue', 'changed'])

const tree = ref([])
const loading = ref(false)

// build tree from flat backend list
function buildTree(list) {
  const map = {}
  list.forEach((f) => { map[f.id] = { ...f, children: [] } })
  const roots = []
  list.forEach((f) => {
    if (f.parent_id && map[f.parent_id]) map[f.parent_id].children.push(map[f.id])
    else roots.push(map[f.id])
  })
  return roots
}

async function load() {
  loading.value = true
  try {
    const resp = await apiListFolders()
    const list = resp.list || resp || []
    tree.value = buildTree(list)
  } catch (e) {
    tree.value = []
  } finally {
    loading.value = false
  }
}

function selectRoot() {
  emit('update:modelValue', null)
}
function onNodeClick(data) {
  emit('update:modelValue', data.id)
}

async function createFolder(parentId) {
  try {
    const { value: name } = await ElMessageBox.prompt('文件夹名称', '新建文件夹', {
      inputPattern: /\S+/,
      inputErrorMessage: '名称不能为空'
    })
    await apiCreateFolder({ name, parent_id: parentId || null })
    ElMessage.success('已新建')
    await load()
    emit('changed')
  } catch (e) {
    // cancel or failure
  }
}

async function renameFolder(data) {
  try {
    const { value: name } = await ElMessageBox.prompt('新名称', '重命名', {
      inputValue: data.name,
      inputPattern: /\S+/,
      inputErrorMessage: '名称不能为空'
    })
    await apiUpdateFolder(data.id, { name })
    ElMessage.success('已重命名')
    await load()
  } catch (e) {
    // cancel
  }
}

async function removeFolder(data) {
  try {
    await ElMessageBox.confirm('删除文件夹「' + data.name + '」？文件夹内文档不会被删除。', '删除文件夹', {
      type: 'warning'
    })
  } catch (e) {
    return
  }
  await apiDeleteFolder(data.id)
  ElMessage.success('已删除')
  if (props.modelValue === data.id) emit('update:modelValue', null)
  await load()
  emit('changed')
}

onMounted(load)
// expose refresh() to parent
defineExpose({ reload: load })
</script>

<style scoped>
.folder-tree {
  width: 200px;
  flex-shrink: 0;
  background: #fff;
  border: 1px solid #e4e7ed;
  padding: 8px;
  overflow: auto;
}
.ft-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}
.ft-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}
.ft-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 6px;
  font-size: 13px;
  cursor: pointer;
  border-radius: 2px;
  color: #606266;
}
.ft-item:hover {
  background: #ecf5ff;
}
.ft-item.active {
  background: #ecf5ff;
  color: #409eff;
}
.ft-node {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}
.ft-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.ft-ops {
  display: none;
  gap: 4px;
  color: #909399;
}
.ft-node:hover .ft-ops {
  display: inline-flex;
}
.ft-ops .el-icon {
  cursor: pointer;
  font-size: 13px;
}
.ft-ops .el-icon:hover {
  color: #409eff;
}
</style>
