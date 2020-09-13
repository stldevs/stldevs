package stldevs

import (
	"fmt"
	"github.com/google/go-github/github"
	"time"
)

type Repository struct {
	Owner            *string
	Name             *string
	FullName         *string              `json:",omitempty"`
	Description      *string              `json:",omitempty"`
	Homepage         *string              `json:",omitempty"`
	DefaultBranch    *string              `json:",omitempty"`
	MasterBranch     *string              `json:",omitempty"`
	CreatedAt        *time.Time           `json:",omitempty"`
	PushedAt         *time.Time           `json:",omitempty"`
	UpdatedAt        *time.Time           `json:",omitempty"`
	HTMLURL          *string              `json:",omitempty"`
	CloneURL         *string              `json:",omitempty"`
	GitURL           *string              `json:",omitempty"`
	MirrorURL        *string              `json:",omitempty"`
	SSHURL           *string              `json:",omitempty"`
	SVNURL           *string              `json:",omitempty"`
	Language         *string              `json:",omitempty"`
	Fork             *bool                `json:",omitempty"`
	ForksCount       *int                 `json:",omitempty"`
	NetworkCount     *int                 `json:",omitempty"`
	OpenIssuesCount  *int                 `db:"open_issues_count" json:",omitempty"`
	StargazersCount  *int                 `json:",omitempty"`
	SubscribersCount *int                 `json:",omitempty"`
	WatchersCount    *int                 `json:",omitempty"`
	Size             *int                 `json:",omitempty"`
	AutoInit         *bool                `json:",omitempty"`
	Organization     *github.Organization `json:",omitempty"`
	Permissions      *map[string]bool     `json:",omitempty"`
	Private          *bool                `json:",omitempty"`
	HasIssues        *bool                `json:",omitempty"`
	HasWiki          *bool                `json:",omitempty"`
	HasDownloads     *bool                `json:",omitempty"`
	TeamID           *int                 `json:",omitempty"`
	RefreshedAt      *time.Time           `json:",omitempty"`

	// API URLs
	URL              *string `json:",omitempty"`
	ArchiveURL       *string `json:",omitempty"`
	AssigneesURL     *string `json:",omitempty"`
	BlobsURL         *string `json:",omitempty"`
	BranchesURL      *string `json:",omitempty"`
	CollaboratorsURL *string `json:",omitempty"`
	CommentsURL      *string `json:",omitempty"`
	CommitsURL       *string `json:",omitempty"`
	CompareURL       *string `json:",omitempty"`
	ContentsURL      *string `json:",omitempty"`
	ContributorsURL  *string `json:",omitempty"`
	DownloadsURL     *string `json:",omitempty"`
	EventsURL        *string `json:",omitempty"`
	ForksURL         *string `json:",omitempty"`
	GitCommitsURL    *string `json:",omitempty"`
	GitRefsURL       *string `json:",omitempty"`
	GitTagsURL       *string `json:",omitempty"`
	HooksURL         *string `json:",omitempty"`
	IssueCommentURL  *string `json:",omitempty"`
	IssueEventsURL   *string `json:",omitempty"`
	IssuesURL        *string `json:",omitempty"`
	KeysURL          *string `json:",omitempty"`
	LabelsURL        *string `json:",omitempty"`
	LanguagesURL     *string `json:",omitempty"`
	MergesURL        *string `json:",omitempty"`
	MilestonesURL    *string `json:",omitempty"`
	NotificationsURL *string `json:",omitempty"`
	PullsURL         *string `json:",omitempty"`
	ReleasesURL      *string `json:",omitempty"`
	StargazersURL    *string `json:",omitempty"`
	StatusesURL      *string `json:",omitempty"`
	SubscribersURL   *string `json:",omitempty"`
	SubscriptionURL  *string `json:",omitempty"`
	TagsURL          *string `json:",omitempty"`
	TreesURL         *string `json:",omitempty"`
	TeamsURL         *string `json:",omitempty"`
}

func (u Repository) String() string {
	return fmt.Sprintf("Repo: %v/%v", *u.Owner, *u.Name)
}
