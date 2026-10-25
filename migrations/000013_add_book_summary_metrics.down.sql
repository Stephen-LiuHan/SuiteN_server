DROP INDEX IF EXISTS idx_books_summary_completed_at;

ALTER TABLE books
    DROP COLUMN IF EXISTS summary_generated_at,
    DROP COLUMN IF EXISTS summary_saved_at,
    DROP COLUMN IF EXISTS summary_edited_at,
    DROP COLUMN IF EXISTS summary_helpfulness,
    DROP COLUMN IF EXISTS summary_capture_count,
    DROP COLUMN IF EXISTS summary_completed_at;
