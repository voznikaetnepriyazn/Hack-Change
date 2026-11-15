import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { logoutUser, loginUser } from "../api/auth.api";
import { getCurrentUser } from "../../account/api/account.api";
import { clearSessionChat } from "../../account/chat/sessionChat";
import { useProfileStore } from "../../stores/profileStore";

export interface AuthContextType {
  isAuthLoading: boolean;
  login: (login: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | null>(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === null) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [isAuthLoading, setIsAuthLoading] = useState(true);
  const setProfileUser = useProfileStore((state) => state.setProfileUser);
  const user = useProfileStore((state) => state.user);

  const login = async (login: string, password: string) => {
    try {
      const data = await loginUser(login, password);
      clearSessionChat();
      if (data) {
        const { user } = data;
        const userProfile = await getCurrentUser(user.role);
        if (userProfile) {
          setProfileUser(userProfile);
        }
      }
    } catch (error) {
      throw new Error("Failed to login");
    }
  };

  const logout = async () => {
    await logoutUser();
    clearSessionChat();
    setProfileUser(null);
  };

  useEffect(() => {
    const restoreAuth = async () => {
      try {
        // if (user) {
          const userProfile = await getCurrentUser();
          setProfileUser(userProfile);
        // }
      } catch {
        clearSessionChat();
        setProfileUser(null);
      } finally {
        setIsAuthLoading(false);
      }
    };

    restoreAuth();
  }, []);

  return (
    <AuthContext.Provider value={{ isAuthLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
