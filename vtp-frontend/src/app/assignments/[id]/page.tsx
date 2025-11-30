"use client";
export const dynamic = 'force-dynamic';
import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { AssignmentService } from '@/services/domain.service';
import type { Assignment, AssignmentSubmission } from '@/types/domains';
import { useAuth } from '@/store';
import { FileText, Calendar, CheckCircle } from 'lucide-react';

export default function AssignmentDetailsPage() {
  const t = useTranslations();
  const { id } = useParams() as { id: string };
  const { user } = useAuth();
  const [assignment, setAssignment] = useState<Assignment| null>(null);
  const [submissions, setSubmissions] = useState<AssignmentSubmission[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string|null>(null);

  async function load() {
    try {
      setLoading(true); setError(null);
      const a = await AssignmentService.getAssignment(id);
      setAssignment(a);
      const subs = await AssignmentService.listSubmissions(id);
      setSubmissions(subs);
    } catch (e:any) { setError(e.message || t('assignments.loadError')); }
    finally { setLoading(false); }
  }

  useEffect(()=>{ if (id) load(); }, [id]);

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        {loading ? (
          <div className="text-gray-300">{t('common.loading')}</div>
        ) : error ? (
          <div className="text-red-400">{error}</div>
        ) : assignment ? (
          <div className="space-y-6">
            <header className="bg-gray-800 rounded-lg p-6">
              <div className="flex items-center gap-3 mb-2">
                <FileText className="w-5 h-5 text-blue-400" />
                <h1 className="text-2xl font-bold text-white">{assignment.title_ar}</h1>
              </div>
              {assignment.description_ar && (
                <p className="text-gray-300">{assignment.description_ar}</p>
              )}
              <div className="flex items-center gap-2 text-sm text-gray-400 mt-3">
                <Calendar className="w-4 h-4" />
                <span>الموعد: {new Date(assignment.due_at).toLocaleString('ar-EG')}</span>
                <span className="ml-auto px-2 py-1 bg-purple-900/30 text-purple-300 rounded">{assignment.max_points} نقطة</span>
              </div>
            </header>

            <section className="bg-gray-800 rounded-lg p-6">
              <h2 className="text-xl font-bold text-white mb-4">{t('assignments.submissions')}</h2>
              {submissions.length === 0 ? (
                <div className="text-gray-400">{t('assignments.noSubmissions')}</div>
              ) : (
                <div className="space-y-4">
                  {submissions.map(s => (
                    <SubmissionRow key={s.id} submission={s} maxPoints={assignment.max_points} canGrade={!!user && user.id === assignment.instructor_id} onGraded={load} />
                  ))}
                </div>
              )}
            </section>
          </div>
        ) : null}
      </div>
    </div>
  );
}

function SubmissionRow({ submission, maxPoints, canGrade, onGraded }: { submission: AssignmentSubmission, maxPoints: number, canGrade: boolean, onGraded: ()=>void }) {
  const t = useTranslations();
  const [grade, setGrade] = useState<number | ''>(submission.grade ?? '');
  const [feedback_ar, setFeedback] = useState<string>(submission.feedback_ar ?? '');
  const [saving, setSaving] = useState(false);
  const graded = submission.graded_at != null;

  async function handleSave() {
    if (!canGrade) return;
    if (grade === '' || Number(grade) < 0 || Number(grade) > maxPoints) return;
    try {
      setSaving(true);
      await AssignmentService.grade(submission.id, Number(grade), feedback_ar);
      onGraded();
    } catch (e) {
      // silent error UI minimal; could expand
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="flex flex-col md:flex-row md:items-center gap-3 bg-gray-700/50 rounded p-4">
      <div className="flex-1">
        <div className="text-white text-sm">طالب: {submission.student_id}</div>
        {submission.file_url && (
          <a href={submission.file_url} target="_blank" rel="noopener" className="text-blue-300 text-sm underline">رابط الملف</a>
        )}
        {submission.notes && (
          <div className="text-gray-300 text-sm mt-1">"{submission.notes}"</div>
        )}
        <div className="text-gray-400 text-xs mt-1">أُرسل في {new Date(submission.submitted_at).toLocaleString('ar-EG')}</div>
      </div>
      <div className="md:w-1/2 w-full">
        {canGrade ? (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3 items-center">
            <input type="number" min={0} max={maxPoints} value={grade} onChange={e=>setGrade(e.target.value === '' ? '' : Number(e.target.value))} className="bg-gray-700 text-white rounded px-3 py-2" placeholder={t('assignments.points')} />
            <input value={feedback_ar} onChange={e=>setFeedback(e.target.value)} className="bg-gray-700 text-white rounded px-3 py-2" placeholder={t('assignments.feedback')} />
            <button onClick={handleSave} disabled={saving || grade === ''} className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded disabled:opacity-50">
              {saving ? 'حفظ…' : t('assignments.grade')}
            </button>
          </div>
        ) : (
          <div className="flex items-center gap-2 text-sm text-gray-300">
            <CheckCircle className="w-4 h-4 text-green-400" />
            <span>{graded ? t('assignments.graded') : t('assignments.pending')}</span>
          </div>
        )}
      </div>
    </div>
  );
}
