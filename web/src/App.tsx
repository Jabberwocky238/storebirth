import Terminal from './components/Terminal'
import GUI from './components/GUI'
import { ModeProvider, useMode } from './context/ModeContext'
import './App.css'

function App() {
  const { mode } = useMode()
  return mode === 'gui' ? <GUI /> : <Terminal />
}

export default function () {
  return <ModeProvider>
    <App />
  </ModeProvider>
}
