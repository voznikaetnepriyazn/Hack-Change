import styles from "./index.module.css";

const user = {
  avatar: null,
  name: "Имя Фамилия",
  position: "Стажёр",
};
const user1 = {
  avatar: "https://thumb.wikimedia.org/wikipedia/ru/thumb/0/0f/Tom_Holland_as_Spider-Man.jpg/330px-Tom_Holland_as_Spider-Man.jpg?utm_source=ru.wikipedia.org&utm_campaign=parser&utm_content=thumbnail",
  name: "John Doe",
  position: "Стажёр",
};

export const ProfileItem = ({
  className,
  isShort
}: {
  className?: string;
  isShort?: boolean;
}) => {
  const getFirstWord = (name: string) => {
    return name[0];
  };

  return (
    <div className={`${styles.profile} ${className} ${isShort ? styles.short : ""}`}>
      <div className={styles.avatar}>
        {user.avatar ? (
          <img src={user.avatar} alt="avatar" />
        ) : (
          getFirstWord(user.name)
        )}
      </div>
      {!isShort && (
      <div className={styles.right}>
        <span className={styles.name}>{user.name}</span>
        <span className={styles.position}>{user.position}</span>
      </div>
      )}
    </div>
  );
};
