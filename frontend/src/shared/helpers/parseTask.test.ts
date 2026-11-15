import { describe, expect, it } from "vitest";
import { parseTask } from "./parseTask";

const markdown = `# Задача: Настроить окружение

**ID:** \`ALPHA-101\`
**Статус:** Завершено
**Дедлайн:** 15.11.2026

## Описание
Сервис отвечает на localhost.
`;

describe("parseTask", () => {
  it("reads the title, code, completed status and deadline from markdown", () => {
    const task = parseTask(4, markdown);

    expect(task).toMatchObject({
      id: 4,
      code: "ALPHA-101",
      title: "Настроить окружение",
      status: "completed",
    });
    expect(task.deadline).toEqual(new Date(2026, 10, 15));
    expect(task.description).toContain("## Описание");
    expect(task.description).not.toContain("ALPHA-101");
    expect(task.description).not.toContain("**Статус:**");
  });

  it("keeps an in-progress task active", () => {
    const task = parseTask(1, markdown.replace("Завершено", "В процессе"));

    expect(task.status).toBe("active");
  });

  it("uses fallbacks when the markdown has no task fields", () => {
    const before = Date.now();
    const task = parseTask(2, "просто текст");

    expect(task).toMatchObject({
      id: 2,
      code: "",
      title: "Задача",
      description: "просто текст",
      status: "active",
    });
    expect(task.deadline.getTime()).toBeGreaterThanOrEqual(before);
  });
});
