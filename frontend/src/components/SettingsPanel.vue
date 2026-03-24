<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { loadSettings, saveSettings, type SettingsData } from '../lib/backend'

const settingsState = ref<SettingsData | null>(null)
const form = reactive({
  token: '',
  masterPassword: '',
})
const status = ref('')

onMounted(async () => {
  try {
    const saved = await loadSettings()
    settingsState.value = saved
    form.token = saved.token ?? ''
    form.masterPassword = saved.masterPassword ?? ''
  } catch (error) {
    status.value = `加载设置失败: ${String(error)}`
  }
})

async function persistSettings(): Promise<void> {
  try {
    const current = settingsState.value ?? {
      token: '',
      masterPassword: '',
      activeProfileId: '',
      profiles: [],
    }
    const next: SettingsData = {
      ...current,
      token: form.token,
      masterPassword: form.masterPassword,
    }
    await saveSettings(next)
    settingsState.value = next
    status.value = '设置保存成功'
  } catch (error) {
    status.value = `保存失败: ${String(error)}`
  }
}
</script>

<template>
  <section class="rounded-xl border border-slate-200 bg-slate-50 p-5 text-left">
    <h2 class="mb-4 text-lg font-semibold text-slate-900">设置</h2>

    <div class="space-y-4">
      <label class="block">
        <span class="mb-1 block text-sm font-medium text-slate-700">GitHub Token</span>
        <input
          v-model="form.token"
          type="password"
          placeholder="请输入 GitHub Personal Access Token"
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-300"
        >
      </label>

      <label class="block">
        <span class="mb-1 block text-sm font-medium text-slate-700">Master Password</span>
        <input
          v-model="form.masterPassword"
          type="password"
          placeholder="请输入主密码"
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-300"
        >
      </label>

      <button
        type="button"
        class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-700"
        @click="persistSettings"
      >
        保存
      </button>

      <p v-if="status" class="text-sm text-slate-600">{{ status }}</p>
    </div>
  </section>
</template>