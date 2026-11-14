import styles from "./index.module.css";
import { useProfileStore } from "../../../stores/profileStore";
import { ProfileItem } from "../../components/ProfileItem";
import { NavLink } from "react-router-dom";

export const ProfilePage = () => {
  const title = useProfileStore((state) => state.title);
  const description = useProfileStore((state) => state.description);
  const user = useProfileStore((state) => state.user);

  const statusRequest = (status: string) => {
    switch (status) {
      case "pending":
        return "На согласовании";
      case "approved":
        return "Одобрено";
      case "rejected":
        return "Отклонено";
      default:
        return "Неизвестно";
    }
  };

  return (
    <div className={styles.page}>
      <div className={styles.top}>
        <h1 className={styles.title}>{title}</h1>
        {user?.role && (
          <p className={styles.description}>{description[user?.role]}</p>
        )}
      </div>
      <div className={styles.content}>
        <ProfileItem size="large" user={user} />
        <div className={styles.info}>
          <h4 className={styles.info__title}>Личная информация</h4>
          <ul className={styles.info__list}>
            {user?.phone && (
              <li className={styles.info__item}>
                <span className={styles.info__itemTitle}>Номер телефона</span>
                <span className={styles.info__itemValue}>{user?.phone}</span>
              </li>
            )}
            {user?.login && (
              <li className={styles.info__item}>
                <span className={styles.info__itemTitle}>Логин</span>
                <span className={styles.info__itemValue}>{user?.login}</span>
              </li>
            )}
            {user?.command && (
              <li className={styles.info__item}>
                <span className={styles.info__itemTitle}>Команда</span>
                <span className={styles.info__itemValue}>{user?.command}</span>
              </li>
            )}
            {user?.post && (
              <li className={styles.info__item}>
                <span className={styles.info__itemTitle}>Должность</span>
                <span className={styles.info__itemValue}>{user?.post}</span>
              </li>
            )}
            {user?.formatWork && (
              <li className={styles.info__item}>
                <span className={styles.info__itemTitle}>Формат работы</span>
                <span className={styles.info__itemValue}>
                  {user?.formatWork}
                </span>
              </li>
            )}
            {user?.office && (
              <li className={styles.info__item}>
                <span className={styles.info__itemTitle}>Офис</span>
                <span className={styles.info__itemValue}>{user?.office}</span>
              </li>
            )}
            {user?.statusVacation && (
              <li className={styles.info__item}>
                <span className={styles.info__itemTitle}>
                  Статус стажировки
                </span>
                <span className={styles.info__itemValue}>
                  {user?.statusVacation}
                </span>
              </li>
            )}
          </ul>
        </div>
        {user?.githubProfile && (
          <div className={styles.info}>
            <h4 className={styles.info__title}>github-профиль</h4>
            <ul className={styles.info__list}>
              {user?.githubProfile?.login && (
                <li className={styles.info__item}>
                  <span className={styles.info__itemTitle}>login</span>
                  <span className={styles.info__itemValue}>
                    {user?.githubProfile?.login}
                  </span>
                </li>
              )}
              {user?.githubProfile?.username && (
                <li className={styles.info__item}>
                  <span className={styles.info__itemTitle}>username</span>
                  <span className={styles.info__itemValue}>
                    {user?.githubProfile?.username}
                  </span>
                </li>
              )}
            </ul>
          </div>
        )}
        {user?.role === "intern" && (
          <div className={styles.requests}>
            <h4 className={styles.info__title}>Мои заявки и статусы</h4>
            {user?.requests && user?.requests.length > 0 ? (
              <ul className={styles.requests__list}>
                <li
                  className={`${styles.requests__item} ${styles.requests__itemHeader}`}
                >
                  <span className={styles.requests__itemValue}>Заявка</span>
                  <span className={styles.requests__itemValue}>Статус</span>
                  <span className={styles.requests__itemValue}>
                    Комментарий
                  </span>
                </li>
                {user?.requests.map((request) => (
                  <li className={`${styles.requests__item}`} key={request.id}>
                    <span className={styles.requests__itemValue}>
                      {request.name}
                    </span>
                    <span
                      className={`${styles.requests__itemValue} ${styles.requests__itemValueStatus}`}
                    >
                      {statusRequest(request.status)}
                    </span>
                    <span className={styles.requests__itemValue}>
                      {request.description}
                    </span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className={styles.requests__vacation}>Нет заявок</p>
            )}
          </div>
        )}
        {user?.role === "intern" && (
          <div className={styles.info}>
            <h4 className={styles.info__title}>Ближайший отпуск</h4>
            {user?.vacationStart && user?.vacationEnd ? (
              <p className={styles.info__vacation}>
                {user?.vacationStart?.toLocaleDateString()} -{" "}
                {user?.vacationEnd?.toLocaleDateString()}
              </p>
            ) : (
              <p className={styles.info__vacation}>
                Не выбрано{" "}
                <NavLink
                  to={"/account/chat"}
                  className={`${styles.text} ${styles.link}`}
                >
                  (Подберите подходящие даты с AI-Buddy)
                </NavLink>
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
