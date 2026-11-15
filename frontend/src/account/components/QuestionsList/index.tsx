import styles from "./index.module.css";
import { useState } from "react";
import { QuestionItem } from "../QuestionItem";
import { Question } from "../../types/account.types";

interface QuestionsListProps {
  questions: Question[];
  items: number[];
  search: string;
}

export const QuestionsList = ({ questions, items, search }: QuestionsListProps) => {
  const [activeQuestion, setActiveQuestion] = useState<number | null>(null);


  const handleClick = (id: number) => {
    setActiveQuestion((prev) => (prev === id ? null : id));
  };


  const filteredQuestions = questions.filter((question) => {
    const matchesTab = items.length === 0 || items.includes(question.id);
  
    const matchesSearch = [
      question.title,
      question.description,
      question.answer,
    ].some((value) =>
      value.toLowerCase().includes(search.toLowerCase())
    );
  
    return matchesTab && matchesSearch;
  });


  return (
    <div className={styles.list}>
      {filteredQuestions.map((question) => (
        <QuestionItem
          key={question.id}
          question={question}
          isActive={activeQuestion === question.id}
          onClick={() => handleClick(question.id)}
        />
      ))}
    </div>
  );
};
