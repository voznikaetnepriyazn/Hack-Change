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
    if (activeQuestion === id) {
      setActiveQuestion(null);
    } else {
    setActiveQuestion(id);
    }
  };

  const filteredQuestions = () => {
    if (items.length === 0) {
      return questions.filter((question) => question.answer.toLowerCase().includes(search.toLowerCase()));
    }
    return questions.filter((question) => items.includes(question.id) && question.answer.toLowerCase().includes(search.toLowerCase()));
  };


  return (
    <div className={styles.list}>
      {filteredQuestions().map((question) => (
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
