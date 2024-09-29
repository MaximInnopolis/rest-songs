-- +goose Up
-- +goose StatementBegin
CREATE TABLE songs (
                       id SERIAL PRIMARY KEY,
                       "group" TEXT NOT NULL,
                       song TEXT NOT NULL,
                       text TEXT NOT NULL,
                       link TEXT NOT NULL,
                       release_date TIMESTAMPTZ NOT NULL,
                       created_at TIMESTAMPTZ DEFAULT NOW(),
                       updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_songs_group ON songs("group");
CREATE INDEX idx_songs_song ON songs(song);
CREATE INDEX idx_songs_release_date ON songs(release_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE songs;
-- +goose StatementEnd