export interface ProfileItem {
  id: string
  sourcePathTemplate: string
  relativePath: string
  enabled: boolean
}

export interface Profile {
  id: string
  name: string
  restoreMode: 'original' | 'rooted'
  restoreRoot: string
  enabled: boolean
  items: ProfileItem[]
}

export interface SettingsData {
  token: string
  masterPassword: string
  activeProfileId: string
  profiles: Profile[]
}

export interface SnapshotMeta {
  id: string
  createdAt: string
}

export interface ApplyConflict {
  itemId: string
  targetPath: string
}

export interface ApplySnapshotRequest {
  profileId: string
  snapshotId: string
  masterPassword: string
  restoreMode: string
  restoreRoot: string
  selectedItemIds: string[]
  overwriteItemIds: string[]
}

export interface ApplyItemResult {
  itemId: string
  targetPath: string
  status: string
  reason: string
}

export interface ApplySnapshotResult {
  applied: number
  skipped: number
  items: ApplyItemResult[]
}

declare global {
  interface Window {
    go: {
      main: {
        App: {
          LoadSettingsV2: () => Promise<SettingsData>
          SaveSettingsV2: (data: SettingsData) => Promise<void>
          CreateProfile: (name: string) => Promise<Profile>
          DeleteProfile: (profileId: string) => Promise<void>
          SetActiveProfile: (profileId: string) => Promise<void>
          ChooseFilesForProfile: (profileId: string) => Promise<string[]>
          RemoveProfileItems: (profileId: string, itemIds: string[]) => Promise<void>
          UploadProfile: (profileId: string, selectedItemIds: string[]) => Promise<{ snapshotId: string; uploaded: number }>
          ListSnapshots: (profileId: string) => Promise<SnapshotMeta[]>
          PreviewApplyConflicts: (req: ApplySnapshotRequest) => Promise<ApplyConflict[]>
          ApplySnapshot: (req: ApplySnapshotRequest) => Promise<ApplySnapshotResult>
          PullProfilesFromCloud: () => Promise<number>
        }
      }
    }
  }
}

const appAPI = () => window.go.main.App

export const loadSettings = (): Promise<SettingsData> => appAPI().LoadSettingsV2()
export const saveSettings = (data: SettingsData): Promise<void> => appAPI().SaveSettingsV2(data)
export const createProfile = (name: string): Promise<Profile> => appAPI().CreateProfile(name)
export const deleteProfile = (profileId: string): Promise<void> => appAPI().DeleteProfile(profileId)
export const setActiveProfile = (profileId: string): Promise<void> => appAPI().SetActiveProfile(profileId)
export const chooseFilesForProfile = (profileId: string): Promise<string[]> => appAPI().ChooseFilesForProfile(profileId)
export const removeProfileItems = (profileId: string, itemIds: string[]): Promise<void> => appAPI().RemoveProfileItems(profileId, itemIds)
export const uploadProfile = (profileId: string, selectedItemIds: string[]): Promise<{ snapshotId: string; uploaded: number }> =>
  appAPI().UploadProfile(profileId, selectedItemIds)
export const listSnapshots = (profileId: string): Promise<SnapshotMeta[]> => appAPI().ListSnapshots(profileId)
export const previewApplyConflicts = (req: ApplySnapshotRequest): Promise<ApplyConflict[]> => appAPI().PreviewApplyConflicts(req)
export const applySnapshot = (req: ApplySnapshotRequest): Promise<ApplySnapshotResult> => appAPI().ApplySnapshot(req)
export const pullProfilesFromCloud = (): Promise<number> => appAPI().PullProfilesFromCloud()
