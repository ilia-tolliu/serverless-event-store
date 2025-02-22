package estypes

type EventPage struct {
	Events               []Event `json:"events"`
	HasMore              bool    `json:"hasMore"`
	LasEvaluatedRevision int     `json:"lasEvaluatedRevision"`
}
