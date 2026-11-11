import styles from "./index.module.css";
import { NavLink } from "react-router-dom";
import { BrandName } from "../../../shared/components/BrandName";
import { ProfileItem } from "../ProfileItem";
import { IconChat } from "../../../assets/images/IconChat";
import { IconAim } from "../../../assets/images/IconAim";
import { IconQuestion } from "../../../assets/images/IconQuestion";
import { IconDocuments } from "../../../assets/images/IconDocuments";

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
  return (
    <aside className={styles.sidebar}>
      <div className={styles.top}>
        <BrandName />
      </div>
      <nav className={styles.menu}>
        {links.map((link) => (
          <NavLink
            to={link.to}
            key={link.to}
            className={({ isActive }: { isActive: boolean }) =>
              ` ${styles.link} ${isActive ? styles.active : ""}`
            }
          >
            {link.icon}
            {link.label}
          </NavLink>
        ))}
      </nav>
      <NavLink to="/account/profile" className={styles.botom}>
        <ProfileItem />
      </NavLink>
    </aside>
  );
};
