import styles from "./index.module.css";
import { Tag } from "../../types/account.types";

interface TabsProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  tags: Tag[];
}

export const Tabs = ({ activeTab, setActiveTab, tags }: TabsProps) => {

  const handleClick = (tab: string) => {
    setActiveTab(tab);
  };

  return (
    <div className={styles.tabs}>
      <div
        className={`${styles.tab} ${activeTab === "all" ? styles.active : ""}`}
        onClick={() => handleClick("all")}
      >
        Все вопросы
      </div>
      {tags.map((tag) => (
        <div
          className={`${styles.tab} ${activeTab === tag.name ? styles.active : ""}`}
          key={tag.id}
          onClick={() => handleClick(tag.name)}
        >
          {tag.name}
        </div>
      ))}
    </div>
  );
};
