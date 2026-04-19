package migrations

import (
    "database/sql"
    "fmt"
)

func Run(db *sql.DB) error {
    fmt.Println("Running database migrations...")
    
    var tableCount int
    err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'wishes'").Scan(&tableCount)
    if err != nil {
        return fmt.Errorf("failed to check tables: %v", err)
    }
    
    if tableCount > 0 {
        fmt.Println("Tables already exist, skipping migrations")
        return nil
    }
    
    fmt.Println("Applying migration 001: create wishes table")
    _, err = db.Exec(`
        CREATE TABLE wishes (
            id BIGSERIAL PRIMARY KEY,
            text TEXT NOT NULL,
            author VARCHAR(100) DEFAULT 'Anonymous',
            priority INTEGER DEFAULT 1,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        );
        CREATE INDEX idx_wishes_deleted_at ON wishes(deleted_at);
        CREATE INDEX idx_wishes_priority ON wishes(priority);
    `)
    if err != nil {
        return fmt.Errorf("migration 001 failed: %v", err)
    }
    

    fmt.Println("Applying migration 002: add priority check constraint")
    _, err = db.Exec(`
        ALTER TABLE wishes ADD CONSTRAINT wishes_priority_check 
            CHECK (priority >= 1 AND priority <= 5);
    `)
    if err != nil {
        return fmt.Errorf("migration 002 failed: %v", err)
    }
    
    fmt.Println("All migrations completed successfully!")
    return nil
}