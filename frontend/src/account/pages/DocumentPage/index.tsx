
import { useParams } from "react-router-dom";
import { useDocumentsStore } from "../../../stores/documentsStore";
import { useEffect, useState } from "react";
import { getDocumentContent } from "../../api/document.api";
import { DocumentContent } from "../../components/ReactMarkdown";
import { Loader } from "../../../shared/components/UI/Loader";
import styles from "./index.module.css";

export const DocumentPage = () => {
  const { documentId } = useParams();
  const id = Number(documentId);

  const documents = useDocumentsStore((state) => state.documents);
  const document = documents.find((item) => item.id === id);

  const [content, setContent] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    setError(null);
    setContent("");

    if (!documentId || !Number.isInteger(id) || id <= 0) {
      setError("Некорректный ID документа");
      setIsLoading(false);

      return () => {
        cancelled = true;
      };
    }

    if (!document) {
      setError("Документ не найден");
      setIsLoading(false);

      return () => {
        cancelled = true;
      };
    }

    setIsLoading(true);

    getDocumentContent(document.link)
      .then((content) => {
        if (!cancelled) {
          setContent(content);
        }
      })
      .catch((error: unknown) => {
        if (!cancelled) {
          setError(
            error instanceof Error
              ? error.message
              : "Неизвестная ошибка"
          );
        }
      })
      .finally(() => {
        if (!cancelled) {
          setIsLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [documentId, id, document]);

  if (isLoading) {
    return <Loader />;
  }

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <div className={styles.page}>
      <DocumentContent content={content} />
    </div>
  );
};
