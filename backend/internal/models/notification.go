package models

import (
    "database/sql"
    "fmt"
    "time"
)

type Notification struct {
    ID               int       `json:"id"`
    RecipientID      int       `json:"recipient_id"`
    NotificationType string    `json:"notification_type"`
    Message          string    `json:"message"`
    IsSent           bool      `json:"is_sent"`
    SentAt           time.Time `json:"sent_at,omitempty"`
    CreatedAt        time.Time `json:"created_at"`
}

func CreateNotification(db *sql.DB, recipientID int, notificationType string, message string) error {
    // Verify that the recipient_id exists in the farmers table
    var exists bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM farmers WHERE id = $1)", recipientID).Scan(&exists)
    if err != nil {
        return fmt.Errorf("CreateNotification: error verifying recipient existence: %w", err)
    }

    if !exists {
        return fmt.Errorf("CreateNotification: recipient_id %d does not exist in farmers table", recipientID)
    }

    stmt, err := db.Prepare(`
        INSERT INTO notifications (recipient_id, notification_type, message, is_sent, created_at)
        VALUES ($1, $2, $3, FALSE, NOW())
    `)
    if err != nil {
        return fmt.Errorf("CreateNotification: error preparing statement: %w", err)
    }
    defer stmt.Close()

    _, err = stmt.Exec(recipientID, notificationType, message)
    if err != nil {
        return fmt.Errorf("CreateNotification: error executing statement: %w", err)
    }

    return nil
}
