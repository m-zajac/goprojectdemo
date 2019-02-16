package app

// Project entity
type Project struct {
	ID         int
	Name       string
	OwnerLogin string
}

// Contributor entity
type Contributor struct {
	ID    int
	Login string
}

// ContributorStats entity
type ContributorStats struct {
	Contributor Contributor
	Commits     int
}
