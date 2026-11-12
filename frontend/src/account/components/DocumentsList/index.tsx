import styles from "./index.module.css";
import { useState } from "react";
import { DocumentItem } from "../DocumentItem";
import { Document } from "../../types/account.types";

interface DocumentsListProps {
  documents: Document[];
  items: number[];
  search: string;
}

export const DocumentsList = ({ documents, items, search }: DocumentsListProps) => {


  const filteredDocuments = documents.filter((document) => {
    const matchesTab = items.length === 0 || items.includes(document.id);
  
    const matchesSearch = [
      document.title,
      document.description,
    ].some((value) =>
      value.toLowerCase().includes(search.toLowerCase())
    );
  
    return matchesTab && matchesSearch;
  });


  return (
    <div className={styles.list}>
      {filteredDocuments.map((document) => (
        <DocumentItem
          key={document.id}
          document={document}
        />
        
      ))}
    </div>
  );
};
