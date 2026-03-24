# cli/ingest.py
"""CLI commands for ingestion."""
import click
from pathlib import Path


@click.group()
def ingest():
    """Manage document ingestion."""
    pass


@ingest.command()
@click.argument("path")
@click.argument("collection")
def add(path: str, collection: str):
    """Add a file or directory to ingestion queue."""
    from saldivia.ingestion_queue import IngestionQueue

    queue = IngestionQueue()
    p = Path(path)

    if p.is_file():
        job = queue.enqueue(str(p), collection)
        click.echo(f"Queued: {p.name} (job {job.id})")
    elif p.is_dir():
        count = 0
        for f in p.glob("**/*.pdf"):
            queue.enqueue(str(f), collection)
            count += 1
        click.echo(f"Queued {count} files")
    else:
        click.echo(f"Path not found: {path}", err=True)


@ingest.command()
def queue():
    """Show ingestion queue status."""
    from saldivia.ingestion_queue import IngestionQueue

    q = IngestionQueue()
    click.echo(f"Pending: {q.pending_count()}")

    jobs = q.list_jobs()[:10]
    for job in jobs:
        status_icon = {"pending": "⏳", "processing": "🔄", "completed": "✅", "failed": "❌"}.get(job.status, "?")
        click.echo(f"  {status_icon} {job.id}: {Path(job.file_path).name} -> {job.collection}")


@ingest.command()
@click.argument("directory")
@click.argument("collection")
def watch(directory: str, collection: str):
    """Watch a directory for new files and auto-ingest."""
    from saldivia.watch import start_watcher
    start_watcher(directory, collection)


@ingest.command()
def clear():
    """Clear completed jobs from history."""
    from saldivia.ingestion_queue import IngestionQueue
    q = IngestionQueue()
    q.clear_completed()
    click.echo("Cleared completed jobs")
