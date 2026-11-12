import styles from "./index.module.css";
import { useProfileStore } from "../../../stores/profileStore";

interface ProfileItemProps {
  className?: string;
  isShort?: boolean;
  size?: "small" | "large";
}

export const ProfileItem = ({className, isShort, size = "small"}: ProfileItemProps) => {
  const user = useProfileStore((state) => state.user);
  const getFirstWord = (name: string) => {
    return name[0];
  };

  return (
    <div className={`${styles.profile} ${className} ${isShort ? styles.short : ""} ${ styles[size]}`}>
      <div className={styles.avatar}>
        {user.avatar ? (
          <img src={user.avatar} alt="avatar" />
        ) : (
          getFirstWord(user.name)
        )}
      </div>
      {!isShort && (
      <div className={styles.right}>
        <span className={styles.name}>{user.name}</span>
        <span className={styles.position}>{user.position}</span>
      </div>
      )}
    </div>
  );
};
