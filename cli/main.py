# cli/main.py
"""RAG Saldivia CLI."""
import click
from cli.collections import collections
from cli.ingest import ingest
from cli.users import users
from cli.areas import areas
from cli.audit import audit


@click.group()
@click.version_option(version="0.1.0")
def cli():
    """RAG Saldivia - Document RAG Platform"""
    pass


cli.add_command(collections)
cli.add_command(ingest)
cli.add_command(users)
cli.add_command(areas)
cli.add_command(audit)


@cli.command()
def status():
    """Show platform status."""
    from saldivia.collections import CollectionManager
    from saldivia.mode_manager import ModeManager

    click.echo("RAG Saldivia Status")
    click.echo("=" * 40)

    # Collections
    cm = CollectionManager()
    health = cm.health()
    click.echo(f"Milvus: {health['status']}")
    if health['status'] == 'healthy':
        click.echo(f"  Collections: {health['collections']}")

    # Mode (if available)
    try:
        import redis
        r = redis.from_url("redis://localhost:6379")
        mode = r.get("mode_manager:status")
        if mode:
            click.echo(f"Mode: {mode.decode()}")
    except:
        click.echo("Mode: unknown (redis not available)")


if __name__ == "__main__":
    cli()
