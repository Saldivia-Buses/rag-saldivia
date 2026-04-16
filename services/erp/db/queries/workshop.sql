-- name: ListCustomerVehicles :many
SELECT cv.*, e.name AS owner_name, d.name AS driver_name
FROM erp_customer_vehicles cv
LEFT JOIN erp_entities e ON e.id = cv.owner_id AND e.tenant_id = cv.tenant_id
LEFT JOIN erp_entities d ON d.id = cv.driver_id AND d.tenant_id = cv.tenant_id
WHERE cv.tenant_id = $1
  AND (sqlc.arg(owner_filter)::TEXT = '' OR cv.owner_id::TEXT = sqlc.arg(owner_filter)::TEXT)
  AND cv.active = true
ORDER BY cv.brand, cv.plate
LIMIT $2 OFFSET $3;

-- name: GetCustomerVehicle :one
SELECT * FROM erp_customer_vehicles WHERE id = $1 AND tenant_id = $2;

-- name: CreateCustomerVehicle :one
INSERT INTO erp_customer_vehicles (tenant_id, owner_id, driver_id, plate, chassis_serial, body_serial, internal_number, brand, model_year, seating_capacity, fuel_type, color, destination, observations, manufacturing_unit_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: UpdateCustomerVehicle :execrows
UPDATE erp_customer_vehicles
SET plate=$3, chassis_serial=$4, body_serial=$5, brand=$6, model_year=$7, seating_capacity=$8, fuel_type=$9, color=$10, destination=$11, observations=$12, updated_at=now()
WHERE id=$1 AND tenant_id=$2;

-- name: ListVehicleIncidentTypes :many
SELECT * FROM erp_vehicle_incident_types WHERE tenant_id=$1 AND active=true ORDER BY name;

-- name: CreateVehicleIncidentType :one
INSERT INTO erp_vehicle_incident_types (tenant_id, name) VALUES ($1, $2) RETURNING *;

-- name: ListVehicleIncidents :many
SELECT vi.*, vit.name AS incident_type_name
FROM erp_vehicle_incidents vi
LEFT JOIN erp_vehicle_incident_types vit ON vit.id = vi.incident_type_id AND vit.tenant_id = vi.tenant_id
WHERE vi.tenant_id=$1
  AND (sqlc.arg(vehicle_filter)::TEXT='' OR vi.vehicle_id::TEXT=sqlc.arg(vehicle_filter)::TEXT)
  AND (sqlc.arg(status_filter)::TEXT='' OR vi.status=sqlc.arg(status_filter)::TEXT)
ORDER BY vi.incident_date DESC
LIMIT $2 OFFSET $3;

-- name: CreateVehicleIncident :one
INSERT INTO erp_vehicle_incidents (tenant_id, vehicle_id, incident_type_id, incident_date, location, responsible, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ResolveVehicleIncident :execrows
UPDATE erp_vehicle_incidents
SET status='resolved', resolved_at=now(), updated_at=now()
WHERE id=$1 AND tenant_id=$2;

-- name: GetWorkshopKPIs :one
SELECT
  COUNT(*) AS total_vehicles,
  COUNT(*) FILTER (WHERE active=true) AS active_vehicles
FROM erp_customer_vehicles WHERE tenant_id=$1;
