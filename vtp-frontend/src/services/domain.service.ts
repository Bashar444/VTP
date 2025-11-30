import { apiClient } from './api.client';
import type {
  Instructor,
  CreateInstructorDTO,
  UpdateInstructorDTO,
  InstructorFilters,
  Subject,
  CreateSubjectDTO,
  Meeting,
  CreateMeetingDTO,
  UpdateMeetingDTO,
  MeetingFilters,
  StudyMaterial,
  CreateMaterialDTO,
  UpdateMaterialDTO,
  MaterialFilters,
  Assignment,
  CreateAssignmentDTO,
  AssignmentSubmission,
  CreateSubmissionDTO,
} from '@/types/domains';

// Instructor Service
export class InstructorService {
  private static baseUrl = '/api/v1/instructors';

  static async getInstructors(filters?: InstructorFilters) {
    const params = new URLSearchParams();
    if (filters?.subject_id) params.append('subject_id', filters.subject_id);
    if (filters?.min_rating) params.append('min_rating', filters.min_rating.toString());
    if (filters?.is_verified !== undefined) params.append('is_verified', filters.is_verified.toString());
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.page_size) params.append('page_size', filters.page_size.toString());

    const response = await apiClient.get<{ instructors: Instructor[] }>(
      `${this.baseUrl}?${params.toString()}`
    );
    return response.data.instructors;
  }

  static async getInstructorById(id: string) {
    const response = await apiClient.get<Instructor>(`${this.baseUrl}/${id}`);
    return response.data;
  }

  static async createInstructor(data: CreateInstructorDTO) {
    const response = await apiClient.post<Instructor>(this.baseUrl, data);
    return response.data;
  }

  static async updateInstructor(id: string, data: UpdateInstructorDTO) {
    const response = await apiClient.put<Instructor>(`${this.baseUrl}/${id}`, data);
    return response.data;
  }

  static async deleteInstructor(id: string) {
    await apiClient.delete(`${this.baseUrl}/${id}`);
  }

  static async getAvailability(id: string, date: string) {
    const response = await apiClient.get<{ date: string; slots: string[] }>(
      `${this.baseUrl}/${id}/availability?date=${date}`
    );
    return response.data;
  }
}

// Subject Service
export class SubjectService {
  private static baseUrl = '/api/v1/subjects';

  static async getSubjects(level?: string, category?: string) {
    const params = new URLSearchParams();
    if (level) params.append('level', level);
    if (category) params.append('category', category);

    const response = await apiClient.get<{ subjects: Subject[] }>(
      `${this.baseUrl}?${params.toString()}`
    );
    return response.data.subjects;
  }

  static async getSubjectById(id: string) {
    const response = await apiClient.get<Subject>(`${this.baseUrl}/${id}`);
    return response.data;
  }

  static async createSubject(data: CreateSubjectDTO) {
    const response = await apiClient.post<Subject>(this.baseUrl, data);
    return response.data;
  }

  static async updateSubject(id: string, data: Partial<CreateSubjectDTO>) {
    const response = await apiClient.put<Subject>(`${this.baseUrl}/${id}`, data);
    return response.data;
  }

  static async deleteSubject(id: string) {
    await apiClient.delete(`${this.baseUrl}/${id}`);
  }
}

// Meeting Service
export class MeetingService {
  private static baseUrl = '/api/v1/meetings';

  static async getMeetings(filters?: MeetingFilters) {
    const params = new URLSearchParams();
    if (filters?.instructor_id) params.append('instructor_id', filters.instructor_id);
    if (filters?.student_id) params.append('student_id', filters.student_id);
    if (filters?.subject_id) params.append('subject_id', filters.subject_id);
    if (filters?.status) params.append('status', filters.status);
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.page_size) params.append('page_size', filters.page_size.toString());

    const response = await apiClient.get<{ meetings: Meeting[] }>(
      `${this.baseUrl}?${params.toString()}`
    );
    return response.data.meetings;
  }

  static async getMeetingById(id: string) {
    const response = await apiClient.get<Meeting>(`${this.baseUrl}/${id}`);
    return response.data;
  }

  static async createMeeting(data: CreateMeetingDTO) {
    const response = await apiClient.post<Meeting>(this.baseUrl, data);
    return response.data;
  }

  static async updateMeeting(id: string, data: UpdateMeetingDTO) {
    const response = await apiClient.put<Meeting>(`${this.baseUrl}/${id}`, data);
    return response.data;
  }

  static async deleteMeeting(id: string) {
    await apiClient.delete(`${this.baseUrl}/${id}`);
  }

  static async cancelMeeting(id: string) {
    const response = await apiClient.post<Meeting>(`${this.baseUrl}/${id}/cancel`, {});
    return response.data;
  }

  static async completeMeeting(id: string) {
    const response = await apiClient.post<Meeting>(`${this.baseUrl}/${id}/complete`, {});
    return response.data;
  }
}

// Study Material Service
export class MaterialService {
  private static baseUrl = '/api/v1/materials';

  static async getMaterials(filters?: MaterialFilters) {
    const params = new URLSearchParams();
    if (filters?.course_id) params.append('course_id', filters.course_id);
    if (filters?.instructor_id) params.append('instructor_id', filters.instructor_id);
    if (filters?.type) params.append('type', filters.type);
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.page_size) params.append('page_size', filters.page_size.toString());

    const response = await apiClient.get<{ materials: StudyMaterial[] }>(
      `${this.baseUrl}?${params.toString()}`
    );
    return response.data.materials;
  }

  static async getMaterialById(id: string) {
    const response = await apiClient.get<StudyMaterial>(`${this.baseUrl}/${id}`);
    return response.data;
  }

  static async createMaterial(data: CreateMaterialDTO) {
    const response = await apiClient.post<StudyMaterial>(this.baseUrl, data);
    return response.data;
  }

  static async updateMaterial(id: string, data: UpdateMaterialDTO) {
    const response = await apiClient.put<StudyMaterial>(`${this.baseUrl}/${id}`, data);
    return response.data;
  }

  static async deleteMaterial(id: string) {
    await apiClient.delete(`${this.baseUrl}/${id}`);
  }

  static getDownloadUrl(id: string) {
    return `${this.baseUrl}/${id}/download`;
  }
}

// Assignment Service
export class AssignmentService {
  private static baseUrl = '/api/v1/assignments';

  static async getAssignments(params?: { instructor_id?: string; subject_id?: string }): Promise<Assignment[]> {
    const q = new URLSearchParams();
    if (params?.instructor_id) q.append('instructor_id', params.instructor_id);
    if (params?.subject_id) q.append('subject_id', params.subject_id);
    const res = await apiClient.get<{ assignments: Assignment[] }>(`${this.baseUrl}?${q.toString()}`);
    return res.data.assignments;
  }

  static async createAssignment(data: CreateAssignmentDTO): Promise<Assignment> {
    const res = await apiClient.post<Assignment>(this.baseUrl, data);
    return res.data;
  }

  static async getAssignment(id: string): Promise<Assignment> {
    const res = await apiClient.get<Assignment>(`${this.baseUrl}/${id}`);
    return res.data;
  }

  static async submit(sub: CreateSubmissionDTO): Promise<AssignmentSubmission> {
    const res = await apiClient.post<AssignmentSubmission>(`${this.baseUrl}/${sub.assignment_id}/submit`, sub);
    return res.data;
  }

  static async listSubmissions(assignmentId: string): Promise<AssignmentSubmission[]> {
    const res = await apiClient.get<{ submissions: AssignmentSubmission[] }>(`${this.baseUrl}/${assignmentId}/submissions`);
    return res.data.submissions;
  }

  static async grade(submissionId: string, grade: number, feedback_ar?: string): Promise<AssignmentSubmission> {
    const res = await apiClient.post<AssignmentSubmission>(`${this.baseUrl}/submissions/${submissionId}/grade`, { grade, feedback_ar });
    return res.data;
  }
}
