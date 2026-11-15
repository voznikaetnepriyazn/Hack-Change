import { create } from "zustand";
import { ContentBlock, Document, Tag } from "../account/types/account.types";
import teamContactsUrl from "../assets/documents/team_alpha_contacts.md?url";
import teamProcessesUrl from "../assets/documents/team_alpha_processes.md?url";
import teamEnvironmentUrl from "../assets/documents/team_alpha_environment.md?url";
import officeUrl from "../assets/documents/office_and_infrastructure.md?url";
import vacationUrl from "../assets/documents/vacation_and_dayoff_rules.md?url";
import healthUrl from "../assets/documents/health_and_sick_leave.md?url";
import financeUrl from "../assets/documents/finance_and_documents.md?url";
import careerUrl from "../assets/documents/career_and_grades.md?url";
import corporateLifeUrl from "../assets/documents/corporate_life_and_events.md?url";

interface DocumentsState {
  documents: Document[];
  tags: Tag[];
  contentBlock: ContentBlock;
  title: string;
  description: string;
}

export const useDocumentsStore = create<DocumentsState>(() => ({
  title: "Документы",
  description: "Инструкции, материалы и полезные ссылки для стажёров.",
  tags: [
    {
      id: 0,
      name: "Все документы",
      tab: "all",
      items: [],
    },
    {
      id: 1,
      name: "Инструкции",
      tab: "instructions",
      items: [4, 5, 6, 7],
    },
    {
      id: 2,
      name: "Обучение",
      tab: "training",
      items: [8, 9],
    },
    {
      id: 3,
      name: "Команда",
      tab: "team",
      items: [1, 2, 3],
    },
  ],
  documents: [
    {
      id: 1,
      title: "Контакты и роли команды Alpha",
      description: "Кто за что отвечает и куда писать с вопросом",
      link: teamContactsUrl,
    },
    {
      id: 2,
      title: "Рабочие процессы команды Alpha",
      description: "Встречи, разработка и релизы команды",
      link: teamProcessesUrl,
    },
    {
      id: 3,
      title: "Техническая среда и доступы команды Alpha",
      description: "Настройка окружения, доступы и код-ревью",
      link: teamEnvironmentUrl,
    },
    {
      id: 4,
      title: "Офис, удалёнка и рабочие инструменты",
      description: "Как устроен офис, удалёнка и доступ к сервисам",
      link: officeUrl,
    },
    {
      id: 5,
      title: "Отпуска, отгулы и отсутствия",
      description: "Как оформить отгул, отпуск и предупредить об отсутствии",
      link: vacationUrl,
    },
    {
      id: 6,
      title: "Здоровье и больничные",
      description: "Что делать, если заболел, и как оформляется больничный",
      link: healthUrl,
    },
    {
      id: 7,
      title: "Финансы и документы",
      description: "Договор, зарплата, реквизиты и выплаты",
      link: financeUrl,
    },
    {
      id: 8,
      title: "Грейды, карьера и развитие",
      description: "Линейка грейдов, повышение и обучение",
      link: careerUrl,
    },
    {
      id: 9,
      title: "Корпоративная жизнь и мероприятия",
      description: "Мероприятия, кружки и культура общения",
      link: corporateLifeUrl,
    },
  ],
  contentBlock: {
    title: "Нужна помощь?",
    description:
      "Попросите AI Buddy кратко пересказать документ или объяснить непонятный термин.",
  },
}));
