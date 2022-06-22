DROP TABLE IF EXISTS news;

CREATE TABLE IF NOT EXISTS news (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
	description TEXT,
    pub_date BIGINT CHECK(pub_date > 0) DEFAULT extract(epoch from now()),
    link TEXT NOT NULL
);