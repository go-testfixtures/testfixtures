DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts_tags;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS assets;

CREATE TABLE posts (
	id         UInt64,
	title      String,
	content    String,
	created_at DateTime DEFAULT '2000-01-01 00:00:00',
	updated_at DateTime DEFAULT '2000-01-01 00:00:00'
) ENGINE = MergeTree ORDER BY id;

CREATE TABLE tags (
	id         UInt64,
	name       String,
	created_at DateTime DEFAULT '2000-01-01 00:00:00',
	updated_at DateTime DEFAULT '2000-01-01 00:00:00'
) ENGINE = MergeTree ORDER BY id;

CREATE TABLE posts_tags (
	post_id    UInt64,
	tag_id     UInt64,
	created_at DateTime DEFAULT now()
) ENGINE = MergeTree ORDER BY (post_id, tag_id);

CREATE TABLE comments (
	id           UInt64,
	post_id      UInt64,
	author_name  String,
	author_email String,
	content      String,
	created_at   DateTime DEFAULT '2000-01-01 00:00:00',
	updated_at   DateTime DEFAULT '2000-01-01 00:00:00'
) ENGINE = MergeTree ORDER BY id;

CREATE TABLE votes (
	id         UInt64,
	comment_id UInt64,
	created_at DateTime DEFAULT '2000-01-01 00:00:00',
	updated_at DateTime DEFAULT '2000-01-01 00:00:00'
) ENGINE = MergeTree ORDER BY id;

CREATE TABLE users (
	id         UInt64,
	attributes String
) ENGINE = MergeTree ORDER BY id;

CREATE TABLE assets (
	id   UInt64,
	data String
) ENGINE = MergeTree ORDER BY id;
