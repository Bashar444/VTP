import api from '@/services/api.client';

export interface RecordingDTO {
  id: string;
  course_id?: string | null;
  instructor_id: string;
  title_ar: string;
  description_ar?: string | null;
  subject_id?: string | null;
  file_url: string;
  duration_seconds: number;
  created_at: string;
  updated_at: string;
  course_title_ar?: string | null;
  subject_name_ar?: string | null;
  instructor_name_ar?: string | null;
}

export interface RecordingFilters {
  course_id?: string;
  instructor_id?: string;
  subject_id?: string;
  page?: number;
  limit?: number;
}

export const RecordingService = {
  async list(filters?: RecordingFilters): Promise<{ items: RecordingDTO[]; total?: number } | RecordingDTO[]> {
    const params = { ...filters };
    const res = await api.get('/recordings', { params });
    // backend returns { items: [], total? }
    return res.data.items !== undefined ? res.data : res.data;
  },
  async create(payload: Partial<RecordingDTO> & { instructor_id: string; title_ar: string; file_url: string }): Promise<RecordingDTO> {
    const res = await api.post('/recordings', payload);
    return res.data;
  },
  async get(id: string): Promise<RecordingDTO> {
    const res = await api.get(`/recordings/${id}`);
    return res.data;
  },
  async delete(id: string): Promise<void> {
    await api.delete(`/recordings/${id}`);
  }
};
