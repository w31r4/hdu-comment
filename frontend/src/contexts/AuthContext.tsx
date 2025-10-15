import type { AxiosError } from 'axios';
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode
} from 'react';
import {
  login as apiLogin,
  register as apiRegister,
  fetchMe,
  refreshTokens,
  logout as apiLogout,
  setAuthTokens,
  clearAuthTokens,
  setRefreshExecutor,
  getRefreshToken
} from '../api/client';
import type { AuthResponse, User } from '../types';

interface AuthContextValue {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, displayName: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshProfile: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

const TOKEN_KEY = 'hdu-food-review-access-token';
const REFRESH_KEY = 'hdu-food-review-refresh-token';

const readLocalToken = (key: string) => {
  try {
    return localStorage.getItem(key);
  } catch (err) {
    console.warn('unable to read from localStorage', err);
    return null;
  }
};

const writeLocalToken = (key: string, value: string | null) => {
  try {
    if (value) {
      localStorage.setItem(key, value);
    } else {
      localStorage.removeItem(key);
    }
  } catch (err) {
    console.warn('unable to write to localStorage', err);
  }
};

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(() => readLocalToken(TOKEN_KEY));
  const [refreshToken, setRefreshToken] = useState<string | null>(() => readLocalToken(REFRESH_KEY));
  const [loading, setLoading] = useState<boolean>(true);

  const refreshRef = useRef<string | null>(refreshToken);
  useEffect(() => {
    refreshRef.current = refreshToken;
  }, [refreshToken]);

  const clearState = useCallback(() => {
    setToken(null);
    setRefreshToken(null);
    setUser(null);
    clearAuthTokens();
    writeLocalToken(TOKEN_KEY, null);
    writeLocalToken(REFRESH_KEY, null);
  }, []);

  const persist = useCallback((auth: AuthResponse) => {
    setToken(auth.access_token);
    setRefreshToken(auth.refresh_token);
    setUser(auth.user);
    setAuthTokens(auth.access_token, auth.refresh_token);
    writeLocalToken(TOKEN_KEY, auth.access_token);
    writeLocalToken(REFRESH_KEY, auth.refresh_token);
  }, []);

  const refreshAccessToken = useCallback(async (): Promise<AuthResponse | null> => {
    const currentRefresh = refreshRef.current;
    if (!currentRefresh) {
      return null;
    }
    try {
      const updated = await refreshTokens(currentRefresh);
      persist(updated);
      return updated;
    } catch (err) {
      console.warn('refresh token failed', err);
      clearState();
      return null;
    }
  }, [persist, clearState]);

  useEffect(() => {
    setAuthTokens(token, refreshToken);
    setRefreshExecutor(refreshAccessToken);
  }, [token, refreshToken, refreshAccessToken]);

  useEffect(() => {
    const bootstrap = async () => {
      if (!token) {
        setLoading(false);
        return;
      }

      try {
        const profile = await fetchMe();
        setUser(profile);
      } catch (err: unknown) {
        const status = (err as AxiosError | undefined)?.response?.status;
        if (status === 401) {
          const refreshed = await refreshAccessToken();
          if (refreshed) {
            try {
              const profile = await fetchMe();
              setUser(profile);
            } catch (innerErr) {
              console.warn('failed to load profile after refresh', innerErr);
              clearState();
            }
          } else {
            clearState();
          }
        } else {
          console.warn('failed to bootstrap auth', err);
          clearState();
        }
      } finally {
        setLoading(false);
      }
    };

    bootstrap();
  }, [token, refreshAccessToken, clearState]);

  const handleLogin = useCallback(async (email: string, password: string) => {
    const auth = await apiLogin(email, password);
    persist(auth);
  }, [persist]);

  const handleRegister = useCallback(
    async (email: string, password: string, displayName: string) => {
      const auth = await apiRegister(email, password, displayName);
      persist(auth);
    },
    [persist]
  );

  const handleLogout = useCallback(async () => {
    try {
      const rt = getRefreshToken();
      if (rt) {
        await apiLogout(rt);
      }
    } catch (err) {
      console.warn('logout request failed', err);
    } finally {
      clearState();
    }
  }, [clearState]);

  const refreshProfile = useCallback(async () => {
    if (!token) return;
    const profile = await fetchMe();
    setUser(profile);
  }, [token]);

  const value = useMemo<AuthContextValue>(() => ({
    user,
    token,
    refreshToken,
    loading,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
    refreshProfile
  }), [user, token, refreshToken, loading, handleLogin, handleRegister, handleLogout, refreshProfile]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuthContext = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuthContext must be used inside AuthProvider');
  }
  return ctx;
};
