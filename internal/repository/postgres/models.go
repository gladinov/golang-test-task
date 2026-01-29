package postgreSQL

var queryCreateNumberTable = `CREATE TABLE IF NOT EXISTS numbers (
    number BIGINT NOT NULL,
	created_at TIMESTAMPTZ DEFAULT now()                            
);`
