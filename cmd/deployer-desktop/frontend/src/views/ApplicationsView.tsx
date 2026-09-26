import { useState } from 'react'
import { InspectPath, PickDirectory, PickZipFile } from '../../wailsjs/go/main/App'

interface AppInfo {
  name: string
  runtime: string
  framework: string
  strategy: string
  port: number
}

export default function ApplicationsView() {
  const [apps, setApps] = useState<AppInfo[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function addFromPath(pick: () => Promise<string>) {
    setError(null)
    const path = await pick()
    if (!path) return
    setLoading(true)
    try {
      const result = await InspectPath(path)
      setApps((prev) => [...prev, ...result])
    } catch (err) {
      setError(String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div className="view-header">
        <h1>Applications</h1>
        <div className="actions">
          <button onClick={() => addFromPath(PickDirectory)}>Add from folder</button>
          <button onClick={() => addFromPath(PickZipFile)}>Add from zip</button>
        </div>
      </div>

      {loading && <p className="muted">Detecting...</p>}
      {error && <p className="error">{error}</p>}

      <div className="card-grid">
        {apps.map((a, i) => (
          <div className="card" key={i}>
            <h3>{a.name}</h3>
            <p>
              {a.runtime}
              {a.framework ? ` · ${a.framework}` : ''}
            </p>
            <p className="muted">
              Strategy: {a.strategy} · Port: {a.port}
            </p>
          </div>
        ))}
        {apps.length === 0 && !loading && <p className="muted">No applications added yet.</p>}
      </div>
    </div>
  )
}