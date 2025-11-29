"use client";
import { useState } from 'react';
import { MeetingService } from '@/services/domain.service';
import type { Subject } from '@/types/domains';
import { useTranslations } from 'next-intl';
import { useAuth } from '@/store';

interface BookingModalProps {
  instructorId: string;
  specializationSubjectIds: string[];
  allSubjects: Subject[];
  availability: Record<string, string[]>;
  onClose: () => void;
  onSuccess: () => void;
}

export function BookingModal({ instructorId, specializationSubjectIds, allSubjects, availability, onClose, onSuccess }: BookingModalProps) {
  const t = useTranslations();
  const { user } = useAuth();
  const [subjectId, setSubjectId] = useState<string>('');
  const [weekday, setWeekday] = useState<string>('');
  const [slot, setSlot] = useState<string>('');
  const [duration, setDuration] = useState<number>(60);
  const [title, setTitle] = useState<string>('جلسة تعليمية');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const subjects = allSubjects.filter(s => specializationSubjectIds.includes(s.id));

  const weekDaysEnglish = Object.keys(availability);

  function nextDateForWeekday(weekdayEnglish: string): Date {
    const target = ['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'].indexOf(weekdayEnglish);
    const now = new Date();
    for (let i=0;i<14;i++) { // search next 2 weeks
      const d = new Date(now.getFullYear(), now.getMonth(), now.getDate()+i);
      if (d.getDay() === target) return d;
    }
    return now;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!subjectId || !weekday || !slot || !user) {
      setError('الرجاء تعبئة جميع الحقول');
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const baseDate = nextDateForWeekday(weekday);
      const [hour, minute] = slot.split(':').map(Number);
      baseDate.setHours(hour, minute, 0, 0);
      const scheduled_at = baseDate.toISOString();
      await MeetingService.createMeeting({
        instructor_id: instructorId,
        student_id: user.id,
        subject_id: subjectId,
        title_ar: title,
        scheduled_at,
        duration
      });
      setSuccess(true);
      onSuccess();
      setTimeout(onClose, 1200);
    } catch (err:any) {
      setError(err.message || 'فشل الحجز');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="bg-gray-800 w-full max-w-lg rounded-lg shadow-xl p-6 relative">
        <button onClick={onClose} className="absolute top-3 left-3 text-gray-400 hover:text-white">✕</button>
        <h2 className="text-2xl font-bold text-white mb-4">{t('booking.title')}</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-gray-300 mb-1">{t('booking.subject')}</label>
            <select value={subjectId} onChange={e=>setSubjectId(e.target.value)} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500">
              <option value="">-- اختر المادة --</option>
              {subjects.map(s=>(<option key={s.id} value={s.id}>{s.name_ar}</option>))}
            </select>
          </div>
          <div>
            <label className="block text-sm text-gray-300 mb-1">{t('booking.date')}</label>
            <select value={weekday} onChange={e=>setWeekday(e.target.value)} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500">
              <option value="">-- اختر اليوم --</option>
              {weekDaysEnglish.map(d=>(<option key={d} value={d}>{translateDay(d)}</option>))}
            </select>
          </div>
          {weekday && (
            <div>
              <label className="block text-sm text-gray-300 mb-1">{t('booking.slot')}</label>
              <select value={slot} onChange={e=>setSlot(e.target.value)} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500">
                <option value="">-- اختر الوقت --</option>
                {availability[weekday].map((s,i)=>(<option key={i} value={s}>{s}</option>))}
              </select>
            </div>
          )}
          <div>
            <label className="block text-sm text-gray-300 mb-1">{t('booking.duration')}</label>
            <input type="number" min={30} step={15} value={duration} onChange={e=>setDuration(Number(e.target.value))} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500" />
          </div>
          <div>
            <label className="block text-sm text-gray-300 mb-1">عنوان الجلسة</label>
            <input type="text" value={title} onChange={e=>setTitle(e.target.value)} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500" />
          </div>
          {error && <div className="text-red-400 text-sm">{error}</div>}
          {success && <div className="text-green-400 text-sm">{t('booking.success')}</div>}
          <div className="flex gap-3 pt-2">
            <button type="submit" disabled={loading} className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-semibold rounded">
              {loading ? 'جارٍ الحجز…' : t('booking.submit')}
            </button>
            <button type="button" onClick={onClose} className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded">
              {t('booking.cancel')}
            </button>
          </div>
        </form>
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
