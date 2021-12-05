DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts_tags;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS assets
(
	id UInt64,
	data String
) ENGINE MergeTree ORDER BY id;

CREATE TABLE IF NOT EXISTS comments
(
	id           UInt64,
	post_id      UInt64,
	author_name  String,
	author_email String,
	content      String,
	created_at   DateTime DEFAULT '2021-05-12 01:02:00',
	updated_at   DateTime DEFAULT '2021-05-12 01:02:00'
) ENGINE MergeTree ORDER BY created_at;

CREATE TABLE IF NOT EXISTS posts
(
	id         UInt64,
	title      String,
	content    String,
	created_at DateTime DEFAULT '2021-05-12 01:02:00',
	updated_at DateTime DEFAULT '2021-05-12 01:02:00'
) ENGINE MergeTree ORDER BY created_at;

CREATE TABLE IF NOT EXISTS posts_tags
(
	post_id    UInt64,
	tag_id     UInt64,
	created_at DateTime DEFAULT now()
) ENGINE MergeTree ORDER BY created_at;

CREATE TABLE IF NOT EXISTS tags
(
	id         UInt64,
	name       String,
	created_at DateTime DEFAULT '2021-05-12 01:02:00',
	updated_at DateTime DEFAULT '2021-05-12 01:02:00'
) ENGINE MergeTree ORDER BY created_at;

CREATE TABLE IF NOT EXISTS users
(
	id         UInt64,
	attributes String,
	created_at DateTime DEFAULT now()
) ENGINE MergeTree ORDER BY created_at;

CREATE TABLE IF NOT EXISTS votes
(
	id         UInt64,
	comment_id UInt64,
	created_at DateTime DEFAULT '2021-05-12 01:02:00',
	updated_at DateTime DEFAULT '2021-05-12 01:02:00'
) ENGINE MergeTree ORDER BY created_at
