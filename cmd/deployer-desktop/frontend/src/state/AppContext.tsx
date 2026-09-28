import { createContext, useCallback, useContext, useState, ReactNode } from 'react'
import { ListApplications, ListServers, ListDeployments } from '../../wailsjs/go/main/App'
import { ApplicationEntry, ServerInfo, DeploymentInfo } from '../api/types'

interface AppState {
  applications: ApplicationEntry[]
  servers: ServerInfo[]
  deployments: DeploymentInfo[]
  selectedApplicationId: string | null
  selectedServerName: string | null
  setSelectedApplicationId: (id: string | null) => void
  setSelectedServerName: (name: string | null) => void
  refreshApplications: () => Promise<void>
  refreshServers: () => Promise<void>
  refreshDeployments: () => Promise<void>
}

const AppStateContext = createContext<AppState | null>(null)

export function AppStateProvider({ children }: { children: ReactNode }) {
  const [applications, setApplications] = useState<ApplicationEntry[]>([])
  const [servers, setServers] = useState<ServerInfo[]>([])
  const [deployments, setDeployments] = useState<DeploymentInfo[]>([])
  const [selectedApplicationId, setSelectedApplicationId] = useState<string | null>(null)
  const [selectedServerName, setSelectedServerName] = useState<string | null>(null)

  const refreshApplications = useCallback(async () => setApplications(await ListApplications()), [])
  const refreshServers = useCallback(async () => setServers(await ListServers()), [])
  const refreshDeployments = useCallback(async () => setDeployments(await ListDeployments()), [])

  return (
    <AppStateContext.Provider value={{
      applications, servers, deployments,
      selectedApplicationId, selectedServerName,
      setSelectedApplicationId, setSelectedServerName,
      refreshApplications, refreshServers, refreshDeployments,
    }}>
      {children}
    </AppStateContext.Provider>
  )
}

export function useAppState() {
  const ctx = useContext(AppStateContext)
  if (!ctx) throw new Error('useAppState must be used within AppStateProvider')
  return ctx
}