package migrations

import (
	"log"
	"gorm.io/gorm"
)

func FixLegacyData(db *gorm.DB) error {
    if err := db.Exec(`
        UPDATE messages
        SET contact_id = NULL
        WHERE contact_id = 0;
    `).Error; err != nil {
        return err
    }

    if err := db.Exec(`
        UPDATE messages
        SET contact_id = NULL
        WHERE contact_id IS NOT NULL
        AND contact_id NOT IN (SELECT id FROM contacts);
    `).Error; err != nil {
        return err
    }

    log.Println("--Data cleanup completed--")
    return nil
}