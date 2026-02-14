CREATE TABLE price_history (
    adj_close  DOUBLE PRECISION NOT NULL,
    close      DOUBLE PRECISION NOT NULL,
    date       DATE             NOT NULL,
    fetched_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    high       DOUBLE PRECISION NOT NULL,
    is_anomaly BOOLEAN          NOT NULL DEFAULT FALSE,
    low        DOUBLE PRECISION NOT NULL,
    open       DOUBLE PRECISION NOT NULL,
    source     TEXT             NOT NULL,
    symbol     TEXT             NOT NULL,
    volume     BIGINT           NOT NULL,

    PRIMARY KEY (symbol, date),

    CONSTRAINT price_positive       CHECK (open > 0 AND high > 0 AND low > 0 AND close > 0 AND adj_close > 0),
    CONSTRAINT high_gte_low         CHECK (high >= low),
    CONSTRAINT volume_non_negative  CHECK (volume >= 0)
);

CREATE TABLE fx_rate (
    date       DATE             NOT NULL,
    fetched_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    pair       TEXT             NOT NULL,
    rate       DOUBLE PRECISION NOT NULL,
    source     TEXT             NOT NULL DEFAULT 'frankfurter',

    PRIMARY KEY (pair, date),

    CONSTRAINT rate_positive CHECK (rate > 0)
);
