CREATE TABLE IF NOT EXISTS matches (
  id SERIAL PRIMARY KEY,
  team_a VARCHAR(100),
  team_b VARCHAR(100),
  score_a INTEGER,
  score_b INTEGER,
  match_date DATE
);