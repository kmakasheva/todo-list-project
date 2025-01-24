package db

const (
	CreateTableQuery = `
		CREATE TABLE IF NOT EXISTS scheduler(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date DATE NOT NULL ,
			title TEXT NOT NULL,
			comment TEXT NOT NULL,
			"repeat" TEXT NOT NULL CHECK (LENGTH("repeat") <= 128)
		);
		`
	CreateIndex = `
	CREATE INDEX scheduler_date 
	    ON scheduler(date);
`
)
