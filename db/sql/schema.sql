CREATE TABLE IF NOT EXISTS agg_meta (
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS agg_user (
    login VARCHAR(255) PRIMARY KEY,
    email TEXT,
    location TEXT,
    hireable BOOLEAN,
    blog TEXT,
    bio TEXT,
    followers INTEGER,
    following INTEGER,
    public_repos INTEGER,
    public_gists INTEGER,
    avatar_url TEXT,
    disk_usage INTEGER,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    type VARCHAR(255),
    name VARCHAR(255),
    hide BOOLEAN NOT NULL DEFAULT FALSE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    refreshed_at TIMESTAMPTZ,
    company TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS agg_repo (
    owner VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    language VARCHAR(255),
    homepage TEXT,
    forks_count INTEGER,
    network_count INTEGER,
    open_issues_count INTEGER,
    stargazers_count INTEGER,
    subscribers_count INTEGER,
    watchers_count INTEGER,
    size INTEGER,
    fork BOOLEAN,
    default_branch TEXT,
    master_branch TEXT,
    created_at TIMESTAMPTZ,
    pushed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    refreshed_at TIMESTAMPTZ,
    PRIMARY KEY (owner, name)
);

CREATE TABLE IF NOT EXISTS migrations (
    name VARCHAR(255) PRIMARY KEY
);
