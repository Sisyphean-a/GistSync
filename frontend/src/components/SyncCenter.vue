<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import ConflictResolverDialog from './ConflictResolverDialog.vue'
import {
  applySnapshot,
  listSnapshots,
  loadSettings,
  previewApplyConflicts,
  setActiveProfile,
  uploadProfile,
  type ApplyConflict,
  type ApplySnapshotRequest,
  type Profile,
  type SettingsData,
  type SnapshotMeta,
} from '../lib/backend'

const state = ref<SettingsData | null>(null)
const snapshots = ref<SnapshotMeta[]>([])
const selectedSnapshotId = ref('')
const selectedUploadItemIds = ref<string[]>([])
const selectedRestoreItemIds = ref<string[]>([])
const status = ref('')
const loading = ref(false)
const conflictVisible = ref(false)
const conflicts = ref<ApplyConflict[]>([])
const pendingApplyRequest = ref<ApplySnapshotRequest | null>(null)

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
  const profile = activeProfile.value
  if (!profile) {
    snapshots.value = []
    selectedSnapshotId.value = ''
    selectedUploadItemIds.value = []
    selectedRestoreItemIds.value = []
    return
  }
  selectedUploadItemIds.value = profile.items.map((item) => item.id)
  selectedRestoreItemIds.value = profile.items.map((item) => item.id)
  await refreshSnapshots(profile.id)
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

async function switchProfile(profileId: string): Promise<void> {
  try {
    await setActiveProfile(profileId)
    await refreshState()
    status.value = ''
  } catch (error) {
    status.value = `切换失败: ${String(error)}`
  }
}

async function uploadSelectedItems(): Promise<void> {
  if (!activeProfile.value) {
    status.value = '请先选择配置集'
    return
  }
  if (selectedUploadItemIds.value.length === 0) {
    status.value = '请至少选择一个上传条目'
    return
  }
  loading.value = true
  try {
    const result = await uploadProfile(activeProfile.value.id, selectedUploadItemIds.value)
    await refreshSnapshots(activeProfile.value.id)
    status.value = `上传完成：快照 ${result.snapshotId}，文件 ${result.uploaded} 个`
  } catch (error) {
    status.value = `上传失败: ${String(error)}`
  } finally {
    loading.value = false
  }
}

async function startApplyFlow(): Promise<void> {
  if (!activeProfile.value) {
    status.value = '请先选择配置集'
    return
  }
  if (selectedRestoreItemIds.value.length === 0) {
    status.value = '请至少选择一个恢复条目'
    return
  }
  const request: ApplySnapshotRequest = {
    profileId: activeProfile.value.id,
    snapshotId: selectedSnapshotId.value,
    masterPassword: '',
    restoreMode: activeProfile.value.restoreMode,
    restoreRoot: activeProfile.value.restoreRoot,
    selectedItemIds: selectedRestoreItemIds.value,
    overwriteItemIds: [],
  }

  loading.value = true
  try {
    const found = await previewApplyConflicts(request)
    if (found.length === 0) {
      const result = await applySnapshot(request)
      status.value = `应用完成：成功 ${result.applied}，跳过 ${result.skipped}`
      return
    }
    conflicts.value = found
    pendingApplyRequest.value = request
    conflictVisible.value = true
    status.value = `检测到 ${found.length} 个冲突，请确认覆盖策略`
  } catch (error) {
    status.value = `应用失败: ${String(error)}`
  } finally {
    loading.value = false
  }
}

async function submitConflictDecision(overwriteItemIds: string[]): Promise<void> {
  const request = pendingApplyRequest.value
  conflictVisible.value = false
  if (!request) {
    return
  }
  loading.value = true
  try {
    const result = await applySnapshot({
      ...request,
      overwriteItemIds,
    })
    status.value = `应用完成：成功 ${result.applied}，跳过 ${result.skipped}`
  } catch (error) {
    status.value = `应用失败: ${String(error)}`
  } finally {
    loading.value = false
    pendingApplyRequest.value = null
    conflicts.value = []
  }
}

function closeConflictDialog(): void {
  conflictVisible.value = false
  pendingApplyRequest.value = null
}
</script>

<template>
  <section class="space-y-4">
    <article class="rounded-xl border border-slate-200 bg-white p-4">
      <div class="mb-4 border-b border-slate-200 pb-3">
        <h2 class="text-base font-semibold text-slate-900">同步中心</h2>
        <p class="mt-1 text-xs text-slate-500">选择配置和快照后，可按条目执行上传或恢复。</p>
      </div>
      <div class="grid gap-3 xl:grid-cols-[280px_1fr]">
        <label class="text-sm text-slate-700">
          <span class="mb-1 block font-medium">当前配置集</span>
          <select
            :value="state?.activeProfileId || ''"
            class="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm"
            @change="switchProfile(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="profile in state?.profiles || []" :key="profile.id" :value="profile.id">{{ profile.name }}</option>
          </select>
        </label>
        <label class="text-sm text-slate-700">
          <span class="mb-1 block font-medium">快照选择</span>
          <select v-model="selectedSnapshotId" class="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm">
            <option value="" disabled>请选择快照（默认最新）</option>
            <option v-for="snapshot in snapshots" :key="snapshot.id" :value="snapshot.id">
              {{ snapshot.createdAt }} - {{ snapshot.id }}
            </option>
          </select>
        </label>
      </div>
    </article>

    <section class="grid gap-4 xl:grid-cols-2">
      <article class="rounded-xl border border-slate-200 bg-white p-4">
        <h3 class="mb-3 text-sm font-semibold text-slate-900">上传到云端（可选择条目）</h3>
        <div class="mb-3 max-h-[320px] overflow-auto rounded-lg border border-slate-200 p-2">
          <div class="grid gap-2">
            <label v-for="item in activeProfile?.items || []" :key="`upload-${item.id}`" class="flex items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm">
              <input v-model="selectedUploadItemIds" :value="item.id" type="checkbox">
              <span class="truncate">{{ item.relativePath || item.sourcePathTemplate }}</span>
            </label>
          </div>
        </div>
        <button class="h-10 min-w-[136px] rounded-lg bg-emerald-700 px-4 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-60" :disabled="loading" @click="uploadSelectedItems">
          上传选中条目
        </button>
      </article>

      <article class="rounded-xl border border-slate-200 bg-white p-4">
        <h3 class="mb-3 text-sm font-semibold text-slate-900">从云端恢复（可选择条目）</h3>
        <div class="mb-3 max-h-[320px] overflow-auto rounded-lg border border-slate-200 p-2">
          <div class="grid gap-2">
            <label v-for="item in activeProfile?.items || []" :key="`restore-${item.id}`" class="flex items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm">
              <input v-model="selectedRestoreItemIds" :value="item.id" type="checkbox">
              <span class="truncate">{{ item.relativePath || item.sourcePathTemplate }}</span>
            </label>
          </div>
        </div>
        <button class="h-10 min-w-[172px] rounded-lg bg-amber-700 px-4 text-sm font-medium text-white hover:bg-amber-600 disabled:opacity-60" :disabled="loading" @click="startApplyFlow">
          预检冲突并应用快照
        </button>
      </article>
    </section>

    <p v-if="status" class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">{{ status }}</p>

    <ConflictResolverDialog
      :visible="conflictVisible"
      :conflicts="conflicts"
      @close="closeConflictDialog"
      @confirm="submitConflictDecision"
    />
  </section>
</template>
