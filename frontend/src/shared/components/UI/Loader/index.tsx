
import styles from "./index.module.css";


export const Loader = () => {
  return (
    <div className={styles.loader}>
      <div className={styles.content}>
        <div className={styles.logo}>
          <span className={styles.ring} />
          <span className={styles.ringInner} />
          <span className={styles.dot} />
        </div>

        <p className={styles.title}>AI-BUDDY</p>

        <p className={styles.message}>Загрузка...</p>

        <div className={styles.progress}>
          <span />
        </div>
      </div>
    </div>
  );
};
