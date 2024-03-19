CREATE DATABASE `geofileshare`;
USE `geofileshare`;

CREATE TABLE `user` (
	id INT auto_increment NOT NULL,
	username varchar(100) NOT NULL,
	email varchar(250) NOT NULL,
	active TINYINT DEFAULT 1 NOT NULL,
	first_name varchar(100) NULL,
	last_name varchar(250) NULL,
	CONSTRAINT user_pk PRIMARY KEY (id)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_general_ci;

-- Insert the basic users
INSERT INTO `user` (username, email, active, first_name, last_name) VALUES ('nikos', 'nikos@geosysta.com', 1, 'Nikos', 'Steiakakis');
