import { computed, ref } from 'vue'
import {
  quickDownload,
  quickUpload,
  type ApplyConflict,
  type QuickOperationResult,
} from '../lib/backend'
import { useSettingsStore } from './useSettingsStore'

type BusyAction = '' | 'download' | 'upload'

const settingsStore = useSettingsStore()
const busyAction = ref<BusyAction>('')
const status = ref('')
const lastResult = ref<QuickOperationResult | null>(null)
const showResultDetails = ref(false)
const conflicts = ref<ApplyConflict[]>([])
const conflictVisible = ref(false)

const selectedProfileId = computed(() => settingsStore.state.value?.activeProfileId ?? '')
const profiles = computed(() => settingsStore.state.value?.profiles ?? [])

async function initialize(): Promise<void> {
  await settingsStore.ensureLoaded()
}

async function switchProfile(profileId: string): Promise<void> {
  await settingsStore.switchActiveProfile(profileId)
}

async function upload(): Promise<void> {
  if (!selectedProfileId.value) {
    status.value = '请先选择配置集'
    return
  }
  busyAction.value = 'upload'
  try {
    const result = await quickUpload({ profileId: selectedProfileId.value })
    lastResult.value = result
    showResultDetails.value = false
    status.value = `上传完成：快照 ${result.snapshotId || '-'}，文件 ${result.summary.uploaded} 个`
  } catch (error) {
    status.value = `上传失败: ${String(error)}`
  } finally {
    busyAction.value = ''
  }
}

async function download(): Promise<void> {
  if (!selectedProfileId.value) {
    status.value = '请先选择配置集'
    return
  }
  busyAction.value = 'download'
  try {
    const result = await quickDownload({
      profileId: selectedProfileId.value,
      conflictPolicy: 'manual',
      overwriteItemIds: [],
    })
    if (result.requiresConflictResolution) {
      conflicts.value = result.conflicts.map((item) => ({ itemId: item.itemId, targetPath: item.targetPath }))
      conflictVisible.value = true
      status.value = `检测到 ${result.summary.conflicts} 个冲突，默认将全部覆盖，可按需修改`
      return
    }
    lastResult.value = result
    showResultDetails.value = false
    status.value = `下载完成：应用 ${result.summary.applied}，跳过 ${result.summary.skipped}`
  } catch (error) {
    status.value = `下载失败: ${String(error)}`
  } finally {
    busyAction.value = ''
  }
}

async function submitConflictDecision(overwriteItemIds: string[]): Promise<void> {
  if (!selectedProfileId.value) {
    return
  }
  busyAction.value = 'download'
  try {
    const result = await quickDownload({
      profileId: selectedProfileId.value,
      conflictPolicy: 'manual',
      overwriteItemIds,
    })
    conflictVisible.value = false
    conflicts.value = []
    lastResult.value = result
    showResultDetails.value = false
    status.value = `下载完成：应用 ${result.summary.applied}，跳过 ${result.summary.skipped}`
  } catch (error) {
    status.value = `下载失败: ${String(error)}`
  } finally {
    busyAction.value = ''
  }
}

function closeConflictDialog(): void {
  conflictVisible.value = false
  conflicts.value = []
}

function toggleDetails(): void {
  showResultDetails.value = !showResultDetails.value
}

export function useQuickSyncStore() {
  return {
    selectedProfileId,
    profiles,
    busyAction,
    status,
    lastResult,
    showResultDetails,
    conflicts,
    conflictVisible,
    initialize,
    switchProfile,
    upload,
    download,
    submitConflictDecision,
    closeConflictDialog,
    toggleDetails,
  }
}
