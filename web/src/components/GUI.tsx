import { BrowserRouter, Routes, Route, Navigate, Outlet, NavLink } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { credentialStore } from '../store'
import { rdbAPI, workerAPI, domainAPI } from '../api'
import { useMode } from '../context/ModeContext'
import AuthPage from './AuthPage'

function RequireAuth() {
  const isAuth = credentialStore.load()
  if (!isAuth) return <Navigate to="/auth" replace />
  return <Outlet />
}

function MainLayout() {
  const { setMode } = useMode()

  return (
    <div className="flex flex-col h-screen bg-zinc-950 text-zinc-100">
      {/* AppBar - top, full width */}
      <header className="h-12 border-b border-zinc-800 flex items-center px-4 shrink-0">
        <div className="text-lg font-bold">Console</div>
        <div className="flex-1" />
        <div className="flex items-center gap-3">
          <button className="text-sm text-zinc-400 hover:text-zinc-200">Lang</button>
          <button className="text-sm text-zinc-400 hover:text-zinc-200">Account</button>
          <button
            onClick={() => setMode('terminal')}
            className="text-sm text-zinc-400 hover:text-zinc-200"
          >
            Terminal
          </button>
        </div>
      </header>

      {/* Drawer + Content below AppBar */}
      <div className="flex flex-1 overflow-hidden">
        <nav className="w-48 border-r border-zinc-800 py-2 shrink-0">
          <NavItem to="/rdb" label="Database" />
          <NavItem to="/domain" label="Domain" />
          <NavItem to="/worker" label="Worker" />
        </nav>
        <main className="flex-1 p-6 overflow-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

function NavItem({ to, label }: { to: string; label: string }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `block px-4 py-2 text-sm ${isActive ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900'}`
      }
    >
      {label}
    </NavLink>
  )
}

function useList<T>(fetcher: () => Promise<T>) {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  useEffect(() => {
    fetcher().then(setData).catch(e => setError(e.message)).finally(() => setLoading(false))
  }, [])
  return { data, error, loading }
}

function formatBytes(bytes: number): string {
  if (bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`
}

function RdbPage() {
  const { data, error, loading } = useList(rdbAPI.list)
  if (loading) return <div className="text-zinc-500">Loading...</div>
  if (error) return <div className="text-red-400">{error}</div>
  const rdbs = data?.rdbs as { id: string; name: string; url: string; size: number }[] | undefined
  return (
    <div>
      <h2 className="text-lg font-semibold mb-4">Database</h2>
      {data?.database_size !== undefined && (
        <p className="text-sm text-zinc-400 mb-4">Total: {formatBytes(data.database_size)}</p>
      )}
      {rdbs && rdbs.length > 0 ? (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-zinc-400 border-b border-zinc-800">
              <th className="pb-2 pr-4">ID</th>
              <th className="pb-2 pr-4">Name</th>
              <th className="pb-2 pr-4">URL</th>
              <th className="pb-2">Size</th>
            </tr>
          </thead>
          <tbody>
            {rdbs.map(r => (
              <tr key={r.id} className="border-b border-zinc-800/50">
                <td className="py-2 pr-4 font-mono text-zinc-300">{r.id}</td>
                <td className="py-2 pr-4">{r.name}</td>
                <td className="py-2 pr-4 text-zinc-400 font-mono text-xs">{r.url}</td>
                <td className="py-2">{formatBytes(r.size)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p className="text-zinc-500">No databases found</p>
      )}
    </div>
  )
}

function DomainPage() {
  const { data, error, loading } = useList(domainAPI.list)
  if (loading) return <div className="text-zinc-500">Loading...</div>
  if (error) return <div className="text-red-400">{error}</div>
  const domains = data?.domains as { id: string; domain: string; target: string; status: string }[] | undefined
  return (
    <div>
      <h2 className="text-lg font-semibold mb-4">Domain</h2>
      {domains && domains.length > 0 ? (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-zinc-400 border-b border-zinc-800">
              <th className="pb-2 pr-4">ID</th>
              <th className="pb-2 pr-4">Domain</th>
              <th className="pb-2 pr-4">Target</th>
              <th className="pb-2">Status</th>
            </tr>
          </thead>
          <tbody>
            {domains.map(d => (
              <tr key={d.id} className="border-b border-zinc-800/50">
                <td className="py-2 pr-4 font-mono text-zinc-300">{d.id}</td>
                <td className="py-2 pr-4">{d.domain}</td>
                <td className="py-2 pr-4 text-zinc-400">{d.target}</td>
                <td className="py-2">{d.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p className="text-zinc-500">No domains found</p>
      )}
    </div>
  )
}

function WorkerPage() {
  const { data, error, loading } = useList(workerAPI.list)
  if (loading) return <div className="text-zinc-500">Loading...</div>
  if (error) return <div className="text-red-400">{error}</div>
  const workers = data as { worker_id: string; worker_name: string; active_version_id: number | null }[] | null
  return (
    <div>
      <h2 className="text-lg font-semibold mb-4">Worker</h2>
      {workers && workers.length > 0 ? (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-zinc-400 border-b border-zinc-800">
              <th className="pb-2 pr-4">ID</th>
              <th className="pb-2 pr-4">Name</th>
              <th className="pb-2">Active Version</th>
            </tr>
          </thead>
          <tbody>
            {workers.map(w => (
              <tr key={w.worker_id} className="border-b border-zinc-800/50">
                <td className="py-2 pr-4 font-mono text-zinc-300">{w.worker_id}</td>
                <td className="py-2 pr-4">{w.worker_name}</td>
                <td className="py-2">{w.active_version_id ?? 'none'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p className="text-zinc-500">No workers found</p>
      )}
    </div>
  )
}

export default function GUI() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/auth" element={<AuthPage />} />
        <Route element={<RequireAuth />}>
          <Route element={<MainLayout />}>
            <Route path="/rdb" element={<RdbPage />} />
            <Route path="/domain" element={<DomainPage />} />
            <Route path="/worker" element={<WorkerPage />} />
            <Route index element={<Navigate to="/rdb" replace />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
