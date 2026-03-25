<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  chooseFilesForProfile,
  createProfile,
  deleteProfile,
  loadSettings,
  pullProfilesFromCloud,
  removeProfileItems,
  saveSettings,
  setActiveProfile,
  type Profile,
  type SettingsData,
} from '../lib/backend'

const state = ref<SettingsData | null>(null)
const profileName = ref('')
const selectedItemIds = ref<string[]>([])
const status = ref('')
const restore = reactive({
  mode: 'original' as 'original' | 'rooted',
  root: '',
})

const activeProfile = computed<Profile | null>(() => {
  const current = state.value
  if (!current) {
    return null
  }
  return current.profiles.find((profile) => profile.id === current.activeProfileId) ?? null
})

onMounted(async () => {
  await refreshState()
})

async function refreshState(): Promise<void> {
  const settings = await loadSettings()
  state.value = settings
  selectedItemIds.value = []
  const profile = activeProfile.value
  if (!profile) {
    restore.mode = 'original'
    restore.root = ''
    return
  }
  restore.mode = profile.restoreMode
  restore.root = profile.restoreRoot
}

async function createNewProfile(): Promise<void> {
  try {
    const created = await createProfile(profileName.value)
    profileName.value = ''
    await setActiveProfile(created.id)
    await refreshState()
    status.value = `已创建配置: ${created.name || created.id}`
  } catch (error) {
    status.value = `创建失败: ${String(error)}`
  }
}

async function switchProfile(profileId: string): Promise<void> {
  try {
    await setActiveProfile(profileId)
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
  if (!activeProfile.value) {
    status.value = '请先创建或选择配置'
    return
  }
  try {
    const selected = await chooseFilesForProfile(activeProfile.value.id)
    await refreshState()
    status.value = selected.length === 0 ? '未选择文件' : `已添加 ${selected.length} 个文件`
  } catch (error) {
    status.value = `添加文件失败: ${String(error)}`
  }
}

async function removeSelectedItems(): Promise<void> {
  if (!activeProfile.value || selectedItemIds.value.length === 0) {
    return
  }
  try {
    await removeProfileItems(activeProfile.value.id, selectedItemIds.value)
    await refreshState()
    status.value = '已移除选中文件'
  } catch (error) {
    status.value = `移除失败: ${String(error)}`
  }
}

async function persistRestoreSettings(): Promise<void> {
  if (!state.value || !activeProfile.value) {
    return
  }
  const profile = state.value.profiles.find((item) => item.id === activeProfile.value?.id)
  if (!profile) {
    return
  }
  profile.restoreMode = restore.mode
  profile.restoreRoot = restore.root
  try {
    await saveSettings(state.value)
  } catch (error) {
    status.value = `保存恢复策略失败: ${String(error)}`
  }
}
</script>

<template>
  <section class="grid gap-4 lg:grid-cols-[300px_1fr]">
    <aside class="rounded-xl border border-slate-200 bg-slate-50 p-4">
      <h2 class="mb-3 text-base font-semibold text-slate-900">配置集</h2>
      <div class="mb-3 flex items-center gap-2">
        <input
          v-model="profileName"
          class="h-10 w-full rounded-lg border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
          placeholder="新配置名（可空）"
        >
        <button class="h-10 rounded-lg bg-slate-900 px-4 text-sm font-medium text-white hover:bg-slate-700" @click="createNewProfile">
          新增
        </button>
      </div>
      <button class="mb-3 w-full rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100" @click="pullFromCloud">
        从云端拉取配置
      </button>

      <div class="space-y-2">
        <button
          v-for="profile in state?.profiles || []"
          :key="profile.id"
          class="w-full rounded-xl border px-3 py-2 text-left text-sm transition"
          :class="profile.id === state?.activeProfileId ? 'border-slate-900 bg-slate-100 text-slate-900' : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-100'"
          @click="switchProfile(profile.id)"
        >
          <div class="flex items-center justify-between gap-2">
            <span class="truncate font-medium">{{ profile.name }}</span>
            <span class="rounded-full bg-slate-200 px-2 py-0.5 text-xs text-slate-700">{{ profile.items.length }}</span>
          </div>
        </button>
      </div>

      <button
        v-if="activeProfile"
        class="mt-4 w-full rounded-lg bg-rose-700 px-4 py-2 text-sm font-medium text-white hover:bg-rose-600"
        @click="removeProfile(activeProfile.id)"
      >
        删除当前配置
      </button>
    </aside>

    <section class="space-y-4">
      <article class="rounded-xl border border-slate-200 p-4">
        <h3 class="mb-3 text-sm font-semibold text-slate-900">恢复策略</h3>
        <div class="grid gap-3 md:grid-cols-[180px_1fr]">
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

      <article class="rounded-xl border border-slate-200 p-4">
        <h3 class="mb-3 text-sm font-semibold text-slate-900">文件条目</h3>
        <div class="mb-3 flex flex-wrap gap-2">
          <button class="rounded-lg bg-slate-800 px-4 py-2 text-sm font-medium text-white hover:bg-slate-700" @click="addFiles">
            添加文件（多选）
          </button>
          <button class="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100" @click="removeSelectedItems">
            移除选中
          </button>
        </div>
        <div class="overflow-auto rounded-lg border border-slate-200">
          <table class="w-full min-w-[680px] text-left text-sm">
            <thead class="bg-slate-50 text-slate-600">
              <tr>
                <th class="w-16 px-3 py-2">选中</th>
                <th class="px-3 py-2">原始路径</th>
                <th class="px-3 py-2">相对层级</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in activeProfile?.items || []" :key="item.id" class="border-t border-slate-200">
                <td class="px-3 py-2"><input v-model="selectedItemIds" :value="item.id" type="checkbox"></td>
                <td class="px-3 py-2 font-mono text-xs text-slate-700">{{ item.sourcePathTemplate }}</td>
                <td class="px-3 py-2 font-mono text-xs text-slate-700">{{ item.relativePath }}</td>
              </tr>
              <tr v-if="(activeProfile?.items?.length || 0) === 0">
                <td colspan="3" class="px-3 py-8 text-center text-sm text-slate-500">当前配置还没有文件，点击“添加文件（多选）”开始。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <p v-if="status" class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">{{ status }}</p>
    </section>
  </section>
</template>
