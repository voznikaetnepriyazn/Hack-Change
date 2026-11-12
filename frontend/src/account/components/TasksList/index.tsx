import styles from "./index.module.css";
import { useState } from "react";
import { TaskItem } from "../TaskItem";
import { Task } from "../../types/account.types";

interface TasksListProps {
  tasks: Task[];
  search: string;
}

export const TasksList = ({ tasks, search }: TasksListProps) => {
  const [activeTask, setActiveTask] = useState<number | null>(null);


  const handleClick = (id: number) => {
    setActiveTask((prev) => (prev === id ? null : id));
  };


  const filteredTasks = tasks.filter((task) => {  
    const matchesSearch = [
      task.title,
      task.description,
    ].some((value) =>
      value.toLowerCase().includes(search.toLowerCase())
    );
  
    return matchesSearch;
  });


  return (
    <div className={styles.list}>
        {filteredTasks.map((task) => (
        <TaskItem
          key={task.id}
          task={task}
          isActive={activeTask === task.id}
          onClick={() => handleClick(task.id)}
        />
      ))}
    </div>
  );
};
