import { useCallback, useEffect, useState } from "react";
import { useAuth } from "../../auth/context/AuthContext";
import { useNavigate } from "react-router-dom";
import { getCurrentUser } from "../../user/api/user.api";
import { AuthUser } from "../../auth/types/auth.types";

export function AccountPage() {
  const [profileError, setProfileError] = useState<string | null>(null);
  const [currentUser, setCurrentUser] = useState<AuthUser | null>(null);
  const { user, logout, accessToken, refreshAccessToken } = useAuth();
  const navigate = useNavigate();

  const loadCurrentUser = useCallback(async () => {
    setProfileError(null);
    try {
      if (accessToken) {
        const user = await getCurrentUser(accessToken, refreshAccessToken);
        setCurrentUser(user);
      }
    } catch (error) {
      setCurrentUser(null);
      setProfileError('Не удалось получить информацию о пользователе');
    }
  }, [accessToken, refreshAccessToken])


  useEffect(() => {
    loadCurrentUser();
  }, [loadCurrentUser]);

  return (
    <main className="page">
      <h1>Аккаунт</h1>
      <section className="panel" aria-live="polite">
        <div className="row">
          <span>Профиль</span>
          <strong className={profileError ? 'bad' : currentUser ? 'good' : ''}>
            {profileError ? 'Недоступен' : currentUser?.email}
          </strong>
        </div>
      </section>
      <div className="row">
        <button type="button" onClick={() => {
          void loadCurrentUser();
        }}>
          Проверить снова
        </button>
      </div>
      {user && (
        <section className="panel">
          <p>Вы вошли как {user.email} {currentUser ? `(${currentUser.email})` : ''}</p>
          <button type="button" onClick={async () => {
            await logout();
            navigate('/auth/login');
          }}>Выйти</button>
        </section>
      )}
    </main>
  )
}