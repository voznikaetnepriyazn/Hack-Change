import { useState, useEffect } from "react";
import styles from "./index.module.css";
import { useTasksStore } from "../../../stores/tasksStore";
import { QuestionBlock } from "../../components/QuestionBlock";
import { TasksList } from "../../components/TasksList";
import { Tabs } from "../../components/Tabs";
import { SearchInput } from "../../../shared/components/SearchInput";

export const TasksPage = () => {
  const [search, setSearch] = useState("");
  const tasks = useTasksStore((state) => state.tasks);
  const tags = useTasksStore((state) => state.tags);
  const contentBlock = useTasksStore((state) => state.contentBlock);
  const title = useTasksStore((state) => state.title);
  const description = useTasksStore((state) => state.description);

  const [activeTab, setActiveTab] = useState("all");

  const filteredTasks =
  activeTab === "all"
    ? tasks
    : tasks.filter((task) => task.status === activeTab);


  return (
    <div className={styles.page}>
      <div className={styles.top}>
        <h1 className={styles.title}>{title}</h1>
        <p className={styles.description}>
          {description}
        </p>
      </div>
      <div className={styles.content}>
        <div className={styles.left}>
          <div className={styles.left__top}>
            <SearchInput search={search} placeholder="Поиск по задачам..." setSearch={setSearch} />
            <Tabs activeTab={activeTab} setActiveTab={setActiveTab} tags={tags} />
          </div>
          <div className={styles.left__content}>
            <TasksList tasks={filteredTasks} search={search} />
          </div>
        </div>
        <div className={styles.right}>
          <QuestionBlock content={contentBlock} className={styles.right__content} />
        </div>
      </div>
    </div>
  );
};
