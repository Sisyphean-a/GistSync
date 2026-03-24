export interface SettingsData {
  token: string
  masterPassword: string
  syncPath: string
}

declare global {
  interface Window {
    go: {
      main: {
        App: {
          LoadSettings: () => Promise<SettingsData>
          SaveSettings: (data: SettingsData) => Promise<void>
          UploadSync: () => Promise<string>
          DownloadSync: (overwrite: boolean) => Promise<string>
        }
      }
    }
  }
}

const appAPI = () => window.go.main.App

export const loadSettings = (): Promise<SettingsData> => appAPI().LoadSettings()

export const saveSettings = (data: SettingsData): Promise<void> => appAPI().SaveSettings(data)

export const uploadSync = (): Promise<string> => appAPI().UploadSync()

export const downloadSync = (overwrite: boolean): Promise<string> => appAPI().DownloadSync(overwrite)