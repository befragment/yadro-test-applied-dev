-- +goose Up
-- +goose StatementBegin
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,

    status TEXT NOT NULL DEFAULT 'parsed',

    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),

    nodes_count INT NOT NULL DEFAULT 0,
    ports_count INT NOT NULL DEFAULT 0
);

CREATE TABLE nodes (
    id BIGSERIAL PRIMARY KEY,

    log_id BIGINT NOT NULL REFERENCES logs(id) ON DELETE CASCADE,

    node_guid TEXT NOT NULL,

    system_image_guid TEXT,
    port_guid TEXT,

    description TEXT NOT NULL,

    -- 1 host, 2 switch
    node_type SMALLINT NOT NULL,

    num_ports INT NOT NULL,

    class_version INT,
    base_version INT,

    -- switch-specific fields
    linear_fdb_cap INT,
    multicast_fdb_cap INT,
    life_time_value INT
);

CREATE TABLE ports (
    id BIGSERIAL PRIMARY KEY,

    node_id BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,

    port_guid TEXT NOT NULL,

    port_num INT NOT NULL,

    lid INT,

    port_state INT,
    port_phy_state INT,

    link_width_active INT,
    link_speed_active INT,

    link_round_trip_latency INT
);

CREATE TABLE nodes_info (
    id BIGSERIAL PRIMARY KEY,

    node_id BIGINT NOT NULL UNIQUE REFERENCES nodes(id) ON DELETE CASCADE,

    serial_number TEXT,
    part_number TEXT,
    revision TEXT,
    product_name TEXT,

    -- SHARP info
    endianness INT,
    enable_endianness_per_job INT,
    reproducibility_disable INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS ports;
DROP TABLE IF EXISTS nodes_info;

-- +goose StatementEnd
