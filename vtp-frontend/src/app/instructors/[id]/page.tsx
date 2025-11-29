'use client';
export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { InstructorService, SubjectService } from '@/services/domain.service';
import type { Instructor, Subject } from '@/types/domains';
import { Star, Award, BookOpen, Clock, CheckCircle, Calendar } from 'lucide-react';

export default function InstructorProfilePage() {
  const params = useParams();
  const router = useRouter();
  const t = useTranslations();
  const instructorId = params.id as string;

  const [instructor, setInstructor] = useState<Instructor | null>(null);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        const instructorData = await InstructorService.getInstructorById(instructorId);
        setInstructor(instructorData);

        // Fetch subjects for specializations
        if (instructorData.specialization.length > 0) {
          const allSubjects = await SubjectService.getSubjects();
          const instructorSubjects = allSubjects.filter((s: Subject) =>
            instructorData.specialization.includes(s.id)
          );
          setSubjects(instructorSubjects);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load instructor');
      } finally {
        setIsLoading(false);
      }
    };

    if (instructorId) {
      fetchData();
    }
  }, [instructorId]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-900 pt-24 pb-12">
        <div className="container mx-auto px-4">
          <div className="animate-pulse space-y-8">
            <div className="h-64 bg-gray-800 rounded-lg" />
            <div className="space-y-4">
              <div className="h-8 bg-gray-800 rounded w-3/4" />
              <div className="h-4 bg-gray-800 rounded w-full" />
              <div className="h-4 bg-gray-800 rounded w-5/6" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !instructor) {
    return (
      <div className="min-h-screen bg-gray-900 pt-24 pb-12">
        <div className="container mx-auto px-4">
          <div className="bg-red-900/20 border border-red-700 rounded-lg p-6">
            <p className="text-red-400">{error || 'المعلم غير موجود'}</p>
            <button
              onClick={() => router.push('/instructors')}
              className="mt-4 text-blue-400 hover:text-blue-300"
            >
              العودة إلى قائمة المعلمين
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        {/* Header Card */}
        <div className="bg-gray-800 rounded-lg p-8 mb-8">
          <div className="flex items-start gap-6">
            {/* Avatar */}
            <div className="flex-shrink-0">
              {instructor.profile_image_url ? (
                <img
                  src={instructor.profile_image_url}
                  alt={instructor.name_ar}
                  className="w-32 h-32 rounded-full object-cover"
                />
              ) : (
                <div className="w-32 h-32 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                  <span className="text-4xl text-white font-bold">
                    {instructor.name_ar.charAt(0)}
                  </span>
                </div>
              )}
            </div>

            {/* Info */}
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <h1 className="text-3xl font-bold text-white">{instructor.name_ar}</h1>
                {instructor.is_verified && (
                  <CheckCircle className="w-6 h-6 text-blue-400" aria-label="معلم موثق" />
                )}
              </div>

              {/* Rating */}
              <div className="flex items-center gap-2 mb-4">
                <div className="flex items-center gap-1">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <Star
                      key={star}
                      className={`w-5 h-5 ${
                        star <= Math.round(instructor.rating)
                          ? 'text-yellow-400 fill-yellow-400'
                          : 'text-gray-600'
                      }`}
                    />
                  ))}
                </div>
                <span className="text-white font-semibold">{instructor.rating.toFixed(1)}</span>
                <span className="text-gray-400">({instructor.total_reviews} تقييم)</span>
              </div>

              {/* Quick Stats */}
              <div className="flex flex-wrap gap-6 text-gray-300">
                <div className="flex items-center gap-2">
                  <Clock className="w-5 h-5 text-blue-400" />
                  <span>{instructor.years_experience} سنوات خبرة</span>
                </div>
                <div className="flex items-center gap-2">
                  <BookOpen className="w-5 h-5 text-green-400" />
                  <span>{subjects.length} مواد تدريسية</span>
                </div>
                <div className="flex items-center gap-2">
                  <Award className="w-5 h-5 text-purple-400" />
                  <span>{instructor.certifications_ar.length} شهادات</span>
                </div>
              </div>

              {/* Hourly Rate */}
              <div className="mt-4">
                <span className="text-2xl font-bold text-white">${instructor.hourly_rate}</span>
                <span className="text-gray-400 mr-2">/ ساعة</span>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex flex-col gap-3">
              <button className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-lg transition-colors">
                حجز جلسة
              </button>
              <button className="px-6 py-3 bg-gray-700 hover:bg-gray-600 text-white font-medium rounded-lg transition-colors">
                إرسال رسالة
              </button>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-8">
            {/* Bio */}
            <div className="bg-gray-800 rounded-lg p-6">
              <h2 className="text-xl font-bold text-white mb-4">نبذة عن المعلم</h2>
              <p className="text-gray-300 leading-relaxed whitespace-pre-wrap">
                {instructor.bio_ar || 'لا توجد نبذة متاحة.'}
              </p>
            </div>

            {/* Subjects */}
            <div className="bg-gray-800 rounded-lg p-6">
              <h2 className="text-xl font-bold text-white mb-4">المواد التدريسية</h2>
              <div className="grid grid-cols-2 gap-4">
                {subjects.map((subject) => (
                  <div
                    key={subject.id}
                    className="bg-gray-700 rounded-lg p-4 hover:bg-gray-600 transition-colors"
                  >
                    <h3 className="text-white font-semibold mb-2">{subject.name_ar}</h3>
                    <div className="flex items-center gap-2 text-sm text-gray-400">
                      <span className="px-2 py-1 bg-gray-600 rounded">
                        {subject.level === 'elementary' && 'ابتدائي'}
                        {subject.level === 'intermediate' && 'متوسط'}
                        {subject.level === 'advanced' && 'متقدم'}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Certifications */}
            {instructor.certifications_ar.length > 0 && (
              <div className="bg-gray-800 rounded-lg p-6">
                <h2 className="text-xl font-bold text-white mb-4">الشهادات والمؤهلات</h2>
                <ul className="space-y-3">
                  {instructor.certifications_ar.map((cert, index) => (
                    <li key={index} className="flex items-start gap-3 text-gray-300">
                      <Award className="w-5 h-5 text-purple-400 flex-shrink-0 mt-1" />
                      <span>{cert}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Availability */}
            <div className="bg-gray-800 rounded-lg p-6">
              <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
                <Calendar className="w-5 h-5 text-blue-400" />
                الأوقات المتاحة
              </h2>
              <div className="space-y-3">
                {Object.entries(instructor.availability).length === 0 ? (
                  <p className="text-gray-400">لا توجد أوقات متاحة حالياً</p>
                ) : (
                  Object.entries(instructor.availability).map(([day, slots]) => (
                    <div key={day} className="border-b border-gray-700 pb-3 last:border-0">
                      <div className="text-white font-semibold mb-2">
                        {translateDay(day)}
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {(slots as string[]).map((slot, idx) => (
                          <span
                            key={idx}
                            className="px-3 py-1 bg-blue-900/30 text-blue-300 text-sm rounded"
                          >
                            {slot}
                          </span>
                        ))}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>

            {/* Stats */}
            <div className="bg-gray-800 rounded-lg p-6">
              <h2 className="text-xl font-bold text-white mb-4">إحصائيات</h2>
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-gray-400">التقييم</span>
                  <span className="text-white font-semibold">{instructor.rating.toFixed(1)}/5</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-gray-400">عدد التقييمات</span>
                  <span className="text-white font-semibold">{instructor.total_reviews}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-gray-400">سنوات الخبرة</span>
                  <span className="text-white font-semibold">{instructor.years_experience}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-gray-400">الحالة</span>
                  <span className={`font-semibold ${instructor.is_active ? 'text-green-400' : 'text-red-400'}`}>
                    {instructor.is_active ? 'نشط' : 'غير نشط'}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function translateDay(day: string): string {
  const days: Record<string, string> = {
    'Sunday': 'الأحد',
    'Monday': 'الاثنين',
    'Tuesday': 'الثلاثاء',
    'Wednesday': 'الأربعاء',
    'Thursday': 'الخميس',
    'Friday': 'الجمعة',
    'Saturday': 'السبت',
  };
  return days[day] || day;
}
