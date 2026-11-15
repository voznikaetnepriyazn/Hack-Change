import { describe, expect, it } from "vitest";
import { useDocumentsStore } from "./documentsStore";
import { useInternStore } from "./internStore";
import { useProfileStore } from "./profileStore";
import { useQuestionsStore } from "./questionsStore";
import { useTasksStore } from "./tasksStore";
import { internUser } from "../test/fixtures";

describe("stores", () => {
  it("exposes questions, documents and profile content", () => {
    expect(useQuestionsStore.getState().questions).toHaveLength(8);
    expect(useQuestionsStore.getState().title).toBe("Частые вопросы");
    expect(useDocumentsStore.getState().documents).toHaveLength(5);
    expect(useProfileStore.getState().description.lead).toContain("руководителя");
    expect(useInternStore.getState().metrics.requests).toBe(20);
  });

  it("updates the profile and intern user", () => {
    useProfileStore.getState().setProfileUser(internUser);
    expect(useProfileStore.getState().user?.id).toBe(internUser.id);

    useProfileStore.getState().setProfileUser(null);
    expect(useProfileStore.getState().user).toBeNull();

    useInternStore.getState().setInternUser(internUser);
    expect(useInternStore.getState().user?.full_name).toBe(internUser.full_name);
    useInternStore.getState().setInternUser(null);
  });

  it("toggles a task status and leaves other tasks untouched", () => {
    const before = useTasksStore.getState().tasks.find((task) => task.id === 1);
    expect(before?.status).toBe("active");

    useTasksStore.getState().changeTaskStatus(1);
    expect(useTasksStore.getState().tasks.find((task) => task.id === 1)?.status).toBe(
      "completed",
    );
    expect(useTasksStore.getState().tasks.find((task) => task.id === 2)?.status).toBe(
      "completed",
    );

    useTasksStore.getState().changeTaskStatus(2);
    expect(useTasksStore.getState().tasks.find((task) => task.id === 2)?.status).toBe(
      "active",
    );

    useTasksStore.getState().changeTaskStatus(999);
    expect(useTasksStore.getState().tasks).toHaveLength(5);

    useTasksStore.getState().changeTaskStatus(1);
    useTasksStore.getState().changeTaskStatus(2);
  });
});
