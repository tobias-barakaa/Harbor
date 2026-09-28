import { useEffect, useRef, useState } from 'react'
import { useAppState } from '../state/AppContext'
import { StreamDeploymentLogs, StopDeploymentLogs } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

interface LogLine { line: string; id: number }

export default function LogsView({ initialDeployment }: { initialDeployment: string | null }) {
  const { deployments, refreshDeployments } = useAppState()
  const [selected, setSelected] = useState<string | null>(initialDeployment)
  const [lines, setLines] = useState<LogLine[]>([])
  const [streaming, setStreaming] = useState(false)
  const [closedError, setClosedError] = useState<string | null>(null)
  const [pinnedToBottom, setPinnedToBottom] = useState(true)
  const scrollRef = useRef<HTMLDivElement>(null)
  const nextId = useRef(0)

  useEffect(() => { refreshDeployments() }, [refreshDeployments])
  useEffect(() => { setSelected(initialDeployment) }, [initialDeployment])

  useEffect(() => {
    if (!selected) return
    setLines([])
    setClosedError(null)
    setStreaming(true)

    const offLine = EventsOn('deployment-log', (payload: { name: string; line: string }) => {
      if (payload.name !== selected) return
      setLines((prev) => [...prev, { line: payload.line, id: nextId.current++ }])
    })
    const offClosed = EventsOn('deployment-log-closed', (payload: { name: string; error: string }) => {
      if (payload.name !== selected) return
      setStreaming(false)
      if (payload.error) setClosedError(payload.error)
    })

    StreamDeploymentLogs(selected).catch(() => setStreaming(false))

    return () => {
      offLine()
      offClosed()
      StopDeploymentLogs(selected!)
    }
  }, [selected])

  useEffect(() => {
    if (pinnedToBottom && scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight
  }, [lines, pinnedToBottom])

  function handleScroll() {
    const el = scrollRef.current
    if (!el) return
    setPinnedToBottom(el.scrollHeight - el.scrollTop - el.clientHeight < 24)
  }

  return (
    <div>
      <div className="view-header">
        <h1>Logs</h1>
        <div className="actions">
          <select value={selected ?? ''} onChange={(e) => setSelected(e.target.value || null)}>
            <option value="">Select deployment...</option>
            {deployments.map((d) => <option key={d.name} value={d.name}>{d.name}</option>)}
          </select>
          <button onClick={() => setLines([])}>Clear</button>
        </div>
      </div>

      {closedError && <p className="error">Log stream ended: {closedError}</p>}

      <div className="log-viewer" ref={scrollRef} onScroll={handleScroll}>
        {lines.length === 0 && <div className="muted">{streaming ? 'Waiting for output...' : 'No logs.'}</div>}
        {lines.map((l) => <div key={l.id} className="log-line">{l.line}</div>)}
      </div>

      {!pinnedToBottom && <button className="scroll-to-bottom" onClick={() => setPinnedToBottom(true)}>↓ Jump to latest</button>}
    </div>
  )
}