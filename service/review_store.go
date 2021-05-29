package service

import "sync"

// ReviewStore is an interface to store mobile ratings
type ReviewStore interface {
	// Add adds a new mobile score to the store and returns its rating
	Add(mobileID string, score float64) (*Review, error)
}

// Review contains the mobile rating information
type Review struct {
	Count uint32
	Sum   float64
}

//InMemoryReviewStore stores mobilr rating inforamtion in memory
type InMemoryReviewStore struct {
	mutex  sync.RWMutex
	review map[string]*Review
}

// NewInMemoryReviewStore returns new InMemoryReviewStore
func NewInMemoryReviewStore() *InMemoryReviewStore {
	return &InMemoryReviewStore{
		review: map[string]*Review{},
	}
}

// Add adds a new mobile score to the store and returns its review
func (store *InMemoryReviewStore) Add(mobileID string, score float64) (*Review, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	review := store.review[mobileID]
	if review == nil {
		review = &Review{
			Count: 1,
			Sum:   score,
		}
	} else {
		review.Count++
		review.Sum += score
	}

	store.review[mobileID] = review
	return review, nil
}
