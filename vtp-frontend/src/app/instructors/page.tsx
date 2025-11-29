'use client';
export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { InstructorService, SubjectService } from '@/services/domain.service';
import type { Instructor, Subject } from '@/types/domains';
import { Search, Star, Filter, CheckCircle, BookOpen } from 'lucide-react';

export default function InstructorsPage() {
  const router = useRouter();
  const t = useTranslations();

  const [instructors, setInstructors] = useState<Instructor[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSubject, setSelectedSubject] = useState<string>('');
  const [minRating, setMinRating] = useState<number>(0);
  const [verifiedOnly, setVerifiedOnly] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const [instructorsData, subjectsData] = await Promise.all([
          InstructorService.getInstructors({
            subject_id: selectedSubject || undefined,
            min_rating: minRating || undefined,
            is_verified: verifiedOnly || undefined,
          }),
          SubjectService.getSubjects(),
        ]);
        setInstructors(instructorsData);
        setSubjects(subjectsData);
      } catch (err) {
        console.error('Failed to load instructors:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [selectedSubject, minRating, verifiedOnly]);

  const filteredInstructors = instructors.filter(instructor =>
    instructor.name_ar.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">المعلمون</h1>
          <p className="text-gray-400">اكتشف أفضل المعلمين المؤهلين في جميع التخصصات</p>
        </div>

        {/* Search and Filters */}
        <div className="bg-gray-800 rounded-lg p-6 mb-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Search */}
            <div className="md:col-span-2">
              <div className="relative">
                <Search className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                <input
                  type="text"
                  placeholder="ابحث عن معلم..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pr-10 pl-4 py-3 bg-gray-700 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            {/* Subject Filter */}
            <div>
              <select
                value={selectedSubject}
                onChange={(e) => setSelectedSubject(e.target.value)}
                className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">كل المواد</option>
                {subjects.map((subject) => (
                  <option key={subject.id} value={subject.id}>
                    {subject.name_ar}
                  </option>
                ))}
              </select>
            </div>

            {/* Rating Filter */}
            <div>
              <select
                value={minRating}
                onChange={(e) => setMinRating(Number(e.target.value))}
                className="w-full px-4 py-3 bg-gray-700 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="0">كل التقييمات</option>
                <option value="4">4+ نجوم</option>
                <option value="4.5">4.5+ نجوم</option>
              </select>
            </div>
          </div>

          {/* Verified Only Toggle */}
          <div className="mt-4">
            <label className="flex items-center gap-2 cursor-pointer w-fit">
              <input
                type="checkbox"
                checked={verifiedOnly}
                onChange={(e) => setVerifiedOnly(e.target.checked)}
                className="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500"
              />
              <span className="text-gray-300">معلمون موثقون فقط</span>
            </label>
          </div>
        </div>

        {/* Results */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div key={i} className="bg-gray-800 rounded-lg p-6 animate-pulse">
                <div className="w-20 h-20 bg-gray-700 rounded-full mb-4" />
                <div className="h-6 bg-gray-700 rounded w-3/4 mb-2" />
                <div className="h-4 bg-gray-700 rounded w-full mb-4" />
                <div className="h-4 bg-gray-700 rounded w-1/2" />
              </div>
            ))}
          </div>
        ) : filteredInstructors.length === 0 ? (
          <div className="bg-gray-800 rounded-lg p-12 text-center">
            <Filter className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <p className="text-gray-400 text-lg">لا توجد نتائج مطابقة للبحث</p>
          </div>
        ) : (
          <>
            <div className="mb-4 text-gray-400">
              عرض {filteredInstructors.length} معلم
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {filteredInstructors.map((instructor) => (
                <div
                  key={instructor.id}
                  className="bg-gray-800 rounded-lg p-6 hover:bg-gray-750 transition-colors cursor-pointer"
                  onClick={() => router.push(`/instructors/${instructor.id}`)}
                >
                  {/* Avatar */}
                  <div className="flex items-start gap-4 mb-4">
                    {instructor.profile_image_url ? (
                      <img
                        src={instructor.profile_image_url}
                        alt={instructor.name_ar}
                        className="w-20 h-20 rounded-full object-cover"
                      />
                    ) : (
                      <div className="w-20 h-20 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center flex-shrink-0">
                        <span className="text-2xl text-white font-bold">
                          {instructor.name_ar.charAt(0)}
                        </span>
                      </div>
                    )}

                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <h3 className="text-xl font-bold text-white">{instructor.name_ar}</h3>
                        {instructor.is_verified && (
                          <CheckCircle className="w-5 h-5 text-blue-400" />
                        )}
                      </div>
                      
                      {/* Rating */}
                      <div className="flex items-center gap-2 mb-2">
                        <div className="flex items-center gap-1">
                          {[1, 2, 3, 4, 5].map((star) => (
                            <Star
                              key={star}
                              className={`w-4 h-4 ${
                                star <= Math.round(instructor.rating)
                                  ? 'text-yellow-400 fill-yellow-400'
                                  : 'text-gray-600'
                              }`}
                            />
                          ))}
                        </div>
                        <span className="text-sm text-gray-400">
                          ({instructor.total_reviews})
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Bio Preview */}
                  <p className="text-gray-300 text-sm mb-4 line-clamp-2">
                    {instructor.bio_ar || 'لا توجد نبذة متاحة.'}
                  </p>

                  {/* Stats */}
                  <div className="flex items-center justify-between text-sm text-gray-400 mb-4">
                    <div className="flex items-center gap-1">
                      <BookOpen className="w-4 h-4" />
                      <span>{instructor.years_experience} سنوات خبرة</span>
                    </div>
                    <div className="text-white font-semibold">
                      ${instructor.hourly_rate}/ساعة
                    </div>
                  </div>

                  {/* Action Button */}
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      router.push(`/instructors/${instructor.id}`);
                    }}
                    className="w-full py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
                  >
                    عرض الملف الشخصي
                  </button>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
