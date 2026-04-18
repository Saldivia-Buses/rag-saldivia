-- 068_drop_default_admin_users.down.sql
-- Re-seed the default admin users — irreversible for real without the
-- original hashes, but kept here for migration-pair symmetry. If you truly
-- need to roll back 068, copy the INSERTs from 057/058.
--
-- This down is a stub because restoring deleted admin credentials from a
-- rollback migration would itself re-ship the backdoor; rollback should be
-- a deliberate operator action via the user-update API, not a DDL path.

-- noop
SELECT 1;
