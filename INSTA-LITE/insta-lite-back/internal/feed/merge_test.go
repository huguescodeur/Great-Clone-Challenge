package feed

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMergeEntries_OrderedByCreatedAtDesc(t *testing.T) {
	now := time.Now()
	t1 := now
	t2 := now.Add(-1 * time.Hour)
	t3 := now.Add(-2 * time.Hour)
	t4 := now.Add(-3 * time.Hour)

	id1, id2, id3, id4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	a := []FeedEntry{{PostID: id1, CreatedAt: t1}, {PostID: id3, CreatedAt: t3}}
	b := []FeedEntry{{PostID: id2, CreatedAt: t2}, {PostID: id4, CreatedAt: t4}}

	got := mergeEntries(a, b, 10)

	if len(got) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(got))
	}

	want := []uuid.UUID{id1, id2, id3, id4}
	for i, e := range got {
		if e.PostID != want[i] {
			t.Errorf("entry[%d]: want %s, got %s", i, want[i], e.PostID)
		}
	}
}

func TestMergeEntries_LimitRespected(t *testing.T) {
	now := time.Now()
	a := []FeedEntry{
		{PostID: uuid.New(), CreatedAt: now},
		{PostID: uuid.New(), CreatedAt: now.Add(-2 * time.Hour)},
	}
	b := []FeedEntry{
		{PostID: uuid.New(), CreatedAt: now.Add(-1 * time.Hour)},
	}

	got := mergeEntries(a, b, 2)

	if len(got) != 2 {
		t.Fatalf("expected 2 entries (limit), got %d", len(got))
	}
}

func TestMergeEntries_BothEmpty(t *testing.T) {
	got := mergeEntries(nil, nil, 10)
	if len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}

func TestMergeEntries_OneEmpty(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	a := []FeedEntry{{PostID: id, CreatedAt: now}}

	got := mergeEntries(a, nil, 10)

	if len(got) != 1 || got[0].PostID != id {
		t.Fatalf("expected single entry %s, got %+v", id, got)
	}
}

func TestMergeEntries_SameTimestamp_BothIncluded(t *testing.T) {
	now := time.Now()
	id1, id2 := uuid.New(), uuid.New()

	a := []FeedEntry{{PostID: id1, CreatedAt: now}}
	b := []FeedEntry{{PostID: id2, CreatedAt: now}}

	got := mergeEntries(a, b, 10)

	if len(got) != 2 {
		t.Fatalf("expected 2 entries with same timestamp, got %d", len(got))
	}
}
