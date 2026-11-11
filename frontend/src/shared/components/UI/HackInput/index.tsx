import styles from "./index.module.css";

export const HackInput = ({
  className,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement>) => {
  return (
    <input
      {...props}
      className={`${styles.hackinput} ${className ?? ""}`}
    />
  );
};