DROP TABLE IF EXISTS news;

-- таблица с rss-новостями
CREATE TABLE IF NOT EXISTS news (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
	description TEXT,
    pub_date BIGINT CHECK(pub_date > 0) DEFAULT extract(epoch from now()),

    -- Согласно RSS 2.0 у новости(item) есть три обязательных атрибута 
    -- (title, description и link).
    -- link - хороший кандидат в качестве ключа поиска новости.
    link TEXT UNIQUE
);

-- индекс для атрибута pub_date.
-- нисходящий B-tree индекс, так как модель данных предполагает
-- выборку последних по дате публикации новостей
CREATE INDEX IF NOT EXISTS pub_date_idx ON news(pub_date DESC);