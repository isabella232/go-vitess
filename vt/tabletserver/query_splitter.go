package tabletserver

import (
	"encoding/binary"
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
// QuerySplitter.splitBoundaries()
type QuerySplitter struct {
	query       *proto.BoundQuery
	splitCount  int
	schemaInfo  *SchemaInfo
	sel         *sqlparser.Select
	tableName   string
	splitColumn string
	rowCount    int64
}

const (
	startBindVarName = "_splitquery_start"
	endBindVarName   = "_splitquery_end"
)

// NewQuerySplitter creates a new QuerySplitter. query is the original query
// to split and splitCount is the desired number of splits. splitCount must
// be a positive int, if not it will be set to 1.
func NewQuerySplitter(
	query *proto.BoundQuery,
	splitColumn string,
	splitCount int,
	schemaInfo *SchemaInfo) *QuerySplitter {
	if splitCount < 1 {
		splitCount = 1
	}
	return &QuerySplitter{
		query:       query,
		splitCount:  splitCount,
		schemaInfo:  schemaInfo,
		splitColumn: splitColumn,
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
	if qs.splitColumn != "" {
		for _, index := range tableInfo.Indexes {
			for _, column := range index.Columns {
				if qs.splitColumn == column {
					return nil
				}
			}
		}
		return fmt.Errorf("split column is not indexed or does not exist in table schema, SplitColumn: %s, TableInfo.Table: %v", qs.splitColumn, tableInfo.Table)
	}
	qs.splitColumn = tableInfo.GetPKColumn(0).Name
	return nil
}

// split splits the query into multiple queries. validateQuery() must return
// nil error before split() is called.
func (qs *QuerySplitter) split(columnType int64, pkMinMax *mproto.QueryResult) ([]proto.QuerySplit, error) {
	boundaries, err := qs.splitBoundaries(columnType, pkMinMax)
	if err != nil {
		return nil, err
	}
	splits := []proto.QuerySplit{}
	// No splits, return the original query as a single split
	if len(boundaries) == 0 {
		split := &proto.QuerySplit{
			Query: *qs.query,
		}
		splits = append(splits, *split)
	} else {
		boundaries = append(boundaries, sqltypes.Value{})
		whereClause := qs.sel.Where
		// Loop through the boundaries and generated modified where clauses
		start := sqltypes.Value{}
		for _, end := range boundaries {
			bindVars := make(map[string]interface{}, len(qs.query.BindVariables))
			for k, v := range qs.query.BindVariables {
				bindVars[k] = v
			}
			qs.sel.Where = qs.getWhereClause(whereClause, bindVars, start, end)
			q := &proto.BoundQuery{
				Sql:           sqlparser.String(qs.sel),
				BindVariables: bindVars,
			}
			split := &proto.QuerySplit{
				Query:    *q,
				RowCount: qs.rowCount,
			}
			splits = append(splits, *split)
			start = end
		}
		qs.sel.Where = whereClause // reset where clause
	}
	return splits, err
}

// getWhereClause returns a whereClause based on desired upper and lower
// bounds for primary key.
func (qs *QuerySplitter) getWhereClause(whereClause *sqlparser.Where, bindVars map[string]interface{}, start, end sqltypes.Value) *sqlparser.Where {
	var startClause *sqlparser.ComparisonExpr
	var endClause *sqlparser.ComparisonExpr
	var clauses sqlparser.BoolExpr
	// No upper or lower bound, just return the where clause of original query
	if start.IsNull() && end.IsNull() {
		return whereClause
	}
	pk := &sqlparser.ColName{
		Name: sqlparser.SQLName(qs.splitColumn),
	}
	if !start.IsNull() {
		startClause = &sqlparser.ComparisonExpr{
			Operator: sqlparser.GreaterEqualStr,
			Left:     pk,
			Right:    sqlparser.ValArg([]byte(":" + startBindVarName)),
		}
		if start.IsNumeric() {
			v, _ := start.ParseInt64()
			bindVars[startBindVarName] = v
		} else if start.IsString() {
			bindVars[startBindVarName] = start.Raw()
		} else if start.IsFractional() {
			v, _ := start.ParseFloat64()
			bindVars[startBindVarName] = v
		}
	}
	// splitColumn < end
	if !end.IsNull() {
		endClause = &sqlparser.ComparisonExpr{
			Operator: sqlparser.LessThanStr,
			Left:     pk,
			Right:    sqlparser.ValArg([]byte(":" + endBindVarName)),
		}
		if end.IsNumeric() {
			v, _ := end.ParseInt64()
			bindVars[endBindVarName] = v
		} else if end.IsString() {
			bindVars[endBindVarName] = end.Raw()
		} else if end.IsFractional() {
			v, _ := end.ParseFloat64()
			bindVars[endBindVarName] = v
		}
	}
	if startClause == nil {
		clauses = endClause
	} else {
		if endClause == nil {
			clauses = startClause
		} else {
			// splitColumn >= start AND splitColumn < end
			clauses = &sqlparser.AndExpr{
				Left:  startClause,
				Right: endClause,
			}
		}
	}
	if whereClause != nil {
		clauses = &sqlparser.AndExpr{
			Left:  &sqlparser.ParenBoolExpr{Expr: whereClause.Expr},
			Right: &sqlparser.ParenBoolExpr{Expr: clauses},
		}
	}
	return &sqlparser.Where{
		Type: sqlparser.WhereStr,
		Expr: clauses,
	}
}

func (qs *QuerySplitter) splitBoundaries(columnType int64, pkMinMax *mproto.QueryResult) ([]sqltypes.Value, error) {
	switch columnType {
	case mproto.VT_TINY, mproto.VT_SHORT, mproto.VT_LONG, mproto.VT_LONGLONG, mproto.VT_INT24:
		return qs.splitBoundariesIntColumn(pkMinMax)
	case mproto.VT_FLOAT, mproto.VT_DOUBLE:
		return qs.splitBoundariesFloatColumn(pkMinMax)
	case mproto.VT_BIT, mproto.VT_VAR_STRING, mproto.VT_STRING:
		return qs.splitBoundariesStringColumn()
	}
	return []sqltypes.Value{}, nil
}

func (qs *QuerySplitter) splitBoundariesIntColumn(pkMinMax *mproto.QueryResult) ([]sqltypes.Value, error) {
	boundaries := []sqltypes.Value{}
	if pkMinMax == nil || len(pkMinMax.Rows) != 1 || pkMinMax.Rows[0][0].IsNull() || pkMinMax.Rows[0][1].IsNull() {
		return boundaries, nil
	}
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

func (qs *QuerySplitter) splitBoundariesFloatColumn(pkMinMax *mproto.QueryResult) ([]sqltypes.Value, error) {
	boundaries := []sqltypes.Value{}
	if pkMinMax == nil || len(pkMinMax.Rows) != 1 || pkMinMax.Rows[0][0].IsNull() || pkMinMax.Rows[0][1].IsNull() {
		return boundaries, nil
	}
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

// TODO(shengzhe): support split based on min, max from the string column.
func (qs *QuerySplitter) splitBoundariesStringColumn() ([]sqltypes.Value, error) {
	splitRange := int64(0xFFFFFFFF) + 1
	splitSize := splitRange / int64(qs.splitCount)
	//TODO(shengzhe): have a better estimated row count based on table size.
	qs.rowCount = int64(splitSize)
	var boundaries []sqltypes.Value
	for i := 1; i < qs.splitCount; i++ {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(splitSize)*uint32(i))
		val, err := sqltypes.BuildValue(buf)
		if err != nil {
			return nil, err
		}
		boundaries = append(boundaries, val)
	}
	return boundaries, nil
}
