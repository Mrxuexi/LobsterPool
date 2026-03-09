export function getLocaleTag(language: string | undefined): string {
  return language?.startsWith('zh') ? 'zh-CN' : 'en-US';
}

export function formatDateTime(value: string, language: string | undefined): string {
  return new Intl.DateTimeFormat(getLocaleTag(language), {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}

export function formatDisplayDate(value: Date, language: string | undefined): string {
  return new Intl.DateTimeFormat(getLocaleTag(language), {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  }).format(value);
}
