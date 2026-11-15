import { Task } from "../../account/types/account.types";

export const parseTask = (id: number, markdown: string): Task => {
  const title = markdown.match(/^#\s+Задача:\s*(.+)$/m)?.[1]?.trim() ?? "Задача";
  const code = markdown.match(/\*\*ID:\*\*\s+`([^`]+)`/)?.[1] ?? "";
  const statusText = markdown.match(/\*\*Статус:\*\*\s+(.+)/)?.[1]?.trim() ?? "";
  const deadlineText = markdown.match(/\*\*Дедлайн:\*\*\s+(.+)/)?.[1]?.trim() ?? "";
  const dateMatch = deadlineText.match(/(\d{2})\.(\d{2})\.(\d{4})/);
  const deadline = dateMatch
    ? new Date(Number(dateMatch[3]), Number(dateMatch[2]) - 1, Number(dateMatch[1]))
    : new Date();

  const description = markdown
    .replace(/^#[^\n]*\n+/, "")
    .replace(/\*\*ID:\*\*[^\n]*\n/, "")
    .replace(/\*\*Статус:\*\*[^\n]*\n/, "")
    .replace(/\*\*Дедлайн:\*\*[^\n]*\n+/, "")
    .trim();

  return {
    id,
    code,
    title,
    description,
    status: /заверш/i.test(statusText) ? "completed" : "active",
    deadline,
  };
};