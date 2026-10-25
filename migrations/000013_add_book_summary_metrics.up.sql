-- Book-level summary metrics for the North Star Metric:
-- monthly count of completed "自分のまとめ" (summary_completed_at)
ALTER TABLE books
    ADD COLUMN summary_generated_at   TIMESTAMP WITH TIME ZONE NULL,
    ADD COLUMN summary_saved_at       TIMESTAMP WITH TIME ZONE NULL,
    ADD COLUMN summary_edited_at      TIMESTAMP WITH TIME ZONE NULL,
    ADD COLUMN summary_helpfulness    SMALLINT NULL CHECK (summary_helpfulness BETWEEN 1 AND 3),
    ADD COLUMN summary_capture_count  INTEGER NULL CHECK (summary_capture_count >= 0),
    ADD COLUMN summary_completed_at   TIMESTAMP WITH TIME ZONE NULL;

CREATE INDEX idx_books_summary_completed_at ON books(summary_completed_at)
    WHERE summary_completed_at IS NOT NULL;
