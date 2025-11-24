-- name: PopularDevs :many
SELECT
        agg_user.login,
        COALESCE(agg_user.name, '')::text AS name,
        agg_user.company,
        COALESCE(agg_user.avatar_url, '')::text AS avatar_url,
        COALESCE(agg_user.followers, 0)::int AS followers,
        COALESCE(agg_user.public_repos, 0)::int AS public_repos,
        repo.stars::int AS stars,
        repo.forks::int AS forks,
        COALESCE(agg_user.type, '')::text AS type
FROM agg_user
JOIN (
        SELECT owner, SUM(stargazers_count) AS stars, SUM(forks_count) AS forks
        FROM agg_repo
        GROUP BY owner
) AS repo ON repo.owner = agg_user.login
WHERE agg_user.type = sqlc.arg(dev_type)
    AND agg_user.hide IS FALSE
    AND (
        sqlc.narg(company_pattern)::text IS NULL OR
        LOWER(agg_user.company) LIKE LOWER(sqlc.narg(company_pattern)::text)
    )
ORDER BY
    CASE WHEN sqlc.arg(sort_by) = 'stars' THEN repo.stars END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'forks' THEN repo.forks END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'followers' THEN agg_user.followers END DESC,
    CASE WHEN sqlc.arg(sort_by) = 'public_repos' THEN agg_user.public_repos END DESC
LIMIT 100;


-- name: UpdateUser :execrows
UPDATE agg_user
SET
    login = $1,
    email = $2,
    name = $3,
    location = $4,
    hireable = $5,
    blog = $6,
    bio = $7,
    followers = $8,
    following = $9,
    public_repos = $10,
    public_gists = $11,
    avatar_url = $12,
    type = $13,
    disk_usage = $14,
    created_at = $15,
    updated_at = $16,
    refreshed_at = $17,
    company = $18
WHERE login = $1;

-- name: InsertUser :exec
INSERT INTO agg_user (
    login,
    email,
    name,
    location,
    hireable,
    blog,
    bio,
    followers,
    following,
    public_repos,
    public_gists,
    avatar_url,
    type,
    disk_usage,
    created_at,
    updated_at,
    refreshed_at,
    company
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15, $16, $17, $18
);

-- name: HideUser :execrows
UPDATE agg_user
SET hide = $1
WHERE login = $2;

-- name: DeleteUser :exec
DELETE FROM agg_user
WHERE login = $1;

-- name: GetUser :one
SELECT
    agg_user.login,
    COALESCE(agg_user.email, '')::text AS email,
    COALESCE(agg_user.name, '')::text AS name,
    COALESCE(agg_user.location, '')::text AS location,
    COALESCE(agg_user.hireable, false)::bool AS hireable,
    COALESCE(agg_user.blog, '')::text AS blog,
    COALESCE(agg_user.bio, '')::text AS bio,
    COALESCE(agg_user.followers, 0)::int AS followers,
    COALESCE(agg_user.following, 0)::int AS following,
    COALESCE(agg_user.public_repos, 0)::int AS public_repos,
    COALESCE(agg_user.public_gists, 0)::int AS public_gists,
    COALESCE(agg_user.avatar_url, '')::text AS avatar_url,
    COALESCE(agg_user.type, '')::text AS type,
    COALESCE(agg_user.disk_usage, 0)::int AS disk_usage,
    agg_user.created_at,
    agg_user.updated_at,
    agg_user.company,
    agg_user.hide,
    agg_user.is_admin,
    COALESCE(repo.stars, 0)::int AS stars,
    COALESCE(repo.forks, 0)::int AS forks
FROM agg_user
LEFT JOIN (
        SELECT owner, SUM(stargazers_count) AS stars, SUM(forks_count) AS forks
        FROM agg_repo
        GROUP BY owner
) AS repo ON repo.owner = agg_user.login
WHERE login = $1;

-- name: SearchUsers :many
SELECT
    agg_user.login,
    COALESCE(agg_user.name, '')::text AS name,
    COALESCE(agg_user.followers, 0)::int AS followers,
    COALESCE(agg_user.public_repos, 0)::int AS public_repos,
    COALESCE(agg_user.public_gists, 0)::int AS public_gists,
    COALESCE(agg_user.avatar_url, '')::text AS avatar_url,
    COALESCE(agg_user.type, '')::text AS type,
    agg_user.hide,
    agg_user.is_admin,
    COALESCE(repo.stars, 0)::int AS stars,
    COALESCE(repo.forks, 0)::int AS forks
FROM agg_user
LEFT JOIN (
    SELECT owner, SUM(stargazers_count) AS stars, SUM(forks_count) AS forks
    FROM agg_repo
    GROUP BY owner
) AS repo ON repo.owner = agg_user.login
WHERE login LIKE $1
ORDER BY stars DESC
LIMIT 50;
