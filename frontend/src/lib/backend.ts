import {
  ApplySnapshot,
  ChooseFilesForProfile,
  CreateProfile,
  DeleteProfile,
  ListSnapshots,
  LoadSettingsV2,
  PreviewApplyConflicts,
  PullProfilesFromCloud,
  QuickDownload,
  QuickUpload,
  RemoveProfileItems,
  SaveSettingsV2,
  SetActiveProfile,
  UploadProfile,
} from '../../wailsjs/go/main/App'
import { settings, syncflow } from '../../wailsjs/go/models'

export interface ProfileItem {
  id: string
  sourcePathTemplate: string
  relativePath: string
  enabled: boolean
}

export interface Profile {
  id: string
  name: string
  restoreMode: string
  restoreRoot: string
  enabled: boolean
  items: ProfileItem[]
}

export interface SettingsData {
  token: string
  masterPassword: string
  activeProfileId: string
  profiles: Profile[]
  cloudBootstrapDone?: boolean
  syncPath?: string
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

export interface UploadProfileResult {
  snapshotId: string
  uploaded: number
}

export type QuickConflictPolicy = 'overwrite_all' | 'manual'

export interface QuickUploadRequest {
  profileId: string
}

export interface QuickDownloadRequest {
  profileId: string
  conflictPolicy: QuickConflictPolicy
  overwriteItemIds: string[]
}

export interface QuickOperationSummary {
  uploaded: number
  applied: number
  skipped: number
  conflicts: number
  errors: number
}

export interface QuickOperationItem {
  itemId: string
  targetPath: string
  status: string
  reason: string
}

export interface QuickOperationResult {
  operationId: string
  action: string
  profileId: string
  snapshotId: string
  requiresConflictResolution: boolean
  summary: QuickOperationSummary
  conflicts: QuickOperationItem[]
  items: QuickOperationItem[]
}

function mapProfileItem(item: settings.ProfileItem): ProfileItem {
  return {
    id: item.id,
    sourcePathTemplate: item.sourcePathTemplate,
    relativePath: item.relativePath,
    enabled: item.enabled,
  }
}

function mapProfile(profile: settings.Profile): Profile {
  return {
    id: profile.id,
    name: profile.name,
    restoreMode: profile.restoreMode,
    restoreRoot: profile.restoreRoot,
    enabled: profile.enabled,
    items: (profile.items ?? []).map(mapProfileItem),
  }
}

function mapSettings(data: settings.Data): SettingsData {
  return {
    token: data.token ?? '',
    masterPassword: data.masterPassword ?? '',
    activeProfileId: data.activeProfileId ?? '',
    profiles: (data.profiles ?? []).map(mapProfile),
    cloudBootstrapDone: data.cloudBootstrapDone,
    syncPath: data.syncPath,
  }
}

function toSettingsModel(data: SettingsData): settings.Data {
  return settings.Data.createFrom(data)
}

function mapSnapshot(item: syncflow.SnapshotMeta): SnapshotMeta {
  return { id: item.id, createdAt: item.createdAt }
}

function mapConflict(item: syncflow.ApplyConflict): ApplyConflict {
  return { itemId: item.itemId, targetPath: item.targetPath }
}

function mapApplyItem(item: syncflow.ApplyItemResult): ApplyItemResult {
  return {
    itemId: item.itemId,
    targetPath: item.targetPath,
    status: item.status,
    reason: item.reason,
  }
}

function mapApplyResult(result: syncflow.ApplySnapshotResult): ApplySnapshotResult {
  return {
    applied: result.applied,
    skipped: result.skipped,
    items: (result.items ?? []).map(mapApplyItem),
  }
}

function toApplyRequestModel(req: ApplySnapshotRequest): syncflow.ApplySnapshotRequest {
  return syncflow.ApplySnapshotRequest.createFrom(req)
}

function mapUploadResult(result: syncflow.UploadProfileResult): UploadProfileResult {
  return { snapshotId: result.snapshotId, uploaded: result.uploaded }
}

export const loadSettings = async (): Promise<SettingsData> => mapSettings(await LoadSettingsV2())
export const saveSettings = async (data: SettingsData): Promise<void> => SaveSettingsV2(toSettingsModel(data))
export const createProfile = async (name: string): Promise<Profile> => mapProfile(await CreateProfile(name))
export const deleteProfile = async (profileId: string): Promise<void> => DeleteProfile(profileId)
export const setActiveProfile = async (profileId: string): Promise<void> => SetActiveProfile(profileId)
export const chooseFilesForProfile = async (profileId: string): Promise<string[]> => ChooseFilesForProfile(profileId)
export const removeProfileItems = async (profileId: string, itemIds: string[]): Promise<void> =>
  RemoveProfileItems(profileId, itemIds)
export const uploadProfile = async (profileId: string, selectedItemIds: string[]): Promise<UploadProfileResult> =>
  mapUploadResult(await UploadProfile(profileId, selectedItemIds))
export const listSnapshots = async (profileId: string): Promise<SnapshotMeta[]> =>
  (await ListSnapshots(profileId)).map(mapSnapshot)
export const previewApplyConflicts = async (req: ApplySnapshotRequest): Promise<ApplyConflict[]> =>
  (await PreviewApplyConflicts(toApplyRequestModel(req))).map(mapConflict)
export const applySnapshot = async (req: ApplySnapshotRequest): Promise<ApplySnapshotResult> =>
  mapApplyResult(await ApplySnapshot(toApplyRequestModel(req)))
export const pullProfilesFromCloud = async (): Promise<number> => PullProfilesFromCloud()
export const quickUpload = async (req: QuickUploadRequest): Promise<QuickOperationResult> => QuickUpload(req)
export const quickDownload = async (req: QuickDownloadRequest): Promise<QuickOperationResult> => QuickDownload(req)
