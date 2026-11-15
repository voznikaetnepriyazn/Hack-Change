import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { BrandName } from "./BrandName";
import { SearchInput } from "./SearchInput";
import { HackButton } from "./UI/HackButton";
import { HackInput } from "./UI/HackInput";
import { HackSelect } from "./UI/HackSelect";
import { Loader } from "./UI/Loader";
import { IconAim } from "./icons/IconAim";
import { IconArrow } from "./icons/IconArrow";
import { IconArrowSelect } from "./icons/IconArrowSelect";
import { IconChat } from "./icons/IconChat";
import { IconCheck } from "./icons/IconCheck";
import { IconDocuments } from "./icons/IconDocuments";
import { IconInfo } from "./icons/IconInfo";
import { IconLogo } from "./icons/IconLogo";
import { IconQuestion } from "./icons/IconQuestion";
import { IconQuestionCircle } from "./icons/IconQuestionCircle";
import { IconRocket } from "./icons/IconRocket";
import { IconSearch } from "./icons/IconSearch";
import { IconSend } from "./icons/IconSend";
import { IconStudent } from "./icons/IconStudent";
import { IconTeam } from "./icons/IconTeam";

describe("shared interface", () => {
  it("renders controls and forwards their events", () => {
    const onClick = vi.fn();
    const onSearch = vi.fn();
    render(
      <>
        <BrandName className="brand" />
        <HackButton onClick={onClick}>Дальше</HackButton>
        <HackButton>Без класса</HackButton>
        <HackInput aria-label="логин" className="field" />
        <HackInput aria-label="пароль" />
        <HackSelect aria-label="роль" defaultValue="intern">
          <option value="intern">Стажер</option>
        </HackSelect>
        <HackSelect aria-label="пустая роль" className="select" />
        <SearchInput search="" placeholder="Найти" setSearch={onSearch} />
        <Loader />
      </>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Дальше" }));
    expect(onClick).toHaveBeenCalledOnce();
    fireEvent.change(screen.getByPlaceholderText("Найти"), { target: { value: "почта" } });
    expect(onSearch).toHaveBeenCalledWith("почта");
    expect(screen.getByText("Загрузка...")).toBeInTheDocument();
    expect(screen.getAllByText("AI-BUDDY").length).toBeGreaterThan(0);
  });

  it("renders icons, including both check states", () => {
    const { container } = render(
      <>
        <IconAim />
        <IconArrow />
        <IconArrowSelect />
        <IconChat />
        <IconCheck />
        <IconCheck status="completed" />
        <IconDocuments />
        <IconInfo />
        <IconLogo />
        <IconQuestion />
        <IconQuestionCircle />
        <IconRocket />
        <IconSearch />
        <IconSend />
        <IconStudent />
        <IconTeam />
      </>,
    );

    expect(container.querySelectorAll("svg").length).toBeGreaterThan(10);
  });
});
