DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts_tags;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;

CREATE TABLE posts (
	id INT PRIMARY KEY AUTO_INCREMENT
	,title VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at TIMESTAMP NOT NULL DEFAULT  '2000-01-01 00:00:00'
	,updated_at TIMESTAMP NOT NULL DEFAULT  '2000-01-01 00:00:00'
);

CREATE TABLE tags (
	id INT PRIMARY KEY AUTO_INCREMENT
	,name VARCHAR(255) NOT NULL
	,created_at TIMESTAMP NOT NULL DEFAULT  '2000-01-01 00:00:00'
	,updated_at TIMESTAMP NOT NULL DEFAULT  '2000-01-01 00:00:00'
);

CREATE TABLE posts_tags (
	post_id INT NOT NULL
	,tag_id INT NOT NULL
	,PRIMARY KEY (post_id, tag_id)
	,FOREIGN KEY (post_id) REFERENCES posts (id)
	,FOREIGN KEY (tag_id) REFERENCES tags (id)
);

CREATE TABLE comments (
	id INT PRIMARY KEY AUTO_INCREMENT
	,post_id INT NOT NULL
	,author_name VARCHAR(255) NOT NULL
	,author_email VARCHAR(255) NOT NULL
	,content TEXT NOT NULL
	,created_at TIMESTAMP NOT NULL DEFAULT  '2000-01-01 00:00:00'
	,updated_at TIMESTAMP NOT NULL DEFAULT  '2000-01-01 00:00:00'
	,FOREIGN KEY (post_id) REFERENCES posts (id)
);

CREATE TABLE users (
	id INT PRIMARY KEY AUTO_INCREMENT
	,attributes TEXT NOT NULL
);

CREATE TABLE favorites (
	id INT PRIMARY KEY AUTO_INCREMENT
	,post_id INT NOT NULL
);
