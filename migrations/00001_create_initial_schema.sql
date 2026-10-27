-- +goose Up

-- users: people who use this DealStalker instance.
-- Login data (email, password hash) is added in a later migration.
CREATE TABLE users (
    id         UUID        PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- products: what a user wants to buy, independent of where it is listed.
CREATE TABLE products (
    id         UUID        PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- listings: a product on one specific source (e.g. its Geizhals page).
-- (source, external_id) is unique, so each product page is fetched only once.
CREATE TABLE listings (
    id              UUID        PRIMARY KEY,
    product_id      UUID        NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    source          TEXT        NOT NULL CHECK (source IN ('geizhals', 'idealo')),
    external_id     TEXT        NOT NULL,
    url             TEXT        NOT NULL,
    last_checked_at TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source, external_id)
);

CREATE INDEX listings_product_id_idx ON listings (product_id);

-- price_snapshots: one successfully fetched price of a listing.
CREATE TABLE price_snapshots (
    id             UUID        PRIMARY KEY,
    listing_id     UUID        NOT NULL REFERENCES listings (id) ON DELETE CASCADE,
    price_cents    BIGINT      NOT NULL CHECK (price_cents >= 0),
    price_currency TEXT        NOT NULL CHECK (char_length(price_currency) = 3),
    fetched_at     TIMESTAMPTZ NOT NULL
);

CREATE INDEX price_snapshots_listing_id_fetched_at_idx ON price_snapshots (listing_id, fetched_at);

-- watches: a user observes a product with a target price.
-- Deleting a product is refused while anyone still watches it.
CREATE TABLE watches (
    id                    UUID        PRIMARY KEY,
    user_id               UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id            UUID        NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    target_price_cents    BIGINT      NOT NULL CHECK (target_price_cents > 0),
    target_price_currency TEXT        NOT NULL CHECK (char_length(target_price_currency) = 3),
    active                BOOLEAN     NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, product_id)
);

CREATE INDEX watches_product_id_idx ON watches (product_id);

-- notification_channels: where a user wants to be notified.
-- Single table for all channel types; the named CHECK ensures that
-- each row has exactly the columns its type needs.
CREATE TABLE notification_channels (
    id                  UUID        PRIMARY KEY,
    user_id             UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type                TEXT        NOT NULL CHECK (type IN ('ntfy', 'discord')),
    enabled             BOOLEAN     NOT NULL DEFAULT true,
    ntfy_server_url     TEXT,
    ntfy_topic          TEXT,
    discord_webhook_url TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT notification_channels_type_config_check CHECK (
        (
            type = 'ntfy'
            AND ntfy_server_url IS NOT NULL
            AND ntfy_topic IS NOT NULL
            AND discord_webhook_url IS NULL
        )
        OR
        (
            type = 'discord'
            AND discord_webhook_url IS NOT NULL
            AND ntfy_server_url IS NULL
            AND ntfy_topic IS NULL
        )
    )
);

CREATE INDEX notification_channels_user_id_idx ON notification_channels (user_id);

-- notifications: a sent alert. Kept when its channel is deleted
-- (channel_id becomes NULL), so the history that prevents repeated
-- alerts is not lost.
CREATE TABLE notifications (
    id             UUID        PRIMARY KEY,
    watch_id       UUID        NOT NULL REFERENCES watches (id) ON DELETE CASCADE,
    channel_id     UUID        REFERENCES notification_channels (id) ON DELETE SET NULL,
    price_cents    BIGINT      NOT NULL CHECK (price_cents >= 0),
    price_currency TEXT        NOT NULL CHECK (char_length(price_currency) = 3),
    sent_at        TIMESTAMPTZ NOT NULL
);

CREATE INDEX notifications_watch_id_sent_at_idx ON notifications (watch_id, sent_at);

-- +goose Down

-- Drop in reverse order of creation: a table can only be dropped
-- once no other table references it anymore.
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_channels;
DROP TABLE IF EXISTS watches;
DROP TABLE IF EXISTS price_snapshots;
DROP TABLE IF EXISTS listings;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;