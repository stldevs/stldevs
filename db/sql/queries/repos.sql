-- name: PopularLanguages :many
SELECT
    language,
    COUNT(*) AS repo_count,
    COUNT(DISTINCT owner) AS user_count
FROM agg_repo
WHERE language IS NOT NULL
  AND fork = FALSE
GROUP BY language
ORDER BY repo_count DESC
LIMIT 50;

-- name: LanguageLeaders :many
WITH ranked_repos AS (
    SELECT
        r1.owner,
        r1.name,
        r1.description,
        r1.forks_count,
        r1.stargazers_count,
        r1.watchers_count,
        r1.fork,
        (
            SELECT SUM(stargazers_count)
            FROM agg_repo
            WHERE LOWER(language) = LOWER($1)
              AND owner = r1.owner
        ) AS total_stars,
        ROW_NUMBER() OVER (PARTITION BY r1.owner ORDER BY r1.stargazers_count DESC) AS rownum
    FROM agg_repo AS r1
    WHERE LOWER(r1.language) = LOWER($1)
)
SELECT
    ranked_repos.owner,
    ranked_repos.name,
    ranked_repos.description,
    ranked_repos.forks_count,
    ranked_repos.stargazers_count,
    ranked_repos.watchers_count,
    ranked_repos.fork,
    ranked_repos.total_stars,
    ranked_repos.rownum,
    agg_user.name AS display_name,
    agg_user.type
FROM ranked_repos
JOIN agg_user ON agg_user.login = ranked_repos.owner
WHERE ranked_repos.rownum < 4
ORDER BY ranked_repos.total_stars DESC, ranked_repos.owner, ranked_repos.stargazers_count DESC;

-- name: ReposForUser :many
SELECT *
FROM agg_repo
WHERE LOWER(owner) = LOWER($1)
ORDER BY language, stargazers_count DESC, name;

-- name: SearchRepos :many
SELECT *
FROM agg_repo
WHERE LOWER(name) LIKE LOWER($1)
   OR LOWER(description) LIKE LOWER($1)
ORDER BY stargazers_count DESC
LIMIT 100;

-- name: DeleteReposByOwner :exec
DELETE FROM agg_repo
WHERE owner = $1;

-- name: DeleteReposByOwnerBefore :execrows
DELETE FROM agg_repo
WHERE owner = $1
  AND refreshed_at < $2;

-- name: InsertRepo :exec
INSERT INTO agg_repo (
    owner,
    name,
    description,
    language,
    homepage,
    forks_count,
    network_count,
    open_issues_count,
    stargazers_count,
    subscribers_count,
    watchers_count,
    size,
    fork,
    default_branch,
    master_branch,
    created_at,
    pushed_at,
    updated_at,
    refreshed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19
);

-- name: UpdateRepo :execrows
UPDATE agg_repo
SET
    owner = $1,
    name = $2,
    description = $3,
    language = $4,
    homepage = $5,
    forks_count = $6,
    network_count = $7,
    open_issues_count = $8,
    stargazers_count = $9,
    subscribers_count = $10,
    watchers_count = $11,
    size = $12,
    fork = $13,
    default_branch = $14,
    master_branch = $15,
    created_at = $16,
    pushed_at = $17,
    updated_at = $18,
    refreshed_at = $19
WHERE owner = $1 AND name = $2;
