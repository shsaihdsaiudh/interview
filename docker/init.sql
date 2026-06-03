-- 用户表
CREATE TABLE users (
    email          TEXT PRIMARY KEY,
    password_hash  TEXT NOT NULL,
    nickname       TEXT NOT NULL,
    student_id     TEXT NOT NULL DEFAULT '',
    department     TEXT NOT NULL DEFAULT '',
    tags           TEXT[] NOT NULL DEFAULT '{}',
    avatar         TEXT NOT NULL DEFAULT '',
    contact_info   TEXT NOT NULL DEFAULT '',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verify_token   TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 空闲时间表
CREATE TABLE availabilities (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(email) ON DELETE CASCADE,
    date       DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time   TIME NOT NULL,
    CONSTRAINT chk_time_range CHECK (end_time > start_time)
);

-- 预约表
CREATE TABLE appointments (
    id            TEXT PRIMARY KEY,
    mentor_id     TEXT NOT NULL REFERENCES users(email),
    student_id    TEXT NOT NULL REFERENCES users(email),
    time_slot_id  TEXT NOT NULL REFERENCES availabilities(id),
    message       TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'pending',
    reject_reason TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_status CHECK (status IN ('pending', 'accepted', 'rejected'))
);

-- 招募卡片表
CREATE TABLE recruitment_cards (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL UNIQUE REFERENCES users(email) ON DELETE CASCADE,
    skills           TEXT[] NOT NULL DEFAULT '{}',
    target_companies TEXT[] NOT NULL DEFAULT '{}',
    role             VARCHAR(20) NOT NULL DEFAULT 'both',
    experience_years INT NOT NULL DEFAULT 0,
    bio              TEXT NOT NULL DEFAULT '',
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_role CHECK (role IN ('interviewee', 'interviewer', 'both'))
);

-- 索引
CREATE INDEX idx_avail_user  ON availabilities(user_id);
CREATE INDEX idx_avail_date  ON availabilities(date);
CREATE INDEX idx_appt_mentor  ON appointments(mentor_id);
CREATE INDEX idx_appt_student ON appointments(student_id);
CREATE INDEX idx_appt_slot    ON appointments(time_slot_id);
CREATE INDEX idx_appt_status  ON appointments(status);
CREATE INDEX idx_card_user    ON recruitment_cards(user_id);
CREATE INDEX idx_card_role    ON recruitment_cards(role);
