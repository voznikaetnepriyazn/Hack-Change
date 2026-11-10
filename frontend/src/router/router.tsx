// src/router.tsx
import { createBrowserRouter, Navigate } from "react-router-dom";
import type { RouteObject } from "react-router-dom";
import { App } from "../App";
import { LoginPage } from "../auth/pages/LoginPage";
import { ProtectedRoute } from "../auth/components/ProtectedRoute";
import { AccountPage } from "../user/pages/AccountPage";

// Аннотация типа гарантирует, что поля path, element, children и т.д. будут проверены
const routes: RouteObject[] = [
  {
    path: "/",
    element: <App />, // Родительский компонент-макет, внутри должен быть <Outlet />
    children: [
      {
        path: "account",
        element: (
            // <ProtectedRoute>
            //   <AccountPage />
            // </ProtectedRoute>
          <AccountPage />
        ),
      },
      { path: "auth/login", element: <LoginPage /> },
    ],
  },
  {
    path: "*", // Маршрут 404
    element: <Navigate to="/account" />,
  },
];

export const router = createBrowserRouter(routes);
