import { createContext, useContext, useState, type ReactNode } from 'react';

type Mode = 'terminal' | 'gui';

const STORAGE_KEY = 'console_mode';

interface ModeContextValue {
  mode: Mode;
  setMode: (mode: Mode) => void;
}

const ModeContext = createContext<ModeContextValue>(null!);

export function ModeProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<Mode>(() => {
    return (localStorage.getItem(STORAGE_KEY) as Mode) || 'terminal';
  });

  const setMode = (newMode: Mode) => {
    localStorage.setItem(STORAGE_KEY, newMode);
    setModeState(newMode);
  };

  return (
    <ModeContext value={{ mode, setMode }}>
      {children}
    </ModeContext>
  );
}

export function useMode() {
  return useContext(ModeContext);
}
