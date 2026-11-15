
import ReactMarkdown from "react-markdown";
import styles from "./index.module.css";

interface DocumentContentProps {
  content: string;
}

export function DocumentContent({ content }: DocumentContentProps) {
  return (
    <article className={styles.markdown}>
      <ReactMarkdown>{content}</ReactMarkdown>
    </article>
  );
}
