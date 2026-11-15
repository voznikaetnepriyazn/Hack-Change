import styles from "./index.module.css";
import { useState } from "react";
import { NavLink } from "react-router-dom";
import { BrandName } from "../../../shared/components/BrandName";
import { ProfileItem } from "../ProfileItem";
import { IconChat } from "../../../shared/components/icons/IconChat";
import { IconAim } from "../../../shared/components/icons/IconAim";
import { IconQuestion } from "../../../shared/components/icons/IconQuestion";
import { IconDocuments } from "../../../shared/components/icons/IconDocuments";
import { IconArrow } from "../../../shared/components/icons/IconArrow";
import { IconStudent } from "../../../shared/components/icons/IconStudent";
import { useProfileStore } from "../../../stores/profileStore";

const links = [
  {
    to: "/account/chat",
    label: "Чат с AI-BUDDY",
    icon: <IconChat />,
    whoLook: "all",
  },
  {
    to: "/account/tasks",
    label: "Задачи",
    icon: <IconAim />,
    whoLook: "intern",
  },
  {
    to: "/account/questions",
    label: "Частые вопросы",
    icon: <IconQuestion />,
    whoLook: "intern",
  },
  {
    to: "/account/intern",
    label: "Мой стажер",
    icon: <IconStudent />,
    whoLook: "lead",
  },
  {
    to: "/account/documents",
    label: "Документы",
    icon: <IconDocuments />,
    whoLook: "all",
  },
];
export const AccountMenu = () => {
  const [isShort, setIsShort] = useState(false);
  const user = useProfileStore((state) => state.user);
  const toggleMenu = () => {
    setIsShort(!isShort);
  };

  return (
    <aside className={`${styles.sidebar} ${isShort ? styles.short : ""}`}>
      <div className={styles.top}>
        <div className={`${styles.brand} ${isShort ? styles.hidden : ""}`}>
          <BrandName />
        </div>
        <IconArrow
          className={`${styles.arrow} ${isShort ? styles.short : ""}`}
          onClick={toggleMenu}
        />
      </div>
      <nav className={styles.menu}>
        {links.filter((link) => link.whoLook === user?.role || link.whoLook === "all").map((link) => (
          <NavLink
            to={link.to}
            key={link.to}
            className={({ isActive }: { isActive: boolean }) =>
              ` ${styles.link} ${isActive ? styles.active : ""} ${isShort ? styles.short : ""}`
            }
          >
            {link.icon}
            <span
              className={`${styles.linkText} ${isShort ? styles.hidden : ""}`}
            >
              {link.label}
            </span>
          </NavLink>
        ))}
      </nav>
      <NavLink to="/account/profile" className={styles.bottom}>
        <ProfileItem isShort={isShort} user={user} />
      </NavLink>
    </aside>
  );
};
