package materials

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"
)

func FetchMaterialTypes(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
		SELECT enumlabel FROM pg_enum pe
		LEFT JOIN pg_type pt ON pt.oid = pe.enumtypid
		WHERE pt.typname = 'material_type';
	`)
	if err != nil {
		return []string{}, err
	}

	var materialTypes []string
	for rows.Next() {
		var materialType string
		if err := rows.Scan(&materialType); err != nil {
			return nil, fmt.Errorf("Error scanning row: %w", err)
		}
		materialTypes = append(materialTypes, materialType)
	}

	return materialTypes, nil
}

func FetchMaterialUsageReasons(db *sql.DB) ([]MaterialUsageReason, error) {
	rows, err := db.Query(`
		SELECT
			reason_id, reason_type, description, code
		FROM
			material_usage_reasons
	`)
	if err != nil {
		return []MaterialUsageReason{}, err
	}

	var reasons []MaterialUsageReason
	for rows.Next() {
		var reason MaterialUsageReason

		err := rows.Scan(
			&reason.ReasonID,
			&reason.ReasonType,
			&reason.Description,
			&reason.Code,
		)

		if err != nil {
			return nil, fmt.Errorf("Error scanning row: %w", err)
		}
		reasons = append(reasons, reason)
	}

	return reasons, nil
}

func SendMaterial(material IncomingMaterial, db *sql.DB) error {
	qty := material.Quantity
	minQty := material.MinQty
	maxQty := material.MaxQty

	_, err := db.Query(`
				INSERT INTO incoming_materials
					(program_id, stock_id, cost, quantity,
					max_required_quantity, min_required_quantity,
					description, material_status, type, owner, user_id)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		material.ProgramID, material.StockID, material.Cost,
		qty, maxQty, minQty,
		material.Description, material.MaterialStatus, material.MaterialType,
		material.Owner,
		material.UserID,
	)

	if err != nil {
		return err
	}
	return nil
}

func GetIncomingMaterials(db *sql.DB, materialId int) ([]IncomingMaterial, error) {
	rows, err := db.Query(`
		SELECT shipping_id, c.program_name, c.program_id, stock_id, cost, quantity,
		min_required_quantity, max_required_quantity, description, im.material_status,
		type, owner, u.user_id, u.username
		FROM incoming_materials im
		LEFT JOIN customer_programs c ON c.program_id = im.program_id
		LEFT JOIN users u ON u.user_id = im.user_id
		WHERE $1 = 0 OR im.shipping_id = $1
		ORDER BY im.shipping_id;
		`, materialId)
	if err != nil {
		return nil, fmt.Errorf("Error querying incoming materials: %w", err)
	}
	defer rows.Close()

	var materials []IncomingMaterial
	for rows.Next() {
		var material IncomingMaterial
		if err := rows.Scan(
			&material.ShippingID,
			&material.ProgramName,
			&material.ProgramID,
			&material.StockID,
			&material.Cost,
			&material.Quantity,
			&material.MinQty,
			&material.MaxQty,
			&material.Description,
			&material.MaterialStatus,
			&material.MaterialType,
			&material.Owner,
			&material.UserID,
			&material.Username,
		); err != nil {
			return nil, fmt.Errorf("Error scanning row: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func GetIncomingWarehouseMaterialsCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM incoming_materials im
		WHERE im.type NOT IN ('CARDS (PVC)', 'CARDS (METAL)', 'CHIPS');
	`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetIncomingVaultMaterialsCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM incoming_materials im
		WHERE im.type IN ('CARDS (PVC)', 'CARDS (METAL)', 'CHIPS');
	`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetMaterials(db *sql.DB, opts *MaterialFilter) ([]Material, error) {
	rows, err := db.Query(`
		SELECT material_id,
		COALESCE(w.name,'None') AS "warehouse_name",
		c.program_name AS "program_name",
		c.program_id,
		c.is_active AS "is_active_program",
		COALESCE(l.location_id, 0) AS "location_id",
		COALESCE(l.name, 'None') AS "location_name",
		stock_id,
		quantity,
		min_required_quantity,
		max_required_quantity,
		m.description,
		COALESCE(notes,'None') AS "notes",
		m.material_status,
		material_type, owner,
		COALESCE(is_primary, false),
		COALESCE(serial_number_range, '')
		FROM materials m
		LEFT JOIN customer_programs c ON c.program_id = m.program_id
		LEFT JOIN locations l ON l.location_id = m.location_id
		LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
		WHERE
			($1 = 0 OR m.material_id = $1) AND
			($2 = '' OR m.stock_id ILIKE '%' || $2 || '%') AND
			($3 = '' OR c.program_name ILIKE '%' || $3 || '%') AND
			($4 = '' OR m.description ILIKE '%' || $4 || '%') AND
			($5 = '' OR l.name ILIKE '%' || $5 || '%')
		ORDER BY m.is_primary DESC, c.program_name, m.stock_id ASC;
		`,
		opts.MaterialId,
		opts.StockId,
		opts.ProgramName,
		opts.Description,
		opts.LocationName,
	)
	if err != nil {
		return nil, fmt.Errorf("Error querying incoming materials: %w", err)
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var material Material
		if err := rows.Scan(
			&material.MaterialID,
			&material.WarehouseName,
			&material.ProgramName,
			&material.ProgramID,
			&material.IsActiveProgram,
			&material.LocationID,
			&material.LocationName,
			&material.StockID,
			&material.Quantity,
			&material.MinQty,
			&material.MaxQty,
			&material.Description,
			&material.Notes,
			&material.MaterialStatus,
			&material.MaterialType,
			&material.Owner,
			&material.IsPrimary,
			&material.SerialNumberRange,
		); err != nil {
			return nil, fmt.Errorf("Error scanning row: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

// Find a Material using the exact search
func GetMaterialsByStockID(db *sql.DB, opts *MaterialFilter) ([]Material, error) {
	if opts.StockId == "" {
		return nil, nil
	}

	rows, err := db.Query(`
		SELECT
			material_id,
			COALESCE(w.name,'None') AS "warehouse_name",
			c.program_name, c.program_id,
			c.is_active AS "is_active_program",
			COALESCE(l.location_id, 0) AS "location_id",
			COALESCE(l.name, 'None') AS "location_name",
			stock_id,
			quantity,
			min_required_quantity,
			max_required_quantity,
			m.description, COALESCE(notes,'None') AS "notes",
			m.material_status,
			material_type, owner,
			COALESCE(is_primary, false),
			COALESCE(serial_number_range, '')
		FROM materials m
		LEFT JOIN customer_programs c ON c.program_id = m.program_id
		LEFT JOIN locations l ON l.location_id = m.location_id
		LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
		WHERE m.stock_id = $1;
		`, opts.StockId,
	)
	if err != nil {
		return nil, fmt.Errorf("Error querying incoming materials: %w", err)
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var material Material
		if err := rows.Scan(
			&material.MaterialID,
			&material.WarehouseName,
			&material.ProgramName,
			&material.ProgramID,
			&material.IsActiveProgram,
			&material.LocationID,
			&material.LocationName,
			&material.StockID,
			&material.Quantity,
			&material.MinQty,
			&material.MaxQty,
			&material.Description,
			&material.Notes,
			&material.MaterialStatus,
			&material.MaterialType,
			&material.Owner,
			&material.IsPrimary,
			&material.SerialNumberRange,
		); err != nil {
			return nil, fmt.Errorf("Error scanning row: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

// The method creates/updates a Material, its Prices, adds a Transaction Log, and deletes the Material from Incoming.
// Method's Context: Material Creation. The Transaction Rollback is executed once an error occurs.
func CreateMaterial(ctx context.Context, db *sql.DB, material MaterialJSON) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Commit() // commit only if the method is done

	var incomingMaterial IncomingMaterial
	err = tx.QueryRow(`
		SELECT program_id, stock_id, quantity, cost, min_required_quantity,
		max_required_quantity, description, material_status, type, owner
		FROM incoming_materials
		WHERE shipping_id = $1`, material.MaterialID).
		Scan(
			&incomingMaterial.ProgramID,
			&incomingMaterial.StockID,
			&incomingMaterial.Quantity,
			&incomingMaterial.Cost,
			&incomingMaterial.MinQty,
			&incomingMaterial.MaxQty,
			&incomingMaterial.Description,
			&incomingMaterial.MaterialStatus,
			&incomingMaterial.MaterialType,
			&incomingMaterial.Owner,
		)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Update material in the current location if location exists
	var materialId int
	rows, err := tx.Query(`
					UPDATE materials
					SET quantity = (quantity + $1),
						notes = $2
					WHERE stock_id = $3
						AND location_id = $4
						AND owner = $5
					RETURNING material_id;
					`, material.Qty, material.Notes, incomingMaterial.StockID, material.LocationID, incomingMaterial.Owner,
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	for rows.Next() {
		err := rows.Scan(&materialId)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	rows.Close()

	// Upsert Prices
	var priceId int
	qty := material.Qty

	if materialId != 0 {
		priceInfo := Price{materialId: materialId, qty: qty, cost: incomingMaterial.Cost}
		priceId, err = upsertPrice(tx, priceInfo)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	} else {
		// If there is no a Material in the chosen Location:
		// 1. Check for a NULL location and if it exists then assign the new location and qty
		rows, err := tx.Query(`
			SELECT material_id FROM materials
			WHERE location_id is NULL
				AND stock_id = $1
				AND owner = $2;
		`, incomingMaterial.StockID, incomingMaterial.Owner)
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		for rows.Next() {
			err := rows.Scan(&materialId)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}

		if materialId != 0 {
			rows, err := tx.Query(`
					UPDATE materials
					SET quantity = $1,
						notes = $2,
						location_id = $3
					WHERE material_id = $4;
					`, material.Qty, material.Notes, material.LocationID, materialId,
			)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
			rows.Close()
		} else {
			// 2. If there is no a NULL Location, then add the material to the new location
			// Insert a new Material
			err = tx.QueryRow(`
						INSERT INTO materials
						(
							stock_id,
							location_id,
							program_id,
							material_type,
							description,
							notes,
							quantity,
							updated_at,
							min_required_quantity,
							max_required_quantity,
							material_status,
							owner,
							is_primary,
							serial_number_range
						)
						VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING material_id;`,
				incomingMaterial.StockID,
				material.LocationID,
				incomingMaterial.ProgramID,
				incomingMaterial.MaterialType,
				incomingMaterial.Description,
				material.Notes,
				material.Qty,
				time.Now(),
				incomingMaterial.MinQty,
				incomingMaterial.MaxQty,
				incomingMaterial.MaterialStatus,
				incomingMaterial.Owner,
				false,
				material.SerialNumberRange,
			).Scan(&materialId)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}

		// Upsert Prices
		priceInfo := Price{materialId: materialId, qty: qty, cost: incomingMaterial.Cost}
		priceId, err = upsertPrice(tx, priceInfo)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// Delete/Update the Material from Incoming
	shippingId := material.MaterialID
	if (incomingMaterial.Quantity == qty) || (incomingMaterial.Quantity < qty) {
		err = deleteIncomingMaterial(tx, shippingId)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	} else {
		_, err := tx.Exec(`
					UPDATE incoming_materials
					SET quantity = (quantity + $2)
					WHERE shipping_id = $1
					`, shippingId, -qty,
		)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// Add a Transaction
	trxInfo := &TransactionInfo{
		priceId:           priceId,
		qty:               qty,
		notes:             material.Notes,
		updatedAt:         time.Now(),
		serialNumberRange: material.SerialNumberRange,
	}
	err = addTranscation(trxInfo, tx)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return materialId, nil
}

func UpdateIncomingMaterial(db *sql.DB, material IncomingMaterial) error {
	_, err := db.Exec(`
		UPDATE incoming_materials
		SET program_id = $2,
			stock_id = $3,
			cost = $4,
			quantity = $5,
			max_required_quantity = $6,
			min_required_quantity = $7,
			description = $8,
			material_status = $9,
			type = $10,
			owner = $11
		WHERE shipping_id = $1;
	`,
		material.ShippingID,
		material.ProgramID,
		material.StockID,
		material.Cost,
		material.Quantity,
		material.MaxQty,
		material.MinQty,
		material.Description,
		material.MaterialStatus,
		material.MaterialType,
		material.Owner,
	)

	if err != nil {
		return err
	}
	return nil
}

func DeleteIncomingMaterial(ctx context.Context, db *sql.DB, shippingId int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Commit()

	_, err = tx.Exec(`
			DELETE FROM incoming_materials WHERE shipping_id = $1;`, shippingId)
	if err != nil {
		return err
	}

	return nil
}

// The method changes the Material quantity at the current and new Location, its Prices, and adds Transaction Logs.
// Method's Context: Material Moving. The Transaction Rollback is executed once an error occurs.
func MoveMaterial(ctx context.Context, db *sql.DB, material MaterialJSON) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Commit() // commit only if the method is done

	materialId := material.MaterialID
	currMaterial, err := getMaterialById(materialId, tx)
	if err != nil {
		return err
	}

	newLocationId := material.LocationID
	quantity := material.Qty
	actualQuantity := currMaterial.Quantity
	currMaterialId := currMaterial.MaterialID
	currentLocationId := currMaterial.LocationID
	stockId := currMaterial.StockID
	owner := currMaterial.Owner
	currNotes := currMaterial.Notes

	// 1. Update the Material in the current Location

	// Check whether remaining quantity exists
	if actualQuantity < quantity {
		return errors.New(
			`The moving quantity (` +
				strconv.Itoa(quantity) + `) is more than the actual one (` +
				strconv.Itoa(actualQuantity) + `)`)
	} else if actualQuantity > quantity {
		// Update material in the current location
		err = tx.QueryRow(`
			UPDATE materials
			SET quantity = (quantity - $1),
				notes = $2
			WHERE material_id = $3 AND location_id = $4
			RETURNING material_id, stock_id, location_id, program_id, material_type,
					description, notes, quantity, updated_at, material_status,
					min_required_quantity, max_required_quantity, owner,
					is_primary, serial_number_range;
			`, quantity, currNotes, currMaterialId, currentLocationId,
		).Scan(
			&currMaterial.MaterialID,
			&currMaterial.StockID,
			&currMaterial.LocationID,
			&currMaterial.ProgramID,
			&currMaterial.MaterialType,
			&currMaterial.Description,
			&currMaterial.Notes,
			&currMaterial.Quantity,
			&currMaterial.UpdatedAt,
			&currMaterial.MaterialStatus,
			&currMaterial.MinQty,
			&currMaterial.MaxQty,
			&currMaterial.Owner,
			&currMaterial.IsPrimary,
			&currMaterial.SerialNumberRange,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else if actualQuantity == quantity {
		defer func() error {
			_, err = tx.Exec(`
				UPDATE materials
				SET location_id = NULL,
					quantity = 0
				WHERE material_id = $1`,
				currMaterialId,
			)
			if err != nil {
				tx.Rollback()
				return err
			}
			return nil
		}()
	}

	// 1.1. Update Prices for the current Location
	autoTicket := "Auto-Ticket: " + strconv.Itoa(rand.IntN(899999)+100000)

	// Remove Prices for the current Material ID
	priceToRemove := PriceToRemove{
		materialId: currMaterialId,
		qty:        quantity,
		notes:      "Moved TO a Location",
		jobTicket:  autoTicket,
	}
	removedPrices, err := removePricesFIFO(tx, priceToRemove)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Update a Material in the new Location
	var newMaterialId int

	// Find an existing Material in the Location
	rows, err := tx.Query(`
			UPDATE materials
			SET quantity = (quantity + $1)
			WHERE
				stock_id = $2 AND
				location_id = $3 AND
				owner = $4
			RETURNING material_id;
				`, quantity, stockId, newLocationId, owner,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	for rows.Next() {
		err := rows.Scan(&newMaterialId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// If there is no a Material in the new Location, then create it
	if newMaterialId == 0 {
		err = tx.QueryRow(`
				INSERT INTO materials
					(
						stock_id, location_id,
						program_id, material_type, description, notes,
						quantity, updated_at,
						material_status, min_required_quantity, max_required_quantity,
						owner, is_primary, serial_number_range
					)
					VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
					RETURNING material_id;`,
			stockId, newLocationId,
			currMaterial.ProgramID, currMaterial.MaterialType, currMaterial.Description,
			currNotes, quantity, time.Now(), currMaterial.MaterialStatus,
			currMaterial.MinQty, currMaterial.MaxQty, currMaterial.Owner,
			currMaterial.IsPrimary, currMaterial.SerialNumberRange).
			Scan(&newMaterialId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// 2.2. Update Prices for the new Location and Material ID

	for i := range removedPrices {
		qty := removedPrices[i].qty
		cost := removedPrices[i].cost
		priceInfo := Price{materialId: newMaterialId, qty: qty, cost: cost}

		priceId, err := upsertPrice(tx, priceInfo)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = addTranscation(&TransactionInfo{
			priceId:           priceId,
			qty:               qty,
			notes:             "Moved FROM a Location",
			jobTicket:         autoTicket,
			updatedAt:         time.Now(),
			serialNumberRange: material.SerialNumberRange,
		}, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

// The method removes a specific Material quantity, its Prices, adds a Transaction Log.
// Method's Context: Material Removing. The Transaction Rollback is executed once an error occurs.
func RemoveMaterial(ctx context.Context, db *sql.DB, material MaterialJSON) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Commit() // commit only if the method is done

	materialId := material.MaterialID
	currMaterial, err := getMaterialById(materialId, tx)
	if err != nil {
		tx.Rollback()
		return errors.New("Unable to get the current material info: " + err.Error())
	}

	quantity := material.Qty
	actualQuantity := currMaterial.Quantity
	jobTicket := material.JobTicket

	if actualQuantity < quantity {
		return errors.New(`The removing quantity (` + strconv.Itoa(quantity) + `) is more than the actual one (` + strconv.Itoa(actualQuantity) + `)`)
	} else if actualQuantity == quantity {
		_, err = tx.Exec(`
			UPDATE materials
			SET location_id = NULL,
				quantity = 0
			WHERE material_id = $1
		`, materialId)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// Update the material quantity
		_, err = tx.Exec(`
				UPDATE materials
				SET quantity = (quantity - $1)
				WHERE material_id = $2;
		`, quantity, materialId,
		)
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	priceToRemove := PriceToRemove{
		materialId:        materialId,
		qty:               quantity,
		notes:             "Removed FROM a Location",
		jobTicket:         jobTicket,
		serialNumberRange: material.SerialNumberRange,
		reasonId:          material.ReasonID,
	}
	_, err = removePricesFIFO(tx, priceToRemove)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func UpdateMaterial(ctx context.Context, db *sql.DB, material MaterialJSON) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Commit()

	materialId := material.MaterialID

	// Checks whether IsPrimary field provided since
	// this field cannot be updated at the same time with other fields
	if material.IsPrimary != nil {
		_, err := tx.Exec(`
		UPDATE materials
		SET is_primary = $2
		WHERE material_id = $1
		`, materialId, *material.IsPrimary)
		if err != nil {
			return err
		}
	} else {
		// Update a Material quantity
		qty := material.Qty

		var updatedMaterialId int

		rows, err := tx.Query(`
			UPDATE materials
			SET quantity = quantity + $2
			WHERE material_id = $1
			RETURNING material_id;
		`, materialId, qty)
		if err != nil {
			return err
		}

		for rows.Next() {
			err := rows.Scan(&updatedMaterialId)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		if updatedMaterialId == 0 {
			return errors.New("No Material Found")
		}

		// Update a Price
		priceInfo := Price{materialId: materialId, qty: qty, cost: 0}
		priceId, err := upsertPrice(tx, priceInfo)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Add a Transaction
		trxInfo := &TransactionInfo{
			priceId:   priceId,
			qty:       qty,
			jobTicket: material.JobTicket,
			updatedAt: time.Now(),
		}
		err = addTranscation(trxInfo, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func RequestMaterials(ctx context.Context, db *sql.DB, requestedMaterials RequestedMaterialsJSON) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Commit()

	query := `
	INSERT INTO requested_materials
		(stock_id, description, quantity_requested, quantity_used, status, notes, updated_at, requested_at, user_id) VALUES `
	args := []any{}
	placeholderCount := 1

	for i, m := range requestedMaterials.Materials {
		qty := m.Qty
		if qty == 0 {
			continue
		}
		if i > 0 {
			query += ", "
		}

		query += fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			placeholderCount, placeholderCount+1,
			placeholderCount+2, placeholderCount+3,
			placeholderCount+4, placeholderCount+5,
			placeholderCount+6, placeholderCount+7,
			placeholderCount+8,
		)

		args = append(
			args,
			m.StockID,
			m.Description,
			m.Qty,
			0,
			"pending",
			"Needs to be delivered",
			time.Now(),
			time.Now(),
			requestedMaterials.UserID,
		)

		placeholderCount += 9
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetRequestedMaterials(db *sql.DB, filterOpts MaterialFilter) ([]Material, error) {
	rows, err := db.Query(`
		SELECT
			request_id,
			COALESCE(u.username, '') AS "username",
			stock_id,
			description,
			quantity_requested,
			quantity_used,
			status,
			notes,
			updated_at,
			requested_at
		FROM requested_materials rm
		LEFT JOIN users u ON u.user_id = rm.user_id
		WHERE ($1 = 0 OR rm.request_id = $1) AND
		      ($2 = '' OR rm.stock_id ILIKE '%' || $2 || '%') AND
			  ($3 = '' OR rm.status::TEXT = $3) AND
			  ($4 = '' OR rm.requested_at::TEXT >= $4) AND
			  ($5 = '' OR rm.requested_at::TEXT <= $5)
		ORDER BY rm.requested_at;
			  `,
		filterOpts.RequestId,
		filterOpts.StockId,
		filterOpts.Status,
		filterOpts.RequestedFrom,
		filterOpts.RequestedTo,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var material Material

		if err := rows.Scan(
			&material.RequestID,
			&material.Username,
			&material.StockID,
			&material.Description,
			&material.QtyRequested,
			&material.QtyUsed,
			&material.Status,
			&material.Notes,
			&material.UpdatedAt,
			&material.RequestedAt,
		); err != nil {
			return nil, fmt.Errorf("Error scanning row: %w", err)
		}
		materials = append(materials, material)
	}
	return materials, nil
}

func GetRequestedMaterialsCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM requested_materials
		WHERE status = 'pending';
	`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func UpdateRequestedMaterial(db *sql.DB, material MaterialJSON) error {
	requestId := material.MaterialID
	quantity := material.Qty

	_, err := db.Exec(`
		UPDATE requested_materials
		SET quantity_used = quantity_used + $2,
			status = $3,
			notes = $4,
			updated_at = $5
		WHERE request_id = $1;
	`, requestId, quantity, material.Status, material.Notes, time.Now())

	if err != nil {
		return err
	}
	return nil
}

func GetMaterialDescription(db *sql.DB, stockId string) (string, error) {
	var description string
	err := db.QueryRow(`
		SELECT description
		FROM materials
		WHERE LOWER(stock_id) = LOWER($1)
		LIMIT 1;
	`,
		stockId,
	).Scan(&description)
	if err != nil {
		return "", err
	}
	return description, nil
}

func GetMaterialTransactions(db *sql.DB, jobTicket string) ([]Transaction, error) {
	rows, err := db.Query(`
		SELECT
			m.material_id,
			m.stock_id,
			l.location_id,
			l.name,
			w.warehouse_id,
			w.name,
			tl.quantity_change,
			tl.job_ticket
		FROM materials m
		LEFT JOIN locations l ON l.location_id = m.location_id
		LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
		LEFT JOIN prices p ON p.material_id = m.material_id
		LEFT JOIN transactions_log tl ON tl.price_id = p.price_id
		WHERE
			tl.quantity_change < 0
			AND tl.job_ticket = $1
	`, jobTicket)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.MaterialID,
			&t.StockID,
			&t.LocationID,
			&t.LocationName,
			&t.WarehouseID,
			&t.WarehouseName,
			&t.Quantity,
			&t.JobTicket,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// UpdateMaterialsStatus updates the material_status field for all materials with the given stockId.
func UpdateMaterialsStatus(db *sql.DB, stockId string, status string) error {
	res, err := db.Exec(`
		UPDATE materials
		SET material_status = $2
		WHERE stock_id = $1;
	`,
		stockId,
		status,
	)
	if err != nil {
		return fmt.Errorf("Failed to update material status for stock_id %s: %w", stockId, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("No materials updated for stock_id %s", stockId)
	}
	return nil
}

func GetMaterialsGroupedByStockID(db *sql.DB) ([]Material, error) {
	rows, err := db.Query(`
		SELECT
			cp.program_name,
			m.stock_id,
			m.description,
			m.material_status,
			SUM(m.quantity) AS total_quantity
		FROM materials m
		LEFT JOIN customer_programs cp ON cp.program_id = m.program_id
		WHERE m.location_id IS NOT NULL
		GROUP BY
			cp.program_name,
			m.stock_id,
			m.description,
			m.material_status
		ORDER BY stock_id;
		`)
	if err != nil {
		return nil, fmt.Errorf("Error querying grouped materials: %w", err)
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var material Material
		if err := rows.Scan(
			&material.ProgramName,
			&material.StockID,
			&material.Description,
			&material.MaterialStatus,
			&material.Quantity,
		); err != nil {
			return nil, fmt.Errorf("Error scanning row: %w", err)
		}
		materials = append(materials, material)
	}

	return materials, nil
}
