import { getAuthState, setAuthState } from './api';

const STORAGE_KEY = 'console_credentials';

interface StoredCredentials {
  userId: string;
  token: string;
  secretKey: string;
}

export const credentialStore = {
  save(userId: string, token: string, secretKey: string) {
    const data: StoredCredentials = { userId, token, secretKey };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
    setAuthState({ currentUser: userId, token, secretKey });
  },

  load(): boolean {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      try {
        const data: StoredCredentials = JSON.parse(stored);
        setAuthState({
          currentUser: data.userId,
          token: data.token,
          secretKey: data.secretKey,
        });
        return true;
      } catch {
        return false;
      }
    }
    return false;
  },

  clear() {
    localStorage.removeItem(STORAGE_KEY);
    setAuthState({ currentUser: null, token: null, secretKey: null });
  },

  getStoredSecretKey(userId: string): string | null {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      try {
        const data: StoredCredentials = JSON.parse(stored);
        if (data.userId === userId && data.secretKey) {
          return data.secretKey;
        }
      } catch { /* ignore */ }
    }
    return null;
  },
};

export function getPromptPrefix(): string {
  const { currentUser } = getAuthState();
  return currentUser ? `${currentUser}@console:~$` : 'guest@console:~$';
}
