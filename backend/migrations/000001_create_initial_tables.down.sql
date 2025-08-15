-- Drop indexes
DROP INDEX IF EXISTS idx_interactions_highlight_id;
DROP INDEX IF EXISTS idx_highlights_session_id;

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS interactions;
DROP TABLE IF EXISTS highlights;
DROP TABLE IF EXISTS sessions;