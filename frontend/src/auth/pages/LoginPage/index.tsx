import { useState, type ComponentProps } from "react";
import { Link, useNavigate } from "react-router-dom";
import ApiError from "../../../shared/errors/ApiError";
import { useAuth } from "../../context/AuthContext";
import image1 from "../../../assets/images/cloud-text.png";
import { IconLogo } from "../../../assets/images/IconLogo";
import styles from "./index.module.css";
import { HackSelect } from "../../../shared/components/UI/HackSelect";
import { HackInput } from "../../../shared/components/UI/HackInput";
import { HackButton } from "../../../shared/components/UI/HackButton";
import { BrandName } from "../../../shared/components/BrandName";

export const LoginPage = () => {
  const [email, setEmail] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [position, setPosition] = useState("");

  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit: NonNullable<ComponentProps<"form">["onSubmit"]> = async (
    e,
  ) => {
    e.preventDefault();
    console.log(password, username, position);
    setTimeout(() => {
      setPassword("");
      setUsername("");
      setPosition("");
      navigate("/account");
    }, 500);
    // setErrorMessage(null);
    // try {
    //   await login(email, password);
    //   navigate("/account");
    // } catch (error) {
    //   if (error instanceof ApiError) {
    //     setErrorMessage(error.message);
    //   } else {
    //     console.error(error);
    //     setErrorMessage(
    //       "An error occurred while logging in. Please try again.",
    //     );
    //   }
    // }
  };

  return (
    <main className="page">
      <aside className={styles.left}>
        <div className={styles.left__container}>
          <div className={styles.left__name}>ВНУТРЕННИЙ СЕРВИС</div>
          <h1 className={styles.left__title}>Твой помощник на стажировке</h1>
          <p className={styles.left__description}>
            Разбирайся в задачах, находи ответы и быстрее осваивайся в команде
          </p>
          <div className={styles.left_images}>
            <IconLogo />
            <img src={image1} alt="image2" />
          </div>
          <p className={styles.left__description}>
            Меньше вопросов в одиночку. <br />
            Больше уверенности в работе.
          </p>
        </div>
      </aside>
      <section className={styles.right}>
        <div className={styles.right__container}>
          <BrandName />
          <div className={styles.right__content}>
            <h2 className={styles.right__title}>Войти в AI-BUDDY</h2>
            <form className={styles.right__form} onSubmit={handleSubmit}>
              <label>
                <span>Логин</span>
                <HackInput
                  type="text"
                  name="login"
                  required
                  placeholder="@username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
              </label>
              <label>
                <span>Пароль</span>
                <HackInput
                  type="password"
                  name="password"
                  required
                  placeholder="********"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </label>
              <label>
                <span>Должность</span>
                <HackSelect
                  name="position"
                  required
                  value={position}
                  onChange={(e) => setPosition(e.target.value)}
                >
                  <option value="">Выберите должность</option>
                  <option value="intern">Стажер</option>
                  <option value="lead">Руководитель</option>
                </HackSelect>
              </label>
              <HackButton type="submit">Продолжить</HackButton>
            </form>
          </div>
          <p className={styles.right__bottom}>
            Продолжая, ты принимаешь правила использования AI Buddy и политику
            конфиденциальности.
          </p>
        </div>
      </section>
    </main>
  );
};
