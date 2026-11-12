import styles from "./index.module.css";
import { Tag } from "../../types/account.types";

interface TabsProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  tags: Tag[];
  setItems: (items: number[]) => void;
}

export const Tabs = ({ activeTab, setActiveTab, tags, setItems }: TabsProps) => {

  const handleClick = (tab: Tag) => {
    setActiveTab(tab.name);
    setItems(tab.items);
  };

  return (
    <div className={styles.tabs}>
      <div
        className={`${styles.tab} ${activeTab === "all" ? styles.active : ""}`}
        onClick={() => handleClick({ id: 0, name: "all", items: [] })}
      >
        Все вопросы
      </div>
      {tags.map((tag) => (
        <div
          className={`${styles.tab} ${activeTab === tag.name ? styles.active : ""}`}
          key={tag.id}
          onClick={() => handleClick(tag)}
        >
          {tag.name}
        </div>
      ))}
    </div>
  );
};
