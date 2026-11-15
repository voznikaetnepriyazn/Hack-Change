import styles from "./index.module.css";
import { User } from "../../types/account.types";
import { useProfileStore } from "../../../stores/profileStore";
interface ProfileItemProps {
  className?: string;
  isShort?: boolean;
  size?: "small" | "large";
  user: User | null;
}

export const ProfileItem = ({
  className,
  isShort,
  size = "small",
  user,
}: ProfileItemProps) => {
  const role = useProfileStore((state) => state.user?.role);
  const getFirstWord = (name: string) => {
    return name[0];
  };

  return (
    <div
      className={`${styles.profile} ${className} ${isShort ? styles.short : ""} ${styles[size]}`}
    >
      <div className={styles.avatar}>
        {user?.avatar ? (
          <img src={user?.avatar} alt="avatar" />
        ) : (
          getFirstWord(user?.name || "")
        )}
      </div>
      <div className={`${styles.right} ${isShort ? styles.hidden : ""}`}>
        <span className={styles.name}>{user?.name}</span>
        <p className={styles.bottom}>
          {size === "large" && user?.role === "intern" && role 
          !== "intern" && <span>{user?.post}</span>}
          <span className={styles.position}>{user?.position}</span>
        </p>
      </div>
    </div>
  );
};
