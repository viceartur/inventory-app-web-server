WITH
    inner_vault AS (
        SELECT
            m.stock_id,
            SUM(m.quantity) AS quantity
        FROM
            materials m
            LEFT JOIN locations l ON l.location_id = m.location_id
            LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
        WHERE
            w.name = 'Inner Vault'
        GROUP BY
            m.stock_id
    ),
    outer_vault AS (
        SELECT
            m.stock_id,
            SUM(m.quantity) AS quantity
        FROM
            materials m
            LEFT JOIN locations l ON l.location_id = m.location_id
            LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
        WHERE
            w.name = 'Outer Vault'
        GROUP BY
            m.stock_id
    )
SELECT
    c.name AS customer_name,
    m.material_type,
    m.stock_id,
    COALESCE(iv.quantity, 0) AS inner_vault_quantity,
    COALESCE(ov.quantity, 0) AS outer_vault_quantity,
    SUM(m.quantity) AS total_quantity
FROM
    materials m
    LEFT JOIN customers c ON c.customer_id = m.customer_id
    LEFT JOIN locations l ON l.location_id = m.location_id
    LEFT JOIN warehouses w ON w.warehouse_id = l.warehouse_id
    LEFT JOIN inner_vault iv ON iv.stock_id = m.stock_id
    LEFT JOIN outer_vault ov ON ov.stock_id = m.stock_id
WHERE
    w.name IN ('Inner Vault', 'Outer Vault')
GROUP BY
    c.name,
    m.material_type,
    m.stock_id,
    iv.quantity,
    ov.quantity
ORDER BY
    customer_name,
    m.stock_id;