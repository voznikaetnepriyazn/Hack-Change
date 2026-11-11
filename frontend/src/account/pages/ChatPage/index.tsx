import styles from "./index.module.css";
import { useEffect, useRef, useState } from "react";
import { IconRocket } from "../../../assets/images/IconRocket";
import { IconQuestionCircle } from "../../../assets/images/IconQuestionCircle";
import { IconInfo } from "../../../assets/images/IconInfo";
import { IconTeam } from "../../../assets/images/IconTeam";
import { IconArrow } from "../../../assets/images/IconArrow";
import { IconSend } from "../../../assets/images/IconSend";

const buttons = [
  {
    icon: <IconRocket />,
    label: "Расскажи о моей стажировке",
  },
  {
    icon: <IconQuestionCircle />,
    label: "Найди ответ на частый вопрос",
  },
  {
    icon: <IconInfo />,
    label: "Дай рекомендации для адаптации",
  },
  {
    icon: <IconTeam />,
    label: "Расскажи о моей команде",
  },
];

interface Message {
  text: string;
}

export const ChatPage = () => {


  const [question, setQuestion] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const chatRef = useRef<HTMLDivElement>(null);


  const handleQuestionSubmit = () => {
    const text = question.trim();
  
    if (!text) return;
  
    setMessages((prev) => [...prev, { text }]);
    setQuestion("");
  };

  useEffect(() => {
    if (chatRef.current) {
      chatRef.current.scrollTop = chatRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <div className={styles.container}>
      <div ref={chatRef} className={styles.chat}>
        <h1 className={styles.title}>Привет, Имя!</h1>
  
        <p className={styles.description}>
          Я AI-Budddy, твой помощник в Авито.
          <br />
          Задай вопрос или выбери, с чего начать:
        </p>
  
        <div className={`${styles.buttons} ${messages.length > 0 ? styles.spaceBottom : ''}`}>
          {buttons.map((button) => (
            <button
              type="button"
              key={button.label}
              className={styles.button}
              onClick={() => setQuestion(button.label)}
            >
              {button.icon}
              <p className={styles.button__label}>{button.label}</p>
              <IconArrow className={styles.arrow} />
            </button>
          ))}
        </div>
  
        {messages.length > 0 && (
          <div className={styles.messages}>
            {messages.map((message, index) => (
              <div key={index} className={styles.message}>
                <p className={styles.message__text}>{message.text}</p>
              </div>
            ))}
          </div>
        )}
      </div>
  
      <div className={styles.bottom__container}>
        <div className={styles.bottom}>
          <input
            type="text"
            name="question"
            placeholder="Напиши свой вопрос"
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
          />
  
          <button type="button" onClick={handleQuestionSubmit}>
            <IconSend className={styles.arrow} />
          </button>
        </div>
      </div>
    </div>
  );
};
