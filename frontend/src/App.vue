<script setup lang="ts">
import { ref } from 'vue'
import SettingsPanel from './components/SettingsPanel.vue'
import SyncCenter from './components/SyncCenter.vue'
import ProfileManager from './components/ProfileManager.vue'

type TabKey = 'sync' | 'profiles' | 'settings'

const activeTab = ref<TabKey>('sync')
</script>

<template>
  <div class="app-shell min-h-screen p-4 md:p-8">
    <div class="mx-auto grid w-full max-w-7xl gap-5 lg:grid-cols-[240px_1fr]">
      <aside class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <div class="mb-5 border-b border-slate-200 pb-4">
          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">GistSync</p>
          <h1 class="mt-1 text-xl font-bold text-slate-900">配置同步控制台</h1>
        </div>
        <nav class="space-y-2">
          <button
            class="w-full rounded-xl px-4 py-2 text-left text-sm font-semibold transition"
            :class="activeTab === 'sync' ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            @click="activeTab = 'sync'"
          >
            同步中心
          </button>
          <button
            class="w-full rounded-xl px-4 py-2 text-left text-sm font-semibold transition"
            :class="activeTab === 'profiles' ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            @click="activeTab = 'profiles'"
          >
            配置管理
          </button>
          <button
            class="w-full rounded-xl px-4 py-2 text-left text-sm font-semibold transition"
            :class="activeTab === 'settings' ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            @click="activeTab = 'settings'"
          >
            安全设置
          </button>
        </nav>
      </aside>

      <main class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm md:p-6">
        <SyncCenter v-if="activeTab === 'sync'" />
        <ProfileManager v-else-if="activeTab === 'profiles'" />
        <SettingsPanel v-else />
      </main>
    </div>
  </div>
</template>
