"use client";
export const dynamic = 'force-dynamic';
import { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { AssignmentService, SubjectService } from '@/services/domain.service';
import type { Assignment, Subject } from '@/types/domains';
import { useAuth } from '@/store';
import { Calendar, FileText, PlusCircle } from 'lucide-react';
import Link from 'next/link';

export default function AssignmentsPage() {
  const t = useTranslations();
  const { user } = useAuth();
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string|null>(null);
  const [subjectFilter, setSubjectFilter] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ title_ar: '', description_ar: '', subject_id: '', due_at: '', max_points: 100 });
  const [page, setPage] = useState(1);
  const [pageSize] = useState(9);
  const [total, setTotal] = useState<number | null>(null);

  async function load() {
    try {
      setLoading(true); setError(null);
      const [asgs, subs] = await Promise.all([
        AssignmentService.getAssignments({
          ...(subjectFilter ? { subject_id: subjectFilter } : {}),
          page,
          limit: pageSize
        } as any),
        SubjectService.getSubjects()
      ]);
      // If backend returns array only, we can't know total; else expect {items,total}
      if (Array.isArray(asgs)) {
        setAssignments(asgs);
        setTotal(null);
      } else {
        const anyAsgs: any = asgs as any;
        setAssignments(anyAsgs.items ?? []);
        setTotal(anyAsgs.total ?? null);
      }
      setSubjects(subs);
    } catch (e:any) { setError(e.message || 'فشل التحميل'); }
    finally { setLoading(false); }
  }

  useEffect(()=>{ load(); }, [subjectFilter, page]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!user) { setError('يجب تسجيل الدخول'); return; }
    if (!form.title_ar || !form.due_at) { setError('اكمل الحقول المطلوبة'); return; }
    try {
      setError(null);
      await AssignmentService.createAssignment({
        instructor_id: user.id,
        title_ar: form.title_ar,
        description_ar: form.description_ar,
        subject_id: form.subject_id || undefined,
        due_at: new Date(form.due_at).toISOString(),
        max_points: form.max_points
      });
      setShowCreate(false);
      setForm({ title_ar: '', description_ar: '', subject_id: '', due_at: '', max_points: 100 });
      await load();
    } catch (e:any) { setError(e.message || 'فشل الإضافة'); }
  }

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-3xl font-bold text-white">{t('assignments.title')}</h1>
          {user && (
            <button onClick={()=>setShowCreate(v=>!v)} className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded flex items-center gap-2">
              <PlusCircle className="w-5 h-5" /> {t('assignments.add')}
            </button>
          )}
        </div>

        {/* Filters */}
        <div className="bg-gray-800 rounded-lg p-4 mb-6">
          <select value={subjectFilter} onChange={e=>setSubjectFilter(e.target.value)} className="bg-gray-700 text-white rounded px-3 py-2">
            <option value="">كل المواد</option>
            {subjects.map(s=> (<option key={s.id} value={s.id}>{s.name_ar}</option>))}
          </select>
        </div>

        {/* Create form */}
        {showCreate && user && (
          <div className="bg-gray-800 rounded-lg p-6 mb-6">
            <h2 className="text-xl font-bold text-white mb-4">{t('assignments.create')}</h2>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm text-gray-300 mb-1">العنوان</label>
                <input value={form.title_ar} onChange={e=>setForm({...form,title_ar:e.target.value})} className="w-full bg-gray-700 text-white rounded px-3 py-2" />
              </div>
              <div>
                <label className="block text-sm text-gray-300 mb-1">الوصف</label>
                <textarea value={form.description_ar} onChange={e=>setForm({...form,description_ar:e.target.value})} className="w-full bg-gray-700 text-white rounded px-3 py-2" />
              </div>
              <div>
                <label className="block text-sm text-gray-300 mb-1">المادة</label>
                <select value={form.subject_id} onChange={e=>setForm({...form,subject_id:e.target.value})} className="w-full bg-gray-700 text-white rounded px-3 py-2">
                  <option value="">—</option>
                  {subjects.map(s=> (<option key={s.id} value={s.id}>{s.name_ar}</option>))}
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-300 mb-1">الموعد النهائي</label>
                <input type="datetime-local" value={form.due_at} onChange={e=>setForm({...form,due_at:e.target.value})} className="w-full bg-gray-700 text-white rounded px-3 py-2" />
              </div>
              <div>
                <label className="block text-sm text-gray-300 mb-1">العلامة القصوى</label>
                <input type="number" min={1} value={form.max_points} onChange={e=>setForm({...form,max_points:Number(e.target.value)})} className="w-full bg-gray-700 text-white rounded px-3 py-2" />
              </div>
              {error && <div className="text-red-400 text-sm">{error}</div>}
              <div className="flex gap-3">
                <button type="submit" className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded">{t('assignments.create')}</button>
                <button type="button" onClick={()=>setShowCreate(false)} className="px-6 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded">{t('assignments.cancel')}</button>
              </div>
            </form>
          </div>
        )}

        {/* List */}
        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_,i)=>(
              <div key={i} className="bg-gray-800 rounded-lg p-6 animate-pulse">
                <div className="h-6 bg-gray-700 rounded w-2/3 mb-3" />
                <div className="h-4 bg-gray-700 rounded w-full mb-2" />
                <div className="h-4 bg-gray-700 rounded w-5/6" />
              </div>
            ))}
          </div>
        ) : assignments.length === 0 ? (
          <div className="bg-gray-800 rounded-lg p-12 text-center text-gray-400">{t('assignments.empty')}</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {assignments.map(a => (
              <div key={a.id} className="bg-gray-800 rounded-lg p-5">
                <div className="flex items-center gap-3 mb-2">
                  <FileText className="w-5 h-5 text-blue-400" />
                  <Link href={`/assignments/${a.id}`} className="text-white font-semibold line-clamp-1 hover:underline">
                    {a.title_ar}
                  </Link>
                </div>
                {a.description_ar && (<p className="text-gray-300 text-sm mb-3 line-clamp-2">{a.description_ar}</p>)}
                <div className="flex items-center gap-2 text-sm text-gray-400 mb-4">
                  <Calendar className="w-4 h-4" />
                  <span>الموعد: {new Date(a.due_at).toLocaleString('ar-EG')}</span>
                  <span className="ml-auto px-2 py-1 bg-purple-900/30 text-purple-300 rounded">{a.max_points} نقطة</span>
                </div>
                <div className="mb-4">
                  <Link href={`/assignments/${a.id}`} className="inline-block px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded">
                    {t('assignments.viewDetails')}
                  </Link>
                </div>
                <AssignmentSubmissionForm assignmentId={a.id} />
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        <div className="flex items-center justify-center gap-3 mt-8">
          <button onClick={()=>setPage(p=>Math.max(1,p-1))} disabled={page<=1} className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded disabled:opacity-50">السابق</button>
          <span className="text-gray-300">صفحة {page}</span>
          <button onClick={()=>setPage(p=>p+1)} className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded">التالي</button>
        </div>
      </div>
    </div>
  );
}

function AssignmentSubmissionForm({ assignmentId }: { assignmentId: string }) {
  const t = useTranslations();
  const { user } = useAuth();
  const [file_url, setFileUrl] = useState('');
  const [notes, setNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState<string|null>(null);
  const [messageType, setMessageType] = useState<'success'|'error'|'info'>('info');

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!user) { setMessage('يجب تسجيل الدخول'); return; }
    try {
      setSubmitting(true); setMessage(null);
      await AssignmentService.submit({ assignment_id: assignmentId, student_id: user.id, file_url, notes });
      setMessage(t('assignments.submitSuccess'));
      setMessageType('success');
      setFileUrl(''); setNotes('');
    } catch (e:any) { setMessage(e.message || t('assignments.submitError')); setMessageType('error'); }
    finally { setSubmitting(false); }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div>
        <label className="block text-sm text-gray-300 mb-1">رابط الملف</label>
        <input value={file_url} onChange={e=>setFileUrl(e.target.value)} className="w-full bg-gray-700 text-white rounded px-3 py-2" />
      </div>
      <div>
        <label className="block text-sm text-gray-300 mb-1">ملاحظات</label>
        <textarea value={notes} onChange={e=>setNotes(e.target.value)} className="w-full bg-gray-700 text-white rounded px-3 py-2" />
      </div>
      {message && (
        <div className={`text-sm ${messageType==='success' ? 'text-green-400' : messageType==='error' ? 'text-red-400' : 'text-gray-300'}`}>{message}</div>
      )}
      <button type="submit" disabled={submitting} className="w-full py-2 bg-blue-600 hover:bg-blue-700 text-white rounded disabled:opacity-50">
        {submitting ? 'جارٍ الإرسال…' : t('assignments.submit')}
      </button>
    </form>
  );
}
