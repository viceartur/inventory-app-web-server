WITH
    reference_date AS (
        SELECT
            DATE '2025-06-04' AS ref_date
    ),
    -- Total future usage grouped by stock_id
    future_usage AS (
        SELECT
            m.stock_id,
            SUM(tl.quantity_change) AS future_quantity_change
        FROM
            transactions_log tl
            JOIN prices p ON p.price_id = tl.price_id
            JOIN materials m ON m.material_id = p.material_id
            JOIN reference_date r ON true
        WHERE
            tl.updated_at > r.ref_date
        GROUP BY
            m.stock_id
    ),
    -- Quantity on reference date grouped by stock_id
    material_quantities AS (
        SELECT
            m.stock_id,
            m.material_type,
            m.customer_id,
            SUM(m.quantity) - COALESCE(fu.future_quantity_change, 0) AS quantity_on_ref_date
        FROM
            materials m
            LEFT JOIN future_usage fu ON fu.stock_id = m.stock_id
        GROUP BY
            m.stock_id,
            m.material_type,
            m.customer_id,
            fu.future_quantity_change
    ),
    -- Filtered transactions (past 6 weeks)
    filtered_transactions AS (
        SELECT
            tl.*,
            m.stock_id
        FROM
            transactions_log tl
            JOIN prices p ON p.price_id = tl.price_id
            JOIN materials m ON m.material_id = p.material_id
            JOIN reference_date r ON true
        WHERE
            tl.updated_at BETWEEN r.ref_date - INTERVAL '6 weeks' AND r.ref_date
    ),
    -- Weekly usage per stock
    weekly_usage AS (
        SELECT
            stock_id,
            DATE_TRUNC ('week', updated_at) AS week_start,
            SUM(quantity_change) AS total_used
        FROM
            filtered_transactions
        WHERE
            quantity_change < 0
            AND notes NOT ILIKE 'moved from a location'
        GROUP BY
            stock_id,
            week_start
    ),
    -- Average usage per stock
    average_usage AS (
        SELECT
            stock_id,
            ABS(AVG(total_used)) AS avg_weekly_usage
        FROM
            weekly_usage
        GROUP BY
            stock_id
    )
    -- Final output
SELECT
    c.name AS customer_name,
    mq.stock_id,
    mq.material_type,
    mq.quantity_on_ref_date,
    ROUND(au.avg_weekly_usage) AS avg_weekly_usage,
    CASE
        WHEN au.avg_weekly_usage = 0 THEN NULL
        ELSE ROUND(mq.quantity_on_ref_date / au.avg_weekly_usage, 2)
    END AS weeks_of_stock_remaining
FROM
    material_quantities mq
    LEFT JOIN average_usage au ON mq.stock_id = au.stock_id
    LEFT JOIN customers c ON c.customer_id = mq.customer_id
WHERE
    mq.quantity_on_ref_date > 0
    AND au.avg_weekly_usage > 0
    AND c.customer_id = 423
    AND mq.stock_id = '0129-001-WAVE'
    AND mq.material_type = 'CARDS (PVC)'
ORDER BY
    c.name,
    mq.material_type,
    mq.stock_id
