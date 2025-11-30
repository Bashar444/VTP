"use client";
export const dynamic = 'force-dynamic';
import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { RecordingService, type RecordingDTO } from '@/services/recording.service';

export default function RecordingDetailsPage() {
  const t = useTranslations();
  const { id } = useParams() as { id: string };
  const router = useRouter();
  const [rec, setRec] = useState<RecordingDTO | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string|null>(null);
  const [deleting, setDeleting] = useState(false);
  const [toast, setToast] = useState<{type:'success'|'error'; message:string}|null>(null);

  async function load() {
    try {
      setLoading(true); setError(null);
      const r = await RecordingService.get(id);
      setRec(r);
    } catch (e:any) { setError(e.message || t('assignments.loadError')); }
    finally { setLoading(false); }
  }

  async function handleDelete() {
    if (!rec) return;
    try {
      setDeleting(true); setToast(null);
      await RecordingService.delete(rec.id);
      setToast({type:'success', message: t('recordings.deleted')});
      setTimeout(()=> router.push('/recordings'), 900);
    } catch (e:any) {
      setToast({type:'error', message: t('recordings.deleteError')});
      setTimeout(()=> setToast(null), 2500);
    }
    finally { setDeleting(false); }
  }

  useEffect(()=>{ if (id) load(); }, [id]);

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        {loading ? (
          <div className="text-gray-300">{t('common.loading')}</div>
        ) : error ? (
          <div className="text-red-400">{error}</div>
        ) : rec ? (
          <div className="space-y-6">
            <header className="bg-gray-800 rounded-lg p-6">
              <h1 className="text-2xl font-bold text-white mb-2">{t('recordings.details')}</h1>
              <h2 className="text-white text-lg">{rec.title_ar}</h2>
              {rec.description_ar && (<p className="text-gray-300 mt-2">{rec.description_ar}</p>)}
              <div className="text-gray-400 text-sm mt-3">{t('recordings.duration')}: {rec.duration_seconds} ثانية</div>
              <div className="flex gap-3 mt-4">
                <a className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded" href={rec.file_url} target="_blank" rel="noopener">{t('recordings.open')}</a>
                <button onClick={handleDelete} disabled={deleting} className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded disabled:opacity-50">{t('recordings.delete')}</button>
              </div>
              {toast && (
                <div className={`fixed top-24 right-4 z-50 px-4 py-2 rounded shadow ${toast.type==='success'?'bg-green-600 text-white':'bg-red-600 text-white'}`}>
                  {toast.message}
                </div>
              )}
            </header>
          </div>
        ) : null}
      </div>
    </div>
  );
}
