package models

import (
    "database/sql"
    "fmt"
    "time"
)

type Notification struct {
    ID               int       `json:"id"`
    RecipientType    string    `json:"recipient_type"`
    RecipientID      int       `json:"recipient_id"`
    NotificationType string    `json:"notification_type"`
    Message          string    `json:"message"`
    IsSent           bool      `json:"is_sent"`
    SentAt           time.Time `json:"sent_at,omitempty"`
    CreatedAt        time.Time `json:"created_at"`
}

func CreateNotification(db *sql.DB, recipientType string, recipientID int, notificationType string, message string) error {
    if recipientType != "admin" && recipientType != "farmer" && recipientType != "buyer" {
        return fmt.Errorf("invalid recipient_type: %s", recipientType)
    }

    stmt, err := db.Prepare(`
        INSERT INTO notifications (recipient_type, recipient_id, notification_type, message, is_sent, created_at)
        VALUES ($1, $2, $3, $4, FALSE, NOW())
    `)
    if err != nil {
        return fmt.Errorf("CreateNotification: error preparing statement: %w", err)
    }
    defer stmt.Close()

    _, err = stmt.Exec(recipientType, recipientID, notificationType, message)
    if err != nil {
        return fmt.Errorf("CreateNotification: error executing statement: %w", err)
    }

    return nil
}
