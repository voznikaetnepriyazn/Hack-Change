import { ArrowSelect } from '../../../../assets/images/IconArrowSelect/index';
import styles from './index.module.css';

type HackSelectProps = React.SelectHTMLAttributes<HTMLSelectElement>;

export function HackSelect({
  children,
  className,
  ...props
}: HackSelectProps) {
  return (
    <div className={styles.hackselectWrap}>
      <select {...props} className={`${styles.hackselect} ${className ?? ""}`}>
        {children}
      </select>

      <span className={styles.hackselectArrow} aria-hidden="true">
        <ArrowSelect />
      </span>
    </div>
  );
}