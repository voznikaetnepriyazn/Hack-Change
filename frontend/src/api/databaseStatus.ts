export const databaseStatuses = ['connected', 'disconnected', 'not_configured'] as const;
export type DatabaseStatus = (typeof databaseStatuses)[number];

type IsDatabaseStatus = (value: unknown) => value is DatabaseStatus;

export const isDatabaseStatus: IsDatabaseStatus = (value): value is DatabaseStatus => {
  return (typeof value === 'string' && databaseStatuses.some(status => status === value));
}