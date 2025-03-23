package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQL struct {
	Address  string
	UserName string
	Password string
	DBName   string
	dbConn   *sql.DB
}

func (db *PostgreSQL) Init() error {
	var err error
	ps := fmt.Sprintf("host=%s user=%s password=%s database=%s sslmode=disable",
		db.Address, db.UserName, db.Password, db.DBName)

	db.dbConn, err = sql.Open("pgx", ps)
	if err != nil {
		return err
	}

	db.dbConn.Exec("DROP TABLE orders;")
	db.dbConn.Exec("DROP TABLE users;")
	db.dbConn.Exec("DROP TABLE order_spend;")
	db.dbConn.Exec("DROP TYPE status_enum;")

	// _, err = db.dbConn.Exec(`CREATE EXTENSION pgcrypto;`)
	// if err != nil {
	// 	return fmt.Errorf("error while creating extension pgcrypto: %w", err)
	// }

	_, err = db.dbConn.Exec(`CREATE TABLE users (id BIGSERIAL PRIMARY KEY,
												login VARCHAR(1000) NOT NULL UNIQUE,
												password VARCHAR(1000) NOT NULL,
	                                            sum FLOAT8,
												CONSTRAINT valid_sum CHECK (sum >= 0), 
												with_drawn FLOAT8);`)
	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TYPE status_enum AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');`)
	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TABLE orders (id BIGINT PRIMARY KEY,
														status Status_Enum,
														UploadedAt TIMESTAMP,
														accrual FLOAT8,
														user_id BIGINT REFERENCES Users (id) ON DELETE CASCADE);`)

	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TABLE order_spend (id BIGINT PRIMARY KEY,
													ProcessedAt TIMESTAMP,
													sum FLOAT8, 
													user_id BIGINT REFERENCES Users (id) ON DELETE CASCADE);`)
	if err != nil {
		return err
	}

	return nil
}
