-- Patient history cache for predictive pre-population
CREATE TABLE patient_history_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_phone VARCHAR(20) NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    field_value TEXT NOT NULL,
    last_session_id UUID NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE,
    occurrence_count INTEGER NOT NULL DEFAULT 1,
    first_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(patient_phone, field_name)
);

CREATE INDEX idx_patient_history_phone ON patient_history_cache(patient_phone);
CREATE INDEX idx_patient_history_phone_seen ON patient_history_cache(patient_phone, last_seen_at DESC);

-- Sentiment logs for emotion detection
CREATE TABLE sentiment_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE,
    emotion_type VARCHAR(20) NOT NULL,
    intensity DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    lexicon_score DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    pattern_score DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    trigger_text TEXT,
    speaker VARCHAR(20) NOT NULL DEFAULT 'patient',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sentiment_session ON sentiment_logs(session_id);
CREATE INDEX idx_sentiment_emotion ON sentiment_logs(emotion_type);
CREATE INDEX idx_sentiment_created ON sentiment_logs(created_at);

-- Triage assessments for urgency scoring
CREATE TABLE triage_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE,
    urgency_level VARCHAR(20) NOT NULL DEFAULT 'LOW',
    composite_score INTEGER NOT NULL DEFAULT 0,
    symptoms JSONB NOT NULL DEFAULT '[]',
    red_flags TEXT[] DEFAULT '{}',
    modifiers_applied JSONB NOT NULL DEFAULT '[]',
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_triage_session ON triage_assessments(session_id);
CREATE INDEX idx_triage_urgency ON triage_assessments(urgency_level);

-- Follow-ups for auto scheduling
CREATE TABLE follow_ups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE,
    follow_up_type VARCHAR(30) NOT NULL,
    description TEXT,
    raw_text TEXT,
    due_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'DETECTED',
    sf_task_id VARCHAR(255),
    confirmed_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_followup_session ON follow_ups(session_id);
CREATE INDEX idx_followup_status ON follow_ups(status);
CREATE INDEX idx_followup_due ON follow_ups(due_date);

-- Add sentiment summary to call_sessions
ALTER TABLE call_sessions ADD COLUMN IF NOT EXISTS sentiment_summary JSONB DEFAULT '{}';

-- Analytics indexes on existing tables
CREATE INDEX IF NOT EXISTS idx_sessions_phone ON call_sessions(patient_phone);
CREATE INDEX IF NOT EXISTS idx_sessions_phone_created ON call_sessions(patient_phone, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_created ON call_sessions(created_at);
CREATE INDEX IF NOT EXISTS idx_fields_source ON extracted_fields(source);
CREATE INDEX IF NOT EXISTS idx_overrides_created ON agent_overrides(created_at);
