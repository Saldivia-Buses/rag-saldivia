import { EntityList } from "../_components/entity-list";

export default function ProveedoresPage() {
  return (
    <EntityList
      entityType="supplier"
      title="Proveedores"
      subtitle="Registro de proveedores"
      codeLabel="Código"
    />
  );
}
