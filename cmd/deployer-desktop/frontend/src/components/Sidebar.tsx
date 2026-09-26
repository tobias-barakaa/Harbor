export type View = 'applications' | 'servers' | 'deployments' | 'logs'

const items: { id: View; label: string }[] = [
  { id: 'applications', label: 'Applications' },
  { id: 'servers', label: 'Servers' },
  { id: 'deployments', label: 'Deployments' },
  { id: 'logs', label: 'Logs' },
]

export default function Sidebar({
  active,
  onSelect,
}: {
  active: View
  onSelect: (v: View) => void
}) {
  return (
    <nav className="sidebar">
      <div className="sidebar-title">Deployer</div>
      {items.map((item) => (
        <button
          key={item.id}
          className={`sidebar-item ${active === item.id ? 'active' : ''}`}
          onClick={() => onSelect(item.id)}
        >
          {item.label}
        </button>
      ))}
    </nav>
  )
}