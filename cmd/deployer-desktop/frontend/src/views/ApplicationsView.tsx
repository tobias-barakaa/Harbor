import { useEffect, useState } from 'react'
import { useAppState } from '../state/AppContext'
import { AddApplication, PickDirectory, PickZipFile, ReinspectApplication, RemoveApplication } from '../../wailsjs/go/main/App'
import { parseError } from '../api/types'

export default function ApplicationsView({ onDeploy }: { onDeploy: (appId: string) => void }) {
  const { applications, refreshApplications } = useAppState()
  const [adding, setAdding] = useState(false)
  const [addError, setAddError] = useState<string | null>(null)
  const [busyIds, setBusyIds] = useState<Set<string>>(new Set())
  const [rowErrors, setRowErrors] = useState<Record<string, string>>({})

  useEffect(() => { refreshApplications() }, [refreshApplications])

  async function addFrom(pick: () => Promise<string>) {
    setAddError(null)
    const path = await pick()
    if (!path) return
    setAdding(true)
    try {
      await AddApplication(path)
      await refreshApplications()
    } catch (err) {
      setAddError(parseError(err).message)
    } finally {
      setAdding(false)
    }
  }

  async function withBusy(id: string, fn: () => Promise<void>) {
    setBusyIds((prev) => new Set(prev).add(id))
    setRowErrors((prev) => { const next = { ...prev }; delete next[id]; return next })
    try {
      await fn()
    } catch (err) {
      setRowErrors((prev) => ({ ...prev, [id]: parseError(err).message }))
    } finally {
      setBusyIds((prev) => { const next = new Set(prev); next.delete(id); return next })
    }
  }

  return (
    <div>
      <div className="view-header">
        <h1>Applications</h1>
        <div className="actions">
          <button disabled={adding} onClick={() => addFrom(PickDirectory)}>Add from folder</button>
          <button disabled={adding} onClick={() => addFrom(PickZipFile)}>Add from zip</button>
        </div>
      </div>

      {adding && <p className="muted">Detecting...</p>}
      {addError && <p className="error">{addError}</p>}

      <div className="card-grid">
        {applications.map((a) => {
          const busy = busyIds.has(a.id)
          return (
            <div className="card" key={a.id}>
              <h3>{a.name}</h3>
              <p>{a.runtime}{a.framework ? ` · ${a.framework}` : ''}</p>
              <p className="muted">Strategy: {a.strategy} · Port: {a.port}</p>
              <p className="muted path">{a.path}</p>
              {rowErrors[a.id] && <p className="error">{rowErrors[a.id]}</p>}
              <div className="card-actions">
                <button disabled={busy} onClick={() => withBusy(a.id, async () => { await ReinspectApplication(a.id); await refreshApplications() })}>
                  {busy ? '...' : 'Inspect'}
                </button>
                <button disabled={busy} onClick={() => onDeploy(a.id)}>Deploy</button>
                <button disabled={busy} className="danger" onClick={() => withBusy(a.id, async () => { await RemoveApplication(a.id); await refreshApplications() })}>
                  Remove
                </button>
              </div>
            </div>
          )
        })}
        {applications.length === 0 && !adding && <p className="muted">No applications added yet.</p>}
      </div>
    </div>
  )
}