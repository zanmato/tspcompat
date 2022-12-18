CREATE TABLE "signs" (
    id int NOT NULL,
    ref_id int NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    video_url TEXT NOT NULL,
    description TEXT NOT NULL,
    frequency TEXT,
    deleted boolean NOT NULL DEFAULT false,
    unusual boolean NOT NULL DEFAULT false,
    PRIMARY KEY (id)
);

CREATE INDEX ON signs USING BTREE (updated_at);

CREATE TABLE "tags" (
  id int NOT NULL,
  name TEXT NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE "examples" (
  id int NOT NULL,
  sign_id int NOT NULL REFERENCES signs (id) ON DELETE CASCADE,
  video_url TEXT NOT NULL,
  description TEXT NOT NULL,
  PRIMARY KEY (id)
);

CREATE INDEX ON examples USING BTREE (sign_id);

CREATE TABLE "words" (
  id int NOT NULL,
  sign_id int NOT NULL REFERENCES signs (id) ON DELETE CASCADE,
  word TEXT NOT NULL,
  PRIMARY KEY (id)
);

CREATE INDEX ON words USING BTREE (sign_id);

CREATE TABLE "signs_tags" (
  sign_id int NOT NULL REFERENCES signs (id) ON DELETE CASCADE,
  tag_id int NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
  PRIMARY KEY (sign_id, tag_id)
);
