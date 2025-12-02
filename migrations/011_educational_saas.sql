-- Migration: Educational SaaS Enhancement
-- Description: Add school terms, grades, attendance, notifications, and meeting integrations

-- School Terms table (academic years/semesters)
CREATE TABLE IF NOT EXISTS school_terms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ar VARCHAR(255) NOT NULL,
    name_en VARCHAR(255),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    academic_year VARCHAR(20) NOT NULL, -- e.g., "2024-2025"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_school_terms_active ON school_terms(is_active);
CREATE INDEX idx_school_terms_dates ON school_terms(start_date, end_date);

-- Grade Levels table (12th grade, etc.)
CREATE TABLE IF NOT EXISTS grade_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ar VARCHAR(100) NOT NULL, -- الصف الثاني عشر
    name_en VARCHAR(100), -- 12th Grade
    level_number INTEGER NOT NULL, -- 12
    education_stage VARCHAR(50) NOT NULL, -- elementary, middle, high
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_grade_levels_number ON grade_levels(level_number);
CREATE INDEX idx_grade_levels_stage ON grade_levels(education_stage);

-- Class Sections table (12A, 12B, etc.)
CREATE TABLE IF NOT EXISTS class_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grade_level_id UUID NOT NULL REFERENCES grade_levels(id) ON DELETE CASCADE,
    school_term_id UUID NOT NULL REFERENCES school_terms(id) ON DELETE CASCADE,
    section_name VARCHAR(50) NOT NULL, -- A, B, C or Arabic equivalent
    homeroom_teacher_id UUID REFERENCES users(id) ON DELETE SET NULL,
    max_students INTEGER DEFAULT 40,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(grade_level_id, school_term_id, section_name)
);

CREATE INDEX idx_class_sections_grade ON class_sections(grade_level_id);
CREATE INDEX idx_class_sections_term ON class_sections(school_term_id);

-- Update Users table with student-specific fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS grade_level_id UUID REFERENCES grade_levels(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS class_section_id UUID REFERENCES class_sections(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS parent_phone VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS parent_email VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS date_of_birth DATE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS enrollment_date DATE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_image_url TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS address TEXT;

CREATE INDEX IF NOT EXISTS idx_users_grade_level ON users(grade_level_id);
CREATE INDEX IF NOT EXISTS idx_users_class_section ON users(class_section_id);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Attendance table
CREATE TABLE IF NOT EXISTS attendance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    class_section_id UUID NOT NULL REFERENCES class_sections(id) ON DELETE CASCADE,
    meeting_id UUID REFERENCES meetings(id) ON DELETE SET NULL,
    subject_id UUID REFERENCES subjects(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('present', 'absent', 'late', 'excused')),
    check_in_time TIMESTAMP WITH TIME ZONE,
    check_out_time TIMESTAMP WITH TIME ZONE,
    recorded_by UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(student_id, meeting_id),
    UNIQUE(student_id, class_section_id, date, subject_id)
);

CREATE INDEX idx_attendance_student ON attendance(student_id);
CREATE INDEX idx_attendance_date ON attendance(date);
CREATE INDEX idx_attendance_status ON attendance(status);
CREATE INDEX idx_attendance_meeting ON attendance(meeting_id);
CREATE INDEX idx_attendance_section ON attendance(class_section_id);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title_ar VARCHAR(255) NOT NULL,
    title_en VARCHAR(255),
    message_ar TEXT NOT NULL,
    message_en TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('info', 'warning', 'success', 'error', 'assignment', 'meeting', 'grade', 'attendance')),
    channel VARCHAR(50) NOT NULL DEFAULT 'in_app' CHECK (channel IN ('in_app', 'email', 'sms', 'push')),
    reference_type VARCHAR(50), -- meeting, assignment, course, etc.
    reference_id UUID,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivery_status VARCHAR(20) DEFAULT 'pending' CHECK (delivery_status IN ('pending', 'sent', 'delivered', 'failed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created ON notifications(created_at DESC);

-- Meeting Integrations (Google Meet, Zoom, Jitsi)
CREATE TABLE IF NOT EXISTS meeting_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('google_meet', 'zoom', 'jitsi', 'internal')),
    external_meeting_id VARCHAR(255),
    meeting_link TEXT NOT NULL,
    host_link TEXT, -- for Zoom host URL
    password VARCHAR(100),
    settings JSONB DEFAULT '{}'::jsonb, -- provider-specific settings
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(meeting_id, provider)
);

CREATE INDEX idx_meeting_integrations_meeting ON meeting_integrations(meeting_id);
CREATE INDEX idx_meeting_integrations_provider ON meeting_integrations(provider);

-- Student Grades table (per subject)
CREATE TABLE IF NOT EXISTS student_grades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    school_term_id UUID NOT NULL REFERENCES school_terms(id) ON DELETE CASCADE,
    assignment_id UUID REFERENCES assignments(id) ON DELETE SET NULL,
    grade_type VARCHAR(50) NOT NULL CHECK (grade_type IN ('exam', 'quiz', 'homework', 'participation', 'project', 'final')),
    max_points INTEGER NOT NULL,
    points_earned INTEGER NOT NULL,
    percentage DECIMAL(5, 2),
    letter_grade VARCHAR(5),
    notes TEXT,
    graded_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_student_grades_student ON student_grades(student_id);
CREATE INDEX idx_student_grades_subject ON student_grades(subject_id);
CREATE INDEX idx_student_grades_term ON student_grades(school_term_id);
CREATE INDEX idx_student_grades_type ON student_grades(grade_type);

-- Timetable/Schedule table
CREATE TABLE IF NOT EXISTS class_schedule (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_section_id UUID NOT NULL REFERENCES class_sections(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    instructor_id UUID NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sunday
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    room_name VARCHAR(100),
    school_term_id UUID NOT NULL REFERENCES school_terms(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(class_section_id, day_of_week, start_time, school_term_id)
);

CREATE INDEX idx_class_schedule_section ON class_schedule(class_section_id);
CREATE INDEX idx_class_schedule_instructor ON class_schedule(instructor_id);
CREATE INDEX idx_class_schedule_day ON class_schedule(day_of_week);

-- Subject-Class mapping (which subjects are taught to which class)
CREATE TABLE IF NOT EXISTS class_subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_section_id UUID NOT NULL REFERENCES class_sections(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    instructor_id UUID REFERENCES instructors(id) ON DELETE SET NULL,
    school_term_id UUID NOT NULL REFERENCES school_terms(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(class_section_id, subject_id, school_term_id)
);

CREATE INDEX idx_class_subjects_section ON class_subjects(class_section_id);
CREATE INDEX idx_class_subjects_subject ON class_subjects(subject_id);

-- Insert 12th grade subjects (Syrian Baccalaureate - Scientific/Literary)
INSERT INTO subjects (name_ar, level, category) VALUES
('الرياضيات - الثالث الثانوي العلمي', 'advanced', 'mathematics'),
('الفيزياء - الثالث الثانوي العلمي', 'advanced', 'sciences'),
('الكيمياء - الثالث الثانوي العلمي', 'advanced', 'sciences'),
('علم الأحياء - الثالث الثانوي العلمي', 'advanced', 'sciences'),
('اللغة العربية - البكالوريا', 'advanced', 'languages'),
('اللغة الإنجليزية - البكالوريا', 'advanced', 'languages'),
('اللغة الفرنسية - البكالوريا', 'advanced', 'languages'),
('التاريخ - الثالث الثانوي الأدبي', 'advanced', 'social_studies'),
('الجغرافيا - الثالث الثانوي الأدبي', 'advanced', 'social_studies'),
('الفلسفة والمنطق', 'advanced', 'humanities')
ON CONFLICT DO NOTHING;

-- Insert default grade levels
INSERT INTO grade_levels (name_ar, name_en, level_number, education_stage) VALUES
('الصف الأول الابتدائي', '1st Grade', 1, 'elementary'),
('الصف الثاني الابتدائي', '2nd Grade', 2, 'elementary'),
('الصف الثالث الابتدائي', '3rd Grade', 3, 'elementary'),
('الصف الرابع الابتدائي', '4th Grade', 4, 'elementary'),
('الصف الخامس الابتدائي', '5th Grade', 5, 'elementary'),
('الصف السادس الابتدائي', '6th Grade', 6, 'elementary'),
('الصف السابع', '7th Grade', 7, 'middle'),
('الصف الثامن', '8th Grade', 8, 'middle'),
('الصف التاسع', '9th Grade', 9, 'middle'),
('الصف الأول الثانوي', '10th Grade', 10, 'high'),
('الصف الثاني الثانوي', '11th Grade', 11, 'high'),
('الصف الثالث الثانوي (البكالوريا)', '12th Grade', 12, 'high')
ON CONFLICT DO NOTHING;

-- Insert default school term
INSERT INTO school_terms (name_ar, name_en, start_date, end_date, is_active, academic_year) VALUES
('الفصل الدراسي الأول 2024-2025', 'First Semester 2024-2025', '2024-09-01', '2025-01-31', TRUE, '2024-2025'),
('الفصل الدراسي الثاني 2024-2025', 'Second Semester 2024-2025', '2025-02-01', '2025-06-30', FALSE, '2024-2025')
ON CONFLICT DO NOTHING;
