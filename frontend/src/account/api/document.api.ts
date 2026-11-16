export async function getDocumentContent(url: string): Promise<string> {
  const response = await fetch(url);

  if (!response.ok) {
    throw new Error("Не удалось загрузить документ");
  }

  return response.text();
}