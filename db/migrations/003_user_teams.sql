CREATE TABLE team_members(
  user_id UUID NOT NULL REFERENCES users(id),
  team_id UUID NOT NULL REFERENCES teams(id),
  PRIMARY KEY (user_id, team_id)
);
---- create above / drop below ----

DROP TABLE team_members;