<script setup lang="ts">
import { ref } from 'vue'
import SettingsPanel from './components/SettingsPanel.vue'
import SyncPanel from './components/SyncPanel.vue'

type TabKey = 'home' | 'settings' | 'sync'

const activeTab = ref<TabKey>('home')
</script>

<template>
  <div class="min-h-screen bg-slate-100/70 p-4 md:p-8">
    <div class="mx-auto w-full max-w-6xl rounded-3xl border border-slate-200 bg-white/95 p-6 shadow-xl md:p-8">
      <header class="mb-8 flex items-center justify-between gap-4 border-b border-slate-200 pb-5">
        <div>
          <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">GistSync</p>
          <h1 class="text-2xl font-bold text-slate-900">配置文件云端同步工具</h1>
        </div>
      </header>

      <nav class="mb-8 flex flex-wrap gap-3">
        <button
          class="rounded-xl px-5 py-2 text-sm font-semibold transition"
          :class="activeTab === 'home' ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
          @click="activeTab = 'home'"
        >
          首页
        </button>
        <button
          class="rounded-xl px-5 py-2 text-sm font-semibold transition"
          :class="activeTab === 'settings' ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
          @click="activeTab = 'settings'"
        >
          设置
        </button>
        <button
          class="rounded-xl px-5 py-2 text-sm font-semibold transition"
          :class="activeTab === 'sync' ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
          @click="activeTab = 'sync'"
        >
          同步
        </button>
      </nav>

      <section v-if="activeTab === 'home'" class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-10 text-left">
        <h2 class="mb-2 text-lg font-semibold text-slate-900">欢迎使用 GistSync</h2>
        <p class="text-sm text-slate-600">
          在“设置”中保存 Token 和主密码，在“同步”中配置路径并执行上传/下载。
        </p>
      </section>

      <SettingsPanel v-else-if="activeTab === 'settings'" />
      <SyncPanel v-else />
    </div>
  </div>
</template>
