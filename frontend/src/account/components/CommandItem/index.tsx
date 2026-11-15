import styles from "./index.module.css";
import { EmployeeCard } from "../../types/account.types";

interface CommandItemProps {
  employee: EmployeeCard;
}

export const CommandItem = ({ employee }: CommandItemProps) => {
  const getFirstWord = (name: string) => {
    return name[0];
  };
  return (
    <div className={styles.employee}>
      <div className={styles.top}>
        <div className={styles.avatar}>
          {employee.avatar ? (
            <img src={employee.avatar} alt="avatar" />
          ) : (
            getFirstWord(employee.full_name || "")
          )}
        </div>
        <div className={styles.right}>
          {employee.full_name && (
            <h2 className={styles.name}>{employee.full_name}</h2>
          )}
          {employee.position && (
            <p className={styles.position}>{employee.position}</p>
          )}
          {employee.email && (
            <p className={styles.email}>{employee.email}</p>
          )}
        </div>
      </div>
      <div className={styles.bottom}>
        <p className={styles.bottomTitle}>С чем можно обратиться</p>
        <div className={styles.bottomItems}>
          {employee.responsibilities && (
            <p className={styles.bottomItem}>{employee.responsibilities}</p>
          )}
        </div>
      </div>
    </div>
  );
};
