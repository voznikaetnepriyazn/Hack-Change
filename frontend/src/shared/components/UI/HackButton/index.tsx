import styles from "./index.module.css";

export const HackButton = ({
  className,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement>) => {
  return (
    <button
      {...props}
      className={`${styles.hackbutton} ${className ?? ""}`}
    />
  );
};