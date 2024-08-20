CREATE TABLE contents (
    id          SERIAL       NOT NULL,
    creator     VARCHAR(20)  NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    image       BYTEA,
    PRIMARY KEY (id)
);
INSERT INTO contents (creator, title, description, image) VALUES ('riota','hoge','moge','image');
