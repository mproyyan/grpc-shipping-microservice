ALTER TABLE IF EXISTS cargos
ADD CONSTRAINT cargos_delivery_id_fkey
FOREIGN KEY (delivery_id)
REFERENCES deliveries (id)