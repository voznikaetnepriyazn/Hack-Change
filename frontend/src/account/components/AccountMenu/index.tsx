import styles from "./index.module.css";
import { useState } from "react";
import { NavLink } from "react-router-dom";
import { BrandName } from "../../../shared/components/BrandName";
import { ProfileItem } from "../ProfileItem";
import { IconChat } from "../../../assets/images/IconChat";
import { IconAim } from "../../../assets/images/IconAim";
import { IconQuestion } from "../../../assets/images/IconQuestion";
import { IconDocuments } from "../../../assets/images/IconDocuments";
import { IconArrow } from "../../../assets/images/IconArrow";

const links = [
  {
    to: "/account/chat",
    label: "Чат с AI-BUDDY",
    icon: <IconChat />,
  },
  {
    to: "/account/tasks",
    label: "Задачи",
    icon: <IconAim />,
  },
  {
    to: "/account/questions",
    label: "Частые вопросы",
    icon: <IconQuestion />,
  },
  {
    to: "/account/documents",
    label: "Документы",
    icon: <IconDocuments />,
  },
];
export const AccountMenu = () => {
  const [isShort, setIsShort] = useState(false);  
  const toggleMenu = () => {
    setIsShort(!isShort);
  };

  return (
    <aside className={`${styles.sidebar} ${isShort ? styles.short : ""}`}>
      <div className={styles.top}>
        {isShort ? '' : <BrandName />}
        <IconArrow className={`${styles.arrow} ${isShort ? styles.short : ""}`} onClick={toggleMenu} />
      </div>
      <nav className={`${styles.menu} ${isShort ? styles.short : ""}`}>
        {links.map((link) => (
          <NavLink
            to={link.to}
            key={link.to}
            className={({ isActive }: { isActive: boolean }) =>
              ` ${styles.link} ${isActive ? styles.active : ""} ${isShort ? styles.short : ""}`
            }
          >
            {link.icon}
            { isShort ? '' : link.label}
          </NavLink>
        ))}
      </nav>
      <NavLink to="/account/profile" className={styles.botom}>
        <ProfileItem isShort={isShort} />
      </NavLink>
    </aside>
  );
};
