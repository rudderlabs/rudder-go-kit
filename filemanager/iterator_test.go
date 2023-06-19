package filemanager

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectIterator(t *testing.T) {
	newMockListSession := func(pages, itemsPerPage int) ListSession {
		pagesArray := make([][]*FileInfo, pages)
		for i := 0; i < pages; i++ {
			page := make([]*FileInfo, itemsPerPage)
			for j := 0; j < itemsPerPage; j++ {
				page[j] = &FileInfo{}
			}
			pagesArray[i] = page
		}
		return &itMockListSession{
			pages: pagesArray,
		}
	}

	t.Run("no results", func(t *testing.T) {
		it := NewListIterator(&mockListSession{})
		if it.Next() {
			t.Error("expected no results")
		}
	})

	t.Run("single page", func(t *testing.T) {
		it := NewListIterator(newMockListSession(1, 10))
		var count int
		for it.Next() {
			require.NotNil(t, it.Get())
			count++
		}
		require.Equal(t, 10, count)
	})

	t.Run("two pages", func(t *testing.T) {
		it := NewListIterator(newMockListSession(2, 10))
		var count int
		for it.Next() {
			require.NotNil(t, it.Get())
			count++
		}
		require.Equal(t, 20, count)
	})
}

type itMockListSession struct {
	page  int
	pages [][]*FileInfo
}

func (m *itMockListSession) Next() (fileObjects []*FileInfo, err error) {
	if m.page < len(m.pages) {
		fileObjects = m.pages[m.page]
		m.page++
	}
	return fileObjects, nil
}
