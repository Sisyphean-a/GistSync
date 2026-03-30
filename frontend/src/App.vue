<script setup lang="ts">
import { onMounted, ref } from 'vue'
import ConflictResolverDialog from './components/ConflictResolverDialog.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import SyncCenter from './components/SyncCenter.vue'
import ProfileManager from './components/ProfileManager.vue'
import { useQuickSyncStore } from './composables/useQuickSyncStore'

type AdvancedTab = 'sync' | 'profiles' | 'settings'

const quickStore = useQuickSyncStore()
const {
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
} = quickStore
const advancedVisible = ref(false)
const advancedTab = ref<AdvancedTab>('sync')

onMounted(async () => {
  await initialize()
})

function openAdvanced(tab: AdvancedTab): void {
  advancedTab.value = tab
  advancedVisible.value = true
}
</script>

<template>
  <div class="app-shell min-h-screen">
    <header class="border-b border-slate-200 bg-white/90 backdrop-blur">
      <div class="app-container flex items-center justify-between gap-4 px-4 py-3 md:px-6">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">GistSync</p>
          <h1 class="mt-1 text-xl font-bold text-slate-900">快速同步</h1>
        </div>
        <button class="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-100" @click="openAdvanced('sync')">
          高级
        </button>
      </div>
    </header>

    <main class="app-container space-y-4 px-4 py-4 md:px-6 md:py-5">
      <section class="rounded-xl border border-slate-200 bg-white p-4">
        <label class="text-sm text-slate-700">
          <span class="mb-1 block font-medium">当前配置集（下载/上传共用）</span>
          <select
            :value="selectedProfileId"
            class="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm"
            @change="switchProfile(($event.target as HTMLSelectElement).value)"
          >
            <option value="" disabled>请选择配置集</option>
            <option v-for="profile in profiles" :key="profile.id" :value="profile.id">{{ profile.name }}</option>
          </select>
        </label>
      </section>

      <section class="grid gap-4 lg:grid-cols-2">
        <article class="rounded-xl border border-slate-200 bg-white p-5">
          <h2 class="text-base font-semibold text-slate-900">下载并更新本地配置</h2>
          <p class="mt-2 text-sm text-slate-600">默认使用最新快照；遇到冲突时会弹窗确认，默认全覆盖。</p>
          <button
            class="mt-5 h-11 min-w-[180px] rounded-lg bg-amber-700 px-5 text-sm font-semibold text-white hover:bg-amber-600 disabled:opacity-60"
            :disabled="busyAction !== ''"
            @click="download"
          >
            一键下载更新
          </button>
        </article>

        <article class="rounded-xl border border-slate-200 bg-white p-5">
          <h2 class="text-base font-semibold text-slate-900">上传新的配置</h2>
          <p class="mt-2 text-sm text-slate-600">默认上传当前配置集全部启用条目，并生成新快照。</p>
          <button
            class="mt-5 h-11 min-w-[180px] rounded-lg bg-emerald-700 px-5 text-sm font-semibold text-white hover:bg-emerald-600 disabled:opacity-60"
            :disabled="busyAction !== ''"
            @click="upload"
          >
            一键上传配置
          </button>
        </article>
      </section>

      <section v-if="status" class="rounded-xl border border-slate-200 bg-slate-50 p-4">
        <p class="text-sm text-slate-700">{{ status }}</p>
        <button
          v-if="lastResult"
          class="mt-2 rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-100"
          @click="toggleDetails"
        >
          {{ showResultDetails ? '收起明细' : '展开明细' }}
        </button>
        <div v-if="lastResult && showResultDetails" class="mt-3 overflow-auto rounded-lg border border-slate-200 bg-white">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-slate-600">
              <tr>
                <th class="px-3 py-2">条目</th>
                <th class="px-3 py-2">状态</th>
                <th class="px-3 py-2">目标路径</th>
                <th class="px-3 py-2">原因</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in lastResult.items" :key="`${item.itemId}-${item.targetPath}`" class="border-t border-slate-200">
                <td class="px-3 py-2 font-mono text-xs">{{ item.itemId }}</td>
                <td class="px-3 py-2">{{ item.status }}</td>
                <td class="px-3 py-2 font-mono text-xs">{{ item.targetPath }}</td>
                <td class="px-3 py-2 text-xs text-slate-600">{{ item.reason }}</td>
              </tr>
              <tr v-if="lastResult.items.length === 0">
                <td colspan="4" class="px-3 py-6 text-center text-sm text-slate-500">本次操作没有明细条目。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </main>

    <div v-if="advancedVisible" class="fixed inset-0 z-40 bg-slate-900/40" @click="advancedVisible = false" />
    <aside class="fixed right-0 top-0 z-50 h-full w-full max-w-[920px] border-l border-slate-200 bg-white shadow-2xl transition-transform duration-200" :class="advancedVisible ? 'translate-x-0' : 'translate-x-full'">
      <header class="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <div class="flex flex-wrap gap-2">
          <button class="tab-btn" :class="advancedTab === 'sync' ? 'tab-btn-active' : 'tab-btn-idle'" @click="advancedTab = 'sync'">同步明细</button>
          <button class="tab-btn" :class="advancedTab === 'profiles' ? 'tab-btn-active' : 'tab-btn-idle'" @click="advancedTab = 'profiles'">配置管理</button>
          <button class="tab-btn" :class="advancedTab === 'settings' ? 'tab-btn-active' : 'tab-btn-idle'" @click="advancedTab = 'settings'">安全设置</button>
        </div>
        <button class="rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-sm font-semibold text-slate-700 hover:bg-slate-100" @click="advancedVisible = false">
          关闭
        </button>
      </header>
      <div class="h-[calc(100%-61px)] overflow-auto p-4">
        <SyncCenter v-if="advancedTab === 'sync'" />
        <ProfileManager v-else-if="advancedTab === 'profiles'" />
        <SettingsPanel v-else />
      </div>
    </aside>

    <ConflictResolverDialog
      :visible="conflictVisible"
      :conflicts="conflicts"
      :default-overwrite-all="true"
      @close="closeConflictDialog"
      @confirm="submitConflictDecision"
    />
  </div>
</template>
