import styles from "./index.module.css";
import { useEffect, useRef, useState } from "react";
import { streamChat } from "../../api/account.api";
import { useProfileStore } from "../../../stores/profileStore";
import { IconRocket } from "../../../shared/components/icons/IconRocket";
import { IconQuestionCircle } from "../../../shared/components/icons/IconQuestionCircle";
import { IconInfo } from "../../../shared/components/icons/IconInfo";
import { IconTeam } from "../../../shared/components/icons/IconTeam";
import { IconArrow } from "../../../shared/components/icons/IconArrow";
import { IconSend } from "../../../shared/components/icons/IconSend";


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
  role?: "user" | "bot";
}

export const ChatPage = () => {
  const user = useProfileStore((state) => state.user);

  const [question, setQuestion] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const chatRef = useRef<HTMLDivElement>(null);
  const stopStreamRef = useRef<(() => void) | null>(null);

  const getMessageBtnCommand = (text: string) => {
    setQuestion(text);
  };

  const handleQuestionSubmit = () => {
    const text = question.trim();

    if (!text || isStreaming) {
      return;
    }

    if (!user || user.role !== "intern") {
      setNotice("Чат доступен только стажёру");
      return;
    }

    setNotice(null);
    setQuestion("");
    setIsStreaming(true);
    setMessages((prev) => [
      ...prev,
      { text, role: "user" },
      { text: "", role: "bot" },
    ]);

    stopStreamRef.current = streamChat(text, user.role, {
      onDelta: (chunk) => {
        setMessages((prev) => {
          const next = [...prev];
          const last = next[next.length - 1];
          if (last?.role !== "bot") {
            return prev;
          }
          next[next.length - 1] = { ...last, text: last.text + chunk };
          return next;
        });
      },
      onDone: () => {
        setIsStreaming(false);
        stopStreamRef.current = null;
      },
      onError: (error) => {
        setIsStreaming(false);
        stopStreamRef.current = null;
        setMessages((prev) => {
          const next = [...prev];
          const last = next[next.length - 1];
          if (last?.role === "bot" && last.text === "") {
            next[next.length - 1] = {
              ...last,
              text: error.message || "Не удалось получить ответ",
            };
          }
          return next;
        });
      },
    });
  };

  useEffect(() => {
    return () => {
      stopStreamRef.current?.();
    };
  }, []);

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
              onClick={() => getMessageBtnCommand(button.label)}
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
              <div key={index} className={`${styles.message} ${message.role === "user" ? styles.user : styles.bot}`}>
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
            disabled={isStreaming}
            onChange={(e) => setQuestion(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                handleQuestionSubmit();
              }
            }}
          />
  
          <button type="button" onClick={handleQuestionSubmit} disabled={isStreaming}>
            <IconSend className={styles.arrow} />
          </button>
        </div>
        {notice && <p className={styles.notice}>{notice}</p>}
      </div>
    </div>
  );
};
