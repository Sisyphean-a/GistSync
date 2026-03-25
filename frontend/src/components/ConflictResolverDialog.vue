<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import type { ApplyConflict } from '../lib/backend'

const props = defineProps<{
  visible: boolean
  conflicts: ApplyConflict[]
}>()

const emit = defineEmits<{
  close: []
  confirm: [overwriteItemIds: string[]]
}>()

const decisions = reactive<Record<string, boolean>>({})

watch(
  () => [props.visible, props.conflicts],
  () => {
    if (!props.visible) {
      return
    }
    for (const conflict of props.conflicts) {
      decisions[conflict.itemId] = false
    }
  },
  { deep: true, immediate: true },
)

const overwriteItemIds = computed(() => {
  return props.conflicts.filter((conflict) => decisions[conflict.itemId]).map((conflict) => conflict.itemId)
})

function setAll(value: boolean): void {
  for (const conflict of props.conflicts) {
    decisions[conflict.itemId] = value
  }
}

function confirm(): void {
  emit('confirm', overwriteItemIds.value)
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4">
    <div class="w-full max-w-3xl rounded-xl border border-slate-300 bg-white shadow-xl">
      <div class="border-b border-slate-200 px-5 py-4">
        <h3 class="text-base font-semibold text-slate-900">检测到冲突文件</h3>
        <p class="mt-1 text-sm text-slate-600">默认全部跳过。仅显式选择覆盖的文件会被写入本地。</p>
      </div>
      <div class="px-5 py-4">
        <div class="mb-3 flex flex-wrap gap-2">
          <button class="rounded-lg bg-rose-700 px-3 py-1.5 text-xs font-semibold text-white hover:bg-rose-600" @click="setAll(true)">
            本次全选覆盖
          </button>
          <button class="rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="setAll(false)">
            本次全选跳过
          </button>
        </div>
        <div class="max-h-72 overflow-auto rounded-lg border border-slate-200">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-slate-600">
              <tr>
                <th class="w-16 px-3 py-2">覆盖</th>
                <th class="px-3 py-2">目标路径</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="conflict in conflicts" :key="conflict.itemId" class="border-t border-slate-200">
                <td class="px-3 py-2">
                  <input v-model="decisions[conflict.itemId]" type="checkbox">
                </td>
                <td class="px-3 py-2 font-mono text-xs text-slate-700">{{ conflict.targetPath }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div class="flex items-center justify-between border-t border-slate-200 px-5 py-4">
        <p class="text-xs text-slate-500">覆盖 {{ overwriteItemIds.length }} 项，跳过 {{ conflicts.length - overwriteItemIds.length }} 项</p>
        <div class="flex gap-2">
          <button class="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100" @click="emit('close')">
            取消
          </button>
          <button class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-700" @click="confirm">
            确认执行
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
