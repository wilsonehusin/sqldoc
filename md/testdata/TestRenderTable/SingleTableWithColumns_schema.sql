DROP TABLE IF EXISTS sqldoc_single_table_with_columns;
CREATE TABLE sqldoc_single_table_with_columns (
    uuid_field UUID PRIMARY KEY,
    varchar_field VARCHAR(255) NOT NULL,
    int_field INT NOT NULL,
    text_field TEXT,
    boolean_field BOOLEAN NOT NULL DEFAULT FALSE,
    json_field JSON NOT NULL DEFAULT '{}',

    timestamp_field TIMESTAMP NOT NULL DEFAULT NOW()
);
