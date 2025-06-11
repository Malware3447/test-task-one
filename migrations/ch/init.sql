CREATE TABLE IF NOT EXISTS events (
                                      id Int32,
                                      project_id Int32,
                                      name String,
                                      description Nullable(String),
    priority Int32,
    removed Bool,
    event_time DateTime
    )
    ENGINE = MergeTree()
    ORDER BY (id, event_time)
