package main

import (
	"context"
	"fmt"

	"github.com/z0ne-dev/mgx/v2"
	"go.uber.org/zap"
)

type migrationLogger struct {
	Logger *zap.Logger
}

func (l *migrationLogger) Log(msg string, data map[string]any) {
	l.Logger.Info(msg, zap.Namespace("migration"), zap.Any("data", data))
}

func migrateSchema(ctx context.Context, cmd mgx.Commands) error {
	// Signs
	if _, err := cmd.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS "signs" (
			"id" int4 NOT NULL,
			"created_at" timestamptz NOT NULL DEFAULT now(),
			"updated_at" timestamptz NOT NULL DEFAULT now(),
			"video_url" text NOT NULL,
			"description" text NOT NULL,
			"frequency" text,
			"vocable" text,
			"transcription" text,
			"deleted" bool NOT NULL DEFAULT false,
			"unusual" bool NOT NULL DEFAULT false,
			"hidden_words" _text,
			PRIMARY KEY ("id")
		)`,
	); err != nil {
		return err
	}

	if _, err := cmd.Exec(
		ctx,
		`CREATE INDEX IF NOT EXISTS "signs_updated_at_idx" ON "signs" ("updated_at")`,
	); err != nil {
		return err
	}

	// Categories
	if _, err := cmd.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS "categories" (
				"id" int4 NOT NULL,
				"name" text NOT NULL,
				"slug" text NOT NULL,
				PRIMARY KEY ("id")
		)`,
	); err != nil {
		return err
	}

	if _, err := cmd.Exec(
		ctx,
		`CREATE INDEX IF NOT EXISTS "categories_slug_idx" ON "categories" ("slug")`,
	); err != nil {
		return err
	}

	// Signs categories relationship
	if _, err := cmd.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS "signs_categories" (
				"sign_id" int4 NOT NULL REFERENCES "signs" ("id") ON DELETE CASCADE,
				"category_id" int4 NOT NULL REFERENCES "categories" ("id") ON DELETE CASCADE,
				PRIMARY KEY ("sign_id","category_id")
		);`,
	); err != nil {
		return err
	}

	// Words
	if _, err := cmd.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS "words" (
			"id" int4 NOT NULL,
			"sign_id" int4 NOT NULL REFERENCES "signs" ("id") ON DELETE CASCADE,
			"word" text NOT NULL,
			PRIMARY KEY ("id")
		);`,
	); err != nil {
		return err
	}

	if _, err := cmd.Exec(
		ctx,
		`CREATE INDEX IF NOT EXISTS "words_sign_id_idx" ON "words" ("sign_id")`,
	); err != nil {
		return err
	}

	// Phrases
	if _, err := cmd.Exec(
		ctx,
		`CREATE TABLE IF NOT EXISTS "phrases" (
			"id" int4 NOT NULL,
			"sign_id" int4 NOT NULL REFERENCES "signs" ("id") ON DELETE CASCADE,
			"video_url" text NOT NULL,
			"phrase" text NOT NULL,
			PRIMARY KEY ("id")
		);`,
	); err != nil {
		return err
	}

	if _, err := cmd.Exec(
		ctx,
		`CREATE INDEX IF NOT EXISTS "phrases_sign_id_idx" ON "phrases" ("sign_id")`,
	); err != nil {
		return err
	}

	return nil
}

func migrateViews(ctx context.Context, cmd mgx.Commands) error {
	// Signs view
	if _, err := cmd.Exec(
		ctx,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS signs_view AS (
				SELECT
				signs.id,
				signs.updated_at,
				jsonb_build_object(
					'id', signs.id,
					'unusual', signs.unusual,
					'video_url', signs.video_url,
					'updated_at', signs.updated_at,
					'description', signs.description,
					'vocable', signs.vocable,
					'transcription', signs.transcription,
					'frequency', signs.frequency,
					'categories', (
						SELECT jsonb_agg(
							jsonb_build_object(
								'id', categories.id,
								'name', categories.name
							)
						)
						FROM categories
						INNER JOIN signs_categories ON signs_categories.sign_id = signs.id AND
							signs_categories.category_id = categories.id
					),
					'words', (
						SELECT jsonb_agg(
							words.word
						)
						FROM words
						WHERE words.sign_id = signs.id
					),
					'phrases', (
						SELECT jsonb_agg(
							jsonb_build_object(
								'id', phrases.id,
								'video_url', phrases.video_url,
								'phrase', phrases.phrase
							)
						)
						FROM phrases
						WHERE phrases.sign_id = signs.id
					)
				) AS sign
				FROM signs
			)`,
	); err != nil {
		return fmt.Errorf("unable to create materialized view: %w", err)
	}

	// Words view
	if _, err := cmd.Exec(
		ctx,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS words_view AS (
				SELECT
				words.id,
				CASE
				WHEN signs.unusual THEN
				jsonb_build_array(
					words.id,
					signs.id,
					words.word,
					true
				)
				ELSE jsonb_build_array(
					words.id,
					signs.id,
					words.word
				) END AS word,
				(
					SELECT ARRAY_AGG(signs_categories.category_id)
					FROM signs_categories
					WHERE signs_categories.sign_id = signs.id
				) AS categories,
				to_tsvector('simple', array_to_string(ARRAY[words.word] || signs.hidden_words, ',')) || to_tsvector('simple', signs.id::text) AS tsv
				FROM signs
				INNER JOIN words ON words.sign_id = signs.id
			)`,
	); err != nil {
		return fmt.Errorf("unable to create materialized view: %w", err)
	}

	// Categories view
	if _, err := cmd.Exec(
		ctx,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS categories_view AS (
				SELECT
				categories.id,
				jsonb_build_array(
					categories.id,
					categories.name,
					(
						SELECT COUNT(*)
						FROM signs_categories
						WHERE signs_categories.category_id = categories.id
					)
				) AS category
				FROM categories
			)`,
	); err != nil {
		return fmt.Errorf("unable to create materialized view: %w", err)
	}

	// Create a unique index on id
	if _, err := cmd.Exec(
		ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS signs_view_id_idx ON signs_view (id)`,
	); err != nil {
		return fmt.Errorf("unable to create index on materialized view: %w", err)
	}

	// Create a unique index on id
	if _, err := cmd.Exec(
		ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS words_view_id_idx ON words_view (id)`,
	); err != nil {
		return fmt.Errorf("unable to create index on materialized view: %w", err)
	}

	if _, err := cmd.Exec(
		ctx,
		`CREATE INDEX IF NOT EXISTS words_view_categories_idx ON words_view USING GIN (categories)`,
	); err != nil {
		return fmt.Errorf("unable to create index on materialized view: %w", err)
	}

	if _, err := cmd.Exec(
		ctx,
		`CREATE INDEX IF NOT EXISTS words_view_tsv_idx ON words_view USING GIN (tsv)`,
	); err != nil {
		return fmt.Errorf("unable to create index on materialized view: %w", err)
	}

	// Create a unique index on id
	if _, err := cmd.Exec(
		ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS categories_view_id_idx ON categories_view (id)`,
	); err != nil {
		return fmt.Errorf("unable to create index on materialized view: %w", err)
	}

	return nil
}

func migrateNormalizedName(ctx context.Context, cmd mgx.Commands) error {
	if _, err := cmd.Exec(
		ctx,
		`CREATE INDEX IF NOT EXISTS words_view_normalized_name_idx ON words_view USING BTREE (LOWER(word->>2) COLLATE "se-SE-x-icu")`,
	); err != nil {
		return fmt.Errorf("unable to create index on materialized view: %w", err)
	}

	if _, err := cmd.Exec(
		ctx,
		`CREATE INDEX IF NOT EXISTS categories_view_normalized_name_idx ON categories_view USING BTREE (LOWER(category->>1) COLLATE "se-SE-x-icu")`,
	); err != nil {
		return fmt.Errorf("unable to create index on materialized view: %w", err)
	}

	return nil
}
