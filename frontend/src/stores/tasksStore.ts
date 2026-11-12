import { create } from "zustand";
import { ContentBlock, Task, Tag } from "../account/types/account.types";


interface TasksState {
  tasks: Task[];
  tags: Tag[];
  contentBlock: ContentBlock;
  title: string;
  description: string;
  changeTaskStatus: (taskId: number) => void;
}


export const useTasksStore = create<TasksState>((set) => ({
  title: "Задачи",
  description: "Список задач, их описание и дедлайны",
  tags: [
    {
      id: 0,
      name: "Все задачи",
      tab: "all",
      items: [],
    },
    {
      id: 1,
      name: "Активные",
      tab: "active",
      items: [1, 2],
    },
    {
      id: 2,
      name: "Завершённые",
      tab: "completed",
      items: [4, 5],
    },
  ],
  tasks: [
    {
      id: 1,
      title: "Задача 1",
      description: "Описание 1 Описание 1 Описание 1 Описание 1 Описание 1 Описание 1 Описание 1 Описание 1 Описание 1 Описание 1 Описание 1 Описание 1",
      status: "active",
      deadline: new Date(2026, 10, 12),
    },
    {
      id: 2,
      title: "Задача 2",
      description: "Описание задачи 2",
      status: "completed",
      deadline: new Date(2026, 11, 15),
    },
    {
      id: 3,
      title: "Задача 3",
      description: "Описание задачи 3",
      status: "active",
      deadline: new Date(2026, 11, 20),
    },
    {
      id: 4,
      title: "Задача 4",
      description: "Описание задачи 4",
      status: "completed",
      deadline: new Date(2026, 10, 14),
    },
    {
      id: 5,
      title: "Задача 5",
      description: "Описание задачи 5",
      status: "active",
      deadline: new Date(2026, 11, 26),
    },
  ],
  contentBlock: {
    title: "Непонятна задача?",
    description: "Попросите AI Buddy помочь разобраться, составить план выполнения или подсказать, с чего начать.",
  },
  changeTaskStatus: (taskId: number) => {
    set((state) => ({
      tasks: state.tasks.map((task) => task.id === taskId ? { ...task, status: task.status === "active" ? "completed" : "active" } : task),
    }));
  },
}));
