package main

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type PaginationActionFn func(tx *gorm.DB) *gorm.DB

type PaginatedData[T any] struct {
	Data       []T  `json:"data"`
	Total      int  `json:"total"`
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Prev       *int `json:"prev"`
	Next       *int `json:"next"`
	PagesCount int  `json:"pages_count"`
	Limit      int  `json:"limit"`
}

func Paginate[T schema.Tabler](c *fiber.Ctx, model T, countStmt *gorm.DB, fns ...PaginationActionFn) (*PaginatedData[T], error) {
	limitVal, pageVal, offsetVal := c.Query("limit"), c.Query("page"), c.Query("offset")
	var limit, page, offset int
	if limitVal == "" {
		limit = 10
	} else {
		limit, _ = strconv.Atoi(limitVal)
	}
	// page or offset is required but not both
	// if neither is set then use the default values
	// if page is set then calculate the offset
	// if offset is set then use it
	if pageVal != "" {
		page, _ = strconv.Atoi(pageVal)
		offset = (page - 1) * limit
	} else if offsetVal != "" {
		offset, _ = strconv.Atoi(offsetVal)
		page = offset/limit + 1
	} else {
		page = 1
		offset = 0
	}

	var total int64
	{
		selects := countStmt.Statement.Selects
		tx := countStmt.Session(&gorm.Session{}).Unscoped()
		if len(selects) > 0 {
			tx = tx.Select("COUNT(*)")
		}
		if err := tx.Table(model.TableName()).Count(&total).Error; err != nil {
			return nil, err
		}
	}

	pagesCount := int(total) / limit
	if int(total)%limit != 0 {
		pagesCount++
	}

	pd := PaginatedData[T]{
		Total:      int(total),
		Page:       page,
		PerPage:    limit,
		PagesCount: pagesCount,
		Limit:      limit,
	}

	if page > 1 {
		prev := page - 1
		pd.Prev = &prev
	}

	if page < pagesCount {
		next := page + 1
		pd.Next = &next
	}

	// if the page is greater than the total pages, return an empty array
	if (page-1)*limit > int(total) {
		pd.Data = []T{}
		return &pd, nil
	}

	stmt := countStmt.Table(model.TableName()).Session(&gorm.Session{}).Unscoped().Model(model)
	if len(fns) > 0 {
		for _, fn := range fns {
			if fn != nil {
				stmt = fn(stmt)
			}
		}
		stmt = stmt.Limit(limit).Offset(offset)
	} else {
		stmt = stmt.Limit(limit).Offset(offset)
	}

	data := []T{}
	if err := stmt.Find(&data).Error; err != nil {
		return nil, err
	}
	pd.Data = data
	return &pd, nil
}

type PDTransformer[T, A any] func(data T) (A, error)

func TransformPaginatedData[T, A any](src *PaginatedData[T], transformer PDTransformer[T, A]) (*PaginatedData[A], error) {
	if src == nil {
		return nil, errors.New("source paginated data is nil")
	}
	if transformer == nil {
		return nil, errors.New("transformer function is nil")
	}
	dst := &PaginatedData[A]{
		Data:       []A{},
		Total:      src.Total,
		Page:       src.Page,
		PerPage:    src.PerPage,
		Prev:       src.Prev,
		Next:       src.Next,
		PagesCount: src.PagesCount,
		Limit:      src.Limit,
	}
	for _, data := range src.Data {
		a, err := transformer(data)
		if err != nil {
			return nil, err
		}
		dst.Data = append(dst.Data, a)

	}
	return dst, nil
}

/*
type PaginatedData struct {
	Data interface{} `json:"data"`

	Total      int  `json:"total"`
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Prev       *int `json:"prev"`
	Next       *int `json:"next"`
	PagesCount int  `json:"pages_count"`
}

func Paginate(c *fiber.Ctx, model interface{}, statement *gorm.DB) (*PaginatedData, error) {
	page := c.QueryInt("page")
	if page <= 0 {
		page = 1
	}
	limit := c.QueryInt("limit")
	if limit <= 0 {
		limit = 10
	}

	var total int64
	if err := countRows(*statement, &total); err != nil {
		return nil, err
	}

	pagesCount := int(total) / limit
	if int(total)%limit != 0 {
		pagesCount++
	}

	pd := PaginatedData{
		Total:      int(total),
		Page:       page,
		PerPage:    limit,
		PagesCount: pagesCount,
	}

	if page > 1 {
		prev := page - 1
		pd.Prev = &prev
	}

	if page < pagesCount {
		next := page + 1
		pd.Next = &next
	}

	// if the page is greater than the total pages, return an empty array
	if (page-1)*limit > int(total) {
		pd.Data = []interface{}{}
		return &pd, nil
	}

	offset := (page - 1) * limit
	rs := statement.Model(model).
		Offset(offset).
		Limit(limit).
		Find(pd.Data)
	if err := rs.Error; err != nil {
		return nil, err
	}
	return &pd, nil
}

func countRows(query gorm.DB, total *int64) error {
	selects := query.Statement.Selects
	tx := query.Session(&gorm.Session{}).Unscoped()
	if len(selects) > 0 {
		tx = tx.Select("COUNT(*)")
	}
	return tx.Count(total).Error
}

*/
