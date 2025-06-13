CREATE TABLE projects (
                          id SERIAL PRIMARY KEY,
                          name VARCHAR(255) NOT NULL,
                          created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO projects (name) VALUES ('первая запись');

CREATE TABLE goods (
                       id SERIAL PRIMARY KEY,
                       project_id INT NOT NULL REFERENCES projects(id),
                       name VARCHAR(255) NOT NULL,
                       description TEXT,
                       priority INT NOT NULL,
                       removed BOOLEAN DEFAULT false,
                       created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX ON goods (project_id, priority);
CREATE INDEX ON goods (removed);