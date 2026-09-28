import { useEffect, useState } from 'react'
import { useAppState } from '../state/AppContext'
import { Deploy, GetDeploymentStatus, RemoveDeployment, RestartDeployment, StopDeployment } from '../../wailsjs/go/main/App'
import { parseError } from '../api/types'
import { StatusBadge } from '../components/StatusBadge'

export default function DeploymentsView({ onViewLogs }: { onViewLogs: (name: string) => void }) {
  const {
    applications, servers, deployments,
    refreshApplications, refreshServers, refreshDeployments,
    selectedApplicationId, setSelectedApplicationId,
    selectedServerName, setSelectedServerName,
  } = useAppState()

  const [deploying, setDeploying] = useState(false)
  const [deployError, setDeployError] = useState<string | null>(null)
  const [busyNames, setBusyNames] = useState<Set<string>>(new Set())
  const [rowErrors, setRowErrors] = useState<Record<string, string>>({})

  useEffect(() => { refreshApplications(); refreshServers(); refreshDeployments() }, [refreshApplications, refreshServers, refreshDeployments])

  async function deploy() {
    if (!selectedApplicationId || !selectedServerName) {
      setDeployError('Select both an application and a server')
      return
    }
    setDeployError(null)
    setDeploying(true)
    try {
      await Deploy({ applicationId: selectedApplicationId, serverName: selectedServerName })
      await refreshDeployments()
    } catch (err) {
      setDeployError(parseError(err).message)
    } finally {
      setDeploying(false)
    }
  }

  async function withBusy(name: string, fn: () => Promise<void>) {
    setBusyNames((prev) => new Set(prev).add(name))
    setRowErrors((prev) => { const next = { ...prev }; delete next[name]; return next })
    try {
      await fn()
    } catch (err) {
      setRowErrors((prev) => ({ ...prev, [name]: parseError(err).message }))
    } finally {
      setBusyNames((prev) => { const next = new Set(prev); next.delete(name); return next })
    }
  }

  async function refreshOne(name: string) {
    try { await GetDeploymentStatus(name); await refreshDeployments() } catch { /* next full-list refresh surfaces it */ }
  }

  return (
    <div>
      <div className="view-header"><h1>Deployments</h1></div>

      <div className="panel">
        <div className="form-grid">
          <select value={selectedApplicationId ?? ''} onChange={(e) => setSelectedApplicationId(e.target.value || null)}>
            <option value="">Select application...</option>
            {applications.map((a) => <option key={a.id} value={a.id}>{a.name} ({a.runtime})</option>)}
          </select>
          <select value={selectedServerName ?? ''} onChange={(e) => setSelectedServerName(e.target.value || null)}>
            <option value="">Select server...</option>
            {servers.map((s) => <option key={s.name} value={s.name}>{s.name}</option>)}
          </select>
        </div>
        {deployError && <p className="error">{deployError}</p>}
        <button disabled={deploying} onClick={deploy}>{deploying ? 'Deploying...' : 'Deploy'}</button>
      </div>

      <div className="card-grid">
        {deployments.map((d) => {
          const busy = busyNames.has(d.name)
          return (
            <div className="card" key={d.name}>
              <h3>{d.name}</h3>
              <p className="muted">{d.serverName} · {d.containerName}</p>
              <p className="muted">{d.ports.map((p) => `${p.host}:${p.container}`).join(', ')}</p>
              <StatusBadge status={d.status} />
              {rowErrors[d.name] && <p className="error">{rowErrors[d.name]}</p>}
              <div className="card-actions">
                <button disabled={busy} onClick={() => refreshOne(d.name)}>Status</button>
                <button disabled={busy} onClick={() => onViewLogs(d.name)}>Logs</button>
                {d.status === 'running' && (
                  <button disabled={busy} onClick={() => withBusy(d.name, async () => { await StopDeployment(d.name); await refreshDeployments() })}>Stop</button>
                )}
                {d.status !== 'running' && d.status !== 'not_found' && (
                  <button disabled={busy} onClick={() => withBusy(d.name, async () => { await RestartDeployment(d.name); await refreshDeployments() })}>Restart</button>
                )}
                <button disabled={busy} className="danger" onClick={() => withBusy(d.name, async () => { await RemoveDeployment(d.name); await refreshDeployments() })}>Remove</button>
              </div>
            </div>
          )
        })}
        {deployments.length === 0 && <p className="muted">No deployments yet.</p>}
      </div>
    </div>
  )
}