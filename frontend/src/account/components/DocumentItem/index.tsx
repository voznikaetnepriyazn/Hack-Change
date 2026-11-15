import styles from "./index.module.css";
import { NavLink } from "react-router-dom";
import { Document } from "../../types/account.types";
import { IconArrow } from "../../../shared/components/icons/IconArrow";

interface DocumentItemProps {
  document: Document;
}
export const DocumentItem = ({
  document,
}: DocumentItemProps) => {
  return (
    <NavLink
      className={styles.wrapper}
      to={document.link}
      target="_blank"
    >
      <div className={styles.document}>
        <h4 className={styles.title}>{document.title}</h4>
        <p className={`${styles.description} ${styles.text}`}>
          {document.description}
        </p>
        <IconArrow className={styles.arrow} />
      </div>
    </NavLink>
  );
};
