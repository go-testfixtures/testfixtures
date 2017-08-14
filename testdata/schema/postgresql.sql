DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts_tags;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;

CREATE TABLE posts (
	id SERIAL PRIMARY KEY
	,title VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at TIMESTAMP NOT NULL
	,updated_at TIMESTAMP NOT NULL
);

CREATE TABLE tags (
	id SERIAL PRIMARY KEY
	,name VARCHAR(255) NOT NULL
	,created_at TIMESTAMP NOT NULL
	,updated_at TIMESTAMP NOT NULL
);

CREATE TABLE posts_tags (
	post_id INTEGER NOT NULL
	,tag_id INTEGER NOT NULL
	,PRIMARY KEY (post_id, tag_id)
	,FOREIGN KEY (post_id) REFERENCES posts (id)
	,FOREIGN KEY (tag_id) REFERENCES tags (id)
);

CREATE TABLE comments (
	id SERIAL PRIMARY KEY NOT NULL
	,post_id INTEGER NOT NULL
	,author_name VARCHAR(255) NOT NULL
	,author_email VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at TIMESTAMP NOT NULL
	,updated_at TIMESTAMP NOT NULL
	,FOREIGN KEY (post_id) REFERENCES posts (id)
);

CREATE TABLE users (
	id SERIAL PRIMARY KEY NOT NULL
	,attributes JSONB NOT NULL
);

DROP TABLE IF EXISTS fixtures.comments;
DROP TABLE IF EXISTS fixtures.posts_tags;
DROP TABLE IF EXISTS fixtures.posts;
DROP TABLE IF EXISTS fixtures.tags;
DROP TABLE IF EXISTS fixtures.users;

DROP SCHEMA IF EXISTS fixtures;

CREATE SCHEMA fixtures;

CREATE TABLE fixtures.posts (
	id SERIAL PRIMARY KEY
	,title VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at TIMESTAMP NOT NULL
	,updated_at TIMESTAMP NOT NULL
);

CREATE TABLE fixtures.tags (
	id SERIAL PRIMARY KEY
	,name VARCHAR(255) NOT NULL
	,created_at TIMESTAMP NOT NULL
	,updated_at TIMESTAMP NOT NULL
);

CREATE TABLE fixtures.posts_tags (
	post_id INTEGER NOT NULL
	,tag_id INTEGER NOT NULL
	,PRIMARY KEY (post_id, tag_id)
	,FOREIGN KEY (post_id) REFERENCES fixtures.posts (id)
	,FOREIGN KEY (tag_id) REFERENCES fixtures.tags (id)
);

CREATE TABLE fixtures.comments (
	id SERIAL PRIMARY KEY NOT NULL
	,post_id INTEGER NOT NULL
	,author_name VARCHAR(255) NOT NULL
	,author_email VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at TIMESTAMP NOT NULL
	,updated_at TIMESTAMP NOT NULL
	,FOREIGN KEY (post_id) REFERENCES fixtures.posts (id)
);

CREATE TABLE fixtures.users (
	id SERIAL PRIMARY KEY NOT NULL
	,attributes JSONB NOT NULL
);

