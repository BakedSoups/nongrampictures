-- Keep fresh and upgraded projects on the same limits.
alter table public.pack_items drop constraint if exists pack_items_position_check;
alter table public.pack_items add constraint pack_items_position_check check (position between 0 and 31);
alter table public.profiles drop constraint if exists profiles_bio_check;
alter table public.profiles add constraint profiles_bio_check check (char_length(bio) <= 50);

-- Repair the enum/column mix-up shipped in migration 028 on databases where it was already applied.
do $$
declare function_sql text; repaired_sql text;
begin
  select pg_get_functiondef('public.update_published_content(text,uuid,text,text,jsonb)'::regprocedure) into function_sql;
  repaired_sql := replace(function_sql,
    'insert into public.levels (owner_id, local_id, title, description, status, current_version)',
    'insert into public.levels (owner_id, local_id, title, description, visibility, status, current_version)');
  repaired_sql := replace(repaired_sql,
    'values (auth.uid(), level_local_id, level_title, level_description, ''pack_only'', 1)',
    'values (auth.uid(), level_local_id, level_title, level_description, ''pack_only'', ''published'', 1)');
  repaired_sql := replace(repaired_sql,
    'set title = level_title, description = level_description, status = case when status = ''published'' then status else ''pack_only'' end, current_version = next_version, updated_at = now()',
    'set title = level_title, description = level_description, status = ''published'', visibility = case when visibility = ''public'' then visibility else ''pack_only'' end, current_version = next_version, updated_at = now()');
  if repaired_sql is distinct from function_sql then execute repaired_sql; end if;
end;
$$;

do $$
declare function_sql text; repaired_sql text;
begin
  select pg_get_functiondef('public.publish_pack(text,text,jsonb)'::regprocedure) into function_sql;
  repaired_sql := replace(function_sql,
    'insert into public.levels (owner_id, local_id, title, description, status, current_version)',
    'insert into public.levels (owner_id, local_id, title, description, visibility, status, current_version)');
  repaired_sql := replace(repaired_sql,
    'values (uid, level_local_id, level_title, level_description, ''pack_only'', 1)',
    'values (uid, level_local_id, level_title, level_description, ''pack_only'', ''published'', 1)');
  repaired_sql := replace(repaired_sql,
    'set title = level_title, description = level_description, status = case when status = ''published'' then status else ''pack_only'' end, current_version = next_version, updated_at = now()',
    'set title = level_title, description = level_description, status = ''published'', visibility = case when visibility = ''public'' then visibility else ''pack_only'' end, current_version = next_version, updated_at = now()');
  if repaired_sql is distinct from function_sql then execute repaired_sql; end if;
end;
$$;

drop policy if exists profile_owner_update on public.profiles;
create policy profile_owner_update on public.profiles for update
  using (id = auth.uid()) with check (id = auth.uid());

-- Child rows must not reveal hidden or pack-only content through direct REST queries.
drop policy if exists pack_versions_public_read on public.pack_versions;
create policy pack_versions_public_read on public.pack_versions for select using (
  exists (select 1 from public.packs p where p.id = pack_id and
    (p.owner_id = auth.uid() or (p.status = 'published' and p.visibility in ('public', 'unlisted'))))
);
drop policy if exists pack_items_public_read on public.pack_items;
create policy pack_items_public_read on public.pack_items for select using (
  exists (select 1 from public.pack_versions pv join public.packs p on p.id = pv.pack_id
    where pv.id = pack_version_id and
      (p.owner_id = auth.uid() or (p.status = 'published' and p.visibility in ('public', 'unlisted'))))
);

-- Direct writes bypass RPC validation. Keep only the direct operations used by the web client.
revoke all on all tables in schema public from anon, authenticated;
grant select on public.profiles, public.levels, public.level_versions, public.packs, public.pack_versions,
  public.pack_items, public.likes, public.pack_likes, public.profile_promotions, public.content_chat_messages
  to anon, authenticated;
grant select, delete on public.drafts to authenticated;
grant update (display_name, avatar_puzzle, bio, social, favorite_palette, favorite_color) on public.profiles to authenticated;
grant select on public.official_submissions, public.notifications, public.pack_progress to authenticated;
revoke all on all sequences in schema public from anon, authenticated;
alter default privileges for role postgres in schema public revoke all on tables from anon, authenticated;
alter default privileges for role postgres in schema public revoke all on sequences from anon, authenticated;
alter default privileges for role postgres in schema public revoke all on functions from public, anon, authenticated;

-- Do not expose mutating security-definer functions to anonymous callers.
revoke execute on function public.post_content_chat(text, uuid, text) from public, anon;
revoke execute on function public.publish_level(text, text, text, text[], jsonb, boolean, boolean) from public, anon;
revoke execute on function public.publish_pack(text, text, text[]) from public, anon;
revoke execute on function public.publish_pack(text, text, jsonb) from public, anon;
revoke execute on function public.review_official_submission(uuid, public.submission_status, text) from public, anon;
revoke execute on function public.save_draft(jsonb) from public, anon;
revoke execute on function public.set_content_preview(text, uuid, jsonb) from public, anon;
revoke execute on function public.set_profile_promotion(text, uuid) from public, anon;
revoke execute on function public.toggle_gallery_like(text, uuid) from public, anon;
revoke execute on function public.unpublish_community_item(text, uuid) from public, anon;
revoke execute on function public.unpublish_community_local_art(text) from public, anon;
revoke execute on function public.update_published_content(text, uuid, text, text, jsonb) from public, anon;

-- One authenticated play per level per minute prevents trivial counter and storage flooding.
create or replace function public.record_level_play(p_level_id uuid, p_completed boolean default false)
returns void language plpgsql security definer set search_path = public as $$
declare uid uuid := auth.uid();
begin
  if uid is null then return; end if;
  if not exists(select 1 from public.levels where id = p_level_id and status = 'published'
    and visibility in ('public', 'pack_only', 'unlisted')) then raise exception 'Level not found'; end if;
  update public.play_events set completed = completed or coalesce(p_completed, false)
    where id = (select id from public.play_events where user_id = uid and level_id = p_level_id
      and created_at > now() - interval '1 minute' order by created_at desc limit 1);
  if found then return; end if;
  insert into public.play_events(user_id, level_id, completed) values(uid, p_level_id, coalesce(p_completed, false));
end;
$$;
revoke execute on function public.record_level_play(uuid, boolean) from public, anon;
grant execute on function public.record_level_play(uuid, boolean) to authenticated;
create index if not exists play_events_user_level_created_idx on public.play_events(user_id, level_id, created_at desc);

-- Chat reads only public content, creates missing signup profiles safely, and rate-limits posting.
drop policy if exists chat_public_read on public.content_chat_messages;
create policy chat_public_read on public.content_chat_messages for select using (
  (level_id is not null and exists(select 1 from public.levels l where l.id = level_id and l.status = 'published' and l.visibility = 'public'))
  or (pack_id is not null and exists(select 1 from public.packs p where p.id = pack_id and p.status = 'published' and p.visibility = 'public'))
);
drop policy if exists chat_author_insert on public.content_chat_messages;

create or replace function public.browse_content_chat(p_kind text, p_content_id uuid)
returns jsonb language sql stable security definer set search_path = public as $$
  select case when
    (p_kind = 'art' and exists(select 1 from public.levels where id = p_content_id and status = 'published' and visibility = 'public')) or
    (p_kind = 'pack' and exists(select 1 from public.packs where id = p_content_id and status = 'published' and visibility = 'public'))
  then coalesce((select jsonb_agg(jsonb_build_object(
    'id', message.id, 'authorId', message.author_id,
    'authorName', coalesce(profile.display_name, 'Creator'), 'avatarPuzzle', profile.avatar_puzzle,
    'body', message.message_body, 'createdAt', message.created_at
  ) order by message.created_at) from (
    select * from public.content_chat_messages where
      (p_kind = 'art' and level_id = p_content_id) or (p_kind = 'pack' and pack_id = p_content_id)
    order by created_at desc limit 40
  ) message left join public.profiles profile on profile.id = message.author_id), '[]'::jsonb)
  else '[]'::jsonb end;
$$;

create or replace function public.post_content_chat(p_kind text, p_content_id uuid, p_body text)
returns jsonb language plpgsql security definer set search_path = public as $$
declare uid uuid := auth.uid(); inserted public.content_chat_messages;
begin
  if uid is null then raise exception 'Sign in to chat'; end if;
  if char_length(trim(coalesce(p_body, ''))) not between 1 and 220 then raise exception 'Message must be 1 to 220 characters'; end if;
  if (select count(*) from public.content_chat_messages where author_id = uid and created_at > now() - interval '1 minute') >= 5 then
    raise exception 'Please wait before sending more messages';
  end if;
  insert into public.profiles(id, display_name) values(uid, 'Creator') on conflict(id) do nothing;
  if p_kind = 'art' then
    if not exists(select 1 from public.levels where id = p_content_id and status = 'published' and visibility = 'public') then raise exception 'Art not found'; end if;
    insert into public.content_chat_messages(author_id, level_id, message_body) values(uid, p_content_id, trim(p_body)) returning * into inserted;
  elsif p_kind = 'pack' then
    if not exists(select 1 from public.packs where id = p_content_id and status = 'published' and visibility = 'public') then raise exception 'Pack not found'; end if;
    insert into public.content_chat_messages(author_id, pack_id, message_body) values(uid, p_content_id, trim(p_body)) returning * into inserted;
  else raise exception 'Invalid chat kind'; end if;
  return jsonb_build_object('id', inserted.id);
end;
$$;
grant execute on function public.browse_content_chat(text, uuid) to anon, authenticated;
grant execute on function public.post_content_chat(text, uuid, text) to authenticated;
create index if not exists content_chat_author_created_idx on public.content_chat_messages(author_id, created_at desc);

create or replace function public.delete_content_chat(p_message_id uuid)
returns void language plpgsql security definer set search_path = public as $$
begin
  if auth.uid() is null then raise exception 'Sign in to manage chat'; end if;
  delete from public.content_chat_messages m where m.id = p_message_id and
    (m.author_id = auth.uid() or exists(select 1 from public.profiles p where p.id = auth.uid() and p.role in ('moderator', 'admin')));
  if not found then raise exception 'Message not found'; end if;
end;
$$;
revoke execute on function public.delete_content_chat(uuid) from public, anon;
grant execute on function public.delete_content_chat(uuid) to authenticated;
