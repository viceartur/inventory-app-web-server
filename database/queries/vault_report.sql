WITH
    inner_vault AS (
        SELECT
            l.name AS location_name,
            m.stock_id,
            SUM(m.quantity) AS quantity
        FROM
            materials m
            LEFT JOIN locations l ON l.location_id = m.location_id
            LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
        WHERE
            w.name = 'Inner Vault'
        GROUP BY
            l.name,
            m.stock_id
    ),
    outer_vault AS (
        SELECT
            l.name AS location_name,
            m.stock_id,
            SUM(m.quantity) AS quantity
        FROM
            materials m
            LEFT JOIN locations l ON l.location_id = m.location_id
            LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
        WHERE
            w.name = 'Outer Vault'
        GROUP BY
            l.name,
            m.stock_id
    )
SELECT
    COALESCE(iv.location_name, '-') AS inner_location,
    COALESCE(ov.location_name, '-') AS outer_location,
    m.stock_id,
    COALESCE(iv.quantity, 0) AS inner_vault_quantity,
    COALESCE(ov.quantity, 0) AS outer_vault_quantity,
    SUM(m.quantity) AS total_quantity
FROM
    materials m
    LEFT JOIN locations l ON l.location_id = m.location_id
    LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
    LEFT JOIN inner_vault iv ON iv.stock_id = m.stock_id
    LEFT JOIN outer_vault ov ON ov.stock_id = m.stock_id
WHERE
    w.name IN ('Inner Vault', 'Outer Vault')
GROUP BY
    iv.location_name,
    ov.location_name,
    m.stock_id,
    iv.quantity,
    ov.quantity
ORDER BY
    iv.location_name,
    ov.location_name,
    m.stock_id;