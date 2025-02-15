package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type MercenaryMarket struct {
	BaseModel

	GUID    uuid.UUID `gorm:"uniqueIndex,size:36" json:"-"`
	Round   uint      `json:"-"`
	Unit    uuid.UUID `json:"unit"`
	Cost    uint      `json:"cost"`
	Expires time.Time `json:"expires"`
}

func (mercenary *MercenaryMarket) BeforeCreate(tx *gorm.DB) (err error) {
	mercenary.GUID = uuid.New()
	return
}

// func GetMercenaryAuction(baseContext context.Context, user *User) *MercenaryMarket {
// 	ctx, span := Tracer.Start(baseContext, "get-mercenary-auction")
// 	defer span.End()

// 	var mercenary *MercenaryMarket
// 	db.WithContext(ctx).Where("round = ?", user.Round.ID).Scan(&mercenary)

// 	return mercenary
// }

func GetMercenaryMarket(baseContext context.Context, user *User) {
	ctx, span := Tracer.Start(baseContext, "get-mercenary-market")
	defer span.End()

	log.Info().Uint("round-id", user.Round.ID).Msg("models.GetMercenaryMarket")

	var mercenary *MercenaryMarket
	db.WithContext(ctx).Where("round = ?", user.RoundID).Find(&mercenary)

	span.SetAttributes(
		attribute.Int("round", int(user.Round.ID)),
		attribute.Int("user", int(user.ID)),
		attribute.String("mercenary", mercenary.Unit.String()),
	)

	response := struct {
		Type      string          `json:"type"`
		Mercenary MercenaryMarket `json:"mercenary"`
	}{
		Type:      "MERCENARY_MARKET",
		Mercenary: *mercenary,
	}

	user.SendMessage(response)

	log.Info().Any("result", response).Msg("Sent response")
}

func BuyMercenary(baseContext context.Context, user *User, quantity int) bool {
	ctx, span := Tracer.Start(baseContext, "buy-mercenary")
	defer span.End()

	round := user.Round

	var mercenary *MercenaryMarket
	db.WithContext(ctx).Where("round = ?", user.Round.ID).Find(&mercenary)

	unit := round.GetUnitByGuid(mercenary.Unit.String())
	log.Info().Uint("round", round.ID).Uint("user", user.ID).Int("quantity", quantity).Str("unit", unit.Name).Int("gold", int(user.RoundData.Gold)).Int("cost", quantity*int(mercenary.Cost)).Msg("Buy Mercenary")

	user.RoundData.Gold -= float64(quantity) * float64(mercenary.Cost)
	if user.RoundData.Gold < 0 {
		return false
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user.AddUnit(ctx, unit, quantity)

		if err := tx.WithContext(ctx).Save(&user).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Error().AnErr("error", err).Msg(("Error buying mercenary"))
		return false
	}

	return true
}
