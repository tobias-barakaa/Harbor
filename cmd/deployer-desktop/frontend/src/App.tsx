import { useState } from 'react'
import Sidebar, { View } from './components/Sidebar'
import ApplicationsView from './views/ApplicationsView'
import ServersView from './views/ServersView'
import DeploymentsView from './views/DeploymentsView'
import LogsView from './views/LogsView'

export default function App() {
  const [view, setView] = useState<View>('applications')

  return (
    <div className="shell">
      <Sidebar active={view} onSelect={setView} />
      <main className="content">
        {view === 'applications' && <ApplicationsView />}
        {view === 'servers' && <ServersView />}
        {view === 'deployments' && <DeploymentsView />}
        {view === 'logs' && <LogsView />}
      </main>
    </div>
  )
}