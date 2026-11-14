package cart

import (
	"errors"
	"time"
)

func quantityDiscount(quantity int) float64 {
	switch {
	case quantity > 1000:
		return 20

	case quantity > 500:
		return 15

	case quantity > 250:
		return 10

	case quantity > 100:
		return 5

	default:
		return 0
	}
}

func fileTypeIncrement(ft FileType) float64 {
	fileTypeIncrement := map[FileType]float64{
		FileTypePDF: 15,
		FileTypePSD: 35,
		FileTypeAI:  25,
	}

	increment, ok := fileTypeIncrement[ft]

	if !ok {
		return 0
	}

	return increment
}

func deliveryIncrementPct(delivery, checkout time.Time) (float64, error) {
	diff := delivery.Sub(checkout)

	if diff < 0 {
		return 0, errors.New("the delivery date is before the checkout date")
	}

	switch {
	// within 24h
	case diff.Hours() <= 24:
		return 30, nil

		// within 48h
	case diff.Hours() <= 48:
		return 20, nil

		// within 72h
	case diff.Hours() <= 72:
		return 10, nil

		// within 1 week
	default:
		return 0, nil
	}
}

func calculatePercentOf(price, pct float64) float64 {
	return price * pct / 100
}

func (p Product) UnitPrice(checkoutDate *time.Time) float64 {
	const basePrice = float64(1)

	qtyDiscountPct := quantityDiscount(p.Quantity)
	qtyDiscount := calculatePercentOf(basePrice, qtyDiscountPct)

	price := basePrice - qtyDiscount

	if checkoutDate != nil {
		dlvIncrement, err := deliveryIncrementPct(p.DeliveryDate, *checkoutDate)

		// a delivery date before checkout gets no delivery increment
		if err == nil {
			price += calculatePercentOf(price, dlvIncrement)
		}

	}

	ftIncrementPct := fileTypeIncrement(p.FileType)
	price += calculatePercentOf(price, ftIncrementPct)

	return price
}

func (c Cart) TotalPrice() float64 {
	totalPrice := float64(0)

	for _, item := range c.ItemList {
		totalPrice += item.UnitPrice(c.DateCheckout) * float64(item.Quantity)
	}

	return totalPrice
}
