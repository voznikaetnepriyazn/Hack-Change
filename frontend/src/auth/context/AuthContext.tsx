import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { loginUser, logoutUser, refreshUserToken } from "../api/auth.api";
import { type AuthUser } from "../types/auth.types";

export interface AuthContextType {
  user: AuthUser | null;
  accessToken: string | null;
  isAuthLoading: boolean;
  updateAccessToken: (token: string) => void;
  refreshAccessToken: () => Promise<string>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | null>(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === null) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [isAuthLoading, setIsAuthLoading] = useState(true);

  const login = async (email: string, password: string) => {
    const response = await loginUser(email, password);
    setUser(response.user);
    setAccessToken(response.accessToken);
  };

  const logout = async () => {
    await logoutUser();
    setUser(null);
    setAccessToken(null);
  };

  const updateAccessToken = (token: string) => {
    setAccessToken(token);
  };

  const refreshAccessToken = async (): Promise<string> => {
    try {
      const response = await refreshUserToken();
    
      setUser(response.user);
      setAccessToken(response.accessToken);
    
      return response.accessToken;
    } catch (error) {
      setUser(null);
      setAccessToken(null);
      throw error;
    } 
  };

  useEffect(() => {
    const restoreAuth = async () => {
      try {
        const response = await refreshUserToken();
  
        setUser(response.user);
        setAccessToken(response.accessToken);
      } catch {
        setUser(null);
        setAccessToken(null);
      } finally {
        setIsAuthLoading(false);
      }
    };
  
    restoreAuth();
  }, []);

  return <AuthContext.Provider value={{ user, accessToken, isAuthLoading, login, logout, updateAccessToken, refreshAccessToken }}>{children}</AuthContext.Provider>
};