import { useAuth } from "../context/AuthContext";
import { Navigate } from "react-router-dom";
import { type ReactNode } from "react";


export const ProtectedRoute = ({ children }: { children: ReactNode }) => {
  const { user , isAuthLoading } = useAuth();
  if (isAuthLoading) {
    return <div>Loading...</div>;
  }
  if (!user) {
    return <Navigate to="/auth/login" replace/>;
  }
  return children;
};