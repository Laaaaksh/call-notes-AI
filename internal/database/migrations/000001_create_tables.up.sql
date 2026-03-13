CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE call_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    talkdesk_call_id VARCHAR(255) NOT NULL,
    agent_id VARCHAR(255) NOT NULL,
    patient_phone VARCHAR(20),
    status VARCHAR(50) NOT NULL DEFAULT 'CREATED',
    parent_session_id UUID REFERENCES call_sessions(id),
    sf_record_id VARCHAR(255),
    language_detected VARCHAR(20),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE,
    submitted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_call_id ON call_sessions(talkdesk_call_id);
CREATE INDEX idx_sessions_agent_id ON call_sessions(agent_id);
CREATE INDEX idx_sessions_status ON call_sessions(status);
CREATE INDEX idx_sessions_parent ON call_sessions(parent_session_id);

CREATE TABLE extracted_fields (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE,
    field_name VARCHAR(100) NOT NULL,
    field_value TEXT NOT NULL,
    confidence DECIMAL(5,4) NOT NULL DEFAULT 0.0,
    source VARCHAR(50) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    previous_value TEXT,
    transcript_ref TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, field_name, version)
);

CREATE INDEX idx_fields_session ON extracted_fields(session_id);
CREATE INDEX idx_fields_name ON extracted_fields(field_name);

CREATE TABLE agent_overrides (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE,
    field_name VARCHAR(100) NOT NULL,
    ai_value TEXT NOT NULL,
    agent_value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_overrides_session ON agent_overrides(session_id);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    actor VARCHAR(255) NOT NULL,
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_session ON audit_logs(session_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
