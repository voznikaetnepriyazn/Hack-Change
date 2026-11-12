import styles from "./index.module.css";
import { IconSearch } from "../icons/IconSearch";

interface SearchInputProps {
  search: string;
  placeholder: string;
  setSearch: (search: string) => void;
}

export const SearchInput = ({ search, placeholder, setSearch }: SearchInputProps) => {
  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  };
  return (
    <label className={styles.search}>
    <IconSearch />
    <input
      type="text"
      value={search}
      className={styles.input}
      onChange={handleSearch}
      placeholder={placeholder}
    />
  </label>
  );
};