-- Recordings metadata schema
CREATE TABLE IF NOT EXISTS recordings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id UUID NULL,
  instructor_id UUID NOT NULL,
  title_ar TEXT NOT NULL,
  description_ar TEXT,
  subject_id UUID NULL,
  file_url TEXT NOT NULL,
  duration_seconds INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recordings_course ON recordings(course_id);
CREATE INDEX IF NOT EXISTS idx_recordings_instructor ON recordings(instructor_id);
CREATE INDEX IF NOT EXISTS idx_recordings_subject ON recordings(subject_id);

-- trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_recordings()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp_recordings
BEFORE UPDATE ON recordings
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_recordings();
