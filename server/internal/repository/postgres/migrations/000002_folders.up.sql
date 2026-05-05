CREATE TABLE folders (
                         id UUID PRIMARY KEY,
                         user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         parent_id UUID REFERENCES folders(id) ON DELETE CASCADE,
                         name TEXT NOT NULL,
                         created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE file_metadata ADD COLUMN folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;