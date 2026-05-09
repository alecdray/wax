package tags

const (
	TagGroupSound = "Sound"
	TagGroupMood  = "Mood"
)

type TagGroupDTO struct {
	ID   string
	Name string
}

type TagDTO struct {
	ID    string
	Name  string
	Group *TagGroupDTO
}

type TagInput struct {
	Name    string
	GroupID string // empty = ungrouped
}
