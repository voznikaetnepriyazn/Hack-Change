// src/router.tsx
import { createBrowserRouter, Navigate } from "react-router-dom";
import type { RouteObject } from "react-router-dom";
import { App } from "../App";
import { LoginPage } from "../auth/pages/LoginPage";
import { ProtectedRoute } from "../auth/components/ProtectedRoute";
import { ChatPage } from "../account/pages/ChatPage";
import { AccountLayout } from "../account/pages/AccountLayout";
import { QuestionsPage } from "../account/pages/QuestionsPage";
import { ProfilePage } from "../account/pages/ProfilePage";
import { DocumentsPage } from "../account/pages/DocumentsPage";
import { TasksPage } from "../account/pages/TasksPage";
import { InternPage } from "../account/pages/InternPage";

const routes: RouteObject[] = [
  {
    path: "/",
    element: <App />,
    children: [
      {
        index: true,
        element: <Navigate to="account" replace />,
      },
      {  path: "auth/login", element: <LoginPage /> },
      {
        path: "account",
        element: (
          <ProtectedRoute>
            <AccountLayout />
          </ProtectedRoute>
        ),
        children: [
          {
            index: true,
            element: <Navigate to="chat" replace />,
          },
          {
            path: "chat",
            element: <ChatPage />,
          },
          {
            path: "questions",
            element: <QuestionsPage />,
          },
          {
            path: "documents",
            element: <DocumentsPage />,
          },
          {
            path: "tasks",
            element: <TasksPage />,
          },
          {
            path: "intern",
            element: <InternPage />,
          },
          {
            path: "profile",
            element: <ProfilePage />,
          },
        ],
      },
    ],
  },
  {
    path: "*",
    element: <Navigate to="/account" />,
  },
];

export const router = createBrowserRouter(routes);
