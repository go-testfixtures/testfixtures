DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts_tags;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS assets;

CREATE SEQUENCE posts_sequence OPTIONS (
  sequence_kind="bit_reversed_positive"
);

CREATE TABLE posts (
	id          INT64 DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE posts_sequence)),
	title       STRING(MAX),
	content     STRING(MAX),
	created_at  TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP()),
	updated_at  TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP())
) PRIMARY KEY (id);

CREATE SEQUENCE tags_sequence OPTIONS (
  sequence_kind="bit_reversed_positive"
);

CREATE TABLE tags (
	id          INT64 DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE tags_sequence)),
	name        STRING(MAX),
	created_at  TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP()),
	updated_at  TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP())
) PRIMARY KEY (id);

CREATE TABLE posts_tags (
	post_id     INT64,
	tag_id      INT64,
	created_at  TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP()),
  CONSTRAINT FK_posts_tags_post_id FOREIGN KEY (post_id) REFERENCES posts (id),
  CONSTRAINT FK_posts_tags_tag_id FOREIGN KEY (tag_id) REFERENCES tags (id)
) PRIMARY KEY (post_id, tag_id);

CREATE SEQUENCE comments_sequence OPTIONS (
  sequence_kind="bit_reversed_positive"
);

CREATE TABLE comments (
	id            INT64 DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE comments_sequence)),
	post_id       INT64,
	author_name   STRING(MAX),
	author_email  STRING(MAX),
	content       STRING(MAX),
	created_at    TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP()),
	updated_at    TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP()),
  CONSTRAINT FK_comments_post_id FOREIGN KEY (post_id) REFERENCES posts (id)
) PRIMARY KEY (id);

CREATE SEQUENCE votes_sequence OPTIONS (
  sequence_kind="bit_reversed_positive"
);

CREATE TABLE votes (
	id          INT64 DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE votes_sequence)),
	comment_id  INT64,
	created_at  TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP()),
	updated_at  TIMESTAMP NOT NULL DEFAULT(CURRENT_TIMESTAMP()),
  CONSTRAINT FK_votes_comment_id FOREIGN KEY (comment_id) REFERENCES comments (id)
) PRIMARY KEY (id);

CREATE SEQUENCE users_sequence OPTIONS (
  sequence_kind="bit_reversed_positive"
);

CREATE TABLE users (
	id          INT64 DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE users_sequence)),
	attributes  STRING(MAX)
) PRIMARY KEY (id);

CREATE SEQUENCE assets_sequence OPTIONS (
  sequence_kind="bit_reversed_positive"
);

CREATE TABLE assets (
	id    INT64 DEFAULT (GET_NEXT_SEQUENCE_VALUE(SEQUENCE assets_sequence)),
	data  BYTES(MAX)
) PRIMARY KEY (id);
