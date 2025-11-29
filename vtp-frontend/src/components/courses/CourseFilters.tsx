import { useState } from 'react';
import { Search, Filter } from 'lucide-react';
import { cn } from '@/utils/cn';
import { useTranslations } from 'next-intl';

interface CourseFiltersProps {
  onFilterChange?: (filters: CourseFilterState) => void;
  onSearch?: (query: string) => void;
  className?: string;
}

export interface CourseFilterState {
  category?: string;
  level?: string;
  search?: string;
  priceRange?: [number, number];
  rating?: number;
  sortBy?: 'newest' | 'popular' | 'highest-rated' | 'price-low' | 'price-high';
}

export const CourseFilters: React.FC<CourseFiltersProps> = ({
  onFilterChange,
  onSearch,
  className,
}) => {
  const t = useTranslations();
  const [filters, setFilters] = useState<CourseFilterState>({});
  const [showMobileFilters, setShowMobileFilters] = useState(false);

  const categories = ['Programming', 'Design', 'Business', 'Science', 'Language', 'Health'];
  const levels = ['Beginner', 'Intermediate', 'Advanced'];
  const sortOptions = [
    { value: 'newest', labelKey: 'filters.sort.newest' },
    { value: 'popular', labelKey: 'filters.sort.popular' },
    { value: 'highest-rated', labelKey: 'filters.sort.highest-rated' },
    { value: 'price-low', labelKey: 'filters.sort.price-low' },
    { value: 'price-high', labelKey: 'filters.sort.price-high' },
  ];

  const updateFilters = (partial: Partial<CourseFilterState>) => {
    const newFilters = { ...filters, ...partial };
    setFilters(newFilters);
    onFilterChange?.(newFilters);
  };

  const handleCategoryChange = (category: string) => {
    updateFilters({ category: filters.category === category ? undefined : category });
  };
  const handleLevelChange = (level: string) => {
    updateFilters({ level: filters.level === level ? undefined : level });
  };
  const handleSortChange = (sortBy: string) => {
    updateFilters({ sortBy: sortBy as CourseFilterState['sortBy'] });
  };
  const handleSearch = (query: string) => {
    updateFilters({ search: query });
    onSearch?.(query);
  };
  const clearFilters = () => {
    setFilters({});
    onFilterChange?.({});
    onSearch?.('');
  };

  return (
    <div className={cn('space-y-4', className)}>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
        <input
          type="text"
          placeholder={t('filters.searchPlaceholder')}
          value={filters.search || ''}
          onChange={e => handleSearch(e.target.value)}
          className="w-full pl-10 pr-4 py-2.5 bg-gray-800 text-white border border-gray-700 rounded-lg focus:outline-none focus:border-blue-500 transition-colors"
        />
      </div>

      <div className="hidden lg:block space-y-4">
        <div>
          <h3 className="font-semibold text-white mb-3">{t('filters.category')}</h3>
          <div className="space-y-2">
            {categories.map(category => (
              <label key={category} className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={filters.category === category.toLowerCase()}
                  onChange={() => handleCategoryChange(category.toLowerCase())}
                  className="w-4 h-4 rounded border-gray-500 bg-gray-800 text-blue-600"
                />
                <span className="text-gray-300 hover:text-white">{category}</span>
              </label>
            ))}
          </div>
        </div>

        <div>
          <h3 className="font-semibold text-white mb-3">{t('filters.level')}</h3>
          <div className="space-y-2">
            {levels.map(level => (
              <label key={level} className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={filters.level === level.toLowerCase()}
                  onChange={() => handleLevelChange(level.toLowerCase())}
                  className="w-4 h-4 rounded border-gray-500 bg-gray-800 text-blue-600"
                />
                <span className="text-gray-300 hover:text-white">{level}</span>
              </label>
            ))}
          </div>
        </div>

        <div>
          <h3 className="font-semibold text-white mb-3">{t('filters.sortBy')}</h3>
          <select
            value={filters.sortBy || 'newest'}
            onChange={e => handleSortChange(e.target.value)}
            className="w-full px-3 py-2 bg-gray-800 text-white border border-gray-700 rounded-lg focus:outline-none focus:border-blue-500"
          >
            {sortOptions.map(option => (
              <option key={option.value} value={option.value}>
                {t(option.labelKey)}
              </option>
            ))}
          </select>
        </div>

        {Object.keys(filters).length > 0 && (
          <button
            onClick={clearFilters}
            className="w-full py-2 px-4 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
          >
            {t('filters.clear')}
          </button>
        )}
      </div>

      <div className="lg:hidden">
        <button
          onClick={() => setShowMobileFilters(!showMobileFilters)}
          className="w-full py-2.5 px-4 bg-gray-800 hover:bg-gray-700 text-white font-semibold rounded-lg transition-colors flex items-center justify-center gap-2"
        >
          <Filter className="w-5 h-5" />
          {t('filters.open')}
        </button>
        {showMobileFilters && (
          <div className="mt-3 p-4 bg-gray-800 rounded-lg space-y-4">
            <div>
              <h3 className="font-semibold text-white mb-2">{t('filters.category')}</h3>
              <div className="grid grid-cols-2 gap-2">
                {categories.map(category => (
                  <button
                    key={category}
                    onClick={() => handleCategoryChange(category.toLowerCase())}
                    className={cn(
                      'py-2 px-3 rounded-lg text-sm font-medium transition-colors',
                      filters.category === category.toLowerCase()
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-700 text-gray-300'
                    )}
                  >
                    {category}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <h3 className="font-semibold text-white mb-2">{t('filters.level')}</h3>
              <div className="grid grid-cols-3 gap-2">
                {levels.map(level => (
                  <button
                    key={level}
                    onClick={() => handleLevelChange(level.toLowerCase())}
                    className={cn(
                      'py-2 px-3 rounded-lg text-sm font-medium transition-colors',
                      filters.level === level.toLowerCase()
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-700 text-gray-300'
                    )}
                  >
                    {level}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <h3 className="font-semibold text-white mb-2">{t('filters.sortBy')}</h3>
              <select
                value={filters.sortBy || 'newest'}
                onChange={e => handleSortChange(e.target.value)}
                className="w-full px-3 py-2 bg-gray-700 text-white border border-gray-600 rounded-lg focus:outline-none focus:border-blue-500"
              >
                {sortOptions.map(option => (
                  <option key={option.value} value={option.value}>
                    {t(option.labelKey)}
                  </option>
                ))}
              </select>
            </div>
            {Object.keys(filters).length > 0 && (
              <button
                onClick={clearFilters}
                className="w-full py-2 px-4 bg-gray-600 hover:bg-gray-500 text-white rounded-lg transition-colors"
              >
                {t('filters.clear')}
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

interface EnrollmentFormProps {
  courseId: string;
  courseName: string;
  coursePrice: number;
  isFree: boolean;
  onEnroll?: () => void;
  onCancel?: () => void;
  isLoading?: boolean;
  className?: string;
}

export const EnrollmentForm: React.FC<EnrollmentFormProps> = ({
  courseId,
  courseName,
  coursePrice,
  isFree,
  onEnroll,
  onCancel,
  isLoading = false,
  className,
}) => {
  const t = useTranslations();
  return (
    <div className={cn('bg-gray-800 rounded-lg p-6', className)}>
      <h3 className="text-xl font-bold text-white mb-4">{t('enroll.form.title')}</h3>
      <div className="mb-6 p-4 bg-gray-900 rounded-lg">
        <p className="text-gray-400 text-sm mb-1">{t('enroll.form.courseLabel')}</p>
        <p className="text-white font-semibold">{courseName}</p>
      </div>
      <div className="mb-6 space-y-3 p-4 bg-gray-900 rounded-lg">
        <div className="flex justify-between">
          <span className="text-gray-400">{t('enroll.form.price')}</span>
          <span className="text-white font-semibold">{isFree ? t('enroll.form.free') : `$${coursePrice}`}</span>
        </div>
        {!isFree && (
          <>
            <div className="flex justify-between text-sm">
              <span className="text-gray-400">{t('enroll.form.discount')}</span>
              <span className="text-green-400">{t('enroll.form.none')}</span>
            </div>
            <div className="border-t border-gray-700 pt-3 flex justify-between">
              <span className="text-gray-300 font-semibold">{t('enroll.form.total')}</span>
              <span className="text-white font-bold text-lg">${coursePrice}</span>
            </div>
          </>
        )}
      </div>
      <div className="mb-6">
        <label className="flex items-start gap-3 cursor-pointer">
          <input
            type="checkbox"
            defaultChecked
            className="w-4 h-4 rounded border-gray-500 bg-gray-800 text-blue-600 mt-1"
          />
          <span className="text-sm text-gray-300">{t('enroll.form.terms')}</span>
        </label>
      </div>
      <div className="space-y-3">
        <button
          onClick={onEnroll}
          disabled={isLoading}
          className="w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white font-bold rounded-lg transition-colors"
        >
          {isLoading ? t('enroll.form.processing') : t('enroll.form.enrollNow')}
        </button>
        <button
          onClick={onCancel}
          className="w-full py-3 px-4 bg-gray-700 hover:bg-gray-600 text-white font-bold rounded-lg transition-colors"
        >
          {t('enroll.form.cancel')}
        </button>
      </div>
    </div>
  );
};

export { EnrollmentForm };