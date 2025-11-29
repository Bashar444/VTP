export const defaultLocale = 'ar';
export const locales = ['ar'];

export function getMessages(locale: string) {
  // For future multi-locale support
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  return require(`./messages/${locale}.json`);
}
