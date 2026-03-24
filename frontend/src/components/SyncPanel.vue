<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { downloadSync, loadSettings, saveSettings, uploadSync, type SettingsData } from '../lib/backend'

const form = reactive<SettingsData>({
  token: '',
  masterPassword: '',
  syncPath: '',
})
const status = ref('')

onMounted(async () => {
  try {
    const saved = await loadSettings()
    form.token = saved.token ?? ''
    form.masterPassword = saved.masterPassword ?? ''
    form.syncPath = saved.syncPath ?? ''
  } catch (error) {
    status.value = `加载同步配置失败: ${String(error)}`
  }
})

async function saveSyncPathOnly(): Promise<void> {
  try {
    await saveSettings({
      token: form.token,
      masterPassword: form.masterPassword,
      syncPath: form.syncPath,
    })
    status.value = '同步路径已保存'
  } catch (error) {
    status.value = `保存同步路径失败: ${String(error)}`
  }
}

async function handleUpload(): Promise<void> {
  await saveSyncPathOnly()
  try {
    const result = await uploadSync()
    status.value = result
  } catch (error) {
    status.value = `上传失败: ${String(error)}`
  }
}

async function handleDownload(): Promise<void> {
  await saveSyncPathOnly()
  try {
    const result = await downloadSync(false)
    status.value = result
  } catch (error) {
    const message = String(error)
    if (message.includes('OVERWRITE_REQUIRED')) {
      const confirmed = window.confirm('文件已存在，是否覆盖？')
      if (!confirmed) {
        status.value = '已取消覆盖下载'
        return
      }
      try {
        const result = await downloadSync(true)
        status.value = result
      } catch (retryError) {
        status.value = `覆盖下载失败: ${String(retryError)}`
      }
      return
    }
    status.value = `下载失败: ${message}`
  }
}
</script>

<template>
  <section class="rounded-xl border border-slate-200 bg-slate-50 p-5 text-left">
    <h2 class="mb-4 text-lg font-semibold text-slate-900">同步</h2>

    <div class="space-y-4">
      <label class="block">
        <span class="mb-1 block text-sm font-medium text-slate-700">同步文件路径</span>
        <input
          v-model="form.syncPath"
          type="text"
          placeholder="例如：{{HOME}}/Desktop/test-config.txt"
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-300"
        >
      </label>

      <div class="flex flex-wrap gap-2">
        <button
          type="button"
          class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-700"
          @click="saveSyncPathOnly"
        >
          保存路径
        </button>
        <button
          type="button"
          class="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-medium text-white transition hover:bg-emerald-600"
          @click="handleUpload"
        >
          上传同步
        </button>
        <button
          type="button"
          class="rounded-lg bg-amber-700 px-4 py-2 text-sm font-medium text-white transition hover:bg-amber-600"
          @click="handleDownload"
        >
          从云端下载同步
        </button>
      </div>

      <p v-if="status" class="text-sm text-slate-600">{{ status }}</p>
    </div>
  </section>
</template>