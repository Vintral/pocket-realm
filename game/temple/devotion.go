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

func RenounceDevotion(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "temple.RenounceDevotion")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	db := baseContext.Value(utils.KeyDB{}).(*gorm.DB)

	log.Info().Uint("user", user.ID).Msg("RenounceDevotion")

	success := false

	var devotion *models.UserDevotion
	if err := db.WithContext(ctx).Table("user_devotions").Where("user_id = ? AND round_id = ?", user.ID, user.RoundID).Find(&devotion).Error; err == nil {
		log.Info().Uint("panthon", devotion.Pantheon).Msg("Retrieved pantheon for user")

		if pantheon := models.GetPantheonById(ctx, devotion.Pantheon); pantheon != nil {
			pantheon.Dump()
		} else {
			log.Error().Uint("pantheon", devotion.Pantheon).Msg("Pantheon not found for id")
		}
	} else {
		log.Error().Err(err).Msg("Error retrieving user's pantheon")
	}

	user.Connection.WriteJSON(devotionResult{
		Type:    "RENOUNCE_DEVOTION",
		Success: success,
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

		var current models.UserDevotion
		if err := db.WithContext(ctx).Table("user_devotions").Where("user_id = ? AND round_id = ?", user.ID, user.RoundID).Find(&current).Error; err == nil {
			current.Dump()
			if current.Pantheon == pantheon.ID {
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

							user.AddBuff(ctx, buff)
							if err := user.Save(ctx); err == nil {
								success = true
							} else {
								log.Error().Err(err).Msg("Error updating user")

								if err := db.Model(&models.UserDevotion{}).Where("user_id = ? AND round_id = ?", user.ID, user.RoundID).Update("level", gorm.Expr("level - ?", 1)).Error; err != nil {
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

	user.Connection.WriteJSON(devotionResult{
		Type:    "RAISE_DEVOTION",
		Success: success,
	})
}
