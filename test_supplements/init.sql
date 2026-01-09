CREATE TABLE pair (
    id          BIGSERIAL           NOT NULL PRIMARY KEY,
    chain_id    VARCHAR             NOT NULL, CHECK(chain_id <> ''),
    contract    VARCHAR             NOT NULL, CHECK(contract <> ''),
    asset0      VARCHAR             NOT NULL, CHECK(asset0 <> ''),
    asset1      VARCHAR             NOT NULL, CHECK(asset1 <> ''),
    lp          VARCHAR             NOT NULL, CHECK(lp <> ''),
    created_at  DOUBLE PRECISION    NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    meta        JSON
);

CREATE UNIQUE INDEX pair_chain_id_contract_key ON pair (chain_id, contract);
CREATE UNIQUE INDEX pair_chain_id_asset0_asset1_key ON pair (chain_id, asset0, asset1);
CREATE UNIQUE INDEX pair_chain_id_lp_key ON pair (chain_id, lp);


CREATE TABLE tokens (
    id         BIGSERIAL                    PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    chain_id   TEXT                         NOT NULL,
    address    TEXT                         NOT NULL,
    protocol   TEXT,
    symbol     TEXT,
    name       TEXT,
    decimals   SMALLINT,
    icon       TEXT,
    verified   BOOLEAN                      DEFAULT FALSE NOT NULL
);

CREATE INDEX idx_tokens_address ON tokens (address);
CREATE INDEX idx_tokens_chain_id ON tokens (chain_id);
CREATE UNIQUE INDEX idx_tokens_chain_id_address_key ON tokens (chain_id, address);
CREATE INDEX idx_tokens_deleted_at ON tokens (deleted_at);


CREATE TABLE pair_stats_30m (
    id                   BIGSERIAL                                                NOT NULL PRIMARY KEY,
    year_utc             SMALLINT                                                 NOT NULL,
    month_utc            SMALLINT                                                 NOT NULL,
    day_utc              SMALLINT                                                 NOT NULL,
    hour_utc             SMALLINT                                                 NOT NULL,
    minute_utc           SMALLINT                                                 NOT NULL,
    pair_id              BIGINT                                                   NOT NULL,
    chain_id             VARCHAR                                                  NOT NULL,
    volume0              NUMERIC                                                  NOT NULL,
    volume1              NUMERIC                                                  NOT NULL,
    volume0_in_price     NUMERIC                                                  NOT NULL,
    volume1_in_price     NUMERIC                                                  NOT NULL,
    last_swap_price      NUMERIC                                                  NOT NULL,
    liquidity0           NUMERIC                                                  NOT NULL,
    liquidity1           NUMERIC                                                  NOT NULL,
    liquidity0_in_price  NUMERIC                                                  NOT NULL,
    liquidity1_in_price  NUMERIC                                                  NOT NULL,
    commission0          NUMERIC                                                  NOT NULL,
    commission1          NUMERIC                                                  NOT NULL,
    commission0_in_price NUMERIC                                                  NOT NULL,
    commission1_in_price NUMERIC                                                  NOT NULL,
    price_token          VARCHAR                                                  NOT NULL,
    tx_cnt               BIGINT                                                   NOT NULL,
    provider_cnt         BIGINT                                                   NOT NULL,
    timestamp            DOUBLE PRECISION                                         NOT NULL,
    created_at           DOUBLE PRECISION DEFAULT date_part('epoch'::text, NOW()) NOT NULL,
    modified_at          DOUBLE PRECISION DEFAULT date_part('epoch'::text, NOW()) NOT NULL
);

CREATE INDEX pair_stats_30m_chain_id_timestamp_uidx ON pair_stats_30m (chain_id, timestamp);
CREATE INDEX pair_stats_30m_pair_id_timestamp_uidx ON pair_stats_30m (pair_id, timestamp);

CREATE TABLE price (
    id              BIGSERIAL NOT NULL PRIMARY KEY,
    height          BIGINT NOT NULL,
    chain_id        TEXT NOT NULL,
    token_id        BIGINT NOT NULL,
    price           NUMERIC NOT NULL,
    price_token_id  BIGINT NOT NULL,
    route_id        BIGINT NOT NULL,
    tx_id           BIGINT NULL,
    created_at      DOUBLE PRECISION DEFAULT date_part('epoch'::text, NOW()) NOT NULL,
    modified_at     DOUBLE PRECISION DEFAULT date_part('epoch'::text, NOW()) NOT NULL
);

-- simplified
CREATE TABLE parsed_tx (
    id          BIGSERIAL NOT NULL PRIMARY KEY,
    chain_id    CHARACTER varying NOT NULL,
    height      BIGINT NOT NULL,
    timestamp   DOUBLE PRECISION NOT NULL,
    created_at  DOUBLE PRECISION NOT NULL DEFAULT date_part('epoch'::text, now()),
    meta        JSON NULL
);

-- simplified
CREATE TABLE pair_stats_recent (
    id                   BIGSERIAL NOT NULL PRIMARY KEY,
    pair_id              BIGSERIAL NOT NULL,
    chain_id             CHARACTER VARYING NOT NULL,
    volume0_in_price     NUMERIC NOT NULL,
    volume1_in_price     NUMERIC NOT NULL,
    commission0_in_price NUMERIC NOT NULL,
    commission1_in_price NUMERIC NOT NULL,
    height               BIGINT NOT NULL,
    timestamp            DOUBLE PRECISION NOT NULL,
    created_at           DOUBLE PRECISION NOT NULL DEFAULT date_part('epoch'::text, now()),
    modified_at          DOUBLE PRECISION NOT NULL DEFAULT date_part('epoch'::text, now())
);