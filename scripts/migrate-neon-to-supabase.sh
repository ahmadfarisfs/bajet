#!/usr/bin/env bash
# Copies Bajet data from the old Neon database (Go backend / GORM tables) into
# the Supabase schema in supabase/schema.sql.
#
#   NEON_DATABASE_URL=postgres://...  SUPABASE_DB_URL=postgres://...  ./scripts/migrate-neon-to-supabase.sh
#
# - Keeps the original ids, so it is safe to run again: rows already copied are
#   skipped (ON CONFLICT DO NOTHING).
# - Cycles arrive unowned with the old Google account id in legacy_user_id; each
#   person's cycles attach to them the first time they sign in to the new app.
# - Everything is imported in one transaction: a failure leaves Supabase untouched.
set -euo pipefail

: "${NEON_DATABASE_URL:?set NEON_DATABASE_URL}"
: "${SUPABASE_DB_URL:?set SUPABASE_DB_URL}"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
export PGTZ=UTC  # old dates are stored as UTC midnight timestamps

echo "→ Exporting from Neon"
psql "$NEON_DATABASE_URL" -v ON_ERROR_STOP=1 -q <<SQL
\copy (select id, user_id, start_date::date, end_date::date, total_budget, coalesce(division_mode, 'equal'), num_periods, created_at, (select count(*) from periods p where p.cycle_id = c.id) from cycles c order by id) to '$work/cycles.csv' csv
\copy (select id, cycle_id, period_number, start_date::date, end_date::date, budget, status, coalesce(result_type, ''), coalesce(result_amount, 0), created_at from periods order by id) to '$work/periods.csv' csv
SQL
# The waitlist table only exists if the landing page ever collected an email.
if [ "$(psql "$NEON_DATABASE_URL" -tAc "select to_regclass('public.waitlists') is not null")" = "t" ]; then
  psql "$NEON_DATABASE_URL" -v ON_ERROR_STOP=1 -q \
    -c "\copy (select lower(trim(email)), coalesce(nullif(source, ''), 'landing'), coalesce(ip, ''), created_at from waitlists order by id) to '$work/waitlist.csv' csv"
else
  : > "$work/waitlist.csv"
fi
echo "  cycles: $(wc -l < "$work/cycles.csv")  periods: $(wc -l < "$work/periods.csv")  waitlist: $(wc -l < "$work/waitlist.csv")"

echo "→ Importing into Supabase"
psql "$SUPABASE_DB_URL" -v ON_ERROR_STOP=1 -q <<SQL
begin;

create temp table in_cycles (id bigint, user_id text, start_date date, end_date date, total_budget numeric,
  division_mode text, num_periods int, created_at timestamptz, period_count int) on commit drop;
create temp table in_periods (id bigint, cycle_id bigint, period_number int, start_date date, end_date date,
  budget numeric, status text, result_type text, result_amount numeric, created_at timestamptz) on commit drop;
create temp table in_waitlist (email text, source text, ip text, created_at timestamptz) on commit drop;

\copy in_cycles from '$work/cycles.csv' csv
\copy in_periods from '$work/periods.csv' csv
\copy in_waitlist from '$work/waitlist.csv' csv

-- Cycles from before sign-in existed have no owner and nobody could open them; skip those.
insert into public.cycles (id, legacy_user_id, start_date, end_date, total_budget, division_mode, num_periods, created_at)
overriding system value
select id, user_id, start_date, end_date, total_budget, division_mode,
       -- num_periods was added later; older rows default to 4 or 0, so trust the real count.
       least(greatest(coalesce(nullif(period_count, 0), num_periods, 4), 1), 12),
       created_at
from in_cycles where coalesce(user_id, '') <> ''
on conflict (id) do nothing;

insert into public.periods (id, cycle_id, period_number, start_date, end_date, budget, status, result_type, result_amount, created_at)
overriding system value
select p.id, p.cycle_id, p.period_number, p.start_date, p.end_date, p.budget,
       case when p.status = 'completed' then 'completed' else 'open' end,
       case when p.status = 'completed' then p.result_type else '' end,
       case when p.status = 'completed' then p.result_amount else 0 end,
       p.created_at
from in_periods p join public.cycles c on c.id = p.cycle_id
on conflict (id) do nothing;

insert into public.waitlist (email, source, ip, created_at)
select email, source, ip, created_at from in_waitlist where position('@' in email) > 1
on conflict (email) do nothing;

-- New rows must get ids after the copied ones.
\o /dev/null
select setval(pg_get_serial_sequence('public.cycles', 'id'), coalesce((select max(id) from public.cycles), 0) + 1, false);
select setval(pg_get_serial_sequence('public.periods', 'id'), coalesce((select max(id) from public.periods), 0) + 1, false);
select setval(pg_get_serial_sequence('public.waitlist', 'id'), coalesce((select max(id) from public.waitlist), 0) + 1, false);

\o
commit;
SQL

echo "→ Supabase now has:"
psql "$SUPABASE_DB_URL" -tA -F ' ' -c "
  select 'cycles', count(*), count(*) filter (where user_id is null) || ' waiting to be claimed' from public.cycles
  union all select 'periods', count(*), '' from public.periods
  union all select 'waitlist', count(*), '' from public.waitlist"
echo "Done. Each person's cycles appear the first time they sign in to the new app."
