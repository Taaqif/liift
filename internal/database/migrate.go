package database

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"liift/internal/models"

	"gorm.io/gorm"
)

// Migrate runs database migrations for all models
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.Equipment{},
		&models.MuscleGroup{},
		&models.Image{},
		&models.Exercise{},
		&models.ExerciseFeature{},
		&models.User{},
		&models.Workout{},
		&models.WorkoutExercise{},
		&models.WorkoutSet{},
		&models.WorkoutSetFeature{},
		&models.WorkoutPlan{},
		&models.WorkoutSession{},
		&models.WorkoutSessionExercise{},
		&models.WorkoutSessionSet{},
		&models.WorkoutSessionSetValue{},
		&models.WorkoutPlanProgress{},
		&models.AISettings{},
		&models.ChatSession{},
		&models.ChatMessage{},
	)
	if err != nil {
		return err
	}

	if err := renameOrderToSortOrder(db); err != nil {
		return err
	}
	if err := workoutSessionExerciseAdHocColumns(db); err != nil {
		return err
	}
	if err := backfillChatSessionSlugs(db); err != nil {
		return err
	}
	if err := migrateAISettingsGlobal(db); err != nil {
		return err
	}
	if err := backfillUserRoles(db); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func renameOrderToSortOrder(db *gorm.DB) error {
	for _, table := range []interface{}{"workout_session_exercises", "workout_session_sets"} {
		if err := db.Migrator().RenameColumn(table, "order", "sort_order"); err != nil {
			log.Printf("Rename order->sort_order in %v (may already be done): %v", table, err)
		} else {
			log.Printf("Renamed column order -> sort_order in %v", table)
		}
	}
	return nil
}

func workoutSessionExerciseAdHocColumns(db *gorm.DB) error {
	m := &models.WorkoutSessionExercise{}
	if !db.Migrator().HasColumn(m, "ExerciseID") {
		if err := db.Migrator().AddColumn(m, "ExerciseID"); err != nil {
			return err
		}
		log.Printf("Added exercise_id to workout_session_exercises")
	}
	if !db.Migrator().HasColumn(m, "RestTimer") {
		if err := db.Migrator().AddColumn(m, "RestTimer"); err != nil {
			return err
		}
		log.Printf("Added rest_timer to workout_session_exercises")
	}
	return nil
}

func backfillChatSessionSlugs(db *gorm.DB) error {
	var sessions []models.ChatSession
	if err := db.Where("slug = ''").Find(&sessions).Error; err != nil {
		return err
	}
	for _, s := range sessions {
		b := make([]byte, 6)
		rand.Read(b)
		slug := hex.EncodeToString(b)
		if err := db.Model(&s).Update("slug", slug).Error; err != nil {
			return err
		}
	}
	if len(sessions) > 0 {
		log.Printf("Backfilled slugs for %d chat sessions", len(sessions))
	}
	// Ensure unique index exists
	if !db.Migrator().HasIndex(&models.ChatSession{}, "idx_chat_sessions_slug") {
		if err := db.Exec("CREATE UNIQUE INDEX idx_chat_sessions_slug ON chat_sessions(slug)").Error; err != nil {
			log.Printf("Could not create unique index on chat_sessions.slug (may already exist): %v", err)
		}
	}
	return nil
}

// migrateAISettingsGlobal drops the user_id column from ai_settings if it still exists.
func migrateAISettingsGlobal(db *gorm.DB) error {
	m := &models.AISettings{}
	if db.Migrator().HasColumn(m, "UserID") {
		// Keep only the first record (lowest ID) and delete the rest before dropping the column
		if err := db.Exec("DELETE FROM ai_settings WHERE id NOT IN (SELECT MIN(id) FROM ai_settings)").Error; err != nil {
			log.Printf("Could not deduplicate ai_settings (may already be clean): %v", err)
		}
		if err := db.Migrator().DropColumn(m, "UserID"); err != nil {
			log.Printf("Could not drop user_id from ai_settings (may already be done): %v", err)
		} else {
			log.Println("Dropped user_id column from ai_settings — settings are now global")
		}
	}
	return nil
}

// backfillUserRoles sets role='admin' for the first user if no admins exist.
func backfillUserRoles(db *gorm.DB) error {
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount > 0 {
		return nil
	}
	// Make the first registered user (lowest ID) the admin
	if err := db.Exec("UPDATE users SET role = 'admin' WHERE id = (SELECT MIN(id) FROM users)").Error; err != nil {
		log.Printf("Could not backfill admin role: %v", err)
		return err
	}
	log.Println("Backfilled admin role for first user")
	return nil
}

// MigrateModels runs migrations for specific models
func MigrateModels(db *gorm.DB, models ...any) error {
	if len(models) == 0 {
		return nil
	}

	log.Printf("Running migrations for %d model(s)...", len(models))
	if err := db.AutoMigrate(models...); err != nil {
		return err
	}
	log.Println("Migrations completed successfully")
	return nil
}
