import styles from "./index.module.css";
import { HackButton } from "../../../shared/components/UI/HackButton";
import { useNavigate } from "react-router-dom";

export const QuestionsPage = () => {
  const navigate = useNavigate();
  return (
    <div className={styles.page}>
      <div className={styles.top}>
        <h1 className={styles.title}>Частые вопросы</h1>
        <p className={styles.description}>
          Быстрые ответы на популярные вопросы о стажировке и работе в Авито.
        </p>
      </div>
      <div className={styles.content}>
        <div className={styles.left}></div>
        <div className={styles.right}>
          <h4 className={styles.right__title}>Не нашли ответ?</h4>
          <p className={styles.right__description}>
            Спросите AI Buddy — он поможет разобраться и подскажет, где найти
            информацию.
          </p>
          <HackButton
            className={styles.right__button}
            onClick={() => {
              navigate("/account/chat");
            }}
          >
            Задать вопрос AI
          </HackButton>
        </div>
      </div>
    </div>
  );
};
