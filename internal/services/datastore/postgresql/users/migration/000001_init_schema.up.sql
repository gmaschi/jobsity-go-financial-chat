CREATE TABLE "users" (
                           "username" varchar PRIMARY KEY,
                           "hashed_password" varchar NOT NULL,
                           "created_at" timestamptz NOT NULL DEFAULT (now()),
                           "updated_at" timestamptz NOT NULL DEFAULT (now())
);