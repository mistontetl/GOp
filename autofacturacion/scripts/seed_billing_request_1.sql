-- Inserta ticket y ticket_lines y los relaciona con billing_requests.
-- Ajusta request_token por el registro deseado (por ejemplo, el "id 1" en tu entorno).

WITH ticket_ins AS (
    INSERT INTO public.tickets (
        id_ticket,
        status,
        cancellation_status,
        forma_pago,
        system_id,
        system_group_id,
        total_amount,
        is_global,
        is_sap,
        objectkey,
        create_date,
        date_created_ticket,
        cliente_id
    ) VALUES (
        'TEST-001',
        'CREATED',
        '0',
        '01',
        'SYS',
        'GRP',
        1000.000000,
        false,
        true,
        gen_random_uuid(),
        now(),
        now(),
        1
    )
    RETURNING tk_id
), line1 AS (
    INSERT INTO public.ticket_lines (
        tk_id,
        descripcion,
        clave_prod_serv,
        clave_unidad,
        no_identificacion,
        porcentaje_descuento,
        tax_rate_type_code,
        cantidad,
        valor_unitario,
        base,
        descuento,
        taxrate,
        amount,
        date_create
    )
    SELECT
        tk_id,
        'CONSUMO DE ALIMENTOS',
        '90101501',
        'E48',
        'CONSUMO',
        '0',
        'Tasa',
        1.000000,
        129.310345,
        129.310345,
        0.000000,
        0.160000,
        150.000000,
        now()
    FROM ticket_ins
)
UPDATE public.billing_requests
SET ticket_id = (SELECT tk_id FROM ticket_ins)
WHERE request_token = 'REEMPLAZA_UUID'::uuid;
