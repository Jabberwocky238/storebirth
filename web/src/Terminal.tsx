import { useState, useRef, useCallback, useEffect } from 'react';
import type { TerminalLine, TerminalAPI } from './types';
import { getPromptPrefix, credentialStore } from './store';
import {
  helpCommand, registerCommand, loginCommand,
  logoutCommand, whoamiCommand, statusCommand,
} from './commands';
import { rdbCommand, kvCommand, workerCommand, domainCommand } from './resourceCommands';

let lineIdCounter = 0;

const ASCII_ART = `
    ____                      __
   / ___|___  _ __  ___  ___ | | ___
  | |   / _ \\| '_ \\/ __|/ _ \\| |/ _ \\
  | |__| (_) | | | \\__ \\ (_) | |  __/
   \\____\\___/|_| |_|___/\\___/|_|\\___|

`;

export default function Terminal() {
  const [lines, setLines] = useState<TerminalLine[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isPassword, setIsPassword] = useState(false);
  const [_commandHistory, setCommandHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [promptPrefix, setPromptPrefix] = useState('guest@console:~$');

  const inputRef = useRef<HTMLInputElement>(null);
  const outputRef = useRef<HTMLDivElement>(null);
  const inputCallbackRef = useRef<((value: string) => void) | null>(null);
  const awaitingInputRef = useRef(false);

  const scrollToBottom = useCallback(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, []);

  useEffect(scrollToBottom, [lines, scrollToBottom]);

  const addLine = useCallback((text: string, className = '', isHTML = false) => {
    setLines(prev => [...prev, { id: lineIdCounter++, text, className, isHTML }]);
  }, []);

  const terminalAPI: TerminalAPI = {
    print: (text: string, className = '') => addLine(text, className),
    printHTML: (html: string, className = '') => addLine(html, className, true),
    clear: () => setLines([]),
    waitForInput: (prompt: string, password = false) => {
      return new Promise<string>((resolve) => {
        addLine(prompt, 'info');
        awaitingInputRef.current = true;
        setIsPassword(password);
        inputCallbackRef.current = resolve;
      });
    },
  };

  const refreshPrompt = useCallback(() => {
    setPromptPrefix(getPromptPrefix());
  }, []);

  // Initialize
  useEffect(() => {
    addLine(ASCII_ART, 'ascii-art', false);
    addLine('');
    addLine('Welcome to Console Terminal v1.0', 'info');
    addLine('Type "help" for available commands', 'info');
    addLine('');

    if (credentialStore.load()) {
      refreshPrompt();
      addLine(`Session restored for ${getPromptPrefix().split('@')[0]}`, 'success');
      addLine('');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleCommand = useCallback(async (cmd: string) => {
    const parts = cmd.trim().split(/\s+/);
    const command = parts[0].toLowerCase();
    const args = parts.slice(1);

    addLine(`${promptPrefix} ${cmd}`);

    if (command === 'clear') {
      setLines([]);
      return;
    }
    if (command === '') return;

    const commandMap: Record<string, (t: TerminalAPI, a: string[]) => void | Promise<void>> = {
      help: (t) => helpCommand(t),
      register: (t) => registerCommand(t),
      login: (t) => loginCommand(t),
      logout: (t) => { logoutCommand(t); refreshPrompt(); },
      whoami: (t) => whoamiCommand(t),
      status: (t) => statusCommand(t),
      rdb: (t, a) => rdbCommand(t, a),
      kv: (t, a) => kvCommand(t, a),
      worker: (t, a) => workerCommand(t, a),
      domain: (t, a) => domainCommand(t, a),
    };

    if (commandMap[command]) {
      await commandMap[command](terminalAPI, args);
      refreshPrompt();
    } else {
      addLine(`Command not found: ${command}`, 'error');
      addLine('Type "help" for available commands', 'info');
    }
  }, [addLine, promptPrefix, refreshPrompt, terminalAPI]);

  const handleKeyDown = useCallback(async (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      const value = inputValue.trim();
      setInputValue('');

      if (awaitingInputRef.current) {
        awaitingInputRef.current = false;
        setIsPassword(false);
        if (inputCallbackRef.current) {
          const cb = inputCallbackRef.current;
          inputCallbackRef.current = null;
          cb(value);
        }
      } else if (value) {
        setCommandHistory(prev => [...prev, value]);
        setHistoryIndex(-1);
        await handleCommand(value);
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setCommandHistory(prev => {
        const newIndex = historyIndex === -1 ? prev.length - 1 : Math.max(0, historyIndex - 1);
        if (prev.length > 0 && newIndex >= 0) {
          setHistoryIndex(newIndex);
          setInputValue(prev[newIndex]);
        }
        return prev;
      });
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setCommandHistory(prev => {
        if (historyIndex === -1) return prev;
        const newIndex = historyIndex + 1;
        if (newIndex >= prev.length) {
          setHistoryIndex(-1);
          setInputValue('');
        } else {
          setHistoryIndex(newIndex);
          setInputValue(prev[newIndex]);
        }
        return prev;
      });
    }
  }, [inputValue, historyIndex, handleCommand]);

  const focusInput = useCallback(() => {
    inputRef.current?.focus();
  }, []);

  return (
    <div id="terminal" onClick={focusInput}>
      <div id="output" ref={outputRef}>
        {lines.map(line =>
          line.isHTML ? (
            <div
              key={line.id}
              className={`line ${line.className}`}
              dangerouslySetInnerHTML={{ __html: line.text }}
            />
          ) : (
            <div key={line.id} className={`line ${line.className}`}>
              {line.text}
            </div>
          )
        )}
      </div>
      <div id="input-line">
        <span id="prompt-symbol">{promptPrefix}</span>
        <input
          ref={inputRef}
          type={isPassword ? 'password' : 'text'}
          id="input"
          autoComplete="off"
          autoFocus
          value={inputValue}
          onChange={e => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
        />
      </div>
    </div>
  );
}
