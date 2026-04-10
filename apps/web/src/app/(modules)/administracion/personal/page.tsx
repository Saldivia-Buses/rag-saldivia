import { EntityList } from "../_components/entity-list";

export default function PersonalPage() {
  return (
    <EntityList
      entityType="employee"
      title="Personal"
      subtitle="Legajos de empleados"
      codeLabel="Legajo"
    />
  );
}
