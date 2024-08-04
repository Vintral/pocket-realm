package utilities

import (
	"fmt"
	"math/rand"

	"github.com/rs/zerolog/log"
)

type pickerItem struct {
	Weight uint
	Item   uint
}

type Picker struct {
	Choices []pickerItem
}

func (picker *Picker) Add(weight uint, itemId uint) {
	if weight == 0 {
		return
	}

	item := pickerItem{Weight: weight, Item: itemId}

	log.Debug().Any("item", item).Msg("Add")
	picker.Choices = append(picker.Choices, item)
}

func (picker *Picker) Choose() uint {
	max := uint(0)

	for _, i := range picker.Choices {
		max += i.Weight
	}

	if max <= 0 {
		log.Warn().Msg("Picker choices have no weight")
	}

	choice := uint(rand.Intn(int(max)))
	log.Debug().Msg("Rolled " + fmt.Sprint(choice) + " out of " + fmt.Sprint(max))

	for _, i := range picker.Choices {
		log.Debug().Msg("Choice " + fmt.Sprint(choice) + " ::: " + fmt.Sprint(i.Weight) + "(" + fmt.Sprint(i.Item) + ")")
		if choice < i.Weight {
			log.Debug().Msg("Picked! " + fmt.Sprint(i.Item))
			return i.Item
		}

		choice -= i.Weight
	}

	log.Warn().Msg("Picker failed")
	return 0
}
