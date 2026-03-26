import { computed, ref } from 'vue'
import {
  loadSettings,
  saveSettings,
  setActiveProfile,
  type Profile,
  type SettingsData,
} from '../lib/backend'

const state = ref<SettingsData | null>(null)
const loading = ref(false)
let opChain: Promise<void> = Promise.resolve()

const activeProfile = computed<Profile | null>(() => {
  const current = state.value
  if (!current) {
    return null
  }
  return current.profiles.find((profile) => profile.id === current.activeProfileId) ?? null
})

function cloneSettings(data: SettingsData): SettingsData {
  return {
    ...data,
    profiles: data.profiles.map((profile) => ({
      ...profile,
      items: profile.items.map((item) => ({ ...item })),
    })),
  }
}

function enqueue<T>(operation: () => Promise<T>): Promise<T> {
  const run = opChain.then(operation, operation)
  opChain = run.then(() => undefined, () => undefined)
  return run
}

async function refreshUnsafe(): Promise<void> {
  loading.value = true
  try {
    state.value = await loadSettings()
  } finally {
    loading.value = false
  }
}

async function refresh(): Promise<void> {
  await enqueue(async () => {
    await refreshUnsafe()
  })
}

async function ensureLoadedUnsafe(): Promise<void> {
  if (state.value) {
    return
  }
  await refreshUnsafe()
}

async function ensureLoaded(): Promise<void> {
  await enqueue(async () => {
    await ensureLoadedUnsafe()
  })
}

async function switchActiveProfile(profileId: string): Promise<void> {
  await enqueue(async () => {
    await setActiveProfile(profileId)
    await refreshUnsafe()
  })
}

async function persistUnsafe(next: SettingsData, rollback: SettingsData): Promise<void> {
  state.value = next
  try {
    await saveSettings(next)
  } catch (error) {
    state.value = rollback
    throw error
  }
}

function updateCredentials(current: SettingsData, token: string, masterPassword: string): SettingsData {
  return {
    ...cloneSettings(current),
    token,
    masterPassword,
  }
}

function updateRestore(current: SettingsData, mode: 'original' | 'rooted', root: string): SettingsData {
  const next = cloneSettings(current)
  const profile = next.profiles.find((item) => item.id === next.activeProfileId)
  if (!profile) {
    return next
  }
  profile.restoreMode = mode
  profile.restoreRoot = root
  return next
}

async function saveWithMutation(buildNext: (current: SettingsData) => SettingsData): Promise<void> {
  await ensureLoadedUnsafe()
  const current = state.value
  if (!current) {
    return
  }
  const rollback = cloneSettings(current)
  const next = buildNext(current)
  await persistUnsafe(next, rollback)
}

async function saveCredentials(token: string, masterPassword: string): Promise<void> {
  await enqueue(async () => {
    await saveWithMutation((current) => updateCredentials(current, token, masterPassword))
  })
}

async function updateActiveProfileRestore(mode: 'original' | 'rooted', root: string): Promise<void> {
  await enqueue(async () => {
    await saveWithMutation((current) => updateRestore(current, mode, root))
  })
}

async function flushPendingOps(): Promise<void> {
  await opChain
}

export function useSettingsStore() {
  return {
    state,
    loading,
    activeProfile,
    refresh,
    ensureLoaded,
    switchActiveProfile,
    saveCredentials,
    updateActiveProfileRestore,
    flushPendingOps,
  }
}
