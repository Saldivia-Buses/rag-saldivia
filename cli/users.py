# cli/users.py
"""CLI commands for user management."""
import click
from saldivia.auth import AuthDB, Role, generate_api_key


@click.group()
def users():
    """Manage users."""
    pass


@users.command("list")
@click.option("--area", type=int, help="Filter by area ID")
def list_users(area: int):
    """List all users."""
    db = AuthDB()
    users_list = db.list_users(area_id=area)

    if not users_list:
        click.echo("No users found")
        return

    click.echo(f"{'ID':<4} {'Email':<30} {'Name':<20} {'Area':<6} {'Role':<12} {'Active'}")
    click.echo("-" * 80)
    for u in users_list:
        status = "Y" if u.active else "N"
        click.echo(f"{u.id:<4} {u.email:<30} {u.name:<20} {u.area_id:<6} {u.role.value:<12} {status}")


@users.command("create")
@click.argument("email")
@click.argument("name")
@click.argument("area_id", type=int)
@click.option("--role", type=click.Choice(["admin", "area_manager", "user"]), default="user")
def create_user(email: str, name: str, area_id: int, role: str):
    """Create a new user. Returns the API key (shown only once)."""
    db = AuthDB()

    # Verify area exists
    area = db.get_area(area_id)
    if not area:
        click.echo(f"Area {area_id} not found", err=True)
        return

    # Generate API key
    api_key, api_key_hash = generate_api_key()

    user = db.create_user(
        email=email,
        name=name,
        area_id=area_id,
        role=Role(role),
        api_key_hash=api_key_hash
    )

    click.echo(f"Created user: {user.email} (ID: {user.id})")
    click.echo(f"Area: {area.name}")
    click.echo(f"Role: {user.role.value}")
    click.echo("")
    click.echo("API Key (save this, it won't be shown again):")
    click.echo(f"   {api_key}")


@users.command("deactivate")
@click.argument("user_id", type=int)
@click.option("--confirm", is_flag=True, help="Confirm deactivation")
def deactivate_user(user_id: int, confirm: bool):
    """Deactivate a user."""
    if not confirm:
        click.echo("Add --confirm to deactivate")
        return

    db = AuthDB()
    db.deactivate_user(user_id)
    click.echo(f"Deactivated user {user_id}")


@users.command("reset-key")
@click.argument("user_id", type=int)
def reset_key(user_id: int):
    """Generate new API key for user."""
    db = AuthDB()
    api_key, api_key_hash = generate_api_key()

    with db._conn() as conn:
        conn.execute("UPDATE users SET api_key_hash = ? WHERE id = ?", (api_key_hash, user_id))

    click.echo(f"New API key for user {user_id}:")
    click.echo(f"   {api_key}")
