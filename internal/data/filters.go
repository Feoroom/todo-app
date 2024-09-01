package data

import (
	"library/internal/validation"
	"math"
	"strings"
)

type Filters struct {
	Page     int
	PageSize int
	// Sort параметр по которому будет происходить сортировка
	// и который имеет вид sort={-}{field}, где {-} символ, который
	// используется для определения сортировки по убыванию
	Sort         string
	SortSafeList []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {

	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}

}

func (f Filters) sortColumn() string {
	for _, safeVal := range f.SortSafeList {
		if f.Sort == safeVal {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter" + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func ValidateFilters(v *validation.Validator, filters Filters) {

	v.Check(filters.Page >= 1, "page", "must be greater than or equal to 1")
	v.Check(filters.Page <= 10_000_000, "page", "must be less than or equal to 10 000 000")

	v.Check(filters.PageSize >= 1, "page_size", "must be greater than or equal to 1")
	v.Check(filters.Page <= 100, "page_size", "must be less than or equal to 100")

	v.Check(validation.In(filters.Sort, filters.SortSafeList...), "sort", "invalid sort value")

}
