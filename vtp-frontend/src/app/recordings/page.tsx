"use client";
export const dynamic = 'force-dynamic';
import { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { RecordingService, type RecordingDTO } from '@/services/recording.service';
import { SubjectService, InstructorService } from '@/services/domain.service';
import Link from 'next/link';

export default function RecordingsPage() {
  const t = useTranslations();
  const [items, setItems] = useState<RecordingDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string|null>(null);
    const [toast, setToast] = useState<{type:'success'|'error'; message:string}|null>(null);
  const [subject, setSubject] = useState('');
  const [instructor, setInstructor] = useState('');
  const [subjects, setSubjects] = useState<Array<{id:string; name_ar:string}>>([]);
  const [instructors, setInstructors] = useState<Array<{id:string; name_ar?:string; full_name?:string}>>([]);
  const [page, setPage] = useState(1);
  const [limit] = useState(12);
  const [total, setTotal] = useState<number|undefined>(undefined);

  async function load() {
    try {
      setLoading(true); setError(null);
      const recs = await RecordingService.list({ subject_id: subject || undefined, instructor_id: instructor || undefined, page, limit });
      if (Array.isArray(recs)) {
        setItems(recs);
        setTotal(undefined);
      } else {
        setItems(recs.items);
        setTotal(recs.total);
      }
    } catch (e:any) { setError(e.message || 'فشل التحميل'); }
    finally { setLoading(false); }
  }

  useEffect(()=>{ load(); }, [subject, instructor, page]);

  useEffect(()=>{
    // preload subjects and instructors for pickers
    (async ()=>{
      try {
        const [subs, insts] = await Promise.all([
          SubjectService.getSubjects(),
          InstructorService.getInstructors()
        ]);
        setSubjects(subs.map(s=>({id:s.id, name_ar:s.name_ar})));
        setInstructors(insts.map(i=>({id:i.id, name_ar:(i as any).name_ar, full_name:(i as any).full_name})));
      } catch {}
    })();
  }, []);

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-3xl font-bold text-white">{t('recordings.title')}</h1>
          <Link href="/dashboard" className="text-indigo-400 hover:underline">لوحة التحكم</Link>
        </div>

        <div className="bg-gray-800 rounded-lg p-4 mb-6 grid grid-cols-1 md:grid-cols-3 gap-3">
          <select value={subject} onChange={e=>setSubject(e.target.value)} className="bg-gray-700 text-white rounded px-3 py-2">
            <option value="">كل المواد</option>
            {subjects.map(s=> (<option key={s.id} value={s.id}>{s.name_ar}</option>))}
          </select>
          <select value={instructor} onChange={e=>setInstructor(e.target.value)} className="bg-gray-700 text-white rounded px-3 py-2">
            <option value="">كل المعلمين</option>
            {instructors.map(i=> (
              <option key={i.id} value={i.id}>{i.name_ar || i.full_name || i.id}</option>
            ))}
          </select>
          <div className="flex items-center gap-3">
            <button onClick={()=>setPage(p=>Math.max(1,p-1))} disabled={page<=1} className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded disabled:opacity-50">السابق</button>
            <span className="text-gray-300">صفحة {page}{typeof total==='number' ? ` / ${Math.max(1, Math.ceil(total/limit))}` : ''}</span>
            <button onClick={()=>setPage(p=>p+1)} disabled={typeof total==='number' && page>=Math.ceil((total||0)/limit)} className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded disabled:opacity-50">التالي</button>
          </div>
        </div>

        {loading ? (
          <div className="text-gray-300">{t('common.loading')}</div>
        ) : error ? (
          <div className="text-red-400">{error}</div>
        ) : items.length === 0 ? (
          <div className="bg-gray-800 rounded-lg p-12 text-center text-gray-400">{t('recordings.empty')}</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {items.map(r => (
              <div key={r.id} className="bg-gray-800 rounded-lg p-5">
                <h3 className="text-white font-semibold mb-1">{r.title_ar}</h3>
                {r.description_ar && (<p className="text-gray-300 text-sm mb-2">{r.description_ar}</p>)}
                <div className="text-gray-400 text-sm mb-2">{t('recordings.duration')}: {r.duration_seconds} ثانية</div>
                <div className="text-gray-400 text-xs mb-3 space-x-3 rtl:space-x-reverse">
                  { (r as any).course_title_ar && (<span>الدورة: {(r as any).course_title_ar}</span>) }
                  { (r as any).subject_name_ar && (<span>المادة: {(r as any).subject_name_ar}</span>) }
                  { (r as any).instructor_name_ar && (<span>المعلم: {(r as any).instructor_name_ar}</span>) }
                </div>
                <div className="flex gap-3">
                  <a className="inline-block px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded" href={r.file_url} target="_blank" rel="noopener">{t('recordings.open')}</a>
                  <Link href={`/recordings/${r.id}`} className="inline-block px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded">{t('recordings.details')}</Link>
                  <button onClick={async()=>{
                    if (confirm('تأكيد الحذف؟')) {
                      try { await RecordingService.delete(r.id); setToast({type:'success', message: t('recordings.deleted')}); await load(); }
                      catch { setToast({type:'error', message: t('recordings.deleteError')}); }
                      setTimeout(()=> setToast(null), 2000);
                    }
                  }} className="inline-block px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded">{t('recordings.delete')}</button>
                </div>
              </div>
            ))}
          </div>
        )}
        {toast && (
          <div className={`fixed bottom-6 left-1/2 -translate-x-1/2 px-4 py-2 rounded ${toast.type==='success' ? 'bg-green-600 text-white' : 'bg-red-600 text-white'}`}>
            {toast.message}
          </div>
        )}
      </div>
    </div>
  );
}
