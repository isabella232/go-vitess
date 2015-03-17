package tabletserver

import (
	"fmt"
	"strconv"

	mproto "github.com/youtube/vitess/go/mysql/proto"
	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/sqlparser"
	"github.com/youtube/vitess/go/vt/tabletserver/proto"
)

// QuerySplitter splits a BoundQuery into equally sized smaller queries.
// QuerySplits are generated by adding primary key range clauses to the
// original query. Only a limited set of queries are supported, see
// QuerySplitter.validateQuery() for details. Also, the table must have at least
// one primary key and the leading primary key must be numeric, see
// QuerySplitter.getSplitBoundaries()
type QuerySplitter struct {
	query      *proto.BoundQuery
	splitCount int
	schemaInfo *SchemaInfo
	sel        *sqlparser.Select
	tableName  string
	pkCol      string
	rowCount   int64
}

// NewQuerySplitter creates a new QuerySplitter. query is the original query
// to split and splitCount is the desired number of splits. splitCount must
// be a positive int, if not it will be set to 1.
func NewQuerySplitter(query *proto.BoundQuery, splitCount int, schemaInfo *SchemaInfo) *QuerySplitter {
	if splitCount < 1 {
		splitCount = 1
	}
	return &QuerySplitter{
		query:      query,
		splitCount: splitCount,
		schemaInfo: schemaInfo,
	}
}

// Ensure that the input query is a Select statement that contains no Join,
// GroupBy, OrderBy, Limit or Distinct operations. Also ensure that the
// source table is present in the schema and has at least one primary key.
func (qs *QuerySplitter) validateQuery() error {
	statement, err := sqlparser.Parse(qs.query.Sql)
	if err != nil {
		return err
	}
	var ok bool
	qs.sel, ok = statement.(*sqlparser.Select)
	if !ok {
		return fmt.Errorf("not a select statement")
	}
	if qs.sel.Distinct != "" || qs.sel.GroupBy != nil ||
		qs.sel.Having != nil || len(qs.sel.From) != 1 ||
		qs.sel.OrderBy != nil || qs.sel.Limit != nil ||
		qs.sel.Lock != "" {
		return fmt.Errorf("unsupported query")
	}
	node, ok := qs.sel.From[0].(*sqlparser.AliasedTableExpr)
	if !ok {
		return fmt.Errorf("unsupported query")
	}
	qs.tableName = sqlparser.GetTableName(node.Expr)
	if qs.tableName == "" {
		return fmt.Errorf("not a simple table expression")
	}
	tableInfo, ok := qs.schemaInfo.tables[qs.tableName]
	if !ok {
		return fmt.Errorf("can't find table in schema")
	}
	if len(tableInfo.PKColumns) == 0 {
		return fmt.Errorf("no primary keys")
	}
	qs.pkCol = tableInfo.GetPKColumn(0).Name
	return nil
}

// split splits the query into multiple queries. validateQuery() must return
// nil error before split() is called.
func (qs *QuerySplitter) split(pkMinMax *mproto.QueryResult) ([]proto.QuerySplit, error) {
	var err error
	splits := []proto.QuerySplit{}
	boundaries, err := qs.getSplitBoundaries(pkMinMax)
	if err != nil {
		return splits, err
	}
	// No splits, return the original query as a single split
	if len(boundaries) == 0 {
		split := &proto.QuerySplit{
			Query: *qs.query,
		}
		splits = append(splits, *split)
	} else {
		// Loop through the boundaries and generated modified where clauses
		start := sqltypes.Value{}
		clauses := []*sqlparser.Where{}
		for _, end := range boundaries {
			clauses = append(clauses, qs.getWhereClause(start, end))
			start.Inner = end.Inner
		}
		clauses = append(clauses, qs.getWhereClause(start, sqltypes.Value{}))
		// Generate one split per clause
		for _, clause := range clauses {
			sel := qs.sel
			sel.Where = clause
			q := &proto.BoundQuery{
				Sql:           sqlparser.String(sel),
				BindVariables: qs.query.BindVariables,
			}
			split := &proto.QuerySplit{
				Query:    *q,
				RowCount: qs.rowCount,
			}
			splits = append(splits, *split)
		}
	}
	return splits, err
}

// getWhereClause returns a whereClause based on desired upper and lower
// bounds for primary key.
func (qs *QuerySplitter) getWhereClause(start, end sqltypes.Value) *sqlparser.Where {
	var startClause *sqlparser.ComparisonExpr
	var endClause *sqlparser.ComparisonExpr
	var clauses sqlparser.BoolExpr
	// No upper or lower bound, just return the where clause of original query
	if start.IsNull() && end.IsNull() {
		return qs.sel.Where
	}
	pk := &sqlparser.ColName{
		Name: []byte(qs.pkCol),
	}
	// pkCol >= start
	if !start.IsNull() {
		startClause = &sqlparser.ComparisonExpr{
			Operator: sqlparser.AST_GE,
			Left:     pk,
			Right:    sqlparser.NumVal((start).Raw()),
		}
	}
	// pkCol < end
	if !end.IsNull() {
		endClause = &sqlparser.ComparisonExpr{
			Operator: sqlparser.AST_LT,
			Left:     pk,
			Right:    sqlparser.NumVal((end).Raw()),
		}
	}
	if startClause == nil {
		clauses = endClause
	} else {
		if endClause == nil {
			clauses = startClause
		} else {
			// pkCol >= start AND pkCol < end
			clauses = &sqlparser.AndExpr{
				Left:  startClause,
				Right: endClause,
			}
		}
	}
	if qs.sel.Where != nil {
		clauses = &sqlparser.AndExpr{
			Left:  qs.sel.Where.Expr,
			Right: clauses,
		}
	}
	return &sqlparser.Where{
		Type: sqlparser.AST_WHERE,
		Expr: clauses,
	}
}

func (qs *QuerySplitter) getSplitBoundaries(pkMinMax *mproto.QueryResult) ([]sqltypes.Value, error) {
	boundaries := []sqltypes.Value{}
	var err error
	// If no min or max values were found, return empty list of boundaries
	if len(pkMinMax.Rows) != 1 || pkMinMax.Rows[0][0].IsNull() || pkMinMax.Rows[0][1].IsNull() {
		return boundaries, err
	}
	switch pkMinMax.Fields[0].Type {
	case mproto.VT_TINY, mproto.VT_SHORT, mproto.VT_LONG, mproto.VT_LONGLONG, mproto.VT_INT24:
		boundaries, err = qs.parseInt(pkMinMax)
	case mproto.VT_FLOAT, mproto.VT_DOUBLE:
		boundaries, err = qs.parseFloat(pkMinMax)
	}
	return boundaries, err
}

func (qs *QuerySplitter) parseInt(pkMinMax *mproto.QueryResult) ([]sqltypes.Value, error) {
	boundaries := []sqltypes.Value{}
	minNumeric := sqltypes.MakeNumeric(pkMinMax.Rows[0][0].Raw())
	maxNumeric := sqltypes.MakeNumeric(pkMinMax.Rows[0][1].Raw())
	if pkMinMax.Rows[0][0].Raw()[0] == '-' {
		// signed values, use int64
		min, err := minNumeric.ParseInt64()
		if err != nil {
			return nil, err
		}
		max, err := maxNumeric.ParseInt64()
		if err != nil {
			return nil, err
		}
		interval := (max - min) / int64(qs.splitCount)
		if interval == 0 {
			return nil, err
		}
		qs.rowCount = interval
		for i := int64(1); i < int64(qs.splitCount); i++ {
			v, err := sqltypes.BuildValue(min + interval*i)
			if err != nil {
				return nil, err
			}
			boundaries = append(boundaries, v)
		}
		return boundaries, nil
	}
	// unsigned values, use uint64
	min, err := minNumeric.ParseUint64()
	if err != nil {
		return nil, err
	}
	max, err := maxNumeric.ParseUint64()
	if err != nil {
		return nil, err
	}
	interval := (max - min) / uint64(qs.splitCount)
	if interval == 0 {
		return nil, err
	}
	qs.rowCount = int64(interval)
	for i := uint64(1); i < uint64(qs.splitCount); i++ {
		v, err := sqltypes.BuildValue(min + interval*i)
		if err != nil {
			return nil, err
		}
		boundaries = append(boundaries, v)
	}
	return boundaries, nil
}

func (qs *QuerySplitter) parseFloat(pkMinMax *mproto.QueryResult) ([]sqltypes.Value, error) {
	boundaries := []sqltypes.Value{}
	min, err := strconv.ParseFloat(pkMinMax.Rows[0][0].String(), 64)
	if err != nil {
		return nil, err
	}
	max, err := strconv.ParseFloat(pkMinMax.Rows[0][1].String(), 64)
	if err != nil {
		return nil, err
	}
	interval := (max - min) / float64(qs.splitCount)
	if interval == 0 {
		return nil, err
	}
	qs.rowCount = int64(interval)
	for i := 1; i < qs.splitCount; i++ {
		boundary := min + interval*float64(i)
		v, err := sqltypes.BuildValue(boundary)
		if err != nil {
			return nil, err
		}
		boundaries = append(boundaries, v)
	}
	return boundaries, nil
}
