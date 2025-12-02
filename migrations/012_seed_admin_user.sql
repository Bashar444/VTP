-- Migration: 012_seed_admin_user.sql
-- Description: Seed admin user and sample courses
-- Date: 2025-12-03

-- Insert admin user (password: Bashar@123 - bcrypt hash)
-- Note: The hash is for 'Bashar@123' using bcrypt with cost 10
INSERT INTO users (email, password_hash, full_name, role, created_at, updated_at)
VALUES (
    'basharalagha21@gmail.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- Bashar@123
    'باشار الأغا',
    'admin',
    NOW(),
    NOW()
) ON CONFLICT (email) DO UPDATE SET
    role = 'admin',
    updated_at = NOW();

-- Create sample courses for the 9 subjects (Arabic names)
-- First get the admin user id
DO $$
DECLARE
    admin_id uuid;
BEGIN
    SELECT id INTO admin_id FROM users WHERE email = 'basharalagha21@gmail.com';
    
    IF admin_id IS NOT NULL THEN
        -- Insert sample courses if they don't exist
        INSERT INTO courses (teacher_id, title_ar, description_ar, visibility, created_at, updated_at)
        SELECT admin_id, title_ar, description_ar, 'public', NOW(), NOW()
        FROM (VALUES
            ('الرياضيات - المستوى الأول', 'دورة شاملة في الرياضيات للمستوى الأول - تشمل الجبر والهندسة والحساب'),
            ('الفيزياء - أساسيات', 'تعلم أساسيات الفيزياء والميكانيكا والحركة والقوى'),
            ('الكيمياء العامة', 'مقدمة في الكيمياء العامة والعضوية والتفاعلات الكيميائية'),
            ('الأحياء - علم الحياة', 'دراسة الكائنات الحية وأنظمتها الحيوية والتطور'),
            ('اللغة العربية - قواعد ونحو', 'تعلم قواعد اللغة العربية والنحو والصرف والبلاغة'),
            ('اللغة الإنجليزية - للمبتدئين', 'دورة اللغة الإنجليزية للمبتدئين - القواعد والمحادثة'),
            ('التاريخ العربي والإسلامي', 'دراسة التاريخ العربي والإسلامي من العصر الجاهلي إلى العصر الحديث'),
            ('الجغرافيا العامة', 'دراسة الجغرافيا الطبيعية والبشرية والمناخ'),
            ('الفلسفة - مقدمة', 'مقدمة في الفلسفة والمنطق والتفكير النقدي')
        ) AS t(title_ar, description_ar)
        WHERE NOT EXISTS (SELECT 1 FROM courses WHERE title_ar = t.title_ar);
    END IF;
END $$;
