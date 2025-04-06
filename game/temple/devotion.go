package temple

import (
	"context"
	"encoding/json"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type raiseDevotionPayload struct {
	Type     string    `json:"type"`
	Pantheon uuid.UUID `json:"pantheon"`
}

type devotionResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

func getOrCreateUserDevotion(baseContext context.Context, db *gorm.DB, user *models.User, pantheon *models.Pantheon) *models.UserDevotion {
	ctx, span := utils.StartSpan(baseContext, "temple.getOrCreateUserDevotion")
	defer span.End()

	var current *models.UserDevotion
	if err := db.WithContext(ctx).Table("user_devotions").Where("user_id = ? AND round_id = ?", user.ID, user.RoundID).Find(&current).Error; err == nil {
		if current.ID == 0 {
			current = &models.UserDevotion{
				UserID:   user.ID,
				RoundID:  uint(user.RoundID),
				Pantheon: pantheon.ID,
				Level:    0,
			}
			if err := db.WithContext(ctx).Save(current).Error; err != nil {
				log.Error().Err(err).Int("user", int(user.ID)).Int("round", user.RoundID).Int("pantheon", int(pantheon.ID)).Msg("Error creating user devotion")
				return nil
			}
		}

		return current
	} else {
		log.Error().Err(err).Int("user", int(user.ID)).Int("round", user.RoundID).Msg("Error getting user devotion")
	}

	return nil
}

func RenounceDevotion(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "temple.RenounceDevotion")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)

	user.Connection.WriteJSON(devotionResult{
		Type:    "RENOUNCE_DEVOTION",
		Success: user.RenounceDevotion(ctx),
	})
}

func RaiseDevotion(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "temple.RaiseDevotion")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	db := baseContext.Value(utils.KeyDB{}).(*gorm.DB)

	log.Info().Uint("user", user.ID).Msg("RaiseDevotion")

	success := false

	var payload raiseDevotionPayload
	if err := json.Unmarshal(baseContext.Value(utils.KeyPayload{}).([]byte), &payload); err == nil {
		pantheon := models.GetPantheonByGuid(ctx, payload.Pantheon)
		pantheon.Dump()

		if current := getOrCreateUserDevotion(ctx, db, user, pantheon); current != nil {
			current.Dump()
			if (current.ID == 0) || (current.Pantheon == pantheon.ID) {
				var next *models.Devotion
				for _, level := range pantheon.Devotions {
					if level.Level == current.Level+1 {
						next = level
					}
				}

				if next != nil {
					var buff *models.Buff
					if err := db.WithContext(ctx).Where("id = ?", next.Buff).Find(&buff).Error; err == nil {
						if err := db.Model(&models.UserDevotion{}).Where("user_id = ? AND round_id = ?", user.ID, user.RoundID).Update("level", gorm.Expr("level + ?", 1)).Error; err != nil {
							log.Error().Err(err).Msg("Error updating UserDevotion")
						} else {

							buff.AddToUser(ctx, user)
							if err := user.Save(ctx); err == nil {
								success = true
							} else {
								log.Error().Err(err).Msg("Error updating user")

								if err := db.WithContext(ctx).Model(&models.UserDevotion{}).Where("user_id = ? AND round_id = ?", user.ID, user.RoundID).Update("level", gorm.Expr("level - ?", 1)).Error; err != nil {
									log.Panic().Err(err).Msg("Error rolling back UserDevotion")
								}
							}
						}
					} else {
						log.Error().Err(err).Msg("Error retrieving buff")
					}
				} else {
					log.Error().Msg("Next level of devotion not found")
				}
			} else {
				log.Error().Msg("Raising devotion of wrong pantheon")
			}
		} else {
			log.Error().Err(err).Msg("Error getting UserDevotion")
		}
	} else {
		log.Error().Err(err).Msg("Error retrieving raiseDevotionPayload")
	}

	if user.Connection != nil {
		user.Connection.WriteJSON(devotionResult{
			Type:    "RAISE_DEVOTION",
			Success: success,
		})
	}
}
