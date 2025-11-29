-- Migration: Create instructors, subjects, meetings, and study materials tables
-- Description: Comprehensive instructor domain with subjects, availability, meetings, and materials

-- Create subjects table
CREATE TABLE IF NOT EXISTS subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ar VARCHAR(255) NOT NULL,
    level VARCHAR(50) NOT NULL CHECK (level IN ('elementary', 'intermediate', 'advanced')),
    category VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_subjects_category ON subjects(category);
CREATE INDEX idx_subjects_level ON subjects(level);

-- Create instructors table
CREATE TABLE IF NOT EXISTS instructors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name_ar VARCHAR(255) NOT NULL,
    bio_ar TEXT,
    specialization JSONB DEFAULT '[]'::jsonb, -- Array of subject IDs
    hourly_rate DECIMAL(10, 2) DEFAULT 0.00,
    rating DECIMAL(3, 2) DEFAULT 0.00 CHECK (rating >= 0 AND rating <= 5),
    total_reviews INTEGER DEFAULT 0,
    years_experience INTEGER DEFAULT 0,
    certifications_ar JSONB DEFAULT '[]'::jsonb, -- Array of certification strings
    availability JSONB DEFAULT '{}'::jsonb, -- {day: [time slots]}
    is_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    profile_image_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX idx_instructors_user_id ON instructors(user_id);
CREATE INDEX idx_instructors_rating ON instructors(rating DESC);
CREATE INDEX idx_instructors_hourly_rate ON instructors(hourly_rate);
CREATE INDEX idx_instructors_is_verified ON instructors(is_verified);
CREATE INDEX idx_instructors_is_active ON instructors(is_active);
CREATE INDEX idx_instructors_specialization ON instructors USING GIN(specialization);

-- Create meetings table
CREATE TABLE IF NOT EXISTS meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instructor_id UUID NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    student_id UUID REFERENCES users(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
    title_ar VARCHAR(255) NOT NULL,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    duration INTEGER NOT NULL, -- minutes
    meeting_url TEXT,
    room_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'cancelled', 'no-show')),
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_meetings_instructor_id ON meetings(instructor_id);
CREATE INDEX idx_meetings_student_id ON meetings(student_id);
CREATE INDEX idx_meetings_subject_id ON meetings(subject_id);
CREATE INDEX idx_meetings_scheduled_at ON meetings(scheduled_at);
CREATE INDEX idx_meetings_status ON meetings(status);
CREATE INDEX idx_meetings_room_id ON meetings(room_id);

-- Create study_materials table
CREATE TABLE IF NOT EXISTS study_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID REFERENCES courses(id) ON DELETE CASCADE,
    instructor_id UUID NOT NULL REFERENCES instructors(id) ON DELETE CASCADE,
    title_ar VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('pdf', 'slides', 'notes', 'worksheet', 'video', 'audio')),
    file_url TEXT NOT NULL,
    file_size BIGINT DEFAULT 0,
    downloads INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_study_materials_course_id ON study_materials(course_id);
CREATE INDEX idx_study_materials_instructor_id ON study_materials(instructor_id);
CREATE INDEX idx_study_materials_type ON study_materials(type);
CREATE INDEX idx_study_materials_created_at ON study_materials(created_at DESC);

-- Insert default subjects (common Syrian curriculum subjects)
INSERT INTO subjects (name_ar, level, category) VALUES
('اللغة العربية', 'elementary', 'languages'),
('اللغة الإنجليزية', 'elementary', 'languages'),
('الرياضيات', 'elementary', 'mathematics'),
('العلوم', 'elementary', 'sciences'),
('الدراسات الاجتماعية', 'elementary', 'social_studies'),
('التربية الإسلامية', 'elementary', 'religious'),
('الرياضيات المتقدمة', 'intermediate', 'mathematics'),
('الفيزياء', 'intermediate', 'sciences'),
('الكيمياء', 'intermediate', 'sciences'),
('الأحياء', 'intermediate', 'sciences'),
('التاريخ', 'intermediate', 'social_studies'),
('الجغرافيا', 'intermediate', 'social_studies'),
('الحاسوب', 'intermediate', 'technology'),
('حساب التفاضل والتكامل', 'advanced', 'mathematics'),
('الجبر الخطي', 'advanced', 'mathematics'),
('الفيزياء المتقدمة', 'advanced', 'sciences'),
('الكيمياء التحليلية', 'advanced', 'sciences'),
('علم الأحياء الجزيئي', 'advanced', 'sciences'),
('البرمجة', 'advanced', 'technology'),
('الإحصاء', 'advanced', 'mathematics')
ON CONFLICT DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_instructors_updated_at BEFORE UPDATE ON instructors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_meetings_updated_at BEFORE UPDATE ON meetings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_study_materials_updated_at BEFORE UPDATE ON study_materials
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
