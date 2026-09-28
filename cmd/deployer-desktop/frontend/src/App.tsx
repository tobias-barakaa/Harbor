import { useState } from 'react'
import Sidebar, { View } from './components/Sidebar'
import { AppStateProvider, useAppState } from './state/AppContext'
import ApplicationsView from './views/ApplicationsView'
import ServersView from './views/ServersView'
import DeploymentsView from './views/DeploymentsView'
import LogsView from './views/LogsView'

function Shell() {
  const [view, setView] = useState<View>('applications')
  const [logsTarget, setLogsTarget] = useState<string | null>(null)
  const { setSelectedApplicationId } = useAppState()

  function goDeploy(appId: string) {
    setSelectedApplicationId(appId)
    setView('deployments')
  }
  function goLogs(deploymentName: string) {
    setLogsTarget(deploymentName)
    setView('logs')
  }

  return (
    <div className="shell">
      <Sidebar active={view} onSelect={setView} />
      <main className="content">
        {view === 'applications' && <ApplicationsView onDeploy={goDeploy} />}
        {view === 'servers' && <ServersView />}
        {view === 'deployments' && <DeploymentsView onViewLogs={goLogs} />}
        {view === 'logs' && <LogsView initialDeployment={logsTarget} />}
      </main>
    </div>
  )
}

export default function App() {
  return <AppStateProvider><Shell /></AppStateProvider>
}