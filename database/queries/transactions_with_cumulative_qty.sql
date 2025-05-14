WITH
    -- Compute starting quantity before the 'from' date for each stock
    starting_quantities AS (
        SELECT
            m.stock_id,
            COALESCE(SUM(tl.quantity_change), 0) AS starting_qty
        FROM
            transactions_log tl
            LEFT JOIN prices p ON p.price_id = tl.price_id
            LEFT JOIN materials m ON m.material_id = p.material_id
            LEFT JOIN customers c ON m.customer_id = c.customer_id
            -- WHERE
            -- ($1 = 0 OR m.customer_id = $1) AND
            -- ($2 = '' OR m.material_type = $2::material_type) AND
            -- ($5 = '' OR m.owner = $5::owner) AND
            -- ($4 = '' OR tl.updated_at < $4::timestamp)
        GROUP BY
            m.stock_id
    ),
    -- Select and enrich all relevant transactions
    ordered_transactions AS (
        SELECT
            m.stock_id,
            m.material_type,
            tl.quantity_change AS quantity,
            p.cost AS unit_cost,
            (tl.quantity_change * p.cost) AS cost,
            tl.updated_at,
            COALESCE(tl.serial_number_range, '') AS serial_number_range,
            tl.transaction_id,
            m.material_id
        FROM
            transactions_log tl
            LEFT JOIN prices p ON p.price_id = tl.price_id
            LEFT JOIN materials m ON m.material_id = p.material_id
            LEFT JOIN customers c ON m.customer_id = c.customer_id
            -- WHERE
            --     ($1 = 0 OR m.customer_id = $1) AND
            --     ($2 = '' OR m.material_type = $2::material_type) AND
            --     ($5 = '' OR m.owner = $5::owner)
    )
    -- Final result: transactions with cumulative quantities
SELECT
    ot.stock_id,
    ot.material_type,
    ot.quantity,
    ot.unit_cost,
    ot.cost,
    ot.updated_at,
    ot.serial_number_range,
    -- Calculate cumulative quantity (balance) per stock_id
    COALESCE(sq.starting_qty, 0) - SUM(ot.quantity) OVER (
        PARTITION BY
            ot.stock_id
        ORDER BY
            ot.updated_at,
            ot.transaction_id ROWS BETWEEN CURRENT ROW
            AND UNBOUNDED FOLLOWING
    ) + ot.quantity AS cumulative_quantity
FROM
    ordered_transactions ot
    LEFT JOIN starting_quantities sq ON sq.stock_id = ot.stock_id
    -- WHERE
    --     ($3 = '' OR ot.updated_at >= $3::timestamp) AND
    --     ($4 = '' OR ot.updated_at <= $4::timestamp)
ORDER BY
    ot.updated_at,
    ot.transaction_id ASC;