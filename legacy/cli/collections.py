# cli/collections.py
"""CLI commands for collection management."""
import click
from saldivia.collections import CollectionManager


@click.group()
def collections():
    """Manage document collections."""
    pass


@collections.command()
def list():
    """List all collections."""
    manager = CollectionManager()
    cols = manager.list()
    if not cols:
        click.echo("No collections found")
        return
    for col in cols:
        stats = manager.stats(col)
        if stats:
            click.echo(f"  {col}: {stats.entity_count} entities, {stats.index_type}")
        else:
            click.echo(f"  {col}")


@collections.command()
@click.argument("name")
@click.option("--schema", default="hybrid", help="Schema type: hybrid or dense")
def create(name: str, schema: str):
    """Create a new collection."""
    manager = CollectionManager()
    if manager.create(name, schema):
        click.echo(f"Created collection: {name}")
    else:
        click.echo(f"Failed to create collection: {name}", err=True)


@collections.command()
@click.argument("name")
@click.option("--confirm", is_flag=True, help="Confirm deletion")
def delete(name: str, confirm: bool):
    """Delete a collection."""
    if not confirm:
        click.echo("Add --confirm to delete")
        return
    manager = CollectionManager()
    if manager.delete(name):
        click.echo(f"Deleted collection: {name}")
    else:
        click.echo(f"Collection not found: {name}", err=True)


@collections.command()
@click.argument("name")
def stats(name: str):
    """Show collection statistics."""
    manager = CollectionManager()
    s = manager.stats(name)
    if not s:
        click.echo(f"Collection not found: {name}", err=True)
        return
    click.echo(f"Collection: {s.name}")
    click.echo(f"  Entities: {s.entity_count}")
    click.echo(f"  Index: {s.index_type}")
    click.echo(f"  Hybrid: {s.has_sparse}")
