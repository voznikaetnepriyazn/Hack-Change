import styles from "./index.module.css";
import { NavLink } from "react-router-dom";
import { Question } from "../../types/account.types";
import { IconArrowSelect } from "../../../shared/components/icons/IconArrowSelect";

interface QuestionItemProps {
  question: Question;
  isActive: boolean;
  onClick: () => void;
}
export const QuestionItem = ({
  question,
  isActive,
  onClick,
}: QuestionItemProps) => {
  return (
    <div className={`${styles.wrapper} ${isActive ? styles.active : ""}`} onClick={onClick}>
      <div className={styles.question}>
        <h4 className={styles.title}>{question.title}</h4>
        <p className={`${styles.description} ${styles.text}`}>
          {question.description}
        </p>
        <IconArrowSelect className={styles.arrow} />
      </div>
      <div className={styles.answer}>
        <div className={styles.answer__content}>
          <p className={`${styles.text} ${styles.answer__text}`}>{question.answer}</p>

          <NavLink
            to={`/account/chat?question=${encodeURIComponent(question.title)}`}
            className={`${styles.text} ${styles.link}`}
            onClick={(e) => e.stopPropagation()}
          >
            Получите более подробный ответ с AI-Buddy
          </NavLink>
        </div>
      </div>
    </div>
  );
};
