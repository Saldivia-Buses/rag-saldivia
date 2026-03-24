# cli/areas.py
"""CLI commands for area management."""
import click
from saldivia.auth import AuthDB, Permission


@click.group()
def areas():
    """Manage areas (departments)."""
    pass


@areas.command("list")
def list_areas():
    """List all areas."""
    db = AuthDB()
    areas_list = db.list_areas()

    if not areas_list:
        click.echo("No areas found")
        return

    click.echo(f"{'ID':<4} {'Name':<25} {'Description'}")
    click.echo("-" * 60)
    for a in areas_list:
        click.echo(f"{a.id:<4} {a.name:<25} {a.description}")


@areas.command("create")
@click.argument("name")
@click.option("--description", "-d", default="", help="Area description")
def create_area(name: str, description: str):
    """Create a new area."""
    db = AuthDB()
    area = db.create_area(name, description)
    click.echo(f"Created area: {area.name} (ID: {area.id})")


@areas.command("grant")
@click.argument("area_id", type=int)
@click.argument("collection")
@click.option("--permission", "-p", type=click.Choice(["read", "write", "admin"]), default="read")
def grant_access(area_id: int, collection: str, permission: str):
    """Grant area access to a collection."""
    db = AuthDB()
    db.grant_collection_access(area_id, collection, Permission(permission))
    click.echo(f"Granted {permission} access to '{collection}' for area {area_id}")


@areas.command("revoke")
@click.argument("area_id", type=int)
@click.argument("collection")
def revoke_access(area_id: int, collection: str):
    """Revoke area access to a collection."""
    db = AuthDB()
    db.revoke_collection_access(area_id, collection)
    click.echo(f"Revoked access to '{collection}' for area {area_id}")


@areas.command("permissions")
@click.argument("area_id", type=int)
def show_permissions(area_id: int):
    """Show collections an area can access."""
    db = AuthDB()
    area = db.get_area(area_id)

    if not area:
        click.echo(f"Area {area_id} not found", err=True)
        return

    click.echo(f"Area: {area.name}")
    click.echo("Collections:")

    perms = db.get_area_collections(area_id)
    if not perms:
        click.echo("  (none)")
        return

    for p in perms:
        click.echo(f"  - {p.collection_name}: {p.permission.value}")
