import { useEffect, useState } from 'react'
import { useAppState } from '../state/AppContext'
import { AddServer, RemoveServer, TestServer } from '../../wailsjs/go/main/App'
import { AddServerInput, parseError } from '../api/types'
import { StatusBadge } from '../components/StatusBadge'

type ConnState = 'unknown' | 'testing' | 'connected' | 'failed'
const emptyForm: AddServerInput = { name: '', host: '', port: 22, username: '', authMethod: 'private_key', privateKeyPath: '', password: '' }

export default function ServersView() {
  const { servers, refreshServers } = useAppState()
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<AddServerInput>(emptyForm)
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [connState, setConnState] = useState<Record<string, ConnState>>({})
  const [connError, setConnError] = useState<Record<string, string>>({})

  useEffect(() => { refreshServers() }, [refreshServers])

  async function submitAdd() {
    setFormError(null)
    setSaving(true)
    try {
      await AddServer(form)
      setForm(emptyForm)
      setShowForm(false)
      await refreshServers()
    } catch (err) {
      setFormError(parseError(err).message)
    } finally {
      setSaving(false)
    }
  }

  async function test(name: string) {
    setConnState((prev) => ({ ...prev, [name]: 'testing' }))
    setConnError((prev) => { const next = { ...prev }; delete next[name]; return next })
    try {
      await TestServer(name)
      setConnState((prev) => ({ ...prev, [name]: 'connected' }))
    } catch (err) {
      setConnState((prev) => ({ ...prev, [name]: 'failed' }))
      setConnError((prev) => ({ ...prev, [name]: parseError(err).message }))
    }
  }

  async function remove(name: string) {
    await RemoveServer(name)
    await refreshServers()
  }

  return (
    <div>
      <div className="view-header">
        <h1>Servers</h1>
        <div className="actions">
          <button onClick={() => setShowForm((v) => !v)}>{showForm ? 'Cancel' : '+ Add Server'}</button>
        </div>
      </div>

      {showForm && (
        <div className="panel">
          <div className="form-grid">
            <input placeholder="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
            <input placeholder="Host" value={form.host} onChange={(e) => setForm({ ...form, host: e.target.value })} />
            <input type="number" placeholder="Port" value={form.port} onChange={(e) => setForm({ ...form, port: Number(e.target.value) })} />
            <input placeholder="Username" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} />
            <select value={form.authMethod} onChange={(e) => setForm({ ...form, authMethod: e.target.value as AddServerInput['authMethod'] })}>
              <option value="private_key">Private key</option>
              <option value="password">Password</option>
            </select>
            {form.authMethod === 'private_key' ? (
              <input placeholder="Private key path" value={form.privateKeyPath} onChange={(e) => setForm({ ...form, privateKeyPath: e.target.value })} />
            ) : (
              <input type="password" placeholder="Password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
            )}
          </div>
          {formError && <p className="error">{formError}</p>}
          <button disabled={saving} onClick={submitAdd}>{saving ? 'Saving...' : 'Save Server'}</button>
        </div>
      )}

      <div className="card-grid">
        {servers.map((s) => {
          const state = connState[s.name] ?? 'unknown'
          return (
            <div className="card" key={s.name}>
              <h3>{s.name}</h3>
              <p className="muted">{s.username}@{s.host}:{s.port}</p>
              <StatusBadge status={state} />
              {connError[s.name] && <p className="error">{connError[s.name]}</p>}
              <div className="card-actions">
                <button disabled={state === 'testing'} onClick={() => test(s.name)}>
                  {state === 'testing' ? 'Testing...' : 'Test Connection'}
                </button>
                <button className="danger" onClick={() => remove(s.name)}>Remove</button>
              </div>
            </div>
          )
        })}
        {servers.length === 0 && <p className="muted">No servers registered.</p>}
      </div>
    </div>
  )
}