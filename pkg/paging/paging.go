// Package paging provides utilities to paginate SQL query results. Only MySQL is supported at this moment.
package paging

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

// SelectContext executes a query using the provided Queryer, and scan each row into dest, which
// must be a slice. A paginated results (containing dest) will be returned. The *sql.Rows are
// closed automatically.
func SelectContext(ctx context.Context, q sqlx.QueryerContext, options PageOptions, dest interface{}, query string, args ...interface{}) (*Page, error) {
	if options.Limit <= 0 {
		options.Limit = DefaultPageLimit
	}

	// Counting all rows with original query before rewriting the query
	var foundRows uint
	if options.CountFoundRows {
		foundRowsQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS t", query)
		if err := sqlx.GetContext(ctx, q, &foundRows, foundRowsQuery, args...); err != nil {
			return nil, errors.Wrapf(err, "query error: %v", foundRowsQuery)
		}
	}

	cursor, err := parsePageCursor(options.PageToken)
	if err != nil {
		return nil, err
	}

	query, args, err = prepare(options, cursor, query, args...)
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "query error: %v", query)
	}
	// if something happens here, we want to make sure the rows are Closed
	defer rows.Close()
	return scanAll(options, cursor, foundRows, rows, dest)
}

// IsValidColumn validates the column name. It only allow column name in a restricted character set ([0-9,a-z,A-Z$_]).
// Some databases support a wider character set but they are not supported by this package.
func IsValidColumn(name string) bool {
	valid, err := regexp.MatchString("\\A[0-9,a-z,A-Z\\$_]+\\z", name)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return valid
}

// prepare takes a PageOptions, query and arguments and returns a new query with a list of args that can be executed
// by a database.
func prepare(options PageOptions, cursor *PageCursor, query string, args ...interface{}) (string, []interface{}, error) {
	// Case 1: ascending order, no page token
	// SELECT * FROM ( ... ) AS t ORDER BY (c, id) ASC LIMIT :limit+1

	// Case 1: ascending order, next page
	// SELECT * FROM ( ... ) AS t WHERE (c, id) > :cursor ORDER BY (c, id) ASC LIMIT :limit+1
	//
	// Case 2: descending order, next page
	// SELECT * FROM ( ... ) AS t WHERE (c, id) < :cursor ORDER BY (c, id) DESC LIMIT :limit+1
	//
	// Case 3: ascending order, previous page (reversed result)
	// SELECT * FROM ( ... ) AS t WHERE (c, id) < :cursor ORDER BY (c, id) DESC LIMIT :limit+1
	//
	// Case 4: descending order, previous page (reversed result)
	// SELECT * FROM ( ... ) AS t WHERE (c, id) > :cursor ORDER BY (c, id) ASC LIMIT :limit+1

	if options.UniqueColumn == "" {
		return "", []interface{}{}, errors.Errorf("options.UniqueColumn is empty")
	}

	if !IsValidColumn(options.UniqueColumn) {
		return "", []interface{}{}, errors.Errorf("options.UniqueColumn contains invalid characters")
	}

	if options.SortColumn != "" && !IsValidColumn(options.SortColumn) {
		return "", []interface{}{}, errors.Errorf("options.SortColumn contains invalid characters")
	}

	if options.Limit > 1000 || options.Limit <= 0 {
		return "", []interface{}{}, errors.Errorf("options.Limit %v is invalid", options.Limit)
	}

	if options.SortDirection != Asc && options.SortDirection != Desc {
		return "", []interface{}{}, errors.Errorf("invalid sort direction")
	}

	// Process previous page cursor as next page in reverse order.
	sortDirection := options.SortDirection
	if cursor != nil && cursor.Direction == Previous {
		sortDirection ^= 1
	}
	query, args, err := buildQuery(options.SortColumn, options.UniqueColumn, sortDirection, options.Limit, cursor, query, args...)
	if err != nil {
		return "", []interface{}{}, err
	}

	return query, args, nil
}

func buildQuery(sortColumn, uniqueColumn string, sortDirection SortDirection, limit uint, cursor *PageCursor, query string, args ...interface{}) (string, []interface{}, error) {
	var cmp, order string
	if sortDirection == Asc {
		cmp = ">"
		order = "ASC"
	} else {
		cmp = "<"
		order = "DESC"
	}

	var whereClause, orderClause string
	var cursorLen int
	if sortColumn != "" && sortColumn != uniqueColumn {
		cursorLen = 2
		whereClause = fmt.Sprintf(" WHERE (`%s`, `%s`) %s (?, ?)", sortColumn, uniqueColumn, cmp)
		orderClause = fmt.Sprintf(" ORDER BY `%s` %s, `%s` %s", sortColumn, order, uniqueColumn, order)
	} else {
		cursorLen = 1
		whereClause = fmt.Sprintf(" WHERE `%s` %s ?", uniqueColumn, cmp)
		orderClause = fmt.Sprintf(" ORDER BY `%s` %s", uniqueColumn, order)
	}

	if cursor == nil {
		whereClause = ""
	} else {
		if len(cursor.Cursor) != cursorLen {
			return "", []interface{}{}, errors.Errorf("page token cursor does not match sort options")
		}
		args = append(args, cursor.Cursor...)
	}

	query = fmt.Sprintf("SELECT * FROM (%s) AS t%s%s LIMIT ?", query, whereClause, orderClause)

	// Query one more row to check if there is a next page.
	args = append(args, limit+1)

	return query, args, nil
}

func scanAll(pageOptions PageOptions, cursor *PageCursor, foundRows uint, rows *sqlx.Rows, dest interface{}) (*Page, error) {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr {
		return nil, errors.New("must pass a pointer, not a value, to StructScan destination")
	}
	if value.IsNil() {
		return nil, errors.New("nil pointer passed to StructScan destination")
	}
	direct := reflect.Indirect(value)
	slice := reflectx.Deref(value.Type())
	if slice.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice but got %s", slice.Kind())
	}
	isPtr := slice.Elem().Kind() == reflect.Ptr
	base := reflectx.Deref(slice.Elem())

	var err error
	var start, end []interface{}
	var items []reflect.Value
	hasNext := false
	for i := uint(0); rows.Next(); i++ {
		if i == 0 {
			start, err = getCursor(pageOptions, rows)
			if err != nil {
				return nil, err
			}
		}
		if i == pageOptions.Limit-1 {
			end, err = getCursor(pageOptions, rows)
			if err != nil {
				return nil, err
			}
		}
		if i == pageOptions.Limit {
			hasNext = true
			break
		}

		vp := reflect.New(base)
		err := rows.StructScan(vp.Interface())
		if err != nil {
			return nil, errors.Wrapf(err, "error scanning rows into dest")
		}
		items = append(items, vp)
	}

	var nextPageToken, prevPageToken PageToken

	// Copy the results back to dest. Reverse the results for previous page query.
	if cursor == nil || cursor.Direction == Next {
		for i := 0; i < len(items); i++ {
			if isPtr {
				direct.Set(reflect.Append(direct, items[i]))
			} else {
				direct.Set(reflect.Append(direct, reflect.Indirect(items[i])))
			}
		}
		if cursor != nil {
			prevPageToken = newPageToken(Previous, start)
		}
		if hasNext {
			nextPageToken = newPageToken(Next, end)
		}
	} else {
		for i := len(items) - 1; i >= 0; i-- {
			if isPtr {
				direct.Set(reflect.Append(direct, items[i]))
			} else {
				direct.Set(reflect.Append(direct, reflect.Indirect(items[i])))
			}
		}
		nextPageToken = newPageToken(Next, start)
		if hasNext {
			prevPageToken = newPageToken(Previous, end)
		}
	}

	return &Page{
		FoundRows:     foundRows,
		NextPageToken: nextPageToken,
		PreviousPageToken: prevPageToken,
	}, nil
}

func getCursor(pageOptions PageOptions, rows *sqlx.Rows) ([]interface{}, error) {
	var columns []string
	if pageOptions.SortColumn != "" && pageOptions.SortColumn != pageOptions.UniqueColumn {
		columns = append(columns, pageOptions.SortColumn)
	}
	columns = append(columns, pageOptions.UniqueColumn)

	return sliceColumns(rows, columns)
}

func sliceColumns(rows *sqlx.Rows, columns []string) ([]interface{}, error) {
	results := make(map[string]interface{})
	err := sqlx.MapScan(rows, results)
	if err != nil {
		return []interface{}{}, err
	}

	var slice []interface{}
	for _, column := range columns {
		if results[column] == nil {
			return []interface{}{}, errors.Errorf("column %v cannot be found in results", column)
		}
		slice = append(slice, results[column])
	}

	return slice, nil
}
