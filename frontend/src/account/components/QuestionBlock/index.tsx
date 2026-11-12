import styles from "./index.module.css";
import { HackButton } from "../../../shared/components/UI/HackButton";
import { useNavigate } from "react-router-dom";

interface QuestionBlockProps {
  content: {
    title: string;
    description: string;
  }
  className?: string;
}

export const QuestionBlock = ({ content, className }: QuestionBlockProps) => {

  const navigate = useNavigate();
  return (
    <div className={`${styles.wrapper} ${className}`}>
      <h4 className={styles.title}>{content.title}</h4>
      <p className={styles.description}>
        {content.description}
      </p>
      <HackButton
        className={styles.button}
        onClick={() => {
          navigate(`/account/chat?question=${content.title}`);
        }}
      >
        Задать вопрос AI
      </HackButton>
    </div>
  );
};
