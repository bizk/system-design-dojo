package migrations

import "gorm.io/gorm"

func Run(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version integer PRIMARY KEY,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`).Error; err != nil {
		return err
	}

	var applied int64
	if err := db.Table("schema_migrations").Where("version = ?", 1).Count(&applied).Error; err != nil {
		return err
	}
	if applied > 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`CREATE TABLE sessions (
			id uuid PRIMARY KEY,
			title text NOT NULL,
			description text NOT NULL DEFAULT '',
			status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'completed')),
			content text NOT NULL DEFAULT '',
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL
		)`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE TABLE session_media (
			id uuid PRIMARY KEY,
			session_id uuid NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			type text NOT NULL CHECK (type IN ('image', 'audio')),
			object_key text NOT NULL UNIQUE,
			filename text NOT NULL,
			content_type text NOT NULL,
			size bigint NOT NULL,
			transcription text,
			created_at timestamptz NOT NULL
		)`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`CREATE INDEX session_media_session_id_idx ON session_media(session_id)`).Error; err != nil {
			return err
		}
		return tx.Exec(`INSERT INTO schema_migrations (version) VALUES (1)`).Error
	})
}
