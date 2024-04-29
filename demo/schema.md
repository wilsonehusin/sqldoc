# `organizations`

|     NAME     |           TYPE           | NULLABLE | DEFAULT |
|--------------|--------------------------|----------|---------|
| `id`         | text                     | NOT NULL |         |
| `name`       | text                     | NOT NULL |         |
| `slug`       | text                     | NOT NULL |         |
| `created_at` | timestamp with time zone | NOT NULL | `now()` |
| `updated_at` | timestamp with time zone | NOT NULL | `now()` |

# `organizations_users`

|       NAME        |           TYPE           | NULLABLE | DEFAULT |
|-------------------|--------------------------|----------|---------|
| `organization_id` | text                     | NOT NULL |         |
| `user_id`         | text                     | NOT NULL |         |
| `permission`      | integer                  | NOT NULL | `0`     |
| `owner`           | boolean                  | NOT NULL | `false` |
| `created_at`      | timestamp with time zone | NOT NULL | `now()` |
| `updated_at`      | timestamp with time zone | NOT NULL | `now()` |

# `schema_migrations`

|   NAME    |  TYPE   | NULLABLE | DEFAULT |
|-----------|---------|----------|---------|
| `version` | bigint  | NOT NULL |         |
| `dirty`   | boolean | NOT NULL |         |

# `users`

|     NAME     |           TYPE           | NULLABLE | DEFAULT |
|--------------|--------------------------|----------|---------|
| `id`         | text                     | NOT NULL |         |
| `email`      | text                     | NOT NULL |         |
| `superuser`  | boolean                  | NOT NULL | `false` |
| `created_at` | timestamp with time zone | NOT NULL | `now()` |
| `updated_at` | timestamp with time zone | NOT NULL | `now()` |
