IF OBJECT_ID('x.empty_table_without_fixtures', 'U') IS NOT NULL DROP TABLE x.empty_table_without_fixtures;
IF EXISTS(SELECT 1 FROM sys.schemas WHERE name = 'x') DROP SCHEMA x;

IF OBJECT_ID('comments', 'U') IS NOT NULL DROP TABLE comments;
IF OBJECT_ID('posts_tags', 'U') IS NOT NULL DROP TABLE posts_tags;
IF OBJECT_ID('posts', 'U') IS NOT NULL DROP TABLE posts;
IF OBJECT_ID('tags', 'U') IS NOT NULL DROP TABLE tags;
IF OBJECT_ID('users', 'U') IS NOT NULL DROP TABLE users;

CREATE TABLE posts (
	id INT IDENTITY PRIMARY KEY
	,title VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at DATETIME NOT NULL
	,updated_at DATETIME NOT NULL
);

CREATE TABLE tags (
	id INT IDENTITY PRIMARY KEY
	,name VARCHAR(255) NOT NULL
	,created_at DATETIME NOT NULL
	,updated_at DATETIME NOT NULL
);
GO

CREATE SCHEMA x AUTHORIZATION dbo;
GO

CREATE TABLE x.empty_table_without_fixtures (
  id INT IDENTITY PRIMARY KEY
  ,name VARCHAR(255) NOT NULL
);

CREATE TABLE posts_tags (
	post_id INTEGER NOT NULL
	,tag_id INTEGER NOT NULL
	,PRIMARY KEY (post_id, tag_id)
	,FOREIGN KEY (post_id) REFERENCES posts (id)
	,FOREIGN KEY (tag_id) REFERENCES tags (id)
);

CREATE TABLE comments (
	id INT IDENTITY PRIMARY KEY NOT NULL
	,post_id INTEGER NOT NULL
	,author_name VARCHAR(255) NOT NULL
	,author_email VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at DATETIME NOT NULL
	,updated_at DATETIME NOT NULL
	,FOREIGN KEY (post_id) REFERENCES posts (id)
);

CREATE TABLE users (
	id INT IDENTITY PRIMARY KEY NOT NULL
	,attributes NVARCHAR(MAX) NOT NULL
);
