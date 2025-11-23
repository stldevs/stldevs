-- name: PopularDevs :many
SELECT
        agg_user.login,
        agg_user.name,
        agg_user.company,
        agg_user.avatar_url,
        agg_user.followers,
        agg_user.public_repos,
        repo.stars,
        repo.forks,
        agg_user.type
FROM agg_user
JOIN (
        SELECT owner, SUM(stargazers_count) AS stars, SUM(forks_count) AS forks
        FROM agg_repo
        GROUP BY owner
) AS repo ON repo.owner = agg_user.login
WHERE agg_user.type = sqlc.arg(dev_type)
    AND agg_user.hide IS FALSE
    AND (
        sqlc.narg(company_pattern) IS NULL OR
        LOWER(agg_user.company) LIKE LOWER(sqlc.narg(company_pattern))
    )
ORDER BY repo.stars DESC
LIMIT 100;

-- name: GetUser :one
SELECT
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
    company,
    hide,
    is_admin
FROM agg_user
WHERE login = $1;

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

-- name: SearchUsers :many
SELECT
    agg_user.login,
    agg_user.name,
    agg_user.followers,
    agg_user.public_repos,
    agg_user.public_gists,
    agg_user.avatar_url,
    agg_user.type,
    agg_user.hide,
    agg_user.is_admin,
    repo.stars,
    repo.forks
FROM agg_user
JOIN (
    SELECT owner, SUM(stargazers_count) AS stars, SUM(forks_count) AS forks
    FROM agg_repo
    GROUP BY owner
) AS repo ON repo.owner = agg_user.login
WHERE agg_user.hide IS FALSE
  AND (
    LOWER(agg_user.login) LIKE LOWER($1) OR
    LOWER(agg_user.name) LIKE LOWER($1) OR
    LOWER(agg_user.bio) LIKE LOWER($1) OR
    LOWER(agg_user.email) LIKE LOWER($1)
  )
ORDER BY repo.stars DESC
LIMIT 100;
