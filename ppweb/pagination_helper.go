package main

type PageHelperLink struct {
	Page   int64
	Active bool
}

type PageHelper struct {
	Current         int64
	MaxPage         int64
	ContentsPerPage int64
	PageLinks       []PageHelperLink
}

// If false returned, redirect to page=1
func NewPageHelper(currentPage, contentsCount, contentsPerPage, choices int64) (*PageHelper, bool) {
	maxPages := contentsCount / contentsPerPage
	if contentsCount%contentsPerPage != 0 {
		maxPages++
	}

	pages := make([]PageHelperLink, 0, choices*2+1)
	for i := currentPage - choices; i <= currentPage+choices; i++ {
		if i < 1 || i > maxPages {
			continue
		}
		pages = append(pages, PageHelperLink{Page: i, Active: i != currentPage})
	}

	if len(pages) == 0 {
		return nil, false
	}

	pageHelper := &PageHelper{
		Current:         currentPage,
		ContentsPerPage: contentsPerPage,
		PageLinks:       pages,
	}

	return pageHelper, true
}

type PaginationHelper struct {
	Page     int
	IsActive bool
}

func NewPaginationHelper(current, max, choices int) []PaginationHelper {
	left := current - choices
	right := current + choices

	if left <= 0 {
		left = 1
	}

	if right > max {
		right = max
	}

	arr := make([]PaginationHelper, 0, right-left+1+4)

	if left != 1 {
		arr = append(arr, PaginationHelper{1, false})
		arr = append(arr, PaginationHelper{-1, false})
	}

	for i := left; i <= right; i++ {
		isActive := false

		if i == current {
			isActive = true
		}

		arr = append(arr, PaginationHelper{i, isActive})
	}

	if right != max {
		arr = append(arr, PaginationHelper{-1, false})
		arr = append(arr, PaginationHelper{max, false})
	}

	return arr
}
