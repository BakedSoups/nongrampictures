create or replace function public.browse_content_chat(p_kind text, p_content_id uuid)
returns jsonb language sql stable security definer set search_path = public as $$
  select coalesce(jsonb_agg(jsonb_build_object(
    'id', message.id,
    'authorId', message.author_id,
    'authorName', coalesce(profile.display_name, 'Creator'),
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
  left join public.profiles profile on profile.id = message.author_id;
$$;

create or replace function public.post_content_chat(p_kind text, p_content_id uuid, p_body text)
returns jsonb language plpgsql security definer set search_path = public as $$
declare
  inserted public.content_chat_messages;
begin
  if auth.uid() is null then raise exception 'Sign in to chat'; end if;
  if char_length(trim(coalesce(p_body, ''))) not between 1 and 220 then
    raise exception 'Message must be 1 to 220 characters';
  end if;

  insert into public.profiles (id, display_name)
  values (auth.uid(), 'Creator')
  on conflict (id) do nothing;

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

grant execute on function public.browse_content_chat(text, uuid) to anon, authenticated;
grant execute on function public.post_content_chat(text, uuid, text) to authenticated;
