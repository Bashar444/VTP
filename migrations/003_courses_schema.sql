-- Phase 3: Course Management System
-- Migration: 003_courses_schema.sql
-- Created: November 24, 2025
-- Purpose: Create course management database tables

-- Courses table
CREATE TABLE IF NOT EXISTS courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    instructor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    department VARCHAR(100),
    semester VARCHAR(20),
    year INT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'archived', 'draft', 'completed')),
    max_students INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT valid_year CHECK (year >= 2020 AND year <= 2100),
    CONSTRAINT valid_max_students CHECK (max_students >= 0)
);

-- Course recordings table (links recordings to courses)
CREATE TABLE IF NOT EXISTS course_recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    recording_id UUID NOT NULL REFERENCES recordings(id) ON DELETE CASCADE,
    lecture_number INT,
    lecture_title VARCHAR(255),
    sequence_order INT NOT NULL,
    is_published BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(course_id, recording_id),
    CONSTRAINT valid_sequence_order CHECK (sequence_order >= 0)
);

-- Course enrollments table (tracks student enrollment)
CREATE TABLE IF NOT EXISTS course_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enrollment_date TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'dropped', 'suspended')),
    UNIQUE(course_id, student_id)
);

-- Course permissions table (manages access control)
CREATE TABLE IF NOT EXISTS course_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'instructor', 'ta', 'student', 'viewer')),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(course_id, user_id)
);

-- Course activity log table (audit trail)
CREATE TABLE IF NOT EXISTS course_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB DEFAULT '{}'::jsonb,
    ip_address VARCHAR(45),
    user_agent TEXT,
    timestamp TIMESTAMP DEFAULT NOW()
);

-- Recording access log table (tracks playback access)
CREATE TABLE IF NOT EXISTS recording_access_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id UUID NOT NULL REFERENCES recordings(id) ON DELETE CASCADE,
    course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('viewed', 'downloaded', 'transcoded', 'shared')),
    ip_address VARCHAR(45),
    user_agent TEXT,
    timestamp TIMESTAMP DEFAULT NOW()
);

-- Create indexes for performance

-- Courses indexes
CREATE INDEX IF NOT EXISTS idx_courses_instructor ON courses(instructor_id);
CREATE INDEX IF NOT EXISTS idx_courses_semester ON courses(semester, year);
CREATE INDEX IF NOT EXISTS idx_courses_status ON courses(status);
CREATE INDEX IF NOT EXISTS idx_courses_code ON courses(code);

-- Course recordings indexes
CREATE INDEX IF NOT EXISTS idx_course_recordings_course ON course_recordings(course_id);
CREATE INDEX IF NOT EXISTS idx_course_recordings_recording ON course_recordings(recording_id);
CREATE INDEX IF NOT EXISTS idx_course_recordings_sequence ON course_recordings(course_id, sequence_order);

-- Course enrollments indexes
CREATE INDEX IF NOT EXISTS idx_course_enrollments_course ON course_enrollments(course_id);
CREATE INDEX IF NOT EXISTS idx_course_enrollments_student ON course_enrollments(student_id);
CREATE INDEX IF NOT EXISTS idx_course_enrollments_status ON course_enrollments(status);
CREATE INDEX IF NOT EXISTS idx_course_enrollments_lookup ON course_enrollments(course_id, student_id);

-- Course permissions indexes
CREATE INDEX IF NOT EXISTS idx_course_permissions_course ON course_permissions(course_id);
CREATE INDEX IF NOT EXISTS idx_course_permissions_user ON course_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_course_permissions_role ON course_permissions(course_id, role);
CREATE INDEX IF NOT EXISTS idx_course_permissions_lookup ON course_permissions(course_id, user_id);

-- Course activity indexes
CREATE INDEX IF NOT EXISTS idx_course_activity_course ON course_activity(course_id);
CREATE INDEX IF NOT EXISTS idx_course_activity_user ON course_activity(user_id);
CREATE INDEX IF NOT EXISTS idx_course_activity_timestamp ON course_activity(timestamp);
CREATE INDEX IF NOT EXISTS idx_course_activity_action ON course_activity(action);

-- Recording access log indexes
CREATE INDEX IF NOT EXISTS idx_recording_access_logs_recording ON recording_access_logs(recording_id);
CREATE INDEX IF NOT EXISTS idx_recording_access_logs_course ON recording_access_logs(course_id);
CREATE INDEX IF NOT EXISTS idx_recording_access_logs_user ON recording_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_recording_access_logs_timestamp ON recording_access_logs(timestamp);

-- Create views for common queries

-- Active courses view
CREATE OR REPLACE VIEW active_courses AS
SELECT 
    c.id,
    c.code,
    c.name,
    c.instructor_id,
    c.semester,
    c.year,
    COUNT(DISTINCT ce.student_id) as enrolled_students,
    COUNT(DISTINCT cr.recording_id) as recording_count
FROM courses c
LEFT JOIN course_enrollments ce ON c.id = ce.course_id AND ce.status = 'active'
LEFT JOIN course_recordings cr ON c.id = cr.course_id
WHERE c.status = 'active'
GROUP BY c.id, c.code, c.name, c.instructor_id, c.semester, c.year;

-- Course dashboard stats view
CREATE OR REPLACE VIEW course_dashboard_stats AS
SELECT 
    c.id as course_id,
    c.name,
    c.code,
    COUNT(DISTINCT ce.student_id) as total_students,
    COUNT(DISTINCT cr.recording_id) as total_recordings,
    COUNT(DISTINCT CASE WHEN cr.is_published = true THEN cr.recording_id END) as published_recordings,
    COUNT(DISTINCT ral.user_id) as unique_viewers,
    COUNT(DISTINCT ral.id) as total_views,
    MAX(ral.timestamp) as last_accessed
FROM courses c
LEFT JOIN course_enrollments ce ON c.id = ce.course_id
LEFT JOIN course_recordings cr ON c.id = cr.course_id
LEFT JOIN recording_access_logs ral ON cr.recording_id = ral.recording_id
GROUP BY c.id, c.name, c.code;

-- Verify migration
SELECT 'Courses table created' as status;
SELECT COUNT(*) as courses_count FROM courses;
SELECT COUNT(*) as enrollments_count FROM course_enrollments;
SELECT COUNT(*) as permissions_count FROM course_permissions;
