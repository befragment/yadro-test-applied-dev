-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_nodes_log_id_id ON nodes (log_id, id);
CREATE INDEX idx_nodes_node_guid ON nodes (node_guid);
CREATE INDEX idx_ports_node_id_port_num ON ports (node_id, port_num);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_ports_node_id_port_num;
DROP INDEX IF EXISTS idx_nodes_node_guid;
DROP INDEX IF EXISTS idx_nodes_log_id_id;
-- +goose StatementEnd
