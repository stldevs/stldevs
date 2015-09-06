package web

import (
	"time"

	"github.com/google/go-github/github"
)

type User struct {
	Login             *string
	AvatarURL         *string
	HTMLURL           *string
	GravatarID        *string
	Name              *string
	Company           *string
	Blog              *string
	Location          *string
	Email             *string
	Hireable          *bool
	Bio               *string
	PublicRepos       *int
	PublicGists       *int
	Followers         *int
	Following         *int
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
	Type              *string
	SiteAdmin         *bool
	TotalPrivateRepos *int
	OwnedPrivateRepos *int
	PrivateGists      *int
	DiskUsage         *int
	Collaborators     *int

	// API URLs
	URL               *string
	EventsURL         *string
	FollowingURL      *string
	FollowersURL      *string
	GistsURL          *string
	OrganizationsURL  *string
	ReceivedEventsURL *string
	ReposURL          *string
	StarredURL        *string
	SubscriptionsURL  *string
}

type Repository struct {
	Owner            *string
	Name             *string
	FullName         *string
	Description      *string
	Homepage         *string
	DefaultBranch    *string
	MasterBranch     *string
	CreatedAt        *time.Time
	PushedAt         *time.Time
	UpdatedAt        *time.Time
	HTMLURL          *string
	CloneURL         *string
	GitURL           *string
	MirrorURL        *string
	SSHURL           *string
	SVNURL           *string
	Language         *string
	Fork             *bool
	ForksCount       *int
	NetworkCount     *int
	OpenIssuesCount  *int
	StargazersCount  *int
	SubscribersCount *int
	WatchersCount    *int
	Size             *int
	AutoInit         *bool
	Parent           *Repository
	Source           *Repository
	Organization     *github.Organization
	Permissions      *map[string]bool
	Private          *bool
	HasIssues        *bool
	HasWiki          *bool
	HasDownloads     *bool
	TeamID           *int

	// API URLs
	URL              *string
	ArchiveURL       *string
	AssigneesURL     *string
	BlobsURL         *string
	BranchesURL      *string
	CollaboratorsURL *string
	CommentsURL      *string
	CommitsURL       *string
	CompareURL       *string
	ContentsURL      *string
	ContributorsURL  *string
	DownloadsURL     *string
	EventsURL        *string
	ForksURL         *string
	GitCommitsURL    *string
	GitRefsURL       *string
	GitTagsURL       *string
	HooksURL         *string
	IssueCommentURL  *string
	IssueEventsURL   *string
	IssuesURL        *string
	KeysURL          *string
	LabelsURL        *string
	LanguagesURL     *string
	MergesURL        *string
	MilestonesURL    *string
	NotificationsURL *string
	PullsURL         *string
	ReleasesURL      *string
	StargazersURL    *string
	StatusesURL      *string
	SubscribersURL   *string
	SubscriptionURL  *string
	TagsURL          *string
	TreesURL         *string
	TeamsURL         *string
}
