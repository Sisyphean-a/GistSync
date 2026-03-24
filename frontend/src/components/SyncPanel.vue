<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  applySnapshot,
  chooseFilesForProfile,
  createProfile,
  deleteProfile,
  listSnapshots,
  loadSettings,
  previewApplyConflicts,
  pullProfilesFromCloud,
  removeProfileItems,
  saveSettings,
  setActiveProfile,
  uploadProfile,
  type ApplySnapshotRequest,
  type Profile,
  type SettingsData,
  type SnapshotMeta,
} from '../lib/backend'

const state = ref<SettingsData | null>(null)
const profileName = ref('')
const selectedItemIds = ref<string[]>([])
const snapshots = ref<SnapshotMeta[]>([])
const selectedSnapshotId = ref('')
const status = ref('')
const restore = reactive({
  mode: 'original' as 'original' | 'rooted',
  root: '',
})

const activeProfile = computed<Profile | null>(() => {
  const current = state.value
  if (!current) {
    return null
  }
  return current.profiles.find((profile) => profile.id === current.activeProfileId) ?? null
})

onMounted(async () => {
  await refreshState()
})

async function refreshState(): Promise<void> {
  const settings = await loadSettings()
  state.value = settings
  selectedItemIds.value = []
  const profile = activeProfile.value
  if (profile) {
    restore.mode = profile.restoreMode
    restore.root = profile.restoreRoot
    await refreshSnapshots(profile.id)
  } else {
    snapshots.value = []
    selectedSnapshotId.value = ''
  }
}

async function refreshSnapshots(profileId: string): Promise<void> {
  try {
    snapshots.value = await listSnapshots(profileId)
    selectedSnapshotId.value = snapshots.value[0]?.id ?? ''
  } catch (error) {
    snapshots.value = []
    selectedSnapshotId.value = ''
    status.value = `快照加载失败: ${String(error)}`
  }
}

async function createNewProfile(): Promise<void> {
  try {
    const created = await createProfile(profileName.value)
    profileName.value = ''
    await setActiveProfile(created.id)
    await refreshState()
    status.value = `已创建配置: ${created.name || created.id}`
  } catch (error) {
    status.value = `创建失败: ${String(error)}`
  }
}

async function pullFromCloud(): Promise<void> {
  try {
    const count = await pullProfilesFromCloud()
    await refreshState()
    status.value = `已从云端同步配置集，本地共 ${count} 个`
  } catch (error) {
    status.value = `从云端拉取失败: ${String(error)}`
  }
}

async function switchProfile(profileId: string): Promise<void> {
  try {
    await setActiveProfile(profileId)
    await refreshState()
    status.value = ''
  } catch (error) {
    status.value = `切换失败: ${String(error)}`
  }
}

async function removeProfile(profileId: string): Promise<void> {
  try {
    await deleteProfile(profileId)
    await refreshState()
    status.value = '配置已删除'
  } catch (error) {
    status.value = `删除失败: ${String(error)}`
  }
}

async function addFiles(): Promise<void> {
  if (!activeProfile.value) {
    status.value = '请先创建或选择配置'
    return
  }
  try {
    const selected = await chooseFilesForProfile(activeProfile.value.id)
    await refreshState()
    status.value = selected.length === 0 ? '未选择文件' : `已添加 ${selected.length} 个文件`
  } catch (error) {
    status.value = `添加文件失败: ${String(error)}`
  }
}

async function removeSelectedItems(): Promise<void> {
  if (!activeProfile.value || selectedItemIds.value.length === 0) {
    return
  }
  try {
    await removeProfileItems(activeProfile.value.id, selectedItemIds.value)
    await refreshState()
    status.value = '已移除选中文件'
  } catch (error) {
    status.value = `移除失败: ${String(error)}`
  }
}

async function pushProfile(): Promise<void> {
  if (!activeProfile.value) {
    status.value = '请先选择配置'
    return
  }
  try {
    const result = await uploadProfile(activeProfile.value.id)
    await refreshSnapshots(activeProfile.value.id)
    status.value = `上传完成：快照 ${result.snapshotId}，文件 ${result.uploaded} 个`
  } catch (error) {
    status.value = `上传失败: ${String(error)}`
  }
}

function collectOverwriteItemIds(conflicts: { itemId: string; targetPath: string }[]): string[] {
  const overwriteIds: string[] = []
  let overwriteAll = false
  let skipAll = false

  for (const conflict of conflicts) {
    if (overwriteAll) {
      overwriteIds.push(conflict.itemId)
      continue
    }
    if (skipAll) {
      continue
    }
    const choice = window.prompt(
      `目标已存在：${conflict.targetPath}\n输入 y 覆盖，n 跳过，a 全部覆盖，s 全部跳过`,
      'y',
    )
    if (choice === 'a') {
      overwriteAll = true
      overwriteIds.push(conflict.itemId)
      continue
    }
    if (choice === 's') {
      skipAll = true
      continue
    }
    if (choice === 'y') {
      overwriteIds.push(conflict.itemId)
    }
  }
  return overwriteIds
}

async function applySelectedSnapshot(): Promise<void> {
  if (!activeProfile.value) {
    status.value = '请先选择配置'
    return
  }
  const request: ApplySnapshotRequest = {
    profileId: activeProfile.value.id,
    snapshotId: selectedSnapshotId.value,
    masterPassword: '',
    restoreMode: restore.mode,
    restoreRoot: restore.root,
    overwriteItemIds: [],
  }

  try {
    const conflicts = await previewApplyConflicts(request)
    request.overwriteItemIds = collectOverwriteItemIds(conflicts)
    const result = await applySnapshot(request)
    status.value = `应用完成：成功 ${result.applied}，跳过 ${result.skipped}`
  } catch (error) {
    status.value = `应用失败: ${String(error)}`
  }
}

async function persistRestoreSettings(): Promise<void> {
  if (!state.value || !activeProfile.value) {
    return
  }
  const profile = state.value.profiles.find((item) => item.id === activeProfile.value?.id)
  if (!profile) {
    return
  }
  profile.restoreMode = restore.mode
  profile.restoreRoot = restore.root
  try {
    await saveSettings(state.value)
  } catch (error) {
    status.value = `保存恢复策略失败: ${String(error)}`
  }
}
</script>

<template>
  <section class="grid gap-5 lg:grid-cols-[280px_1fr]">
    <aside class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <h3 class="mb-3 text-sm font-semibold text-slate-800">配置集</h3>
      <div class="mb-3 flex items-center gap-2">
        <input
          v-model="profileName"
          class="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
          placeholder="新配置名（可空）"
        >
        <button class="h-10 rounded-lg bg-slate-900 px-4 text-sm font-medium text-white hover:bg-slate-700" @click="createNewProfile">
          新增
        </button>
      </div>
      <button class="mb-3 w-full rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50" @click="pullFromCloud">
        从云端拉取配置
      </button>

      <div class="space-y-2">
        <button
          v-for="profile in state?.profiles || []"
          :key="profile.id"
          class="w-full rounded-xl border px-3 py-2 text-left text-sm transition"
          :class="profile.id === state?.activeProfileId ? 'border-slate-900 bg-slate-100 text-slate-900' : 'border-slate-200 text-slate-700 hover:bg-slate-50'"
          @click="switchProfile(profile.id)"
        >
          <div class="flex items-center justify-between">
            <span class="truncate font-medium">{{ profile.name }}</span>
            <span class="rounded-full bg-slate-200 px-2 py-0.5 text-xs text-slate-700">{{ profile.items.length }}</span>
          </div>
        </button>
      </div>

      <button
        v-if="activeProfile"
        class="mt-4 w-full rounded-lg bg-rose-700 px-4 py-2 text-sm font-medium text-white hover:bg-rose-600"
        @click="removeProfile(activeProfile.id)"
      >
        删除当前配置
      </button>
    </aside>

    <div class="space-y-4">
      <section class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <h3 class="mb-4 text-base font-semibold text-slate-900">文件条目</h3>
        <div class="mb-4 flex flex-wrap gap-2">
          <button class="rounded-lg bg-indigo-700 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-600" @click="addFiles">
            添加文件（多选）
          </button>
          <button class="rounded-lg bg-slate-800 px-4 py-2 text-sm font-medium text-white hover:bg-slate-700" @click="removeSelectedItems">
            移除选中
          </button>
          <button class="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600" @click="pushProfile">
            上传当前配置
          </button>
        </div>

        <div class="overflow-auto rounded-xl border border-slate-200">
          <table class="w-full min-w-[680px] text-left text-sm">
            <thead class="bg-slate-50 text-slate-600">
              <tr>
                <th class="w-16 px-3 py-2">选中</th>
                <th class="px-3 py-2">原始路径</th>
                <th class="px-3 py-2">相对层级</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in activeProfile?.items || []" :key="item.id" class="border-t border-slate-200">
                <td class="px-3 py-2"><input v-model="selectedItemIds" :value="item.id" type="checkbox"></td>
                <td class="px-3 py-2 font-mono text-xs text-slate-700">{{ item.sourcePathTemplate }}</td>
                <td class="px-3 py-2 font-mono text-xs text-slate-700">{{ item.relativePath }}</td>
              </tr>
              <tr v-if="(activeProfile?.items?.length || 0) === 0">
                <td colspan="3" class="px-3 py-8 text-center text-sm text-slate-500">当前配置还没有文件，点击“添加文件（多选）”开始。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <h3 class="mb-4 text-base font-semibold text-slate-900">快照应用</h3>
        <div class="grid gap-3 md:grid-cols-[180px_1fr_auto]">
          <select v-model="restore.mode" class="h-10 rounded-lg border border-slate-300 px-3 text-sm" @change="persistRestoreSettings">
            <option value="original">原始路径恢复</option>
            <option value="rooted">指定根目录恢复</option>
          </select>
          <input
            v-model="restore.root"
            class="h-10 rounded-lg border border-slate-300 px-3 text-sm"
            placeholder="rooted 模式下填写恢复根目录"
            @blur="persistRestoreSettings"
          >
          <button class="h-10 rounded-lg bg-amber-700 px-4 text-sm font-medium text-white hover:bg-amber-600" @click="applySelectedSnapshot">
            应用快照
          </button>
        </div>

        <div class="mt-3">
          <label class="mb-1 block text-xs font-medium text-slate-500">选择快照</label>
          <select v-model="selectedSnapshotId" class="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm">
            <option value="" disabled>请选择快照（默认最新）</option>
            <option v-for="snapshot in snapshots" :key="snapshot.id" :value="snapshot.id">
              {{ snapshot.createdAt }} - {{ snapshot.id }}
            </option>
          </select>
        </div>
      </section>

      <p v-if="status" class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">{{ status }}</p>
    </div>
  </section>
</template>
