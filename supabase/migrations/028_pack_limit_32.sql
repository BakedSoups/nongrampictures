create or replace function public.publish_pack(p_title text, p_description text, p_level_local_ids text[])
returns jsonb
language plpgsql
security definer
set search_path = public
as $$
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

create or replace function public.publish_pack(p_title text, p_description text, p_levels jsonb)
returns jsonb
language plpgsql
security definer
set search_path = public
as $$
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
    level_tags := coalesce((select array_agg(trim(tag)) from jsonb_array_elements_text(coalesce(level_payload->'tags', '[]'::jsonb)) tag where trim(tag) <> ''), '{}');
    level_puzzle := level_payload->'puzzle';
    width := (level_puzzle->>'width')::integer;
    height := (level_puzzle->>'height')::integer;

    if level_local_id is null or level_local_id = '' then raise exception 'Pack level is missing an id'; end if;
    if level_title = '' or char_length(level_title) > 80 then raise exception 'Level title must be 1 to 80 characters'; end if;
    if level_puzzle is null or width not in (8, 10, 15, 20, 32) or height not in (8, 10, 15, 20, 32) then raise exception 'Invalid puzzle size'; end if;
    if jsonb_array_length(level_puzzle->'solution') <> height then raise exception 'Invalid solution height'; end if;
    if jsonb_array_length(level_puzzle->'skeleton') <> height or jsonb_array_length(level_puzzle->'reveal') <> height then raise exception 'Invalid artwork height'; end if;

    filled_cells := 0;
    for row_text in select value #>> '{}' from jsonb_array_elements(level_puzzle->'solution') value loop
      if char_length(row_text) <> width or row_text !~ '^[01]+$' then raise exception 'Invalid solution row'; end if;
      filled_cells := filled_cells + char_length(replace(row_text, '0', ''));
    end loop;
    if filled_cells = 0 then raise exception 'Puzzle must contain filled cells'; end if;

    for pixel_row in select value from jsonb_array_elements(level_puzzle->'skeleton') value loop
      if jsonb_array_length(pixel_row) <> width then raise exception 'Invalid before art row'; end if;
      for pixel_value in select value #>> '{}' from jsonb_array_elements(pixel_row) value loop
        if pixel_value <> '' and pixel_value !~ '^#[0-9a-fA-F]{8}$' then raise exception 'Invalid before art color'; end if;
      end loop;
    end loop;

    for pixel_row in select value from jsonb_array_elements(level_puzzle->'reveal') value loop
      if jsonb_array_length(pixel_row) <> width then raise exception 'Invalid final art row'; end if;
      for pixel_value in select value #>> '{}' from jsonb_array_elements(pixel_row) value loop
        if pixel_value <> '' and pixel_value !~ '^#[0-9a-fA-F]{8}$' then raise exception 'Invalid final art color'; end if;
      end loop;
    end loop;

    select * into target_level
    from public.levels
    where owner_id = uid and local_id = level_local_id;

    if target_level.id is null then
      insert into public.levels (owner_id, local_id, title, description, status, current_version)
      values (uid, level_local_id, level_title, level_description, 'pack_only', 1)
      returning * into target_level;
      next_version := 1;
    else
      next_version := target_level.current_version + 1;
      update public.levels
      set title = level_title, description = level_description, status = case when status = 'published' then status else 'pack_only' end, current_version = next_version, updated_at = now()
      where id = target_level.id
      returning * into target_level;
    end if;

    insert into public.level_versions (level_id, version, title, description, tags, puzzle)
    values (target_level.id, next_version, level_title, level_description, level_tags, level_puzzle)
    returning * into target_version;

    insert into public.pack_items (pack_version_id, level_version_id, position)
    values (target_pack_version.id, target_version.id, position_index);
    position_index := position_index + 1;
  end loop;

  return jsonb_build_object('packId', target_pack.id, 'packVersionId', target_pack_version.id);
end;
$$;

create or replace function public.update_published_content(p_kind text, p_content_id uuid, p_title text, p_description text, p_levels jsonb default null::jsonb)
returns jsonb
language plpgsql
security definer
set search_path = public
as $$
declare
  previous_level_version public.level_versions;
  previous_pack_version public.pack_versions;
  next_version integer;
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
  if auth.uid() is null then raise exception 'Sign in before updating'; end if;
  if trim(p_title) = '' or char_length(p_title) > 80 then raise exception 'Name must be 1 to 80 characters'; end if;

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
        level_tags := coalesce((select array_agg(trim(tag)) from jsonb_array_elements_text(coalesce(level_payload->'tags', '[]'::jsonb)) tag where trim(tag) <> ''), '{}');
        level_puzzle := level_payload->'puzzle';
        width := (level_puzzle->>'width')::integer;
        height := (level_puzzle->>'height')::integer;

        if level_local_id is null or level_local_id = '' then raise exception 'Pack level is missing an id'; end if;
        if level_title = '' or char_length(level_title) > 80 then raise exception 'Level title must be 1 to 80 characters'; end if;
        if level_puzzle is null or width not in (8, 10, 15, 20, 32) or height not in (8, 10, 15, 20, 32) then raise exception 'Invalid puzzle size'; end if;
        if jsonb_array_length(level_puzzle->'solution') <> height then raise exception 'Invalid solution height'; end if;
        if jsonb_array_length(level_puzzle->'skeleton') <> height or jsonb_array_length(level_puzzle->'reveal') <> height then raise exception 'Invalid artwork height'; end if;

        filled_cells := 0;
        for row_text in select value #>> '{}' from jsonb_array_elements(level_puzzle->'solution') value loop
          if char_length(row_text) <> width or row_text !~ '^[01]+$' then raise exception 'Invalid solution row'; end if;
          filled_cells := filled_cells + char_length(replace(row_text, '0', ''));
        end loop;
        if filled_cells = 0 then raise exception 'Puzzle must contain filled cells'; end if;

        for pixel_row in select value from jsonb_array_elements(level_puzzle->'skeleton') value loop
          if jsonb_array_length(pixel_row) <> width then raise exception 'Invalid before art row'; end if;
          for pixel_value in select value #>> '{}' from jsonb_array_elements(pixel_row) value loop
            if pixel_value <> '' and pixel_value !~ '^#[0-9a-fA-F]{8}$' then raise exception 'Invalid before art color'; end if;
          end loop;
        end loop;

        for pixel_row in select value from jsonb_array_elements(level_puzzle->'reveal') value loop
          if jsonb_array_length(pixel_row) <> width then raise exception 'Invalid final art row'; end if;
          for pixel_value in select value #>> '{}' from jsonb_array_elements(pixel_row) value loop
            if pixel_value <> '' and pixel_value !~ '^#[0-9a-fA-F]{8}$' then raise exception 'Invalid final art color'; end if;
          end loop;
        end loop;

        select * into target_level
        from public.levels
        where owner_id = auth.uid() and local_id = level_local_id;

        if target_level.id is null then
          insert into public.levels (owner_id, local_id, title, description, status, current_version)
          values (auth.uid(), level_local_id, level_title, level_description, 'pack_only', 1)
          returning * into target_level;
          next_version := 1;
        else
          next_version := target_level.current_version + 1;
          update public.levels
          set title = level_title, description = level_description, status = case when status = 'published' then status else 'pack_only' end, current_version = next_version, updated_at = now()
          where id = target_level.id
          returning * into target_level;
        end if;

        insert into public.level_versions (level_id, version, title, description, tags, puzzle)
        values (target_level.id, next_version, level_title, level_description, level_tags, level_puzzle)
        returning * into target_version;

        insert into public.pack_items (pack_version_id, level_version_id, position)
        values (new_pack_version_id, target_version.id, position_index);
        position_index := position_index + 1;
      end loop;
    end if;

    update public.packs
    set title = p_title, description = p_description, current_version = next_version, updated_at = now()
    where id = p_content_id and owner_id = auth.uid();
  else
    raise exception 'Unknown published content type';
  end if;

  return jsonb_build_object('ok', true);
end;
$$;
