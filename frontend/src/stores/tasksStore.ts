import { create } from "zustand";
import { ContentBlock, Task, Tag } from "../account/types/account.types";
import { parseTask } from "../shared/helpers/parseTask";
import task1 from "../assets/tasks/task1.md?raw";
import task2 from "../assets/tasks/task2.md?raw";
import task3 from "../assets/tasks/task3.md?raw";
import task4 from "../assets/tasks/task4.md?raw";
import task5 from "../assets/tasks/task5.md?raw";

interface TasksState {
  tasks: Task[];
  tags: Tag[];
  contentBlock: ContentBlock;
  title: string;
  description: string;
  changeTaskStatus: (taskId: number) => void;
}



const tasks = [task1, task2, task3, task4, task5].map((markdown, index) =>
  parseTask(index + 1, markdown),
);

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
      items: tasks.filter((task) => task.status === "active").map((task) => task.id),
    },
    {
      id: 2,
      name: "Завершённые",
      tab: "completed",
      items: tasks.filter((task) => task.status === "completed").map((task) => task.id),
    },
  ],
  tasks,
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
