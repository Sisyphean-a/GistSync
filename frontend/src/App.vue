<script setup lang="ts">
import { ref } from 'vue'
import SettingsPanel from './components/SettingsPanel.vue'
import SyncCenter from './components/SyncCenter.vue'
import ProfileManager from './components/ProfileManager.vue'

type TabKey = 'sync' | 'profiles' | 'settings'

const activeTab = ref<TabKey>('sync')
</script>

<template>
  <div class="app-shell min-h-screen">
    <header class="border-b border-slate-200 bg-white/85 backdrop-blur">
      <div class="app-container px-4 py-3 md:px-6">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">GistSync</p>
            <h1 class="mt-1 text-xl font-bold text-slate-900">配置同步控制台</h1>
          </div>
          <nav class="flex flex-wrap items-center gap-2">
            <button
              class="tab-btn"
              :class="activeTab === 'sync' ? 'tab-btn-active' : 'tab-btn-idle'"
              @click="activeTab = 'sync'"
            >
              同步中心
            </button>
            <button
              class="tab-btn"
              :class="activeTab === 'profiles' ? 'tab-btn-active' : 'tab-btn-idle'"
              @click="activeTab = 'profiles'"
            >
              配置管理
            </button>
            <button
              class="tab-btn"
              :class="activeTab === 'settings' ? 'tab-btn-active' : 'tab-btn-idle'"
              @click="activeTab = 'settings'"
            >
              安全设置
            </button>
          </nav>
        </div>
      </div>
    </header>

    <main class="app-container px-4 py-4 md:px-6 md:py-5">
      <SyncCenter v-if="activeTab === 'sync'" />
      <ProfileManager v-else-if="activeTab === 'profiles'" />
      <SettingsPanel v-else />
    </main>
  </div>
</template>
