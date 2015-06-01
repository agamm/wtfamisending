CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP TABLE IF EXISTS requests ;
CREATE TABLE requests
(
    id text PRIMARY KEY NOT NULL,
    raw_request text NOT NULL,
    ip inet NOT NULL,
    timestamp timestamp default current_timestamp
);

CREATE UNIQUE INDEX ON requests(id);
ALTER TABLE requests ADD CONSTRAINT requests_id_len CHECK (length(id) = 64);
ALTER TABLE requests ADD CONSTRAINT requests_req_len CHECK (length(raw_request) < 16384);