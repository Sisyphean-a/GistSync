<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import {
  chooseFilesForProfile,
  createProfile,
  deleteProfile,
  pullProfilesFromCloud,
  removeProfileItems,
} from '../lib/backend'
import { useSettingsStore } from '../composables/useSettingsStore'

const store = useSettingsStore()
const profileName = ref('')
const selectedItemIds = ref<string[]>([])
const status = ref('')
const restore = reactive({
  mode: 'original' as 'original' | 'rooted',
  root: '',
})

onMounted(async () => {
  await refreshState()
})

async function refreshState(): Promise<void> {
  await store.refresh()
  selectedItemIds.value = []
  const profile = store.activeProfile.value
  if (!profile) {
    restore.mode = 'original'
    restore.root = ''
    return
  }
  restore.mode = profile.restoreMode as 'original' | 'rooted'
  restore.root = profile.restoreRoot
}

async function createNewProfile(): Promise<void> {
  try {
    const created = await createProfile(profileName.value)
    profileName.value = ''
    await store.switchActiveProfile(created.id)
    await refreshState()
    status.value = `已创建配置: ${created.name || created.id}`
  } catch (error) {
    status.value = `创建失败: ${String(error)}`
  }
}

async function switchProfile(profileId: string): Promise<void> {
  try {
    await store.switchActiveProfile(profileId)
    await refreshState()
    status.value = ''
  } catch (error) {
    status.value = `切换失败: ${String(error)}`
  }
}

async function pullFromCloud(): Promise<void> {
  try {
    const count = await pullProfilesFromCloud()
    await refreshState()
    status.value = `已从云端同步配置集，本地共 ${count} 个`
  } catch (error) {
    status.value = `从云端拉取失败: ${String(error)}`
  }
}

async function removeProfile(profileId: string): Promise<void> {
  try {
    await deleteProfile(profileId)
    await refreshState()
    status.value = '配置已删除'
  } catch (error) {
    status.value = `删除失败: ${String(error)}`
  }
}

async function addFiles(): Promise<void> {
  const profile = store.activeProfile.value
  if (!profile) {
    status.value = '请先创建或选择配置'
    return
  }
  try {
    const selected = await chooseFilesForProfile(profile.id)
    await refreshState()
    status.value = selected.length === 0 ? '未选择文件' : `已添加 ${selected.length} 个文件`
  } catch (error) {
    status.value = `添加文件失败: ${String(error)}`
  }
}

async function removeSelectedItems(): Promise<void> {
  const profile = store.activeProfile.value
  if (!profile || selectedItemIds.value.length === 0) {
    return
  }
  try {
    await removeProfileItems(profile.id, selectedItemIds.value)
    await refreshState()
    status.value = '已移除选中文件'
  } catch (error) {
    status.value = `移除失败: ${String(error)}`
  }
}

async function persistRestoreSettings(): Promise<void> {
  try {
    await store.updateActiveProfileRestore(restore.mode, restore.root)
  } catch (error) {
    status.value = `保存恢复策略失败: ${String(error)}`
  }
}
</script>

<template>
  <section class="space-y-4">
    <section class="grid gap-4 xl:grid-cols-[360px_1fr]">
      <aside class="rounded-xl border border-slate-200 bg-white p-4">
        <div class="mb-4 flex items-center justify-between gap-2 border-b border-slate-200 pb-3">
          <h2 class="text-base font-semibold text-slate-900">配置集</h2>
          <span class="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-600">
            {{ store.state.value?.profiles?.length || 0 }} 组
          </span>
        </div>

        <div class="mb-3 flex flex-wrap items-center gap-2">
          <input
            v-model="profileName"
            class="h-10 min-w-[220px] flex-1 rounded-lg border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
            placeholder="新配置名（可空）"
          >
          <button class="h-10 min-w-[88px] rounded-lg bg-slate-900 px-4 text-sm font-medium text-white hover:bg-slate-700" @click="createNewProfile">
            新增配置
          </button>
        </div>

        <button class="mb-4 h-10 w-full rounded-lg border border-slate-300 bg-white px-4 text-sm font-medium text-slate-700 hover:bg-slate-100" @click="pullFromCloud">
          从云端拉取配置
        </button>

        <div class="max-h-[320px] space-y-2 overflow-auto pr-1">
          <button
            v-for="profile in store.state.value?.profiles || []"
            :key="profile.id"
            class="w-full rounded-xl border px-3 py-2 text-left text-sm transition"
            :class="profile.id === store.state.value?.activeProfileId ? 'border-slate-900 bg-slate-100 text-slate-900' : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-100'"
            @click="switchProfile(profile.id)"
          >
            <div class="flex items-center justify-between gap-2">
              <span class="truncate font-medium">{{ profile.name }}</span>
              <span class="rounded-full bg-slate-200 px-2 py-0.5 text-xs text-slate-700">{{ profile.items.length }}</span>
            </div>
          </button>
        </div>

        <button
          v-if="store.activeProfile.value"
          class="mt-4 h-10 w-full rounded-lg bg-rose-700 px-4 text-sm font-medium text-white hover:bg-rose-600"
          @click="removeProfile(store.activeProfile.value.id)"
        >
          删除当前配置
        </button>
      </aside>

      <article class="rounded-xl border border-slate-200 bg-white p-4">
        <div class="mb-4 border-b border-slate-200 pb-3">
          <h3 class="text-sm font-semibold text-slate-900">恢复策略</h3>
          <p class="mt-1 text-xs text-slate-500">恢复模式会自动保存到当前配置。</p>
        </div>
        <div class="grid gap-3 lg:grid-cols-[220px_1fr]">
          <select v-model="restore.mode" class="h-10 rounded-lg border border-slate-300 px-3 text-sm" @change="persistRestoreSettings">
            <option value="original">原始路径恢复</option>
            <option value="rooted">指定根目录恢复</option>
          </select>
          <input
            v-model="restore.root"
            class="h-10 rounded-lg border border-slate-300 px-3 text-sm"
            placeholder="rooted 模式下填写恢复根目录"
            @blur="persistRestoreSettings"
          >
        </div>
      </article>
    </section>

    <article class="rounded-xl border border-slate-200 bg-white p-4">
      <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
        <h3 class="text-sm font-semibold text-slate-900">文件条目</h3>
        <div class="flex flex-wrap gap-2">
          <button class="h-10 min-w-[136px] rounded-lg bg-slate-800 px-4 text-sm font-medium text-white hover:bg-slate-700" @click="addFiles">
            添加文件（多选）
          </button>
          <button class="h-10 min-w-[104px] rounded-lg border border-slate-300 bg-white px-4 text-sm font-medium text-slate-700 hover:bg-slate-100" @click="removeSelectedItems">
            移除选中
          </button>
        </div>
      </div>
      <div class="overflow-auto rounded-lg border border-slate-200">
        <table class="w-full min-w-[860px] text-left text-sm">
          <thead class="bg-slate-50 text-slate-600">
            <tr>
              <th class="w-16 px-3 py-2">选中</th>
              <th class="px-3 py-2">原始路径</th>
              <th class="px-3 py-2">相对层级</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in store.activeProfile.value?.items || []" :key="item.id" class="border-t border-slate-200">
              <td class="px-3 py-2"><input v-model="selectedItemIds" :value="item.id" type="checkbox"></td>
              <td class="px-3 py-2 font-mono text-xs text-slate-700">{{ item.sourcePathTemplate }}</td>
              <td class="px-3 py-2 font-mono text-xs text-slate-700">{{ item.relativePath }}</td>
            </tr>
            <tr v-if="(store.activeProfile.value?.items?.length || 0) === 0">
              <td colspan="3" class="px-3 py-8 text-center text-sm text-slate-500">当前配置还没有文件，点击“添加文件（多选）”开始。</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <p v-if="status" class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">{{ status }}</p>
  </section>
</template>
