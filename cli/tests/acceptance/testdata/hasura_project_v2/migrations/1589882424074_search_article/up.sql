CREATE FUNCTION search_article(search text)
RETURNS SETOF article AS $$
    SELECT *
    FROM article
    WHERE
      title ilike ('%' || search || '%')
      OR content ilike ('%' || search || '%')
$$ LANGUAGE sql STABLE;
