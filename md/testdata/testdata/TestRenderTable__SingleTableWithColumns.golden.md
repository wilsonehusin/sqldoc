# `sqldoc_single_table_with_columns`

|       NAME        |            TYPE             | NULLABLE |   DEFAULT    |
|-------------------|-----------------------------|----------|--------------|
| `uuid_field`      | uuid                        | NOT NULL |              |
| `varchar_field`   | character varying           | NOT NULL |              |
| `int_field`       | integer                     | NOT NULL |              |
| `text_field`      | text                        |          |              |
| `boolean_field`   | boolean                     | NOT NULL | `false`      |
| `json_field`      | json                        | NOT NULL | `'{}'::json` |
| `timestamp_field` | timestamp without time zone | NOT NULL | `now()`      |
