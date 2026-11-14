import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { loginUser, logoutUser } from "../api/auth.api";
import { getCurrentUser } from "../../account/api/account.api";
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

  const login = async (login: string, password: string) => {
    // const data = await loginUser(login, password);
    // if(data) {
    const userProfile = await getCurrentUser();
    console.log("Введенные данные:", login, password);

    console.log("Полученный пользователь:", userProfile);
    if (userProfile) {
      setProfileUser(userProfile);
    }
    // }
  };

  const logout = async () => {
    await logoutUser();
    setProfileUser(null);
  };

  useEffect(() => {
    const restoreAuth = async () => {
      try {
        const userProfile = await getCurrentUser();
        setProfileUser(userProfile);
      } catch {
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
