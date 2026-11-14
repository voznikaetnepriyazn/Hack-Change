import styles from "./index.module.css";
import { useInternStore } from "../../../stores/internStore";
import { ProfileItem } from "../../components/ProfileItem";
import { getIntern } from "../../api/account.api";
import { useEffect, useState } from "react";
import { Loader } from "../../../shared/components/UI/Loader";

export const InternPage = () => {
  const title = useInternStore((state) => state.title);
  const description = useInternStore((state) => state.description);
  const user = useInternStore((state) => state.user);
  const setIntern = useInternStore((state) => state.setInternUser);
  const progress = useInternStore((state) => state.progress);
  const metrics = useInternStore((state) => state.metrics);
  const themesinQuestions = useInternStore((state) => state.themesinQuestions);
  const [isLoading, setIsLoading] = useState(() => !user);
  const [error, setError] = useState<string | null>(null);

  const getInternData = async () => {
    if (user) return;
  
    try {
      setIsLoading(true);
      setError(null);
  
      const intern = await getIntern();
  
      if (!intern) {
        throw new Error("Данные стажёра не найдены");
      }
  
      setIntern(intern);
    } catch (error) {
      setError(
        error instanceof Error
          ? error.message
          : "Не удалось загрузить данные стажёра"
      );
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    getInternData();
  }, [user]);

  if (isLoading) {
    return <Loader />;
  }
  
  if (error) {
    return <p>{error}</p>;
  }
  
  if (!user) {
    return <p>Данные стажёра отсутствуют.</p>;
  }

  return (
    <div className={styles.page}>
      <div className={styles.top}>
        <h1 className={styles.title}>{title}</h1>
        {user?.role && <p className={styles.description}>{description}</p>}
      </div>
      <div className={styles.content}>
        <ProfileItem size="large" user={user} />
        <div className={styles.info}>
          <h4 className={styles.info__title}>Прогресс стажёра</h4>
          <ul className={styles.progress__list}>
            <li className={styles.progress__item}>
              <span>День стажировки:</span>
              <span>{progress.day}</span>
            </li>
            <li className={styles.progress__item}>
              <span>Текущий этап:</span>
              <span>{progress.step}</span>
            </li>
          </ul>
        </div>
        <div className={styles.info}>
          <h4 className={styles.info__title}>Метрики агента</h4>
          <div className={styles.metrics__items}>
            <div className={styles.metrics__item}>
              <h5 className={styles.metricsTitle}>Запросов:</h5>
              <p className={styles.metricsValue}>{metrics.requests}</p>
            </div>
            <div className={styles.metrics__item}>
              <h5 className={styles.metricsTitle}>Ответов без эскалации:</h5>
              <p className={styles.metricsValue}>{metrics.answersWithoutEscalation}</p>
            </div>
            <div className={styles.metrics__item}>
              <h5 className={styles.metricsTitle}>Эскалаций:</h5>
              <p className={styles.metricsValue}>{metrics.escalation}</p>
            </div>
          </div>
        </div>
        <h4 className={styles.info__title}>
          Самые частые темы запросов:{" "}
          {themesinQuestions?.map((theme) => theme.name).join(", ")}
        </h4>
      </div>
    </div>
  );
};
