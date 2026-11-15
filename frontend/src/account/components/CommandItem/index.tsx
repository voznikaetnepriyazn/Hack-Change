import styles from "./index.module.css";
import { Teammate } from "../../types/account.types";


interface CommandItemProps {
  teammate: Teammate;
}

export const CommandItem = ({ teammate }: CommandItemProps) => {
  const getFirstWord = (name: string) => {
    return name[0];
  };
  return (
    <div>
      <div className={styles.avatar}>
        {user?.avatar ? (
          <img src={user?.avatar} alt="avatar" />
        ) : (
          getFirstWord(user?.name || "")
        )}
      </div>
    </div>
  );
};
