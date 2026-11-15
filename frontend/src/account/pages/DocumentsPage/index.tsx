import { useState } from "react";
import styles from "./index.module.css";
import { useDocumentsStore } from "../../../stores/documentsStore";
import { QuestionBlock } from "../../components/QuestionBlock";
import { DocumentsList } from "../../components/DocumentsList";
import { Tabs } from "../../components/Tabs";
import { SearchInput } from "../../../shared/components/SearchInput";

export const DocumentsPage = () => {
  const [search, setSearch] = useState("");
  const documents = useDocumentsStore((state) => state.documents);
  const tags = useDocumentsStore((state) => state.tags);
  const contentBlock = useDocumentsStore((state) => state.contentBlock);
  const title = useDocumentsStore((state) => state.title);
  const description = useDocumentsStore((state) => state.description);

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
            <SearchInput search={search} placeholder="Поиск по документам..." setSearch={setSearch} />
            <Tabs activeTab={activeTab} setActiveTab={setActiveTab} tags={tags} />
          </div>
          <div className={styles.left__content}>
            <DocumentsList documents={documents} items={items} search={search} />
          </div>
        </div>
        <div className={styles.right}>
          <QuestionBlock content={contentBlock} className={styles.right__content} />
        </div>
      </div>
    </div>
  );
};
