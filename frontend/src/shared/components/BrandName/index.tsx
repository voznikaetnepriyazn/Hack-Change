import styles from "./index.module.css";

export const BrandName = ({ className }: { className?: string }) => {
  return (
    <div className={`${styles.brandName} ${className}`}>
      <span>AI-BUDDY</span>
      <span>AVITO</span>
    </div>
  );
};
