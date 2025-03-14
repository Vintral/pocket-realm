package temple

import (
	"context"

	"github.com/Vintral/pocket-realm/models"
	"github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type getPantheonsResult struct {
	Type      string               `json:"type"`
	Pantheons []*models.Pantheon   `json:"pantheons"`
	Current   *models.UserDevotion `json:"current"`
}

func getUserDevotion(baseContext context.Context, db *gorm.DB, user *models.User, ch chan *models.UserDevotion) {
	ctx, span := utils.StartSpan(baseContext, "temple.getUserDevotion")
	defer span.End()

	var devotion *models.UserDevotion
	if err := db.WithContext(ctx).
		Select("user_devotions.id", "user_id", "round_id", "pantheon", "pantheons.category AS pantheon_name", "level").
		Joins("INNER JOIN pantheons ON pantheons.id = pantheon").
		Where("user_id = ? AND round_id = ?", user.ID, user.RoundID).
		Find(&devotion).Error; err != nil {
		log.Error().Err(err).Msg("Error retrieving user devotion")
	}

	devotion.Dump()

	ch <- devotion
}

func GetPantheons(baseContext context.Context) {
	ctx, span := utils.StartSpan(baseContext, "temple.GetPantheons")
	defer span.End()

	user := baseContext.Value(utils.KeyUser{}).(*models.User)
	db := baseContext.Value(utils.KeyDB{}).(*gorm.DB)

	log.Info().Uint("user", user.ID).Msg("GetPantheons")

	pc := make(chan []*models.Pantheon)
	go models.GetPantheons(ctx, pc)

	dc := make(chan *models.UserDevotion)
	go getUserDevotion(ctx, db, user, dc)

	pantheons, current := <-pc, <-dc

	user.Connection.WriteJSON(getPantheonsResult{
		Type:      "GET_PANTHEONS",
		Pantheons: pantheons,
		Current:   current,
	})
}
