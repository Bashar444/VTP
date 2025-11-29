"use client";
export const dynamic = 'force-dynamic';
import { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { MaterialService } from '@/services/domain.service';
import type { StudyMaterial } from '@/types/domains';
import { useAuth } from '@/store';
import { FileText, UploadCloud, Download, Trash2, Filter } from 'lucide-react';

const typeLabels: Record<string,string> = {
  pdf: 'ملف PDF',
  slides: 'شرائح عرض',
  notes: 'ملاحظات',
  worksheet: 'ورقة عمل',
  video: 'فيديو',
  audio: 'صوت',
};

export default function MaterialsPage() {
  const t = useTranslations();
  const { user } = useAuth();
  const [materials, setMaterials] = useState<StudyMaterial[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string|null>(null);
  const [typeFilter, setTypeFilter] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ title_ar: '', type: 'pdf', file_url: '', file_size: 0 });
  const [submitting, setSubmitting] = useState(false);

  async function load() {
    try {
      setLoading(true); setError(null);
      const data = await MaterialService.getMaterials(typeFilter ? { type: typeFilter } : undefined);
      setMaterials(data);
    } catch (e:any) {
      setError(e.message || 'فشل التحميل');
    } finally { setLoading(false); }
  }

  useEffect(()=>{ load(); }, [typeFilter]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!user) { setError('يجب تسجيل الدخول'); return; }
    if (!form.title_ar || !form.file_url) { setError('اكمل الحقول المطلوبة'); return; }
    try {
      setSubmitting(true); setError(null);
      await MaterialService.createMaterial({
        instructor_id: user.id,
        title_ar: form.title_ar,
        type: form.type as any,
        file_url: form.file_url,
        file_size: Number(form.file_size),
      });
      setShowForm(false);
      setForm({ title_ar: '', type: 'pdf', file_url: '', file_size: 0 });
      await load();
    } catch (e:any) { setError(e.message || 'فشل الإضافة'); }
    finally { setSubmitting(false); }
  }

  return (
    <div className="min-h-screen bg-gray-900 pt-24 pb-12">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-3xl font-bold text-white">{t('materials.title')}</h1>
          {user && (
            <button onClick={()=>setShowForm(v=>!v)} className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded flex items-center gap-2">
              <UploadCloud className="w-5 h-5" /> {t('materials.add')}
            </button>
          )}
        </div>

        {/* Filters */}
        <div className="bg-gray-800 rounded-lg p-4 mb-6 flex flex-wrap gap-4 items-center">
          <div className="flex items-center gap-2 text-gray-300">
            <Filter className="w-4 h-4" /> {t('materials.filter.type')}
          </div>
          <select value={typeFilter} onChange={e=>setTypeFilter(e.target.value)} className="bg-gray-700 text-white rounded px-3 py-2">
            <option value="">كل الأنواع</option>
            {Object.keys(typeLabels).map(k=>(<option key={k} value={k}>{typeLabels[k]}</option>))}
          </select>
        </div>

        {/* Upload Form */}
        {showForm && user && (
          <div className="bg-gray-800 rounded-lg p-6 mb-6">
            <h2 className="text-xl font-bold text-white mb-4">{t('materials.upload')}</h2>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm text-gray-300 mb-1">العنوان</label>
                <input value={form.title_ar} onChange={e=>setForm({...form,title_ar:e.target.value})} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label className="block text-sm text-gray-300 mb-1">{t('materials.type')}</label>
                <select value={form.type} onChange={e=>setForm({...form,type:e.target.value})} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500">
                  {Object.keys(typeLabels).map(k=>(<option key={k} value={k}>{typeLabels[k]}</option>))}
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-300 mb-1">رابط الملف</label>
                <input value={form.file_url} onChange={e=>setForm({...form,file_url:e.target.value})} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label className="block text-sm text-gray-300 mb-1">{t('materials.size')} (بايت)</label>
                <input type="number" value={form.file_size} onChange={e=>setForm({...form,file_size:Number(e.target.value)})} className="w-full bg-gray-700 text-white rounded px-3 py-2 focus:ring-2 focus:ring-blue-500" />
              </div>
              {error && <div className="text-red-400 text-sm">{error}</div>}
              <div className="flex gap-3">
                <button type="submit" disabled={submitting} className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded disabled:opacity-50">
                  {submitting ? 'جارٍ الرفع…' : t('materials.upload')}
                </button>
                <button type="button" onClick={()=>setShowForm(false)} className="px-6 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded">إلغاء</button>
              </div>
            </form>
          </div>
        )}

        {/* List */}
        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_,i)=>(
              <div key={i} className="bg-gray-800 rounded-lg p-6 animate-pulse">
                <div className="h-6 bg-gray-700 rounded w-2/3 mb-4" />
                <div className="h-4 bg-gray-700 rounded w-full mb-2" />
                <div className="h-4 bg-gray-700 rounded w-5/6" />
              </div>
            ))}
          </div>
        ) : materials.length === 0 ? (
          <div className="bg-gray-800 rounded-lg p-12 text-center text-gray-400">
            {t('materials.noItems')}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {materials.map(m => (
              <div key={m.id} className="bg-gray-800 rounded-lg p-5 hover:bg-gray-750 transition-colors">
                <div className="flex items-center gap-3 mb-3">
                  <FileText className="w-5 h-5 text-blue-400" />
                  <h3 className="text-white font-semibold line-clamp-1">{m.title_ar}</h3>
                </div>
                <div className="text-sm text-gray-400 mb-2">{typeLabels[m.type] || m.type}</div>
                <div className="text-xs text-gray-500 mb-4">{t('materials.size')}: {Math.round(m.file_size/1024)} كيلوبايت</div>
                <div className="flex items-center justify-between text-sm">
                  <a href={m.file_url} target="_blank" className="flex items-center gap-1 text-blue-400 hover:text-blue-300">
                    <Download className="w-4 h-4" /> {t('materials.download')}
                  </a>
                  {user?.id === m.instructor_id && (
                    <button onClick={()=>handleDelete(m.id)} className="flex items-center gap-1 text-red-400 hover:text-red-300 text-sm">
                      <Trash2 className="w-4 h-4" /> {t('materials.delete')}
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

async function handleDelete(id: string) {
  if (!confirm('تأكيد الحذف؟')) return;
  try {
    await MaterialService.deleteMaterial(id);
    // naive reload after deletion
    window.location.reload();
  } catch (e:any) {
    alert(e.message || 'فشل الحذف');
  }
}
