package materials

import (
	"database/sql"
	"time"
)

// Internal Methods that helps to implement the basic Business Logic.

func getMaterialPrices(tx *sql.Tx, materialId int) (map[int]Price, error) {
	rows, err := tx.Query(`
		SELECT * FROM prices
		WHERE material_id = $1
		AND quantity > 0
		ORDER BY price_id ASC;
	`, materialId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make(map[int]Price)

	for rows.Next() {
		var price PriceDB
		err := rows.Scan(&price.PriceID, &price.MaterialID, &price.Qty, &price.Cost)
		if err != nil {
			return nil, err
		}

		prices[price.PriceID] = Price{qty: price.Qty, cost: price.Cost}
	}
	return prices, nil
}

// The internal method changes the Prices quantity and returns updated Prices
func updatePriceQty(tx *sql.Tx, priceInfo *Price) (float64, error) {
	var updatedCost float64
	err := tx.QueryRow(`
					UPDATE prices
					SET quantity = (quantity + $2)
					WHERE price_id = $1
					RETURNING cost;
					`, priceInfo.priceId, priceInfo.qty,
	).Scan(&updatedCost)
	if err != nil {
		return 0, err
	}

	return updatedCost, nil
}

func upsertPrice(tx *sql.Tx, priceInfo Price) (int, error) {
	var priceId int
	rows, err := tx.Query(`
					INSERT INTO prices (material_id, quantity, cost)
						VALUES ($1, $2, $3)
					ON CONFLICT (material_id, cost)
						DO UPDATE
							SET quantity = (prices.quantity + EXCLUDED.quantity)
					RETURNING price_id;
					`, priceInfo.materialId, priceInfo.qty, priceInfo.cost,
	)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		err := rows.Scan(&priceId)
		if err != nil {
			return 0, err
		}
	}
	return priceId, nil
}

func addTranscation(trx *TransactionInfo, tx *sql.Tx) error {
	// Check whether the Reason ID provided
	var reasonId any
	if trx.reasonId > 0 {
		reasonId = trx.reasonId
	} else {
		reasonId = nil
	}

	_, err := tx.Exec(`
		INSERT INTO transactions_log (
			price_id, quantity_change, notes, job_ticket, updated_at,
			serial_number_range, reason_id
		)
		VALUES($1, $2, $3, $4, $5, $6, $7);
	`, trx.priceId,
		trx.qty,
		trx.notes,
		trx.jobTicket,
		trx.updatedAt,
		trx.serialNumberRange,
		reasonId,
	)
	if err != nil {
		return err
	}
	return nil
}

func removePricesFIFO(tx *sql.Tx, priceToRemove PriceToRemove) ([]Price, error) {
	materialId := priceToRemove.materialId
	qty := priceToRemove.qty
	notes := priceToRemove.notes
	jobTicket := priceToRemove.jobTicket

	materialPrices, err := getMaterialPrices(tx, materialId)
	if err != nil {
		return nil, err
	}

	removedPrices := []Price{}

	remainingQty := qty
	for priceId, priceInfo := range materialPrices {
		if remainingQty <= priceInfo.qty {
			qtyToRemove := remainingQty
			priceInfo := &Price{
				priceId: priceId,
				qty:     -qtyToRemove,
			}

			cost, err := updatePriceQty(tx, priceInfo)
			if err != nil {
				return nil, err
			}

			err = addTranscation(&TransactionInfo{
				priceId:           priceId,
				qty:               -qtyToRemove,
				notes:             notes,
				jobTicket:         jobTicket,
				updatedAt:         time.Now(),
				serialNumberRange: priceToRemove.serialNumberRange,
				reasonId:          priceToRemove.reasonId,
			}, tx)
			if err != nil {
				return nil, err
			}

			remainingQty = 0
			removedPrices = append(removedPrices, Price{qty: qtyToRemove, cost: cost})
			break
		} else {
			qtyToRemove := priceInfo.qty
			priceInfo := &Price{
				priceId: priceId,
				qty:     -qtyToRemove,
			}

			cost, err := updatePriceQty(tx, priceInfo)
			if err != nil {
				return nil, err
			}

			err = addTranscation(&TransactionInfo{
				priceId:           priceId,
				qty:               -qtyToRemove,
				notes:             notes,
				jobTicket:         jobTicket,
				updatedAt:         time.Now(),
				serialNumberRange: priceToRemove.serialNumberRange,
				reasonId:          priceToRemove.reasonId,
			}, tx)
			if err != nil {
				return nil, err
			}

			remainingQty -= qtyToRemove
			removedPrices = append(removedPrices, Price{qty: qtyToRemove, cost: cost})
		}
	}
	return removedPrices, nil
}

// Delete an Incoming Material once it's accepted
func deleteIncomingMaterial(tx *sql.Tx, shippingId int) error {
	if _, err := tx.Exec(`
			DELETE FROM incoming_materials WHERE shipping_id = $1;`,
		shippingId); err != nil {
		return err
	}

	return nil
}

func getMaterialById(materialId int, tx *sql.Tx) (MaterialDB, error) {
	var currMaterial MaterialDB
	err := tx.QueryRow(`SELECT
							material_id, stock_id, location_id,
							customer_id, material_type, description, notes,
							quantity, updated_at,
							is_active, min_required_quantity, max_required_quantity,
							owner, is_primary, COALESCE(serial_number_range, '')
						FROM materials
						WHERE material_id = $1`,
		materialId,
	).Scan(
		&currMaterial.MaterialID,
		&currMaterial.StockID,
		&currMaterial.LocationID,
		&currMaterial.CustomerID,
		&currMaterial.MaterialType,
		&currMaterial.Description,
		&currMaterial.Notes,
		&currMaterial.Quantity,
		&currMaterial.UpdatedAt,
		&currMaterial.IsActive,
		&currMaterial.MinQty,
		&currMaterial.MaxQty,
		&currMaterial.Owner,
		&currMaterial.IsPrimary,
		&currMaterial.SerialNumberRange,
	)
	if err != nil {
		return MaterialDB{}, err
	}

	return currMaterial, nil
}
