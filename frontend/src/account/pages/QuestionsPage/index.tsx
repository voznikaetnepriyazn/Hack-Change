import { useState } from "react";
import styles from "./index.module.css";
import { useQuestionsStore } from "../../../stores/questionsStore";
import { QuestionBlock } from "../../components/QuestionBlock";
import { QuestionsList } from "../../components/QuestionsList";
import { Tabs } from "../../components/Tabs";
import { SearchInput } from "../../../shared/components/SearchInput";

export const QuestionsPage = () => {
  const [search, setSearch] = useState("");
  const questions = useQuestionsStore((state) => state.questions);
  const tags = useQuestionsStore((state) => state.tags);
  const contentBlock = useQuestionsStore((state) => state.contentBlock);
  const title = useQuestionsStore((state) => state.title);
  const description = useQuestionsStore((state) => state.description);

  const [activeTab, setActiveTab] = useState("all");
  const items =
  activeTab === "all"
    ? []
    : tags.find((tag) => tag.tab === activeTab)?.items ?? [];

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
            <SearchInput search={search} placeholder="Найти ответ на вопрос..." setSearch={setSearch} />
            <Tabs activeTab={activeTab} setActiveTab={setActiveTab} tags={tags} />
          </div>
          <div className={styles.left__content}>
            <QuestionsList questions={questions} items={items} search={search} />
          </div>
        </div>
        <div className={styles.right}>
          <QuestionBlock content={contentBlock} className={styles.right__content} />
        </div>
      </div>
    </div>
  );
};
