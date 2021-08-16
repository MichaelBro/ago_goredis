CREATE TABLE news
(
    id       BIGSERIAL PRIMARY KEY,
    title    TEXT      NOT NULL,
    text     TEXT      NOT NULL,
    created  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


