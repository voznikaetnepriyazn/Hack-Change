import styles from "./index.module.css";
import ReactMarkdown from "react-markdown";
import { Task } from "../../types/account.types";
import { IconCheck } from "../../../shared/components/icons/IconCheck";
import { NavLink } from "react-router-dom";
import { IconArrow } from "../../../shared/components/icons/IconArrow";
import { useTasksStore } from "../../../stores/tasksStore";

interface TaskItemProps {
  task: Task;
  isActive: boolean;
  onClick: () => void;
}
export const TaskItem = ({ task, isActive, onClick }: TaskItemProps) => {
  const changeTaskStatus = useTasksStore((state) => state.changeTaskStatus);
  const deadline = () => {
    const today = new Date();
    const deadline = task.deadline;
    if (deadline.getDate() === today.getDate()) {
      return "Дедлайн: сегодня";
    }
    if (deadline < today) {
      return "Дедлайн: просрочено";
    }
    return "Дедлайн: " + deadline.toLocaleDateString("ru-RU");
  };
  return (
    <div className={styles.wrapper}>
      <div className={styles.head}>
        <p className={styles.code}>{task.code}</p>
        <h4 className={styles.title}>{task.title}</h4>
      </div>
      {isActive && (
        <div className={styles.markdown}>
          <ReactMarkdown>{task.description}</ReactMarkdown>
        </div>
      )}
      <p className={`${styles.deadline} ${styles.text}`}>{deadline()}</p>
      {!isActive && (
        <IconCheck
          status={task.status}
          className={styles.status}
          onClick={() => changeTaskStatus(task.id)}
        />
      )}
      <div className={styles.bottom}>
        <div className={styles.hide} onClick={onClick}>
          {!isActive && "Открыть задачу"}
          <IconArrow
            className={`${styles.arrow} ${isActive ? styles.active : ""}`}
          />
          {isActive && "Скрыть задачу"}
        </div>
        {isActive && (
          <NavLink
            to={`/account/chat?task=${encodeURIComponent(task.id)}`}
            className={`${styles.text} ${styles.link}`}
            onClick={(e) => e.stopPropagation()}
          >
            Попросите помочь AI-Buddy
          </NavLink>
        )}
      </div>
    </div>
  );
};
