package aggregator

import "time"

type User struct {
	Login             *string
	ID                *int
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
