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
      {tags.map((tag) => (
        <div
          className={`${styles.tab} ${activeTab === tag.tab ? styles.active : ""}`}
          key={tag.id}
          onClick={() => handleClick(tag.tab)}
        >
          {tag.name}
        </div>
      ))}
    </div>
  );
};
