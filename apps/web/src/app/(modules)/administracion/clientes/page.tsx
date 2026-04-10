import { EntityList } from "../_components/entity-list";

export default function ClientesPage() {
  return (
    <EntityList
      entityType="customer"
      title="Clientes"
      subtitle="Registro de clientes"
      codeLabel="CUIT"
    />
  );
}
