IF OBJECT_ID('non_default_schema.empty_table_without_fixtures', 'U') IS NOT NULL
	DROP TABLE non_default_schema.empty_table_without_fixtures;
IF EXISTS(SELECT 1 FROM sys.schemas WHERE name = 'non_default_schema')
	DROP SCHEMA non_default_schema;


IF OBJECT_ID('transactions', 'U') IS NOT NULL DROP TABLE transactions;
IF OBJECT_ID('accounts', 'U') IS NOT NULL DROP TABLE accounts;
IF OBJECT_ID('votes', 'U') IS NOT NULL DROP TABLE votes;
IF OBJECT_ID('comments', 'U') IS NOT NULL DROP TABLE comments;
IF OBJECT_ID('posts_tags', 'U') IS NOT NULL DROP TABLE posts_tags;
IF OBJECT_ID('posts', 'U') IS NOT NULL DROP TABLE posts;
IF OBJECT_ID('tags', 'U') IS NOT NULL DROP TABLE tags;
IF OBJECT_ID('assets', 'U') IS NOT NULL DROP TABLE assets;
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

CREATE SCHEMA non_default_schema AUTHORIZATION dbo;
GO

CREATE TABLE non_default_schema.empty_table_without_fixtures (
	id INT IDENTITY PRIMARY KEY
	,name VARCHAR(255) NOT NULL
);

CREATE TABLE posts_tags (
	post_id INTEGER NOT NULL
	,tag_id INTEGER NOT NULL
	,PRIMARY KEY (post_id, tag_id)
	,FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE
	,FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
);

CREATE TABLE comments (
	id INT IDENTITY PRIMARY KEY NOT NULL
	,post_id INTEGER NOT NULL
	,author_name VARCHAR(255) NOT NULL
	,author_email VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at DATETIME NOT NULL
	,updated_at DATETIME NOT NULL
	,FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE
);

CREATE TABLE votes (
	id INT IDENTITY PRIMARY KEY
	,comment_id INTEGER NOT NULL
	,created_at DATETIME NOT NULL
	,updated_at DATETIME NOT NULL
	,FOREIGN KEY (comment_id) REFERENCES comments (id) ON DELETE CASCADE
);

CREATE TABLE users (
	id INT IDENTITY PRIMARY KEY NOT NULL
	,attributes NVARCHAR(MAX) NOT NULL
);

CREATE TABLE assets (
	id INT IDENTITY PRIMARY KEY NOT NULL
	,data VARBINARY(MAX) NOT NULL
);

CREATE TABLE accounts (
	id INT IDENTITY PRIMARY KEY NOT NULL
	,user_id INT NOT NULL
	,currency VARCHAR(3) NOT NULL
	,balance INT NOT NULL
	,created_at DATETIME NOT NULL
	,updated_at DATETIME NOT NULL
	,FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE transactions (
	id INT IDENTITY PRIMARY KEY NOT NULL
	,account_id INT NOT NULL
	,user_id INT NOT NULL
	,currency VARCHAR(3) NOT NULL
	,amount INT NOT NULL
	,created_at DATETIME NOT NULL
	,updated_at DATETIME NOT NULL
	,FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE
	,FOREIGN KEY (user_id) REFERENCES users (id)
);