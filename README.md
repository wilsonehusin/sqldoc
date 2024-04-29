# SQLDoc

SQLDoc is a markdown documentation for SQL tables. Inspired by Rails ActiveRecord `schema.rb` and [drwl/annotaterb](https://github.com/drwl/annotaterb).

![Demo of SQLDoc](demo/demo.gif)

## Why?

Projects often manages their schema roll out through migrations (e.g. [golang-migrate/migrate](https://github.com/golang-migrate/migrate)). As a project matures, it's quite common to encounter several `ALTER TABLE` commands, which makes it difficult to have a near-instant idea of what the schema looks like.

In Rails, `schema.rb` provides that insight, while extensions like [drwl/annotaterb](https://github.com/drwl/annotaterb) goes further to co-locate the schema documentation with the model definitions.

SQLDoc makes it easy to look at Markdown documentation for SQL tables, which I prefer over running `psql -c "\d+ table_name"`.
