CREATE TABLE IF NOT EXISTS matches (
  id SERIAL PRIMARY KEY,
  team_a VARCHAR(100),
  team_b VARCHAR(100),
  match_date DATE,
  goals INT DEFAULT 0,
  yellow_cards INT DEFAULT 0,
  red_cards INT DEFAULT 0,
  extra_time VARCHAR(20) DEFAULT ''
);