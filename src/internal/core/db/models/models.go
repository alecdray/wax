package models

type FeedKind string

const (
	FeedKindSpotify      FeedKind = "spotify"
	FeedKindSpotifyRadar FeedKind = "spotify_radar"
)

type FeedSyncStatus string

const (
	FeedSyncStatusNone    FeedSyncStatus = "none"
	FeedSyncStatusPending FeedSyncStatus = "pending"
	FeedSyncStatusSuccess FeedSyncStatus = "success"
	FeedSyncStatusFailure FeedSyncStatus = "failure"
)

func (f FeedSyncStatus) IsUnsyned() bool {
	return f == FeedSyncStatusNone
}

func (f FeedSyncStatus) IsSyncing() bool {
	return f == FeedSyncStatusPending
}

func (f FeedSyncStatus) IsSynced() bool {
	return f == FeedSyncStatusSuccess
}

func (f FeedSyncStatus) IsSyncFailed() bool {
	return f == FeedSyncStatusFailure
}

type ReleaseFormat string

const (
	ReleaseFormatDigital  ReleaseFormat = "digital"
	ReleaseFormatVinyl    ReleaseFormat = "vinyl"
	ReleaseFormatCD       ReleaseFormat = "cd"
	ReleaseFormatCassette ReleaseFormat = "cassette"
)
