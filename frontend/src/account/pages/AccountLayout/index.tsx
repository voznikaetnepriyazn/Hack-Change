import { Outlet } from "react-router-dom";

import { AccountMenu } from "../../components/AccountMenu";

import styles from "./index.module.css";
import { IconLogo } from "../../../assets/images/IconLogo";

export const AccountLayout = () => {
  return (
    <div className={styles.account}>
      <AccountMenu />

      <main className={styles.content}>
        <IconLogo className={styles.logo} />
        <Outlet />
      </main>
    </div>
  );
};
