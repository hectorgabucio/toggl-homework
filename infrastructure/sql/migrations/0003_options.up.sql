CREATE TABLE IF NOT EXISTS option(
    id INTEGER PRIMARY KEY,
    body TEXT NOT NULL,
    correct BOOL NOT NULL,
    question_id INTEGER NOT NULL,
    FOREIGN KEY(question_id) REFERENCES question(id) ON DELETE CASCADE ON UPDATE CASCADE
)