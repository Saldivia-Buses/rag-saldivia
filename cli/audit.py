# cli/audit.py
"""CLI commands for audit log."""
import click
from saldivia.auth import AuthDB


@click.group()
def audit():
    """View audit logs."""
    pass


@audit.command("show")
@click.option("--user", type=int, help="Filter by user ID")
@click.option("--limit", "-n", default=50, help="Number of entries")
def show_audit(user: int, limit: int):
    """Show recent audit log entries."""
    db = AuthDB()
    entries = db.get_audit_log(user_id=user, limit=limit)

    if not entries:
        click.echo("No audit entries found")
        return

    click.echo(f"{'Time':<20} {'User':<6} {'Action':<10} {'Collection':<15} {'Preview'}")
    click.echo("-" * 80)

    for e in entries:
        time_str = e.timestamp.strftime("%Y-%m-%d %H:%M") if e.timestamp else "?"
        preview = (e.query_preview or "")[:30]
        click.echo(f"{time_str:<20} {e.user_id:<6} {e.action:<10} {e.collection or '-':<15} {preview}")


@audit.command("export")
@click.argument("output", type=click.Path())
@click.option("--format", "-f", type=click.Choice(["csv", "json"]), default="csv")
def export_audit(output: str, format: str):
    """Export audit log to file."""
    import json as json_lib
    import csv

    db = AuthDB()
    entries = db.get_audit_log(limit=10000)

    if format == "json":
        with open(output, "w") as f:
            json_lib.dump([{
                "id": e.id,
                "user_id": e.user_id,
                "action": e.action,
                "collection": e.collection,
                "query_preview": e.query_preview,
                "ip_address": e.ip_address,
                "timestamp": e.timestamp.isoformat() if e.timestamp else None
            } for e in entries], f, indent=2)
    else:
        with open(output, "w", newline="") as f:
            writer = csv.writer(f)
            writer.writerow(["id", "user_id", "action", "collection", "query_preview", "ip_address", "timestamp"])
            for e in entries:
                writer.writerow([
                    e.id, e.user_id, e.action, e.collection,
                    e.query_preview, e.ip_address,
                    e.timestamp.isoformat() if e.timestamp else ""
                ])

    click.echo(f"Exported {len(entries)} entries to {output}")
