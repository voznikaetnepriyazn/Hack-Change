import styles from "./index.module.css";
import { useProfileStore } from "../../../stores/profileStore";
import { ProfileItem } from "../../components/ProfileItem";

export const ProfilePage = () => {
  const title = useProfileStore((state) => state.title);
  const description = useProfileStore((state) => state.description);
  const user = useProfileStore((state) => state.user);

  return (
    <div className={styles.page}>
      <div className={styles.top}>
        <h1 className={styles.title}>{title}</h1>
        <p className={styles.description}>{description}</p>
      </div>
      <div className={styles.content}>
        <ProfileItem size="large" />
        <div className={styles.info}>
          <h4 className={styles.info__title}>Личная информация</h4>
          <ul className={styles.info__list}>
            <li className={styles.info__item}>
              <span className={styles.info__itemTitle}>Номер телефона</span>
              <span className={styles.info__itemValue}>{user.phone}</span>
            </li>
            <li className={styles.info__item}>
              <span className={styles.info__itemTitle}>Логин</span>
              <span className={styles.info__itemValue}>{user.login}</span>
            </li>
            <li className={styles.info__item}>
              <span className={styles.info__itemTitle}>Отдел</span>
              <span className={styles.info__itemValue}>{user.department}</span>
            </li>
            <li className={styles.info__item}>
              <span className={styles.info__itemTitle}>Должность</span>
              <span className={styles.info__itemValue}>{user.position}</span>
            </li>
          </ul>
        </div>
        <div className={styles.info}>
          <h4 className={styles.info__title}>Ближайший отпуск</h4>
          <p className={styles.info__vacation}>
            {user.vacationStart.toLocaleDateString()} - {user.vacationEnd.toLocaleDateString()}
          </p>
        </div>
      </div>
    </div>
  );
};
