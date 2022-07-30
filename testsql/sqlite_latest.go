package testsql

const SQLITE_LATEST = `
CREATE TABLE IF NOT EXISTS "crud_object" (
  "tid" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "user_id" INTEGER NOT NULL DEFAULT 0,
  "type" TEXT NOT NULL DEFAULT '',
  "level" INTEGER NOT NULL DEFAULT 0,
  "title" TEXT NOT NULL,
  "image" TEXT,
  "description" TEXT,
  "data" TEXT NOT NULL DEFAULT '{}',
  "int_value" INT4 NOT NULL DEFAULT 0,
  "int_ptr" INT4,
  "int_array" TEXT NOT NULL DEFAULT '[]',
  "int64_value" INTEGER NOT NULL DEFAULT 0,
  "int64_ptr" INTEGER,
  "int64_array" TEXT NOT NULL DEFAULT '[]',
  "float64_value" DOUBLE NOT NULL DEFAULT 0,
  "float64_ptr" DOUBLE,
  "float64_array" TEXT NOT NULL DEFAULT '[]',
  "string_value" TEXT NOT NULL DEFAULT '',
  "string_ptr" TEXT,
  "string_array" TEXT NOT NULL DEFAULT '[]',
  "map_value" TEXT NOT NULL DEFAULT '{}',
  "map_array" TEXT NOT NULL DEFAULT '[]',
  "time_value" DATE NOT NULL,
  "update_time" DATE NOT NULL,
  "create_time" DATE NOT NULL,
  "status" INT4 NOT NULL
);
`

const SQLITE_DROP = `
DROP TABLE IF EXISTS "crud_object";
`

const SQLITE_CLEAR = `
DELETE FROM "crud_object";
`
