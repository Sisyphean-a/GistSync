import {
  ApplySnapshot,
  ChooseFilesForProfile,
  CreateProfile,
  DeleteProfile,
  ListSnapshots,
  LoadSettingsV2,
  PreviewApplyConflicts,
  PullProfilesFromCloud,
  RemoveProfileItems,
  SaveSettingsV2,
  SetActiveProfile,
  UploadProfile,
} from '../../wailsjs/go/main/App'
import type { settings, syncflow } from '../../wailsjs/go/models'

type Plain<T> =
  T extends Array<infer U>
    ? Plain<U>[]
    : T extends object
      ? { [K in keyof T as T[K] extends (...args: unknown[]) => unknown ? never : K]: Plain<T[K]> }
      : T

export type ProfileItem = Plain<settings.ProfileItem>
export type Profile = Plain<settings.Profile>
export type SettingsData = Plain<settings.Data>
export type SnapshotMeta = Plain<syncflow.SnapshotMeta>
export type ApplyConflict = Plain<syncflow.ApplyConflict>
export type ApplySnapshotRequest = Plain<syncflow.ApplySnapshotRequest>
export type ApplyItemResult = Plain<syncflow.ApplyItemResult>
export type ApplySnapshotResult = Plain<syncflow.ApplySnapshotResult>
export type UploadProfileResult = Plain<syncflow.UploadProfileResult>

export const loadSettings = async (): Promise<SettingsData> => LoadSettingsV2() as unknown as SettingsData
export const saveSettings = (data: SettingsData): Promise<void> => SaveSettingsV2(data as unknown as settings.Data)
export const createProfile = async (name: string): Promise<Profile> => CreateProfile(name) as unknown as Profile
export const deleteProfile = (profileId: string): Promise<void> => DeleteProfile(profileId)
export const setActiveProfile = (profileId: string): Promise<void> => SetActiveProfile(profileId)
export const chooseFilesForProfile = (profileId: string): Promise<string[]> => ChooseFilesForProfile(profileId)
export const removeProfileItems = (profileId: string, itemIds: string[]): Promise<void> => RemoveProfileItems(profileId, itemIds)
export const uploadProfile = async (profileId: string, selectedItemIds: string[]): Promise<UploadProfileResult> =>
  UploadProfile(profileId, selectedItemIds) as unknown as UploadProfileResult
export const listSnapshots = async (profileId: string): Promise<SnapshotMeta[]> =>
  ListSnapshots(profileId) as unknown as SnapshotMeta[]
export const previewApplyConflicts = async (req: ApplySnapshotRequest): Promise<ApplyConflict[]> =>
  PreviewApplyConflicts(req as unknown as syncflow.ApplySnapshotRequest) as unknown as ApplyConflict[]
export const applySnapshot = async (req: ApplySnapshotRequest): Promise<ApplySnapshotResult> =>
  ApplySnapshot(req as unknown as syncflow.ApplySnapshotRequest) as unknown as ApplySnapshotResult
export const pullProfilesFromCloud = (): Promise<number> => PullProfilesFromCloud()
