<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import ConflictResolverDialog from './ConflictResolverDialog.vue'
import {
  applySnapshot,
  listSnapshots,
  previewApplyConflicts,
  uploadProfile,
  type ApplyConflict,
  type ApplySnapshotRequest,
  type SnapshotMeta,
} from '../lib/backend'
import {
  advancedDownloadButtonLabel,
  advancedUploadButtonLabel,
  describeSyncActivity,
  isBusyActivity,
  type SyncActivity,
} from '../lib/syncActivity'
import { useSettingsStore } from '../composables/useSettingsStore'

const store = useSettingsStore()
const snapshots = ref<SnapshotMeta[]>([])
const selectedSnapshotId = ref('')
const selectedUploadItemIds = ref<string[]>([])
const selectedRestoreItemIds = ref<string[]>([])
const status = ref('')
const activity = ref<SyncActivity>('')
const conflictVisible = ref(false)
const conflicts = ref<ApplyConflict[]>([])
const pendingApplyRequest = ref<ApplySnapshotRequest | null>(null)
const busy = computed(() => isBusyActivity(activity.value))
const busyDescription = computed(() => describeSyncActivity(activity.value))
const uploadButtonLabel = computed(() => advancedUploadButtonLabel(activity.value))
const downloadButtonLabel = computed(() => advancedDownloadButtonLabel(activity.value))

onMounted(async () => {
  activity.value = 'loading_snapshots'
  const pendingStatus = describeSyncActivity(activity.value)
  status.value = pendingStatus
  try {
    await refreshState()
    if (status.value === pendingStatus) {
      status.value = ''
    }
  } finally {
    activity.value = ''
  }
})

async function refreshState(): Promise<void> {
  await store.refresh()
  const profile = store.activeProfile.value
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
  activity.value = 'switching_profile'
  status.value = describeSyncActivity(activity.value)
  try {
    await store.switchActiveProfile(profileId)
    activity.value = 'loading_snapshots'
    const pendingStatus = describeSyncActivity(activity.value)
    status.value = pendingStatus
    await refreshState()
    if (status.value === pendingStatus) {
      status.value = ''
    }
  } catch (error) {
    status.value = `切换失败: ${String(error)}`
  } finally {
    activity.value = ''
  }
}

async function uploadSelectedItems(): Promise<void> {
  const profile = store.activeProfile.value
  if (!profile) {
    status.value = '请先选择配置集'
    return
  }
  if (selectedUploadItemIds.value.length === 0) {
    status.value = '请至少选择一个上传条目'
    return
  }
  activity.value = 'uploading'
  status.value = describeSyncActivity(activity.value)
  try {
    const result = await uploadProfile(profile.id, selectedUploadItemIds.value)
    activity.value = 'loading_snapshots'
    status.value = describeSyncActivity(activity.value)
    await refreshSnapshots(profile.id)
    status.value = `上传完成：快照 ${result.snapshotId}，文件 ${result.uploaded} 个`
  } catch (error) {
    status.value = `上传失败: ${String(error)}`
  } finally {
    activity.value = ''
  }
}

async function startApplyFlow(): Promise<void> {
  const profile = store.activeProfile.value
  if (!profile) {
    status.value = '请先选择配置集'
    return
  }
  if (selectedRestoreItemIds.value.length === 0) {
    status.value = '请至少选择一个恢复条目'
    return
  }
  const request: ApplySnapshotRequest = {
    profileId: profile.id,
    snapshotId: selectedSnapshotId.value,
    masterPassword: '',
    restoreMode: profile.restoreMode,
    restoreRoot: profile.restoreRoot,
    selectedItemIds: selectedRestoreItemIds.value,
    overwriteItemIds: [],
  }

  activity.value = 'checking_conflicts'
  status.value = describeSyncActivity(activity.value)
  try {
    const found = await previewApplyConflicts(request)
    if (found.length === 0) {
      activity.value = 'applying_snapshot'
      status.value = describeSyncActivity(activity.value)
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
    activity.value = ''
  }
}

async function submitConflictDecision(overwriteItemIds: string[]): Promise<void> {
  const request = pendingApplyRequest.value
  if (!request) {
    return
  }
  activity.value = 'applying_snapshot'
  status.value = describeSyncActivity(activity.value)
  try {
    const result = await applySnapshot({
      ...request,
      overwriteItemIds,
    })
    conflictVisible.value = false
    pendingApplyRequest.value = null
    conflicts.value = []
    status.value = `应用完成：成功 ${result.applied}，跳过 ${result.skipped}`
  } catch (error) {
    status.value = `应用失败: ${String(error)}`
  } finally {
    activity.value = ''
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
            :value="store.state.value?.activeProfileId || ''"
            class="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm"
            :disabled="busy"
            @change="switchProfile(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="profile in store.state.value?.profiles || []" :key="profile.id" :value="profile.id">{{ profile.name }}</option>
          </select>
        </label>
        <label class="text-sm text-slate-700">
          <span class="mb-1 block font-medium">快照选择</span>
          <select v-model="selectedSnapshotId" class="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm" :disabled="busy">
            <option value="" disabled>请选择快照（默认最新）</option>
            <option v-for="snapshot in snapshots" :key="snapshot.id" :value="snapshot.id">
              {{ snapshot.createdAt }} - {{ snapshot.id }}
            </option>
          </select>
        </label>
      </div>
    </article>

    <article v-if="busy" class="rounded-xl border border-sky-200 bg-sky-50 p-4">
      <div class="flex items-center gap-3">
        <span class="h-4 w-4 animate-spin rounded-full border-2 border-sky-200 border-t-sky-700" />
        <div>
          <p class="text-sm font-semibold text-sky-900">{{ busyDescription }}</p>
          <p class="text-xs text-sky-700">同步过程中会暂时锁定当前页面操作。</p>
        </div>
      </div>
    </article>

    <section class="grid gap-4 xl:grid-cols-2">
      <article class="rounded-xl border border-slate-200 bg-white p-4">
        <h3 class="mb-3 text-sm font-semibold text-slate-900">上传到云端（可选择条目）</h3>
        <div class="mb-3 max-h-[320px] overflow-auto rounded-lg border border-slate-200 p-2">
          <div class="grid gap-2">
            <label v-for="item in store.activeProfile.value?.items || []" :key="`upload-${item.id}`" class="flex items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm">
              <input v-model="selectedUploadItemIds" :disabled="busy" :value="item.id" type="checkbox">
              <span class="truncate">{{ item.relativePath || item.sourcePathTemplate }}</span>
            </label>
          </div>
        </div>
        <button class="h-10 min-w-[136px] rounded-lg bg-emerald-700 px-4 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-60" :disabled="busy" @click="uploadSelectedItems">
          <span class="inline-flex items-center gap-2">
            <span v-if="activity === 'uploading'" class="h-4 w-4 animate-spin rounded-full border-2 border-emerald-200 border-t-white" />
            <span>{{ uploadButtonLabel }}</span>
          </span>
        </button>
      </article>

      <article class="rounded-xl border border-slate-200 bg-white p-4">
        <h3 class="mb-3 text-sm font-semibold text-slate-900">从云端恢复（可选择条目）</h3>
        <div class="mb-3 max-h-[320px] overflow-auto rounded-lg border border-slate-200 p-2">
          <div class="grid gap-2">
            <label v-for="item in store.activeProfile.value?.items || []" :key="`restore-${item.id}`" class="flex items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm">
              <input v-model="selectedRestoreItemIds" :disabled="busy" :value="item.id" type="checkbox">
              <span class="truncate">{{ item.relativePath || item.sourcePathTemplate }}</span>
            </label>
          </div>
        </div>
        <button class="h-10 min-w-[172px] rounded-lg bg-amber-700 px-4 text-sm font-medium text-white hover:bg-amber-600 disabled:opacity-60" :disabled="busy" @click="startApplyFlow">
          <span class="inline-flex items-center gap-2">
            <span v-if="activity === 'checking_conflicts' || activity === 'applying_snapshot'" class="h-4 w-4 animate-spin rounded-full border-2 border-amber-200 border-t-white" />
            <span>{{ downloadButtonLabel }}</span>
          </span>
        </button>
      </article>
    </section>

    <p v-if="status" class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">{{ status }}</p>

    <ConflictResolverDialog
      :visible="conflictVisible"
      :conflicts="conflicts"
      :submitting="activity === 'applying_snapshot'"
      @close="closeConflictDialog"
      @confirm="submitConflictDecision"
    />
  </section>
</template>
