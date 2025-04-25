WITH
    reference_date AS (
        SELECT
            DATE '2025-03-20' AS ref_date
    ),
    -- Get total quantity_change *after* the reference date
    future_usage AS (
        SELECT
            p.material_id,
            SUM(tl.quantity_change) AS future_quantity_change
        FROM
            transactions_log tl
            JOIN prices p ON p.price_id = tl.price_id
            JOIN reference_date r ON true
        WHERE
            tl.updated_at > r.ref_date
        GROUP BY
            p.material_id
    ),
    -- Quantity on the reference date = current quantity - future usage
    material_quantities AS (
        SELECT
            m.material_id,
            m.stock_id,
            m.customer_id,
            m.quantity - COALESCE(fu.future_quantity_change, 0) AS quantity_on_ref_date
        FROM
            materials m
            LEFT JOIN future_usage fu ON fu.material_id = m.material_id
    ),
    -- Filter transactions within the 6-week window before reference date
    filtered_transactions AS (
        SELECT
            tl.*,
            p.material_id
        FROM
            transactions_log tl
            JOIN prices p ON p.price_id = tl.price_id
            JOIN reference_date r ON true
        WHERE
            tl.updated_at BETWEEN r.ref_date - INTERVAL '6 weeks' AND r.ref_date
    ),
    -- Weekly usage per material
    weekly_usage AS (
        SELECT
            material_id,
            DATE_TRUNC ('week', updated_at) AS week_start,
            SUM(quantity_change) AS total_used
        FROM
            filtered_transactions
        WHERE
            -- Only actual materials used are in the query
            quantity_change < 0
            AND notes NOT ILIKE 'moved from a location'
        GROUP BY
            material_id,
            week_start
    ),
    -- Average usage per material over 6 weeks
    average_usage AS (
        SELECT
            material_id,
            ABS(AVG(total_used)) AS avg_weekly_usage
        FROM
            weekly_usage
        GROUP BY
            material_id
    )
    -- Final output:
SELECT
    c.name as customer_name,
    mq.stock_id,
    mq.quantity_on_ref_date,
    ROUND(au.avg_weekly_usage) AS avg_weekly_usage,
    CASE
        WHEN au.avg_weekly_usage = 0 THEN NULL
        ELSE ROUND(mq.quantity_on_ref_date / au.avg_weekly_usage, 2)
    END AS weeks_of_stock_remaining
FROM
    material_quantities mq
    LEFT JOIN average_usage au ON mq.material_id = au.material_id
    LEFT JOIN customers c ON c.customer_id = mq.customer_id
WHERE
    mq.quantity_on_ref_date > 0
    AND avg_weekly_usage > 0;
