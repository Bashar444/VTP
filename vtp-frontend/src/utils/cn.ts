/**
 * Merge classNames conditionally
 * Utility function for combining Tailwind CSS classes
 */
export function cn(...classes: (string | undefined | null | false)[]): string {
  return classes
    .filter((cls): cls is string => typeof cls === 'string' && cls.length > 0)
    .join(' ');
}
