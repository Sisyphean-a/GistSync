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

async function refresh(): Promise<void> {
  loading.value = true
  try {
    state.value = await loadSettings()
  } finally {
    loading.value = false
  }
}

async function ensureLoaded(): Promise<void> {
  if (state.value) {
    return
  }
  await refresh()
}

async function switchActiveProfile(profileId: string): Promise<void> {
  await setActiveProfile(profileId)
  await refresh()
}

async function persist(next: SettingsData): Promise<void> {
  await saveSettings(next)
  state.value = next
}

async function saveCredentials(token: string, masterPassword: string): Promise<void> {
  await ensureLoaded()
  const current = state.value
  if (!current) {
    return
  }
  await persist({
    ...cloneSettings(current),
    token,
    masterPassword,
  })
}

async function updateActiveProfileRestore(mode: 'original' | 'rooted', root: string): Promise<void> {
  await ensureLoaded()
  const current = state.value
  if (!current) {
    return
  }
  const next = cloneSettings(current)
  const profile = next.profiles.find((item) => item.id === next.activeProfileId)
  if (!profile) {
    return
  }
  profile.restoreMode = mode
  profile.restoreRoot = root
  await persist(next)
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
  }
}
