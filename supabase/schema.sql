


SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


CREATE SCHEMA IF NOT EXISTS "public";


ALTER SCHEMA "public" OWNER TO "pg_database_owner";


COMMENT ON SCHEMA "public" IS 'standard public schema';



CREATE TYPE "public"."content_status" AS ENUM (
    'draft',
    'published',
    'hidden',
    'removed'
);


ALTER TYPE "public"."content_status" OWNER TO "postgres";


CREATE TYPE "public"."content_visibility" AS ENUM (
    'draft',
    'public',
    'pack_only',
    'unlisted'
);


ALTER TYPE "public"."content_visibility" OWNER TO "postgres";


CREATE TYPE "public"."report_status" AS ENUM (
    'open',
    'reviewed',
    'dismissed',
    'actioned'
);


ALTER TYPE "public"."report_status" OWNER TO "postgres";


CREATE TYPE "public"."submission_status" AS ENUM (
    'submitted',
    'in_review',
    'changes_requested',
    'approved',
    'declined'
);


ALTER TYPE "public"."submission_status" OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."browse_completed_levels"() RETURNS "jsonb"
    LANGUAGE "sql" STABLE SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
  select coalesce(jsonb_agg(completed_item.item order by (completed_item.item->>'completedAt')::timestamptz desc), '[]'::jsonb)
  from (
    select distinct on (level.id) jsonb_build_object(
      'kind', 'art', 'id', level.id, 'localId', level.local_id, 'ownerId', level.owner_id,
      'creatorName', profile.display_name, 'creatorBio', profile.bio, 'avatarPuzzle', profile.avatar_puzzle,
      'title', version.title, 'description', version.description,
      'plays', (select count(*) from public.play_events where level_id = level.id),
      'likes', (select count(*) from public.likes where level_id = level.id),
      'liked', exists(select 1 from public.likes where level_id = level.id and user_id = auth.uid()),
      'owned', level.owner_id = auth.uid(),
      'completed', true,
      'promoted', exists(select 1 from public.profile_promotions where owner_id = level.owner_id and level_id = level.id),
      'previewPixels', level.preview_pixels,
      'puzzle', version.puzzle,
      'publishedAt', version.published_at,
      'completedAt', play.created_at
    ) item
    from public.play_events play
    join public.levels level on level.id = play.level_id
    join public.profiles profile on profile.id = level.owner_id
    join public.level_versions version on version.level_id = level.id and version.version = level.current_version
    where play.user_id = auth.uid()
      and play.completed
      and level.status = 'published'
      and level.visibility in ('public', 'pack_only', 'unlisted')
    order by level.id, play.created_at desc
    limit 200
  ) completed_item;
$$;


ALTER FUNCTION "public"."browse_completed_levels"() OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."browse_content_chat"("p_kind" "text", "p_content_id" "uuid") RETURNS "jsonb"
    LANGUAGE "sql" STABLE SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
  select coalesce(jsonb_agg(jsonb_build_object(
    'id', message.id,
    'authorId', message.author_id,
    'authorName', profile.display_name,
    'avatarPuzzle', profile.avatar_puzzle,
    'body', message.message_body,
    'createdAt', message.created_at
  ) order by message.created_at), '[]'::jsonb)
  from (
    select * from public.content_chat_messages
    where (p_kind = 'art' and level_id = p_content_id)
       or (p_kind = 'pack' and pack_id = p_content_id)
    order by created_at desc
    limit 40
  ) message
  join public.profiles profile on profile.id = message.author_id;
$$;


ALTER FUNCTION "public"."browse_content_chat"("p_kind" "text", "p_content_id" "uuid") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."browse_creators"() RETURNS "jsonb"
    LANGUAGE "sql" STABLE SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
  select coalesce(jsonb_agg(creator order by creator->>'displayName'), '[]'::jsonb)
  from (
    select jsonb_build_object(
      'id', profile.id, 'displayName', profile.display_name, 'bio', profile.bio, 'social', profile.social,
      'palette', profile.favorite_palette, 'favoriteColor', profile.favorite_color, 'avatarPuzzle', profile.avatar_puzzle,
      'featured', coalesce((select jsonb_agg(promoted_item.featured) from (
        select jsonb_build_object(
          'kind', 'art', 'id', level.id, 'ownerId', level.owner_id, 'creatorName', profile.display_name,
          'title', version.title, 'likes', (select count(*) from public.likes where level_id = level.id),
          'puzzle', version.puzzle, 'promoted', true, 'publishedAt', version.published_at
        ) featured from public.profile_promotions promotion
        join public.levels level on level.id = promotion.level_id
        join public.level_versions version on version.level_id = level.id and version.version = level.current_version
        where promotion.owner_id = profile.id
        union all
        select jsonb_build_object(
          'kind', 'pack', 'id', pack.id, 'ownerId', pack.owner_id, 'creatorName', profile.display_name,
          'title', pack_version.title, 'likes', (select count(*) from public.pack_likes where pack_id = pack.id),
          'levels', coalesce((select jsonb_agg(jsonb_build_object('id', lv.id, 'levelId', l.id, 'version', lv.version, 'title', lv.title, 'puzzle', lv.puzzle, 'publishedAt', lv.published_at) order by pi.position)
            from public.pack_items pi join public.level_versions lv on lv.id = pi.level_version_id join public.levels l on l.id = lv.level_id where pi.pack_version_id = pack_version.id), '[]'::jsonb),
          'promoted', true, 'publishedAt', pack_version.published_at
        ) featured from public.profile_promotions promotion
        join public.packs pack on pack.id = promotion.pack_id
        join public.pack_versions pack_version on pack_version.pack_id = pack.id and pack_version.version = pack.current_version
        where promotion.owner_id = profile.id
      ) promoted_item), '[]'::jsonb),
      'levels', coalesce((select jsonb_agg(jsonb_build_object(
        'id', version.id, 'levelId', level.id, 'version', version.version, 'title', version.title,
        'description', version.description, 'tags', version.tags, 'puzzle', version.puzzle, 'publishedAt', version.published_at
      ) order by version.published_at desc)
      from public.levels level join public.level_versions version on version.level_id = level.id and version.version = level.current_version
      where level.owner_id = profile.id and level.status = 'published' and level.visibility = 'public'), '[]'::jsonb)
    ) creator from public.profiles profile order by profile.created_at desc limit 100
  ) creators;
$$;


ALTER FUNCTION "public"."browse_creators"() OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."browse_gallery"("p_kind" "text" DEFAULT 'art'::"text", "p_sort" "text" DEFAULT 'new'::"text") RETURNS "jsonb"
    LANGUAGE "plpgsql" STABLE SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
declare result jsonb;
begin
  if p_kind not in ('art', 'pack') then raise exception 'Invalid gallery kind'; end if;
  if p_sort not in ('new', 'top', 'played') then raise exception 'Invalid gallery sort'; end if;

  if p_kind = 'art' then
    select coalesce(jsonb_agg(item order by
      case when p_sort = 'top' then (item->>'likes')::integer end desc,
      case when p_sort in ('top', 'played') then (item->>'plays')::integer end desc,
      (item->>'publishedAt')::timestamptz desc), '[]'::jsonb) into result
    from (
      select jsonb_build_object(
        'kind', 'art', 'id', level.id, 'localId', level.local_id, 'ownerId', level.owner_id,
        'creatorName', profile.display_name, 'creatorBio', profile.bio, 'avatarPuzzle', profile.avatar_puzzle,
        'title', version.title, 'description', version.description,
        'plays', (select count(*) from public.play_events where level_id = level.id),
        'likes', (select count(*) from public.likes where level_id = level.id),
        'liked', exists(select 1 from public.likes where level_id = level.id and user_id = auth.uid()),
        'owned', level.owner_id = auth.uid(),
        'completed', exists(select 1 from public.play_events where level_id = level.id and user_id = auth.uid() and completed),
        'promoted', exists(select 1 from public.profile_promotions where owner_id = level.owner_id and level_id = level.id),
        'previewPixels', level.preview_pixels,
        'puzzle', version.puzzle, 'publishedAt', version.published_at
      ) item
      from public.levels level
      join public.profiles profile on profile.id = level.owner_id
      join public.level_versions version on version.level_id = level.id and version.version = level.current_version
      where level.status = 'published' and level.visibility = 'public'
      limit 200
    ) gallery;
  else
    select coalesce(jsonb_agg(item order by
      case when p_sort = 'top' then (item->>'likes')::integer end desc,
      case when p_sort in ('top', 'played') then (item->>'plays')::integer end desc,
      (item->>'publishedAt')::timestamptz desc), '[]'::jsonb) into result
    from (
      select jsonb_build_object(
        'kind', 'pack', 'id', pack.id, 'ownerId', pack.owner_id,
        'creatorName', profile.display_name, 'creatorBio', profile.bio, 'avatarPuzzle', profile.avatar_puzzle,
        'title', pack_version.title, 'description', pack_version.description,
        'plays', coalesce((select count(*) from public.play_events play
          join public.level_versions played_version on played_version.level_id = play.level_id
          join public.pack_items pack_item on pack_item.level_version_id = played_version.id
          where pack_item.pack_version_id = pack_version.id), 0),
        'likes', (select count(*) from public.pack_likes where pack_id = pack.id),
        'liked', exists(select 1 from public.pack_likes where pack_id = pack.id and user_id = auth.uid()),
        'owned', pack.owner_id = auth.uid(),
        'promoted', exists(select 1 from public.profile_promotions where owner_id = pack.owner_id and pack_id = pack.id),
        'previewPixels', pack.preview_pixels,
        'levels', coalesce((select jsonb_agg(jsonb_build_object(
          'id', level_version.id, 'levelId', level.id, 'localId', level.local_id, 'version', level_version.version,
          'title', level_version.title, 'description', level_version.description,
          'tags', level_version.tags, 'puzzle', level_version.puzzle,
          'completed', exists(select 1 from public.play_events where level_id = level.id and user_id = auth.uid() and completed),
          'publishedAt', level_version.published_at
        ) order by pack_item.position)
        from public.pack_items pack_item
        join public.level_versions level_version on level_version.id = pack_item.level_version_id
        join public.levels level on level.id = level_version.level_id
        where pack_item.pack_version_id = pack_version.id), '[]'::jsonb),
        'publishedAt', pack_version.published_at
      ) item
      from public.packs pack
      join public.profiles profile on profile.id = pack.owner_id
      join public.pack_versions pack_version on pack_version.pack_id = pack.id and pack_version.version = pack.current_version
      where pack.status = 'published' and pack.visibility = 'public'
      limit 200
    ) gallery;
  end if;
  return result;
end;
$$;


ALTER FUNCTION "public"."browse_gallery"("p_kind" "text", "p_sort" "text") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."browse_my_published"() RETURNS "jsonb"
    LANGUAGE "sql" STABLE SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
  select coalesce(jsonb_agg(item order by item->>'publishedAt' desc), '[]'::jsonb)
  from (
    select value as item from jsonb_array_elements(public.browse_gallery('art', 'new'))
    where value->>'ownerId' = auth.uid()::text
    union all
    select value as item from jsonb_array_elements(public.browse_gallery('pack', 'new'))
    where value->>'ownerId' = auth.uid()::text
  ) owned;
$$;


ALTER FUNCTION "public"."browse_my_published"() OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."create_profile_for_user"() RETURNS "trigger"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  insert into public.profiles (id, display_name)
  values (
    new.id,
    left(coalesce(nullif(new.raw_user_meta_data->>'full_name', ''), nullif(split_part(new.email, '@', 1), ''), 'Creator'), 40)
  )
  on conflict do nothing;
  return new;
end;
$$;


ALTER FUNCTION "public"."create_profile_for_user"() OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."post_content_chat"("p_kind" "text", "p_content_id" "uuid", "p_body" "text") RETURNS "jsonb"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
declare
  inserted public.content_chat_messages;
begin
  if auth.uid() is null then raise exception 'Sign in to chat'; end if;
  if char_length(trim(coalesce(p_body, ''))) not between 1 and 220 then raise exception 'Message must be 1 to 220 characters'; end if;
  if p_kind = 'art' then
    if not exists(select 1 from public.levels where id = p_content_id and status = 'published' and visibility = 'public') then
      raise exception 'Art not found';
    end if;
    insert into public.content_chat_messages(author_id, level_id, message_body)
    values(auth.uid(), p_content_id, trim(p_body)) returning * into inserted;
  elsif p_kind = 'pack' then
    if not exists(select 1 from public.packs where id = p_content_id and status = 'published' and visibility = 'public') then
      raise exception 'Pack not found';
    end if;
    insert into public.content_chat_messages(author_id, pack_id, message_body)
    values(auth.uid(), p_content_id, trim(p_body)) returning * into inserted;
  else
    raise exception 'Invalid chat kind';
  end if;
  return jsonb_build_object('id', inserted.id);
end;
$$;


ALTER FUNCTION "public"."post_content_chat"("p_kind" "text", "p_content_id" "uuid", "p_body" "text") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."publish_level"("p_local_id" "text", "p_title" "text", "p_description" "text", "p_tags" "text"[], "p_puzzle" "jsonb", "p_submit_official" boolean DEFAULT false, "p_rights_confirmed" boolean DEFAULT false) RETURNS "jsonb"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $_$
declare
  uid uuid := auth.uid();
  target_level public.levels;
  target_version public.level_versions;
  next_version integer;
  width integer := (p_puzzle->>'width')::integer;
  height integer := (p_puzzle->>'height')::integer;
  row_text text;
  pixel_row jsonb;
  pixel_value text;
  filled_cells integer := 0;
begin
  if uid is null then raise exception 'Sign in before publishing'; end if;
  if trim(p_title) = '' or char_length(p_title) > 80 then raise exception 'Title must be 1 to 80 characters'; end if;
  if width not in (8, 10, 15, 20, 32) or height not in (8, 10, 15, 20, 32) then raise exception 'Unsupported puzzle dimensions'; end if;
  if jsonb_typeof(p_puzzle->'solution') <> 'array' or jsonb_array_length(p_puzzle->'solution') <> height then raise exception 'Invalid solution rows'; end if;
  for row_text in select jsonb_array_elements_text(p_puzzle->'solution') loop
    if char_length(row_text) <> width or row_text !~ '^[01]+$' then raise exception 'Invalid solution row'; end if;
    filled_cells := filled_cells + char_length(replace(row_text, '0', ''));
  end loop;
  if filled_cells = 0 then raise exception 'Puzzle must contain at least one filled cell'; end if;
  if jsonb_typeof(p_puzzle->'skeletonPixels') <> 'array' or jsonb_array_length(p_puzzle->'skeletonPixels') <> height then raise exception 'Invalid Before layer'; end if;
  for pixel_row in select jsonb_array_elements(p_puzzle->'skeletonPixels') loop
    if jsonb_typeof(pixel_row) <> 'array' or jsonb_array_length(pixel_row) <> width then raise exception 'Invalid Before row'; end if;
    for pixel_value in select jsonb_array_elements_text(pixel_row) loop
      if lower(pixel_value) not in ('', 'transparent', '#000000ff') then raise exception 'Before art must be black or transparent'; end if;
    end loop;
  end loop;
  if jsonb_typeof(p_puzzle->'revealPixels') <> 'array' or jsonb_array_length(p_puzzle->'revealPixels') <> height then raise exception 'Invalid After layer'; end if;
  for pixel_row in select jsonb_array_elements(p_puzzle->'revealPixels') loop
    if jsonb_typeof(pixel_row) <> 'array' or jsonb_array_length(pixel_row) <> width then raise exception 'Invalid After row'; end if;
  end loop;
  if p_submit_official and not p_rights_confirmed then raise exception 'Rights confirmation is required'; end if;

  insert into public.levels (owner_id, local_id, title, description, tags)
  values (uid, p_local_id, trim(p_title), coalesce(p_description, ''), coalesce(p_tags, '{}'))
  on conflict (owner_id, local_id) do update set
    title = excluded.title, description = excluded.description, tags = excluded.tags,
    status = 'published', visibility = 'public',
    current_version = public.levels.current_version + 1, updated_at = now()
  returning * into target_level;

  next_version := target_level.current_version;
  insert into public.level_versions (level_id, version, title, description, tags, puzzle)
  values (target_level.id, next_version, target_level.title, target_level.description, target_level.tags, p_puzzle)
  returning * into target_version;

  if p_submit_official then
    insert into public.official_submissions (owner_id, level_id, level_version_id, rights_confirmed)
    values (uid, target_level.id, target_version.id, true);
  end if;
  return jsonb_build_object('levelId', target_level.id, 'levelVersionId', target_version.id, 'version', next_version);
end;
$_$;


ALTER FUNCTION "public"."publish_level"("p_local_id" "text", "p_title" "text", "p_description" "text", "p_tags" "text"[], "p_puzzle" "jsonb", "p_submit_official" boolean, "p_rights_confirmed" boolean) OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_level_local_ids" "text"[]) RETURNS "jsonb"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
declare
	uid uuid := auth.uid();
	target_pack public.packs;
	target_pack_version public.pack_versions;
	current_local_id text;
	level_version_id uuid;
	position_index integer := 0;
begin
	if uid is null then raise exception 'Sign in before publishing'; end if;
	if trim(p_title) = '' or char_length(p_title) > 80 then raise exception 'Pack title must be 1 to 80 characters'; end if;
	if coalesce(cardinality(p_level_local_ids), 0) not between 1 and 32 then raise exception 'Packs must contain 1 to 32 levels'; end if;
	if (select count(distinct value) from unnest(p_level_local_ids) value) <> cardinality(p_level_local_ids) then raise exception 'A level can appear only once'; end if;

	insert into public.packs (owner_id, title, description) values (uid, trim(p_title), coalesce(p_description, '')) returning * into target_pack;
	insert into public.pack_versions (pack_id, version, title, description)
	values (target_pack.id, 1, target_pack.title, target_pack.description) returning * into target_pack_version;

	foreach current_local_id in array p_level_local_ids loop
		select lv.id into level_version_id
		from public.levels l join public.level_versions lv on lv.level_id = l.id and lv.version = l.current_version
		where l.owner_id = uid and l.local_id = current_local_id and l.status = 'published';
		if level_version_id is null then raise exception 'Publish every pack level before publishing the pack'; end if;
		insert into public.pack_items (pack_version_id, level_version_id, position)
		values (target_pack_version.id, level_version_id, position_index);
		position_index := position_index + 1;
	end loop;
	return jsonb_build_object('packId', target_pack.id, 'packVersionId', target_pack_version.id);
end;
$$;


ALTER FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_level_local_ids" "text"[]) OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_levels" "jsonb") RETURNS "jsonb"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $_$
declare
  uid uuid := auth.uid();
  target_pack public.packs;
  target_pack_version public.pack_versions;
  level_payload jsonb;
  target_level public.levels;
  target_version public.level_versions;
  next_version integer;
  level_local_id text;
  level_title text;
  level_description text;
  level_tags text[];
  level_puzzle jsonb;
  width integer;
  height integer;
  row_text text;
  pixel_row jsonb;
  pixel_value text;
  filled_cells integer;
  position_index integer := 0;
begin
  if uid is null then raise exception 'Sign in before publishing'; end if;
  if trim(p_title) = '' or char_length(p_title) > 80 then raise exception 'Pack title must be 1 to 80 characters'; end if;
  if jsonb_typeof(p_levels) <> 'array' or jsonb_array_length(p_levels) not between 1 and 32 then raise exception 'Packs must contain 1 to 32 levels'; end if;
  if (select count(distinct value->>'id') from jsonb_array_elements(p_levels) value) <> jsonb_array_length(p_levels) then raise exception 'A level can appear only once'; end if;

  insert into public.packs (owner_id, title, description)
  values (uid, trim(p_title), coalesce(p_description, ''))
  returning * into target_pack;
  insert into public.pack_versions (pack_id, version, title, description)
  values (target_pack.id, 1, target_pack.title, target_pack.description)
  returning * into target_pack_version;

  for level_payload in select value from jsonb_array_elements(p_levels) value loop
    level_local_id := level_payload->>'id';
    level_title := trim(coalesce(level_payload->>'title', ''));
    level_description := coalesce(level_payload->>'description', '');
    level_tags := coalesce(array(select jsonb_array_elements_text(coalesce(level_payload->'tags', '[]'::jsonb))), '{}');
    level_puzzle := level_payload->'puzzle';
    width := (level_puzzle->>'width')::integer;
    height := (level_puzzle->>'height')::integer;
    filled_cells := 0;

    if coalesce(level_local_id, '') = '' then raise exception 'Pack level is missing an id'; end if;
    if level_title = '' or char_length(level_title) > 80 then raise exception 'Pack level title must be 1 to 80 characters'; end if;
    if width not in (8, 10, 15, 20, 32) or height not in (8, 10, 15, 20, 32) then raise exception 'Unsupported puzzle dimensions'; end if;
    if jsonb_typeof(level_puzzle->'solution') <> 'array' or jsonb_array_length(level_puzzle->'solution') <> height then raise exception 'Invalid solution rows'; end if;
    for row_text in select jsonb_array_elements_text(level_puzzle->'solution') loop
      if char_length(row_text) <> width or row_text !~ '^[01]+$' then raise exception 'Invalid solution row'; end if;
      filled_cells := filled_cells + char_length(replace(row_text, '0', ''));
    end loop;
    if filled_cells = 0 then raise exception 'Puzzle must contain at least one filled cell'; end if;
    if jsonb_typeof(level_puzzle->'skeletonPixels') <> 'array' or jsonb_array_length(level_puzzle->'skeletonPixels') <> height then raise exception 'Invalid Before layer'; end if;
    for pixel_row in select jsonb_array_elements(level_puzzle->'skeletonPixels') loop
      if jsonb_typeof(pixel_row) <> 'array' or jsonb_array_length(pixel_row) <> width then raise exception 'Invalid Before row'; end if;
      for pixel_value in select jsonb_array_elements_text(pixel_row) loop
        if lower(pixel_value) not in ('', 'transparent', '#000000ff') then raise exception 'Before art must be black or transparent'; end if;
      end loop;
    end loop;
    if jsonb_typeof(level_puzzle->'revealPixels') <> 'array' or jsonb_array_length(level_puzzle->'revealPixels') <> height then raise exception 'Invalid After layer'; end if;
    for pixel_row in select jsonb_array_elements(level_puzzle->'revealPixels') loop
      if jsonb_typeof(pixel_row) <> 'array' or jsonb_array_length(pixel_row) <> width then raise exception 'Invalid After row'; end if;
    end loop;

    insert into public.levels (owner_id, local_id, title, description, tags, visibility, status)
    values (uid, level_local_id, level_title, level_description, level_tags, 'pack_only'::public.content_visibility, 'published'::public.content_status)
    on conflict (owner_id, local_id) do update set
      title = excluded.title,
      description = excluded.description,
      tags = excluded.tags,
      status = 'published'::public.content_status,
      visibility = case
        when public.levels.visibility = 'public'::public.content_visibility then 'public'::public.content_visibility
        else 'pack_only'::public.content_visibility
      end,
      current_version = public.levels.current_version + 1,
      updated_at = now()
    returning * into target_level;

    next_version := target_level.current_version;
    insert into public.level_versions (level_id, version, title, description, tags, puzzle)
    values (target_level.id, next_version, target_level.title, target_level.description, target_level.tags, level_puzzle)
    returning * into target_version;

    insert into public.pack_items (pack_version_id, level_version_id, position)
    values (target_pack_version.id, target_version.id, position_index);
    position_index := position_index + 1;
  end loop;

  return jsonb_build_object('packId', target_pack.id, 'packVersionId', target_pack_version.id);
end;
$_$;


ALTER FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_levels" "jsonb") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."record_level_play"("p_level_id" "uuid", "p_completed" boolean DEFAULT false) RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  if not exists(select 1 from public.levels where id = p_level_id and status = 'published' and visibility in ('public', 'pack_only', 'unlisted')) then
    raise exception 'Level not found';
  end if;
  insert into public.play_events(user_id, level_id, completed) values(auth.uid(), p_level_id, coalesce(p_completed, false));
end;
$$;


ALTER FUNCTION "public"."record_level_play"("p_level_id" "uuid", "p_completed" boolean) OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."reject_version_mutation"() RETURNS "trigger"
    LANGUAGE "plpgsql"
    AS $$ begin raise exception 'published versions are immutable'; end; $$;


ALTER FUNCTION "public"."reject_version_mutation"() OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."review_official_submission"("p_submission_id" "uuid", "p_status" "public"."submission_status", "p_note" "text" DEFAULT ''::"text") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
declare submission public.official_submissions;
begin
  if not exists (select 1 from public.profiles where id = auth.uid() and role in ('moderator', 'admin')) then
    raise exception 'Admin access required';
  end if;
  if p_status not in ('in_review', 'changes_requested', 'approved', 'declined') then raise exception 'Invalid review status'; end if;
  update public.official_submissions set status = p_status, reviewer_note = coalesce(p_note, ''), updated_at = now()
  where id = p_submission_id returning * into submission;
  if submission.id is null then raise exception 'Submission not found'; end if;
  insert into public.notifications (user_id, kind, body)
  values (submission.owner_id, 'official_submission', 'Your main-game submission is now ' || replace(p_status::text, '_', ' ') || '.');
end;
$$;


ALTER FUNCTION "public"."review_official_submission"("p_submission_id" "uuid", "p_status" "public"."submission_status", "p_note" "text") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  if auth.uid() is null then
    raise exception 'Sign in before saving a profile';
  end if;
  update public.profiles
  set avatar_puzzle = p_avatar_puzzle
  where id = auth.uid();
end;
$$;


ALTER FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text" DEFAULT ''::"text") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  if auth.uid() is null then
    raise exception 'Sign in before saving a profile';
  end if;
  update public.profiles
  set avatar_puzzle = p_avatar_puzzle,
      bio = left(coalesce(p_bio, ''), 120)
  where id = auth.uid();
end;
$$;


ALTER FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text" DEFAULT ''::"text", "p_display_name" "text" DEFAULT 'Creator'::"text") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  if auth.uid() is null then raise exception 'Sign in before saving a profile'; end if;
  if trim(coalesce(p_display_name, '')) = '' then raise exception 'Name is required'; end if;
  update public.profiles
  set avatar_puzzle = p_avatar_puzzle,
      bio = left(coalesce(p_bio, ''), 120),
      display_name = left(trim(p_display_name), 40)
  where id = auth.uid();
end;
$$;


ALTER FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text" DEFAULT ''::"text", "p_display_name" "text" DEFAULT 'Creator'::"text", "p_social" "text" DEFAULT ''::"text") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  if auth.uid() is null then raise exception 'Sign in before saving a profile'; end if;
  if trim(coalesce(p_display_name, '')) = '' then raise exception 'Name is required'; end if;
  if p_social ~* 'https?://' or p_social ~* 'www\.' or p_social ~ '[/\\]' then
    raise exception 'Social must use supported profile entries';
  end if;
  update public.profiles
  set avatar_puzzle = p_avatar_puzzle,
      bio = left(coalesce(p_bio, ''), 120),
      display_name = left(trim(p_display_name), 40),
      social = left(trim(coalesce(p_social, '')), 160)
  where id = auth.uid();
end;
$$;


ALTER FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text", "p_social" "text") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text" DEFAULT ''::"text", "p_display_name" "text" DEFAULT 'Creator'::"text", "p_social" "text" DEFAULT ''::"text", "p_palette" "text" DEFAULT ''::"text", "p_favorite_color" "text" DEFAULT ''::"text") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $_$
begin
  if auth.uid() is null then raise exception 'Sign in before saving a profile'; end if;
  if trim(coalesce(p_display_name, '')) = '' then raise exception 'Name is required'; end if;
  if p_social ~* 'https?://' or p_social ~* 'www\.' or p_social ~ '[/\\]' then
    raise exception 'Social must use supported profile entries';
  end if;
  if p_favorite_color <> '' and p_favorite_color !~ '^#[0-9A-Fa-f]{6}$' then
    raise exception 'Favorite color must be a hex color';
  end if;
  update public.profiles
  set avatar_puzzle = p_avatar_puzzle,
      bio = left(coalesce(p_bio, ''), 50),
      display_name = left(trim(p_display_name), 40),
      social = left(trim(coalesce(p_social, '')), 160),
      favorite_palette = left(trim(coalesce(p_palette, '')), 40),
      favorite_color = left(trim(coalesce(p_favorite_color, '')), 9)
  where id = auth.uid();
end;
$_$;


ALTER FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text", "p_social" "text", "p_palette" "text", "p_favorite_color" "text") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."save_draft"("p_draft" "jsonb") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
declare
	uid uuid := auth.uid();
	incoming_updated_at timestamptz := coalesce((p_draft->>'updatedAt')::timestamptz, now());
begin
	if uid is null then raise exception 'Sign in before cloud saving'; end if;
	if coalesce(p_draft->>'id', '') = '' then raise exception 'Draft ID is required'; end if;
	if char_length(coalesce(p_draft->>'title', '')) not between 1 and 80 then raise exception 'Invalid title'; end if;
	insert into public.drafts (owner_id, local_id, title, description, tags, puzzle, playtested, updated_at)
	values (
		uid, p_draft->>'id', p_draft->>'title', coalesce(p_draft->>'description', ''),
		coalesce(array(select jsonb_array_elements_text(p_draft->'tags')), '{}'), p_draft->'puzzle',
		coalesce((p_draft->>'playtested')::boolean, false), incoming_updated_at
	)
	on conflict (owner_id, local_id) do update set
		title = excluded.title, description = excluded.description, tags = excluded.tags,
		puzzle = excluded.puzzle, playtested = excluded.playtested, updated_at = excluded.updated_at
	where public.drafts.updated_at <= excluded.updated_at;
end;
$$;


ALTER FUNCTION "public"."save_draft"("p_draft" "jsonb") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."set_content_preview"("p_kind" "text", "p_content_id" "uuid", "p_preview_pixels" "jsonb") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  if auth.uid() is null then raise exception 'Sign in before uploading a cover'; end if;
  if jsonb_typeof(p_preview_pixels) <> 'array' then raise exception 'Invalid cover image'; end if;
  if p_kind = 'art' then
    update public.levels set preview_pixels = p_preview_pixels, updated_at = now()
    where id = p_content_id and owner_id = auth.uid();
  elsif p_kind = 'pack' then
    update public.packs set preview_pixels = p_preview_pixels, updated_at = now()
    where id = p_content_id and owner_id = auth.uid();
  else
    raise exception 'Invalid content kind';
  end if;
  if not found then raise exception 'Published content not found'; end if;
end;
$$;


ALTER FUNCTION "public"."set_content_preview"("p_kind" "text", "p_content_id" "uuid", "p_preview_pixels" "jsonb") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."set_profile_promotion"("p_kind" "text", "p_content_id" "uuid") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  if auth.uid() is null then raise exception 'Sign in to promote work'; end if;
  if p_kind = 'art' then
    if not exists(select 1 from public.levels where id = p_content_id and owner_id = auth.uid() and status = 'published' and visibility = 'public') then raise exception 'Published art not found'; end if;
    delete from public.profile_promotions where owner_id = auth.uid();
    insert into public.profile_promotions(owner_id, level_id) values(auth.uid(), p_content_id);
  elsif p_kind = 'pack' then
    if not exists(select 1 from public.packs where id = p_content_id and owner_id = auth.uid() and status = 'published' and visibility = 'public') then raise exception 'Published pack not found'; end if;
    delete from public.profile_promotions where owner_id = auth.uid();
    insert into public.profile_promotions(owner_id, pack_id) values(auth.uid(), p_content_id);
  else
    raise exception 'Invalid gallery kind';
  end if;
end;
$$;


ALTER FUNCTION "public"."set_profile_promotion"("p_kind" "text", "p_content_id" "uuid") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."toggle_gallery_like"("p_kind" "text", "p_content_id" "uuid") RETURNS boolean
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
declare now_liked boolean;
begin
  if auth.uid() is null then raise exception 'Sign in to like published work'; end if;
  if p_kind = 'art' then
    if not exists(select 1 from public.levels where id = p_content_id and status = 'published' and visibility = 'public') then raise exception 'Art not found'; end if;
    delete from public.likes where user_id = auth.uid() and level_id = p_content_id;
    if found then return false; end if;
    insert into public.likes(user_id, level_id) values(auth.uid(), p_content_id);
  elsif p_kind = 'pack' then
    if not exists(select 1 from public.packs where id = p_content_id and status = 'published' and visibility = 'public') then raise exception 'Pack not found'; end if;
    delete from public.pack_likes where user_id = auth.uid() and pack_id = p_content_id;
    if found then return false; end if;
    insert into public.pack_likes(user_id, pack_id) values(auth.uid(), p_content_id);
  else
    raise exception 'Invalid gallery kind';
  end if;
  return true;
end;
$$;


ALTER FUNCTION "public"."toggle_gallery_like"("p_kind" "text", "p_content_id" "uuid") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."unpublish_community_item"("p_kind" "text", "p_content_id" "uuid") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
begin
  if auth.uid() is null then raise exception 'Sign in to manage published work'; end if;
  if p_kind = 'art' then
    update public.levels set status = 'hidden', updated_at = now()
    where id = p_content_id and owner_id = auth.uid() and status = 'published';
  elsif p_kind = 'pack' then
    update public.packs set status = 'hidden', updated_at = now()
    where id = p_content_id and owner_id = auth.uid() and status = 'published';
  else
    raise exception 'Invalid published content kind';
  end if;
  if not found then raise exception 'Published item not found'; end if;
  delete from public.profile_promotions
  where owner_id = auth.uid() and (level_id = p_content_id or pack_id = p_content_id);
end;
$$;


ALTER FUNCTION "public"."unpublish_community_item"("p_kind" "text", "p_content_id" "uuid") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."unpublish_community_local_art"("p_local_id" "text") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $$
declare
  target_id uuid;
begin
  if auth.uid() is null then raise exception 'Sign in to manage published work'; end if;

  update public.levels set status = 'hidden', updated_at = now()
  where local_id = p_local_id and owner_id = auth.uid() and status = 'published'
  returning id into target_id;

  if target_id is null then raise exception 'Published item not found'; end if;

  delete from public.profile_promotions
  where owner_id = auth.uid() and level_id = target_id;
end;
$$;


ALTER FUNCTION "public"."unpublish_community_local_art"("p_local_id" "text") OWNER TO "postgres";


CREATE OR REPLACE FUNCTION "public"."update_published_content"("p_kind" "text", "p_content_id" "uuid", "p_title" "text", "p_description" "text" DEFAULT ''::"text", "p_levels" "jsonb" DEFAULT NULL::"jsonb") RETURNS "void"
    LANGUAGE "plpgsql" SECURITY DEFINER
    SET "search_path" TO 'public'
    AS $_$
declare
  next_version integer;
  previous_level_version public.level_versions%rowtype;
  previous_pack_version public.pack_versions%rowtype;
  new_pack_version_id uuid;
  level_payload jsonb;
  target_level public.levels;
  target_version public.level_versions;
  level_local_id text;
  level_title text;
  level_description text;
  level_tags text[];
  level_puzzle jsonb;
  width integer;
  height integer;
  row_text text;
  pixel_row jsonb;
  pixel_value text;
  filled_cells integer;
  position_index integer := 0;
begin
  if auth.uid() is null then raise exception 'Sign in to manage published work'; end if;

  p_title := left(trim(coalesce(p_title, '')), 80);
  p_description := left(coalesce(p_description, ''), 500);
  if p_title = '' then raise exception 'Name is required'; end if;

  if p_kind = 'art' then
    select version.* into previous_level_version
    from public.levels level
    join public.level_versions version on version.level_id = level.id and version.version = level.current_version
    where level.id = p_content_id and level.owner_id = auth.uid() and level.status = 'published';

    if previous_level_version.id is null then raise exception 'Published art not found'; end if;

    next_version := previous_level_version.version + 1;
    insert into public.level_versions(level_id, version, title, description, tags, puzzle)
    values (p_content_id, next_version, p_title, p_description, previous_level_version.tags, previous_level_version.puzzle);

    update public.levels
    set title = p_title, description = p_description, current_version = next_version, updated_at = now()
    where id = p_content_id and owner_id = auth.uid();
  elsif p_kind = 'pack' then
    select version.* into previous_pack_version
    from public.packs pack
    join public.pack_versions version on version.pack_id = pack.id and version.version = pack.current_version
    where pack.id = p_content_id and pack.owner_id = auth.uid() and pack.status = 'published';

    if previous_pack_version.id is null then raise exception 'Published pack not found'; end if;

    if p_levels is not null then
      if jsonb_typeof(p_levels) <> 'array' or jsonb_array_length(p_levels) not between 1 and 32 then
        raise exception 'Packs must contain 1 to 32 levels';
      end if;
      if (select count(distinct value->>'id') from jsonb_array_elements(p_levels) value) <> jsonb_array_length(p_levels) then
        raise exception 'A level can appear only once';
      end if;
    end if;

    next_version := previous_pack_version.version + 1;
    insert into public.pack_versions(pack_id, version, title, description)
    values (p_content_id, next_version, p_title, p_description)
    returning id into new_pack_version_id;

    if p_levels is null then
      insert into public.pack_items(pack_version_id, level_version_id, position)
      select new_pack_version_id, level_version_id, position
      from public.pack_items
      where pack_version_id = previous_pack_version.id
      order by position;
    else
      for level_payload in select value from jsonb_array_elements(p_levels) value loop
        level_local_id := level_payload->>'id';
        level_title := trim(coalesce(level_payload->>'title', ''));
        level_description := coalesce(level_payload->>'description', '');
        level_tags := coalesce(array(select jsonb_array_elements_text(coalesce(level_payload->'tags', '[]'::jsonb))), '{}');
        level_puzzle := level_payload->'puzzle';
        width := (level_puzzle->>'width')::integer;
        height := (level_puzzle->>'height')::integer;
        filled_cells := 0;

        if coalesce(level_local_id, '') = '' then raise exception 'Pack level is missing an id'; end if;
        if level_title = '' or char_length(level_title) > 80 then raise exception 'Pack level title must be 1 to 80 characters'; end if;
        if width not in (8, 10, 15, 20, 32) or height not in (8, 10, 15, 20, 32) then raise exception 'Unsupported puzzle dimensions'; end if;
        if jsonb_typeof(level_puzzle->'solution') <> 'array' or jsonb_array_length(level_puzzle->'solution') <> height then raise exception 'Invalid solution rows'; end if;
        for row_text in select jsonb_array_elements_text(level_puzzle->'solution') loop
          if char_length(row_text) <> width or row_text !~ '^[01]+$' then raise exception 'Invalid solution row'; end if;
          filled_cells := filled_cells + char_length(replace(row_text, '0', ''));
        end loop;
        if filled_cells = 0 then raise exception 'Puzzle must contain at least one filled cell'; end if;
        if jsonb_typeof(level_puzzle->'skeletonPixels') <> 'array' or jsonb_array_length(level_puzzle->'skeletonPixels') <> height then raise exception 'Invalid Before layer'; end if;
        for pixel_row in select jsonb_array_elements(level_puzzle->'skeletonPixels') loop
          if jsonb_typeof(pixel_row) <> 'array' or jsonb_array_length(pixel_row) <> width then raise exception 'Invalid Before row'; end if;
          for pixel_value in select jsonb_array_elements_text(pixel_row) loop
            if lower(pixel_value) not in ('', 'transparent', '#000000ff') then raise exception 'Before art must be black or transparent'; end if;
          end loop;
        end loop;
        if jsonb_typeof(level_puzzle->'revealPixels') <> 'array' or jsonb_array_length(level_puzzle->'revealPixels') <> height then raise exception 'Invalid After layer'; end if;

        insert into public.levels (owner_id, local_id, title, description, tags, visibility, status)
        values (auth.uid(), level_local_id, level_title, level_description, level_tags, 'pack_only'::public.content_visibility, 'published'::public.content_status)
        on conflict (owner_id, local_id) do update set
          title = excluded.title,
          description = excluded.description,
          tags = excluded.tags,
          status = 'published'::public.content_status,
          visibility = case
            when public.levels.visibility = 'public'::public.content_visibility then 'public'::public.content_visibility
            else 'pack_only'::public.content_visibility
          end,
          current_version = public.levels.current_version + 1,
          updated_at = now()
        returning * into target_level;

        insert into public.level_versions(level_id, version, title, description, tags, puzzle)
        values (target_level.id, target_level.current_version, target_level.title, target_level.description, target_level.tags, level_puzzle)
        returning * into target_version;

        insert into public.pack_items(pack_version_id, level_version_id, position)
        values (new_pack_version_id, target_version.id, position_index);
        position_index := position_index + 1;
      end loop;
    end if;

    update public.packs
    set title = p_title, description = p_description, current_version = next_version, updated_at = now()
    where id = p_content_id and owner_id = auth.uid();
  else
    raise exception 'Invalid published content kind';
  end if;
end;
$_$;


ALTER FUNCTION "public"."update_published_content"("p_kind" "text", "p_content_id" "uuid", "p_title" "text", "p_description" "text", "p_levels" "jsonb") OWNER TO "postgres";

SET default_tablespace = '';

SET default_table_access_method = "heap";


CREATE TABLE IF NOT EXISTS "public"."content_chat_messages" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "author_id" "uuid" NOT NULL,
    "level_id" "uuid",
    "pack_id" "uuid",
    "message_body" "text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "content_chat_messages_check" CHECK ((("level_id" IS NULL) <> ("pack_id" IS NULL))),
    CONSTRAINT "content_chat_messages_message_body_check" CHECK ((("char_length"(TRIM(BOTH FROM "message_body")) >= 1) AND ("char_length"(TRIM(BOTH FROM "message_body")) <= 220)))
);


ALTER TABLE "public"."content_chat_messages" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."drafts" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "owner_id" "uuid" NOT NULL,
    "local_id" "text",
    "title" "text" NOT NULL,
    "description" "text" DEFAULT ''::"text" NOT NULL,
    "tags" "text"[] DEFAULT '{}'::"text"[] NOT NULL,
    "puzzle" "jsonb" NOT NULL,
    "playtested" boolean DEFAULT false NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "drafts_description_check" CHECK (("char_length"("description") <= 500)),
    CONSTRAINT "drafts_title_check" CHECK ((("char_length"("title") >= 1) AND ("char_length"("title") <= 80)))
);


ALTER TABLE "public"."drafts" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."level_versions" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "level_id" "uuid" NOT NULL,
    "version" integer NOT NULL,
    "title" "text" NOT NULL,
    "description" "text" DEFAULT ''::"text" NOT NULL,
    "tags" "text"[] DEFAULT '{}'::"text"[] NOT NULL,
    "puzzle" "jsonb" NOT NULL,
    "published_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "level_versions_version_check" CHECK (("version" > 0))
);


ALTER TABLE "public"."level_versions" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."levels" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "owner_id" "uuid" NOT NULL,
    "local_id" "text",
    "title" "text" NOT NULL,
    "description" "text" DEFAULT ''::"text" NOT NULL,
    "tags" "text"[] DEFAULT '{}'::"text"[] NOT NULL,
    "visibility" "public"."content_visibility" DEFAULT 'public'::"public"."content_visibility" NOT NULL,
    "status" "public"."content_status" DEFAULT 'published'::"public"."content_status" NOT NULL,
    "current_version" integer DEFAULT 1 NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "preview_pixels" "jsonb",
    CONSTRAINT "levels_current_version_check" CHECK (("current_version" > 0)),
    CONSTRAINT "levels_description_check" CHECK (("char_length"("description") <= 500)),
    CONSTRAINT "levels_title_check" CHECK ((("char_length"("title") >= 1) AND ("char_length"("title") <= 80)))
);


ALTER TABLE "public"."levels" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."likes" (
    "user_id" "uuid" NOT NULL,
    "level_id" "uuid" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."likes" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."notifications" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "user_id" "uuid" NOT NULL,
    "kind" "text" NOT NULL,
    "body" "text" NOT NULL,
    "read_at" timestamp with time zone,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."notifications" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."official_submissions" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "owner_id" "uuid" NOT NULL,
    "level_id" "uuid" NOT NULL,
    "level_version_id" "uuid" NOT NULL,
    "status" "public"."submission_status" DEFAULT 'submitted'::"public"."submission_status" NOT NULL,
    "rights_confirmed" boolean NOT NULL,
    "creator_note" "text" DEFAULT ''::"text" NOT NULL,
    "reviewer_note" "text" DEFAULT ''::"text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "official_submissions_creator_note_check" CHECK (("char_length"("creator_note") <= 1000)),
    CONSTRAINT "official_submissions_reviewer_note_check" CHECK (("char_length"("reviewer_note") <= 1000)),
    CONSTRAINT "official_submissions_rights_confirmed_check" CHECK ("rights_confirmed")
);


ALTER TABLE "public"."official_submissions" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."pack_items" (
    "pack_version_id" "uuid" NOT NULL,
    "level_version_id" "uuid" NOT NULL,
    "position" integer NOT NULL,
    CONSTRAINT "pack_items_position_check" CHECK ((("position" >= 0) AND ("position" <= 19)))
);


ALTER TABLE "public"."pack_items" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."pack_likes" (
    "user_id" "uuid" NOT NULL,
    "pack_id" "uuid" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."pack_likes" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."pack_progress" (
    "user_id" "uuid" NOT NULL,
    "pack_id" "uuid" NOT NULL,
    "completed_level_ids" "uuid"[] DEFAULT '{}'::"uuid"[] NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."pack_progress" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."pack_versions" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "pack_id" "uuid" NOT NULL,
    "version" integer NOT NULL,
    "title" "text" NOT NULL,
    "description" "text" DEFAULT ''::"text" NOT NULL,
    "published_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."pack_versions" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."packs" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "owner_id" "uuid" NOT NULL,
    "title" "text" NOT NULL,
    "description" "text" DEFAULT ''::"text" NOT NULL,
    "tags" "text"[] DEFAULT '{}'::"text"[] NOT NULL,
    "visibility" "public"."content_visibility" DEFAULT 'public'::"public"."content_visibility" NOT NULL,
    "status" "public"."content_status" DEFAULT 'published'::"public"."content_status" NOT NULL,
    "current_version" integer DEFAULT 1 NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "preview_pixels" "jsonb",
    CONSTRAINT "packs_description_check" CHECK (("char_length"("description") <= 500)),
    CONSTRAINT "packs_title_check" CHECK ((("char_length"("title") >= 1) AND ("char_length"("title") <= 80)))
);


ALTER TABLE "public"."packs" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."play_events" (
    "id" bigint NOT NULL,
    "user_id" "uuid",
    "level_id" "uuid" NOT NULL,
    "completed" boolean DEFAULT false NOT NULL,
    "duration_seconds" integer,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "play_events_duration_seconds_check" CHECK (("duration_seconds" >= 0))
);


ALTER TABLE "public"."play_events" OWNER TO "postgres";


ALTER TABLE "public"."play_events" ALTER COLUMN "id" ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME "public"."play_events_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);



CREATE TABLE IF NOT EXISTS "public"."profile_promotions" (
    "owner_id" "uuid" NOT NULL,
    "level_id" "uuid",
    "pack_id" "uuid",
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "profile_promotions_check" CHECK ((("level_id" IS NULL) <> ("pack_id" IS NULL)))
);


ALTER TABLE "public"."profile_promotions" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."profiles" (
    "id" "uuid" NOT NULL,
    "display_name" "text" DEFAULT 'Creator'::"text" NOT NULL,
    "role" "text" DEFAULT 'creator'::"text" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "avatar_puzzle" "jsonb",
    "bio" "text" DEFAULT ''::"text" NOT NULL,
    "social" "text" DEFAULT ''::"text" NOT NULL,
    "favorite_palette" "text" DEFAULT ''::"text" NOT NULL,
    "favorite_color" "text" DEFAULT ''::"text" NOT NULL,
    CONSTRAINT "profiles_bio_check" CHECK (("char_length"("bio") <= 120)),
    CONSTRAINT "profiles_display_name_check" CHECK ((("char_length"("display_name") >= 1) AND ("char_length"("display_name") <= 40))),
    CONSTRAINT "profiles_favorite_color_check" CHECK ((("favorite_color" = ''::"text") OR ("favorite_color" ~ '^#[0-9A-Fa-f]{6}$'::"text"))),
    CONSTRAINT "profiles_favorite_palette_check" CHECK (("char_length"("favorite_palette") <= 40)),
    CONSTRAINT "profiles_role_check" CHECK (("role" = ANY (ARRAY['creator'::"text", 'moderator'::"text", 'admin'::"text"]))),
    CONSTRAINT "profiles_social_check" CHECK ((("char_length"("social") <= 160) AND ("social" !~* 'https?://'::"text") AND ("social" !~* 'www\.'::"text") AND ("social" !~ '[/\\]'::"text")))
);


ALTER TABLE "public"."profiles" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."reports" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "reporter_id" "uuid",
    "level_id" "uuid",
    "pack_id" "uuid",
    "reason" "text" NOT NULL,
    "status" "public"."report_status" DEFAULT 'open'::"public"."report_status" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "reports_check" CHECK ((("level_id" IS NULL) <> ("pack_id" IS NULL))),
    CONSTRAINT "reports_reason_check" CHECK ((("char_length"("reason") >= 3) AND ("char_length"("reason") <= 500)))
);


ALTER TABLE "public"."reports" OWNER TO "postgres";


ALTER TABLE ONLY "public"."content_chat_messages"
    ADD CONSTRAINT "content_chat_messages_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."drafts"
    ADD CONSTRAINT "drafts_owner_id_local_id_key" UNIQUE ("owner_id", "local_id");



ALTER TABLE ONLY "public"."drafts"
    ADD CONSTRAINT "drafts_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."level_versions"
    ADD CONSTRAINT "level_versions_level_id_version_key" UNIQUE ("level_id", "version");



ALTER TABLE ONLY "public"."level_versions"
    ADD CONSTRAINT "level_versions_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."levels"
    ADD CONSTRAINT "levels_owner_id_local_id_key" UNIQUE ("owner_id", "local_id");



ALTER TABLE ONLY "public"."levels"
    ADD CONSTRAINT "levels_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."likes"
    ADD CONSTRAINT "likes_pkey" PRIMARY KEY ("user_id", "level_id");



ALTER TABLE ONLY "public"."notifications"
    ADD CONSTRAINT "notifications_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."official_submissions"
    ADD CONSTRAINT "official_submissions_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."pack_items"
    ADD CONSTRAINT "pack_items_pack_version_id_level_version_id_key" UNIQUE ("pack_version_id", "level_version_id");



ALTER TABLE ONLY "public"."pack_items"
    ADD CONSTRAINT "pack_items_pkey" PRIMARY KEY ("pack_version_id", "position");



ALTER TABLE ONLY "public"."pack_likes"
    ADD CONSTRAINT "pack_likes_pkey" PRIMARY KEY ("user_id", "pack_id");



ALTER TABLE ONLY "public"."pack_progress"
    ADD CONSTRAINT "pack_progress_pkey" PRIMARY KEY ("user_id", "pack_id");



ALTER TABLE ONLY "public"."pack_versions"
    ADD CONSTRAINT "pack_versions_pack_id_version_key" UNIQUE ("pack_id", "version");



ALTER TABLE ONLY "public"."pack_versions"
    ADD CONSTRAINT "pack_versions_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."packs"
    ADD CONSTRAINT "packs_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."play_events"
    ADD CONSTRAINT "play_events_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."profile_promotions"
    ADD CONSTRAINT "profile_promotions_pkey" PRIMARY KEY ("owner_id");



ALTER TABLE ONLY "public"."profiles"
    ADD CONSTRAINT "profiles_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."reports"
    ADD CONSTRAINT "reports_pkey" PRIMARY KEY ("id");



CREATE INDEX "content_chat_level_created_idx" ON "public"."content_chat_messages" USING "btree" ("level_id", "created_at" DESC);



CREATE INDEX "content_chat_pack_created_idx" ON "public"."content_chat_messages" USING "btree" ("pack_id", "created_at" DESC);



CREATE INDEX "levels_browse_idx" ON "public"."levels" USING "btree" ("status", "visibility", "updated_at" DESC);



CREATE INDEX "pack_likes_pack_idx" ON "public"."pack_likes" USING "btree" ("pack_id");



CREATE INDEX "packs_browse_idx" ON "public"."packs" USING "btree" ("status", "visibility", "updated_at" DESC);



CREATE INDEX "play_events_level_idx" ON "public"."play_events" USING "btree" ("level_id");



CREATE INDEX "submissions_review_idx" ON "public"."official_submissions" USING "btree" ("status", "created_at");



CREATE OR REPLACE TRIGGER "immutable_level_versions" BEFORE DELETE OR UPDATE ON "public"."level_versions" FOR EACH ROW EXECUTE FUNCTION "public"."reject_version_mutation"();



CREATE OR REPLACE TRIGGER "immutable_pack_versions" BEFORE DELETE OR UPDATE ON "public"."pack_versions" FOR EACH ROW EXECUTE FUNCTION "public"."reject_version_mutation"();



ALTER TABLE ONLY "public"."content_chat_messages"
    ADD CONSTRAINT "content_chat_messages_author_id_fkey" FOREIGN KEY ("author_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."content_chat_messages"
    ADD CONSTRAINT "content_chat_messages_level_id_fkey" FOREIGN KEY ("level_id") REFERENCES "public"."levels"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."content_chat_messages"
    ADD CONSTRAINT "content_chat_messages_pack_id_fkey" FOREIGN KEY ("pack_id") REFERENCES "public"."packs"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."drafts"
    ADD CONSTRAINT "drafts_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."level_versions"
    ADD CONSTRAINT "level_versions_level_id_fkey" FOREIGN KEY ("level_id") REFERENCES "public"."levels"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."levels"
    ADD CONSTRAINT "levels_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."likes"
    ADD CONSTRAINT "likes_level_id_fkey" FOREIGN KEY ("level_id") REFERENCES "public"."levels"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."likes"
    ADD CONSTRAINT "likes_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."notifications"
    ADD CONSTRAINT "notifications_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."official_submissions"
    ADD CONSTRAINT "official_submissions_level_id_fkey" FOREIGN KEY ("level_id") REFERENCES "public"."levels"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."official_submissions"
    ADD CONSTRAINT "official_submissions_level_version_id_fkey" FOREIGN KEY ("level_version_id") REFERENCES "public"."level_versions"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."official_submissions"
    ADD CONSTRAINT "official_submissions_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."pack_items"
    ADD CONSTRAINT "pack_items_level_version_id_fkey" FOREIGN KEY ("level_version_id") REFERENCES "public"."level_versions"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."pack_items"
    ADD CONSTRAINT "pack_items_pack_version_id_fkey" FOREIGN KEY ("pack_version_id") REFERENCES "public"."pack_versions"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."pack_likes"
    ADD CONSTRAINT "pack_likes_pack_id_fkey" FOREIGN KEY ("pack_id") REFERENCES "public"."packs"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."pack_likes"
    ADD CONSTRAINT "pack_likes_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."pack_progress"
    ADD CONSTRAINT "pack_progress_pack_id_fkey" FOREIGN KEY ("pack_id") REFERENCES "public"."packs"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."pack_progress"
    ADD CONSTRAINT "pack_progress_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."pack_versions"
    ADD CONSTRAINT "pack_versions_pack_id_fkey" FOREIGN KEY ("pack_id") REFERENCES "public"."packs"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."packs"
    ADD CONSTRAINT "packs_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."play_events"
    ADD CONSTRAINT "play_events_level_id_fkey" FOREIGN KEY ("level_id") REFERENCES "public"."levels"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."play_events"
    ADD CONSTRAINT "play_events_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."profiles"("id") ON DELETE SET NULL;



ALTER TABLE ONLY "public"."profile_promotions"
    ADD CONSTRAINT "profile_promotions_level_id_fkey" FOREIGN KEY ("level_id") REFERENCES "public"."levels"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."profile_promotions"
    ADD CONSTRAINT "profile_promotions_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "public"."profiles"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."profile_promotions"
    ADD CONSTRAINT "profile_promotions_pack_id_fkey" FOREIGN KEY ("pack_id") REFERENCES "public"."packs"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."profiles"
    ADD CONSTRAINT "profiles_id_fkey" FOREIGN KEY ("id") REFERENCES "auth"."users"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."reports"
    ADD CONSTRAINT "reports_level_id_fkey" FOREIGN KEY ("level_id") REFERENCES "public"."levels"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."reports"
    ADD CONSTRAINT "reports_pack_id_fkey" FOREIGN KEY ("pack_id") REFERENCES "public"."packs"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."reports"
    ADD CONSTRAINT "reports_reporter_id_fkey" FOREIGN KEY ("reporter_id") REFERENCES "public"."profiles"("id") ON DELETE SET NULL;



CREATE POLICY "chat_author_insert" ON "public"."content_chat_messages" FOR INSERT WITH CHECK (("author_id" = "auth"."uid"()));



CREATE POLICY "chat_public_read" ON "public"."content_chat_messages" FOR SELECT USING (true);



ALTER TABLE "public"."content_chat_messages" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."drafts" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "drafts_owner_all" ON "public"."drafts" USING (("owner_id" = "auth"."uid"())) WITH CHECK (("owner_id" = "auth"."uid"()));



ALTER TABLE "public"."level_versions" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "level_versions_public_read" ON "public"."level_versions" FOR SELECT USING ((EXISTS ( SELECT 1
   FROM "public"."levels" "l"
  WHERE (("l"."id" = "level_versions"."level_id") AND (("l"."owner_id" = "auth"."uid"()) OR (("l"."status" = 'published'::"public"."content_status") AND ("l"."visibility" = ANY (ARRAY['public'::"public"."content_visibility", 'unlisted'::"public"."content_visibility"]))))))));



ALTER TABLE "public"."levels" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "levels_owner_write" ON "public"."levels" USING (("owner_id" = "auth"."uid"())) WITH CHECK (("owner_id" = "auth"."uid"()));



CREATE POLICY "levels_public_read" ON "public"."levels" FOR SELECT USING (((("status" = 'published'::"public"."content_status") AND ("visibility" = ANY (ARRAY['public'::"public"."content_visibility", 'unlisted'::"public"."content_visibility"]))) OR ("owner_id" = "auth"."uid"())));



ALTER TABLE "public"."likes" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "likes_owner_write" ON "public"."likes" USING (("user_id" = "auth"."uid"())) WITH CHECK (("user_id" = "auth"."uid"()));



CREATE POLICY "likes_public_read" ON "public"."likes" FOR SELECT USING (true);



ALTER TABLE "public"."notifications" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "notifications_owner_read" ON "public"."notifications" FOR SELECT USING (("user_id" = "auth"."uid"()));



CREATE POLICY "notifications_owner_update" ON "public"."notifications" FOR UPDATE USING (("user_id" = "auth"."uid"())) WITH CHECK (("user_id" = "auth"."uid"()));



ALTER TABLE "public"."official_submissions" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."pack_items" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "pack_items_public_read" ON "public"."pack_items" FOR SELECT USING ((EXISTS ( SELECT 1
   FROM ("public"."pack_versions" "pv"
     JOIN "public"."packs" "p" ON (("p"."id" = "pv"."pack_id")))
  WHERE (("pv"."id" = "pack_items"."pack_version_id") AND (("p"."owner_id" = "auth"."uid"()) OR ("p"."status" = 'published'::"public"."content_status"))))));



ALTER TABLE "public"."pack_likes" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "pack_likes_owner_write" ON "public"."pack_likes" USING (("user_id" = "auth"."uid"())) WITH CHECK (("user_id" = "auth"."uid"()));



CREATE POLICY "pack_likes_public_read" ON "public"."pack_likes" FOR SELECT USING (true);



ALTER TABLE "public"."pack_progress" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."pack_versions" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "pack_versions_public_read" ON "public"."pack_versions" FOR SELECT USING ((EXISTS ( SELECT 1
   FROM "public"."packs" "p"
  WHERE (("p"."id" = "pack_versions"."pack_id") AND (("p"."owner_id" = "auth"."uid"()) OR ("p"."status" = 'published'::"public"."content_status"))))));



ALTER TABLE "public"."packs" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "packs_owner_write" ON "public"."packs" USING (("owner_id" = "auth"."uid"())) WITH CHECK (("owner_id" = "auth"."uid"()));



CREATE POLICY "packs_public_read" ON "public"."packs" FOR SELECT USING (((("status" = 'published'::"public"."content_status") AND ("visibility" = ANY (ARRAY['public'::"public"."content_visibility", 'unlisted'::"public"."content_visibility"]))) OR ("owner_id" = "auth"."uid"())));



ALTER TABLE "public"."play_events" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "plays_insert" ON "public"."play_events" FOR INSERT WITH CHECK ((("user_id" IS NULL) OR ("user_id" = "auth"."uid"())));



CREATE POLICY "profile_owner_update" ON "public"."profiles" FOR UPDATE USING (("id" = "auth"."uid"())) WITH CHECK ((("id" = "auth"."uid"()) AND ("role" = ( SELECT "profiles_1"."role"
   FROM "public"."profiles" "profiles_1"
  WHERE ("profiles_1"."id" = "auth"."uid"())))));



ALTER TABLE "public"."profile_promotions" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."profiles" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "profiles_public_read" ON "public"."profiles" FOR SELECT USING (true);



CREATE POLICY "progress_owner_all" ON "public"."pack_progress" USING (("user_id" = "auth"."uid"())) WITH CHECK (("user_id" = "auth"."uid"()));



CREATE POLICY "promotions_owner_write" ON "public"."profile_promotions" USING (("owner_id" = "auth"."uid"())) WITH CHECK (("owner_id" = "auth"."uid"()));



CREATE POLICY "promotions_public_read" ON "public"."profile_promotions" FOR SELECT USING (true);



ALTER TABLE "public"."reports" ENABLE ROW LEVEL SECURITY;


CREATE POLICY "reports_admin_read" ON "public"."reports" FOR SELECT USING ((EXISTS ( SELECT 1
   FROM "public"."profiles"
  WHERE (("profiles"."id" = "auth"."uid"()) AND ("profiles"."role" = ANY (ARRAY['moderator'::"text", 'admin'::"text"]))))));



CREATE POLICY "reports_owner_insert" ON "public"."reports" FOR INSERT WITH CHECK ((("reporter_id" IS NULL) OR ("reporter_id" = "auth"."uid"())));



CREATE POLICY "submissions_owner_read" ON "public"."official_submissions" FOR SELECT USING ((("owner_id" = "auth"."uid"()) OR (EXISTS ( SELECT 1
   FROM "public"."profiles"
  WHERE (("profiles"."id" = "auth"."uid"()) AND ("profiles"."role" = ANY (ARRAY['moderator'::"text", 'admin'::"text"])))))));



GRANT USAGE ON SCHEMA "public" TO "postgres";
GRANT USAGE ON SCHEMA "public" TO "anon";
GRANT USAGE ON SCHEMA "public" TO "authenticated";
GRANT USAGE ON SCHEMA "public" TO "service_role";



GRANT ALL ON FUNCTION "public"."browse_completed_levels"() TO "anon";
GRANT ALL ON FUNCTION "public"."browse_completed_levels"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."browse_completed_levels"() TO "service_role";



GRANT ALL ON FUNCTION "public"."browse_content_chat"("p_kind" "text", "p_content_id" "uuid") TO "anon";
GRANT ALL ON FUNCTION "public"."browse_content_chat"("p_kind" "text", "p_content_id" "uuid") TO "authenticated";
GRANT ALL ON FUNCTION "public"."browse_content_chat"("p_kind" "text", "p_content_id" "uuid") TO "service_role";



GRANT ALL ON FUNCTION "public"."browse_creators"() TO "anon";
GRANT ALL ON FUNCTION "public"."browse_creators"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."browse_creators"() TO "service_role";



GRANT ALL ON FUNCTION "public"."browse_gallery"("p_kind" "text", "p_sort" "text") TO "anon";
GRANT ALL ON FUNCTION "public"."browse_gallery"("p_kind" "text", "p_sort" "text") TO "authenticated";
GRANT ALL ON FUNCTION "public"."browse_gallery"("p_kind" "text", "p_sort" "text") TO "service_role";



GRANT ALL ON FUNCTION "public"."browse_my_published"() TO "anon";
GRANT ALL ON FUNCTION "public"."browse_my_published"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."browse_my_published"() TO "service_role";



GRANT ALL ON FUNCTION "public"."create_profile_for_user"() TO "anon";
GRANT ALL ON FUNCTION "public"."create_profile_for_user"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."create_profile_for_user"() TO "service_role";



GRANT ALL ON FUNCTION "public"."post_content_chat"("p_kind" "text", "p_content_id" "uuid", "p_body" "text") TO "anon";
GRANT ALL ON FUNCTION "public"."post_content_chat"("p_kind" "text", "p_content_id" "uuid", "p_body" "text") TO "authenticated";
GRANT ALL ON FUNCTION "public"."post_content_chat"("p_kind" "text", "p_content_id" "uuid", "p_body" "text") TO "service_role";



GRANT ALL ON FUNCTION "public"."publish_level"("p_local_id" "text", "p_title" "text", "p_description" "text", "p_tags" "text"[], "p_puzzle" "jsonb", "p_submit_official" boolean, "p_rights_confirmed" boolean) TO "anon";
GRANT ALL ON FUNCTION "public"."publish_level"("p_local_id" "text", "p_title" "text", "p_description" "text", "p_tags" "text"[], "p_puzzle" "jsonb", "p_submit_official" boolean, "p_rights_confirmed" boolean) TO "authenticated";
GRANT ALL ON FUNCTION "public"."publish_level"("p_local_id" "text", "p_title" "text", "p_description" "text", "p_tags" "text"[], "p_puzzle" "jsonb", "p_submit_official" boolean, "p_rights_confirmed" boolean) TO "service_role";



GRANT ALL ON FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_level_local_ids" "text"[]) TO "anon";
GRANT ALL ON FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_level_local_ids" "text"[]) TO "authenticated";
GRANT ALL ON FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_level_local_ids" "text"[]) TO "service_role";



GRANT ALL ON FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_levels" "jsonb") TO "anon";
GRANT ALL ON FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_levels" "jsonb") TO "authenticated";
GRANT ALL ON FUNCTION "public"."publish_pack"("p_title" "text", "p_description" "text", "p_levels" "jsonb") TO "service_role";



GRANT ALL ON FUNCTION "public"."record_level_play"("p_level_id" "uuid", "p_completed" boolean) TO "anon";
GRANT ALL ON FUNCTION "public"."record_level_play"("p_level_id" "uuid", "p_completed" boolean) TO "authenticated";
GRANT ALL ON FUNCTION "public"."record_level_play"("p_level_id" "uuid", "p_completed" boolean) TO "service_role";



GRANT ALL ON FUNCTION "public"."reject_version_mutation"() TO "anon";
GRANT ALL ON FUNCTION "public"."reject_version_mutation"() TO "authenticated";
GRANT ALL ON FUNCTION "public"."reject_version_mutation"() TO "service_role";



GRANT ALL ON FUNCTION "public"."review_official_submission"("p_submission_id" "uuid", "p_status" "public"."submission_status", "p_note" "text") TO "anon";
GRANT ALL ON FUNCTION "public"."review_official_submission"("p_submission_id" "uuid", "p_status" "public"."submission_status", "p_note" "text") TO "authenticated";
GRANT ALL ON FUNCTION "public"."review_official_submission"("p_submission_id" "uuid", "p_status" "public"."submission_status", "p_note" "text") TO "service_role";



GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb") TO "anon";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb") TO "authenticated";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb") TO "service_role";



GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text") TO "anon";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text") TO "authenticated";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text") TO "service_role";



GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text") TO "anon";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text") TO "authenticated";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text") TO "service_role";



GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text", "p_social" "text") TO "anon";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text", "p_social" "text") TO "authenticated";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text", "p_social" "text") TO "service_role";



GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text", "p_social" "text", "p_palette" "text", "p_favorite_color" "text") TO "anon";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text", "p_social" "text", "p_palette" "text", "p_favorite_color" "text") TO "authenticated";
GRANT ALL ON FUNCTION "public"."save_creator_profile"("p_avatar_puzzle" "jsonb", "p_bio" "text", "p_display_name" "text", "p_social" "text", "p_palette" "text", "p_favorite_color" "text") TO "service_role";



GRANT ALL ON FUNCTION "public"."save_draft"("p_draft" "jsonb") TO "anon";
GRANT ALL ON FUNCTION "public"."save_draft"("p_draft" "jsonb") TO "authenticated";
GRANT ALL ON FUNCTION "public"."save_draft"("p_draft" "jsonb") TO "service_role";



GRANT ALL ON FUNCTION "public"."set_content_preview"("p_kind" "text", "p_content_id" "uuid", "p_preview_pixels" "jsonb") TO "anon";
GRANT ALL ON FUNCTION "public"."set_content_preview"("p_kind" "text", "p_content_id" "uuid", "p_preview_pixels" "jsonb") TO "authenticated";
GRANT ALL ON FUNCTION "public"."set_content_preview"("p_kind" "text", "p_content_id" "uuid", "p_preview_pixels" "jsonb") TO "service_role";



GRANT ALL ON FUNCTION "public"."set_profile_promotion"("p_kind" "text", "p_content_id" "uuid") TO "anon";
GRANT ALL ON FUNCTION "public"."set_profile_promotion"("p_kind" "text", "p_content_id" "uuid") TO "authenticated";
GRANT ALL ON FUNCTION "public"."set_profile_promotion"("p_kind" "text", "p_content_id" "uuid") TO "service_role";



GRANT ALL ON FUNCTION "public"."toggle_gallery_like"("p_kind" "text", "p_content_id" "uuid") TO "anon";
GRANT ALL ON FUNCTION "public"."toggle_gallery_like"("p_kind" "text", "p_content_id" "uuid") TO "authenticated";
GRANT ALL ON FUNCTION "public"."toggle_gallery_like"("p_kind" "text", "p_content_id" "uuid") TO "service_role";



GRANT ALL ON FUNCTION "public"."unpublish_community_item"("p_kind" "text", "p_content_id" "uuid") TO "anon";
GRANT ALL ON FUNCTION "public"."unpublish_community_item"("p_kind" "text", "p_content_id" "uuid") TO "authenticated";
GRANT ALL ON FUNCTION "public"."unpublish_community_item"("p_kind" "text", "p_content_id" "uuid") TO "service_role";



GRANT ALL ON FUNCTION "public"."unpublish_community_local_art"("p_local_id" "text") TO "anon";
GRANT ALL ON FUNCTION "public"."unpublish_community_local_art"("p_local_id" "text") TO "authenticated";
GRANT ALL ON FUNCTION "public"."unpublish_community_local_art"("p_local_id" "text") TO "service_role";



GRANT ALL ON FUNCTION "public"."update_published_content"("p_kind" "text", "p_content_id" "uuid", "p_title" "text", "p_description" "text", "p_levels" "jsonb") TO "anon";
GRANT ALL ON FUNCTION "public"."update_published_content"("p_kind" "text", "p_content_id" "uuid", "p_title" "text", "p_description" "text", "p_levels" "jsonb") TO "authenticated";
GRANT ALL ON FUNCTION "public"."update_published_content"("p_kind" "text", "p_content_id" "uuid", "p_title" "text", "p_description" "text", "p_levels" "jsonb") TO "service_role";



GRANT ALL ON TABLE "public"."content_chat_messages" TO "anon";
GRANT ALL ON TABLE "public"."content_chat_messages" TO "authenticated";
GRANT ALL ON TABLE "public"."content_chat_messages" TO "service_role";



GRANT ALL ON TABLE "public"."drafts" TO "anon";
GRANT ALL ON TABLE "public"."drafts" TO "authenticated";
GRANT ALL ON TABLE "public"."drafts" TO "service_role";



GRANT ALL ON TABLE "public"."level_versions" TO "anon";
GRANT ALL ON TABLE "public"."level_versions" TO "authenticated";
GRANT ALL ON TABLE "public"."level_versions" TO "service_role";



GRANT ALL ON TABLE "public"."levels" TO "anon";
GRANT ALL ON TABLE "public"."levels" TO "authenticated";
GRANT ALL ON TABLE "public"."levels" TO "service_role";



GRANT ALL ON TABLE "public"."likes" TO "anon";
GRANT ALL ON TABLE "public"."likes" TO "authenticated";
GRANT ALL ON TABLE "public"."likes" TO "service_role";



GRANT ALL ON TABLE "public"."notifications" TO "anon";
GRANT ALL ON TABLE "public"."notifications" TO "authenticated";
GRANT ALL ON TABLE "public"."notifications" TO "service_role";



GRANT ALL ON TABLE "public"."official_submissions" TO "anon";
GRANT ALL ON TABLE "public"."official_submissions" TO "authenticated";
GRANT ALL ON TABLE "public"."official_submissions" TO "service_role";



GRANT ALL ON TABLE "public"."pack_items" TO "anon";
GRANT ALL ON TABLE "public"."pack_items" TO "authenticated";
GRANT ALL ON TABLE "public"."pack_items" TO "service_role";



GRANT ALL ON TABLE "public"."pack_likes" TO "anon";
GRANT ALL ON TABLE "public"."pack_likes" TO "authenticated";
GRANT ALL ON TABLE "public"."pack_likes" TO "service_role";



GRANT ALL ON TABLE "public"."pack_progress" TO "anon";
GRANT ALL ON TABLE "public"."pack_progress" TO "authenticated";
GRANT ALL ON TABLE "public"."pack_progress" TO "service_role";



GRANT ALL ON TABLE "public"."pack_versions" TO "anon";
GRANT ALL ON TABLE "public"."pack_versions" TO "authenticated";
GRANT ALL ON TABLE "public"."pack_versions" TO "service_role";



GRANT ALL ON TABLE "public"."packs" TO "anon";
GRANT ALL ON TABLE "public"."packs" TO "authenticated";
GRANT ALL ON TABLE "public"."packs" TO "service_role";



GRANT ALL ON TABLE "public"."play_events" TO "anon";
GRANT ALL ON TABLE "public"."play_events" TO "authenticated";
GRANT ALL ON TABLE "public"."play_events" TO "service_role";



GRANT ALL ON SEQUENCE "public"."play_events_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."play_events_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."play_events_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."profile_promotions" TO "anon";
GRANT ALL ON TABLE "public"."profile_promotions" TO "authenticated";
GRANT ALL ON TABLE "public"."profile_promotions" TO "service_role";



GRANT ALL ON TABLE "public"."profiles" TO "anon";
GRANT ALL ON TABLE "public"."profiles" TO "authenticated";
GRANT ALL ON TABLE "public"."profiles" TO "service_role";



GRANT ALL ON TABLE "public"."reports" TO "anon";
GRANT ALL ON TABLE "public"."reports" TO "authenticated";
GRANT ALL ON TABLE "public"."reports" TO "service_role";



ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "service_role";
