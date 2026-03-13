DROP INDEX IF EXISTS idx_overrides_created;
DROP INDEX IF EXISTS idx_fields_source;
DROP INDEX IF EXISTS idx_sessions_created;
DROP INDEX IF EXISTS idx_sessions_phone_created;
DROP INDEX IF EXISTS idx_sessions_phone;

ALTER TABLE call_sessions DROP COLUMN IF EXISTS sentiment_summary;

DROP TABLE IF EXISTS follow_ups;
DROP TABLE IF EXISTS triage_assessments;
DROP TABLE IF EXISTS sentiment_logs;
DROP TABLE IF EXISTS patient_history_cache;
