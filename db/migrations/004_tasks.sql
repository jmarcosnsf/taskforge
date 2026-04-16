CREATE TABLE tasks(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR(255) NOT NULL,
  description VARCHAR(255),
  status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'done')),
  team_id UUID NOT NULL REFERENCES teams(id),
  user_id UUID REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

---- create above / drop below ----

DROP TABLE tasks;
