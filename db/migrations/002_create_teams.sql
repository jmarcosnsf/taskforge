CREATE TABLE teams(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description VARCHAR(255),
  owner_id UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP not null DEFAULT NOW(),
  updated_at TIMESTAMP not null DEFAULT NOW()
);

---- create above / drop below ----

DROP TABLE teams;