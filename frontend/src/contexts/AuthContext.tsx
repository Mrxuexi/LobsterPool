import type { ReactElement, ReactNode } from 'react';
import { createContext, useContext, useEffect, useState } from 'react';
import type { User } from '../types';
import * as api from '../api/client';

interface AuthContextType {
  user: User | null;
  isAdmin: boolean;
  loading: boolean;
  login: (username: string, password: string) => Promise<User>;
  register: (username: string, password: string) => Promise<User>;
  changePassword: (newPassword: string) => Promise<User>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);
const TOKEN_STORAGE_KEY = 'token';
const USER_STORAGE_KEY = 'user';

interface AuthProviderProps {
  children: ReactNode;
}

function readStoredUser(): User | null {
  const storedUser = localStorage.getItem(USER_STORAGE_KEY);
  return storedUser ? JSON.parse(storedUser) as User : null;
}

function hasStoredToken(): boolean {
  return localStorage.getItem(TOKEN_STORAGE_KEY) !== null;
}

function storeAuthSession(token: string, user: User): void {
  localStorage.setItem(TOKEN_STORAGE_KEY, token);
  localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user));
}

function clearAuthSession(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY);
  localStorage.removeItem(USER_STORAGE_KEY);
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

export function AuthProvider({ children }: AuthProviderProps): ReactElement {
  const [user, setUser] = useState<User | null>(readStoredUser);
  const [loading, setLoading] = useState(hasStoredToken);
  const isAdmin = user?.role === 'admin';

  useEffect(() => {
    if (!hasStoredToken()) {
      return;
    }

    api.getMe()
      .then((currentUser) => {
        setUser(currentUser);
        localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(currentUser));
      })
      .catch(() => {
        clearAuthSession();
        setUser(null);
      })
      .finally(() => setLoading(false));
  }, []);

  async function login(username: string, password: string): Promise<User> {
    const res = await api.login(username, password);
    storeAuthSession(res.token, res.user);
    setUser(res.user);
    return res.user;
  }

  async function register(username: string, password: string): Promise<User> {
    const res = await api.register(username, password);
    storeAuthSession(res.token, res.user);
    setUser(res.user);
    return res.user;
  }

  async function changePassword(newPassword: string): Promise<User> {
    const updatedUser = await api.changePassword(newPassword);
    setUser(updatedUser);
    localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(updatedUser));
    return updatedUser;
  }

  function logout(): void {
    clearAuthSession();
    setUser(null);
  }

  return (
    <AuthContext.Provider value={{ user, isAdmin, loading, login, register, changePassword, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
