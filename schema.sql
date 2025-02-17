CREATE TABLE IF NOT EXISTS games (
  id INTEGER PRIMARY KEY,
  game_id Text NOT NULL,
  game_info TEXT NOT NULL,
  created_dt_tm TEXT NOT NULL,
  created_user_id TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS gamestate (
  id INTEGER PRIMARY KEY,
  user_id TEXT NOT NULL,
  game_id INTEGER NOT NULL,
  game_state TEXT NOT NULL,
  created_dt_tm TEXT NOT NULL,
  FOREIGN KEY(game_id) REFERENCES games(id)
);
