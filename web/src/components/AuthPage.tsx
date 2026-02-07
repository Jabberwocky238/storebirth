import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { authAPI } from '../api'
import { credentialStore } from '../store'
import { useMode } from '../context/ModeContext'

type AuthTab = 'login' | 'register' | 'recover'

export default function AuthPage() {
  const navigate = useNavigate()
  const { setMode } = useMode()
  const [tab, setTab] = useState<AuthTab>(() => {
    const h = location.hash.replace('#', '') as AuthTab
    return ['login', 'register', 'recover'].includes(h) ? h : 'login'
  })

  useEffect(() => {
    location.hash = tab
  }, [tab])

  useEffect(() => {
    const onChange = () => {
      const h = location.hash.replace('#', '') as AuthTab
      if (['login', 'register', 'recover'].includes(h)) setTab(h)
    }
    window.addEventListener('hashchange', onChange)
    return () => window.removeEventListener('hashchange', onChange)
  }, [])

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-950 text-zinc-100">
      <div className="w-full max-w-sm p-6">
        <h1 className="text-2xl font-bold text-center mb-6">Console</h1>

        {tab === 'login' && <LoginForm onSuccess={() => navigate('/')} />}
        {tab === 'register' && <RegisterForm onSuccess={() => navigate('/')} />}
        {tab === 'recover' && <RecoverForm onSuccess={() => setTab('login')} />}

        <div className="mt-6 text-center text-sm text-zinc-400 space-y-2">
          {tab !== 'login' && (
            <button onClick={() => setTab('login')} className="block w-full hover:text-zinc-200">
              Back to Login
            </button>
          )}
          {tab === 'login' && (
            <>
              <button onClick={() => setTab('register')} className="block w-full hover:text-zinc-200">
                Create an account
              </button>
              <button onClick={() => setTab('recover')} className="block w-full hover:text-zinc-200">
                Forgot password? Recover secret key
              </button>
            </>
          )}
          <button
            onClick={() => setMode('terminal')}
            className="block w-full mt-4 text-zinc-500 hover:text-zinc-300"
          >
            Switch to Terminal
          </button>
        </div>
      </div>
    </div>
  )
}

function LoginForm({ onSuccess }: { onSuccess: () => void }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const result = await authAPI.login(email, password)
      const secretKey = credentialStore.getStoredSecretKey(result.user_id)
      if (!secretKey) {
        setError('No stored secret key. Please recover your key first.')
        setLoading(false)
        return
      }
      credentialStore.save(result.user_id, result.token, secretKey)
      onSuccess()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input label="Email" type="email" value={email} onChange={setEmail} />
      <Input label="Password" type="password" value={password} onChange={setPassword} />
      {error && <p className="text-red-400 text-sm">{error}</p>}
      <SubmitButton loading={loading}>Login</SubmitButton>
    </form>
  )
}

function RegisterForm({ onSuccess }: { onSuccess: () => void }) {
  const [step, setStep] = useState<'email' | 'verify'>('email')
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [secretKeyDisplay, setSecretKeyDisplay] = useState('')

  const handleSendCode = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await authAPI.sendCode(email)
      setStep('verify')
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const result = await authAPI.register(email, code, password)
      credentialStore.save(result.user_id, result.token, result.secret_key)
      setSecretKeyDisplay(result.secret_key)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  if (secretKeyDisplay) {
    return (
      <div className="space-y-4">
        <p className="text-green-400 text-sm font-semibold">Registration successful!</p>
        <div className="bg-zinc-800 p-3 rounded">
          <p className="text-yellow-400 text-xs mb-2 font-semibold">
            IMPORTANT: Backup your secret key!
          </p>
          <p className="text-zinc-300 text-xs break-all font-mono">{secretKeyDisplay}</p>
        </div>
        <SubmitButton onClick={onSuccess}>Continue</SubmitButton>
      </div>
    )
  }

  if (step === 'email') {
    return (
      <form onSubmit={handleSendCode} className="space-y-4">
        <Input label="Email" type="email" value={email} onChange={setEmail} />
        {error && <p className="text-red-400 text-sm">{error}</p>}
        <SubmitButton loading={loading}>Send Verification Code</SubmitButton>
      </form>
    )
  }

  return (
    <form onSubmit={handleRegister} className="space-y-4">
      <p className="text-zinc-400 text-sm">Code sent to {email}</p>
      <Input label="Verification Code" value={code} onChange={setCode} />
      <Input label="Password" type="password" value={password} onChange={setPassword} />
      {error && <p className="text-red-400 text-sm">{error}</p>}
      <SubmitButton loading={loading}>Register</SubmitButton>
    </form>
  )
}

function RecoverForm({ onSuccess }: { onSuccess: () => void }) {
  const [step, setStep] = useState<'email' | 'verify'>('email')
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [secretKey, setSecretKey] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSendCode = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await authAPI.sendCode(email)
      setStep('verify')
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  const handleRecover = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const result = await authAPI.login(email, code)
      credentialStore.save(result.user_id, result.token, secretKey)
      onSuccess()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  if (step === 'email') {
    return (
      <form onSubmit={handleSendCode} className="space-y-4">
        <Input label="Email" type="email" value={email} onChange={setEmail} />
        {error && <p className="text-red-400 text-sm">{error}</p>}
        <SubmitButton loading={loading}>Send Verification Code</SubmitButton>
      </form>
    )
  }

  return (
    <form onSubmit={handleRecover} className="space-y-4">
      <p className="text-zinc-400 text-sm">Code sent to {email}</p>
      <Input label="Verification Code" value={code} onChange={setCode} />
      <Input label="Secret Key (sk_...)" value={secretKey} onChange={setSecretKey} />
      {error && <p className="text-red-400 text-sm">{error}</p>}
      <SubmitButton loading={loading}>Recover</SubmitButton>
    </form>
  )
}

function Input({ label, type = 'text', value, onChange }: {
  label: string; type?: string; value: string; onChange: (v: string) => void
}) {
  return (
    <div>
      <label className="block text-sm text-zinc-400 mb-1">{label}</label>
      <input
        type={type}
        value={value}
        onChange={e => onChange(e.target.value)}
        required
        className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded text-zinc-100 text-sm focus:outline-none focus:border-zinc-500"
      />
    </div>
  )
}

function SubmitButton({ children, loading, onClick }: {
  children: React.ReactNode; loading?: boolean; onClick?: () => void
}) {
  return (
    <button
      type={onClick ? 'button' : 'submit'}
      onClick={onClick}
      disabled={loading}
      className="w-full py-2 bg-zinc-700 hover:bg-zinc-600 disabled:opacity-50 rounded text-sm font-medium"
    >
      {loading ? 'Loading...' : children}
    </button>
  )
}
