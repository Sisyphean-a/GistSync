<script setup lang="ts">
import { reactive } from 'vue'

interface SettingsForm {
  token: string
  masterPassword: string
}

const form = reactive<SettingsForm>({
  token: '',
  masterPassword: '',
})

function saveSettings(): void {
  const payload = {
    token: form.token,
    masterPassword: form.masterPassword,
  }
  console.log('[Mock Save] Settings payload:', payload)
  window.alert(`已模拟保存\nToken: ${payload.token}\nMaster Password: ${payload.masterPassword}`)
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
        @click="saveSettings"
      >
        保存
      </button>
    </div>
  </section>
</template>