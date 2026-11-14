import { useAuth } from "../context/AuthContext";
import { Navigate } from "react-router-dom";
import { type ReactNode } from "react";
import { useProfileStore } from "../../stores/profileStore";
import { Loader } from "../../shared/components/UI/Loader";


export const ProtectedRoute = ({ children }: { children: ReactNode }) => {
  const { isAuthLoading } = useAuth();
  const user = useProfileStore((state) => state.user);
  if (isAuthLoading) {
    return <Loader />;
  }
  if (!user) {
    return <Navigate to="/auth/login" replace/>;
  }
  return children;
};