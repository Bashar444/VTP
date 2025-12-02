-- Migration: 013_seed_courses.sql
-- Description: Seed sample courses for the 9 subjects
-- Date: 2025-12-03

-- Get admin user id and create sample courses
DO $$
DECLARE
    admin_id uuid;
BEGIN
    SELECT id INTO admin_id FROM users WHERE role = 'admin' LIMIT 1;
    
    IF admin_id IS NOT NULL THEN
        -- Insert sample courses for all 9 subjects
        INSERT INTO courses (code, name, description, instructor_id, department, semester, year, status, max_students)
        VALUES 
            ('MATH101', 'الرياضيات - المستوى الأول', 'دورة شاملة في الرياضيات للمستوى الأول - تشمل الجبر والهندسة والحساب', admin_id, 'الرياضيات', 'الأول', 2025, 'active', 100),
            ('PHYS101', 'الفيزياء - أساسيات', 'تعلم أساسيات الفيزياء والميكانيكا والحركة والقوى', admin_id, 'العلوم', 'الأول', 2025, 'active', 100),
            ('CHEM101', 'الكيمياء العامة', 'مقدمة في الكيمياء العامة والعضوية والتفاعلات الكيميائية', admin_id, 'العلوم', 'الأول', 2025, 'active', 100),
            ('BIO101', 'الأحياء - علم الحياة', 'دراسة الكائنات الحية وأنظمتها الحيوية والتطور', admin_id, 'العلوم', 'الأول', 2025, 'active', 100),
            ('ARAB101', 'اللغة العربية - قواعد ونحو', 'تعلم قواعد اللغة العربية والنحو والصرف والبلاغة', admin_id, 'اللغات', 'الأول', 2025, 'active', 100),
            ('ENG101', 'اللغة الإنجليزية - للمبتدئين', 'دورة اللغة الإنجليزية للمبتدئين - القواعد والمحادثة', admin_id, 'اللغات', 'الأول', 2025, 'active', 100),
            ('HIST101', 'التاريخ العربي والإسلامي', 'دراسة التاريخ العربي والإسلامي من العصر الجاهلي إلى العصر الحديث', admin_id, 'الدراسات الاجتماعية', 'الأول', 2025, 'active', 100),
            ('GEO101', 'الجغرافيا العامة', 'دراسة الجغرافيا الطبيعية والبشرية والمناخ', admin_id, 'الدراسات الاجتماعية', 'الأول', 2025, 'active', 100),
            ('PHIL101', 'الفلسفة - مقدمة', 'مقدمة في الفلسفة والمنطق والتفكير النقدي', admin_id, 'العلوم الإنسانية', 'الأول', 2025, 'active', 100)
        ON CONFLICT (code) DO NOTHING;
    END IF;
END $$;

-- Show result
SELECT code, name, department FROM courses;
