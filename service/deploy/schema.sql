-- blog 库如果已经存在，不要再执行 CREATE DATABASE。
-- 部分云数据库会把 “database exists” 当成错误而不是警告。
-- 直接在已有的 blog 库里执行下面的建表语句即可。

CREATE TABLE IF NOT EXISTS app_meta (
  name VARCHAR(64) NOT NULL,
  value VARCHAR(255) NOT NULL,
  PRIMARY KEY (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS categories (
  name VARCHAR(191) NOT NULL,
  description TEXT NOT NULL,
  PRIMARY KEY (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS articles (
  id VARCHAR(64) NOT NULL,
  slug VARCHAR(191) NOT NULL,
  title VARCHAR(500) NOT NULL,
  excerpt TEXT NOT NULL,
  published_at CHAR(10) NOT NULL,
  category VARCHAR(191) NOT NULL,
  read_minutes INT NOT NULL,
  tags JSON NOT NULL,
  featured_code TEXT NULL,
  content MEDIUMTEXT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_articles_slug (slug),
  KEY idx_articles_published (published_at, slug),
  KEY idx_articles_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
