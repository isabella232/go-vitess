// Copyright 2016, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package planbuilder

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/youtube/vitess/go/vt/sqlparser"
)

// This file contains routines for processing the FROM
// clause. Functions in this file manipulate various data
// structures. If they return an error, one should assume
// that the data structures may be in an inconsistent state.
// In general, the error should just be returned back to the
// application.

// planBuilder represents any object that's used to
// build a plan. The top-level planBuilder will be a
// tree that points to other planBuilder objects.
// Currently, joinBuilder and routeBuilder are the
// only two supported planBuilder objects. More will be
// added as we extend the functionality.
// Each Builder object builds a Plan object, and they
// will mirror the same tree. Once all the plans are built,
// the builder objects will be discarded, and only
// the Plan objects will remain.
type planBuilder interface {
	// Order is a number that signifies execution order.
	// A lower Order number Route is executed before a
	// higher one. For a node that contains other nodes,
	// the Order represents the highest order of the leaf
	// nodes.
	Order() int
}

// joinBuilder is used to build a Join primitive.
// It's used to buid a normal join or a left join
// operation.
// TODO(sougou): struct is incomplete.
type joinBuilder struct {
	LeftOrder, RightOrder int
	// Left and Right are the nodes for the join.
	Left, Right planBuilder
	// Join is the join plan.
	Join *Join
}

// Order returns the order of the node.
func (jb *joinBuilder) Order() int {
	return jb.RightOrder
}

// MarshalJSON marshals joinBuilder into a readable form.
// It's used for testing and diagnostics. The representation
// cannot be used to reconstruct a joinBuilder.
func (jb *joinBuilder) MarshalJSON() ([]byte, error) {
	marshalJoin := struct {
		LeftOrder   int
		RightOrder  int
		Left, Right planBuilder
		Join        *Join
	}{
		LeftOrder:  jb.LeftOrder,
		RightOrder: jb.RightOrder,
		Left:       jb.Left,
		Right:      jb.Right,
		Join: &Join{
			IsLeft:    jb.Join.IsLeft,
			LeftCols:  jb.Join.LeftCols,
			RightCols: jb.Join.RightCols,
		},
	}
	return json.Marshal(marshalJoin)
}

// Join is the join plan.
type Join struct {
	IsLeft              bool        `json:",omitempty"`
	Left, Right         interface{} `json:",omitempty"`
	LeftCols, RightCols []int       `json:",omitempty"`
}

// Len returns the number of columns in the join
func (jn *Join) Len() int {
	return len(jn.LeftCols) + len(jn.RightCols)
}

// routeBuilder is used to build a Route primitive.
// It's used to build one of the Select routes like
// SelectScatter, etc. Portions of the original Select AST
// are moved into this node, which will be used to build
// the final SQL for this route.
// TODO(sougou): struct is incomplete.
type routeBuilder struct {
	// IsRHS is true if the routeBuilder is the RHS of a
	// LEFT JOIN. If so, many restrictions come into play.
	IsRHS bool
	// Select is the AST for the query fragment that will be
	// executed by this route.
	Select sqlparser.Select
	order  int
	Colsym []*colsym
	// Route is the plan object being built. It will contain all the
	// information necessary to execute the route operation.
	Route *Route
}

// Order returns the order of the node.
func (rtb *routeBuilder) Order() int {
	return rtb.order
}

func (rtb *routeBuilder) SupplyJoinVar(col *sqlparser.ColName, varname string) {
	switch meta := col.Metadata.(type) {
	case *colsym:
		for i, colsym := range rtb.Colsym {
			if meta == colsym {
				rtb.Route.SupplyVars[varname] = i
				return
			}
		}
		panic("unexpected")
	case *tableAlias:
		for i, colsym := range rtb.Colsym {
			if colsym.Underlying != nil {
				if colsym.Underlying.Metadata == col.Metadata && colsym.Underlying.Name == col.Name {
					rtb.Route.SupplyVars[varname] = i
					return
				}
			}
		}
		rtb.Route.SupplyVars[varname] = len(rtb.Colsym)
		rtb.Colsym = append(rtb.Colsym, &colsym{
			Alias:      sqlparser.SQLName(sqlparser.String(col)),
			Underlying: col,
		})
		rtb.Select.SelectExprs = append(
			rtb.Select.SelectExprs,
			&sqlparser.NonStarExpr{
				Expr: &sqlparser.ColName{
					Metadata:  col.Metadata,
					Qualifier: meta.Alias,
					Name:      col.Name,
				},
			},
		)
	}
}

// MarshalJSON marshals routeBuilder into a readable form.
// It's used for testing and diagnostics. The representation
// cannot be used to reconstruct a routeBuilder.
func (rtb *routeBuilder) MarshalJSON() ([]byte, error) {
	marshalRoute := struct {
		IsRHS  bool   `json:",omitempty"`
		Select string `json:",omitempty"`
		Order  int
		Route  *Route
	}{
		IsRHS:  rtb.IsRHS,
		Select: sqlparser.String(&rtb.Select),
		Order:  rtb.order,
		Route:  rtb.Route,
	}
	return json.Marshal(marshalRoute)
}

// IsSingle returns true if the route targets only one database.
func (rtb *routeBuilder) IsSingle() bool {
	return rtb.Route.PlanID == SelectUnsharded || rtb.Route.PlanID == SelectEqualUnique
}

// Route is a Plan object that represents a route.
// It can be any one of the Select primitives from PlanID.
// Some plan ids correspond to a multi-shard query,
// and some are for a single-shard query. The rules
// of what can be merged, or what can be pushed down
// depend on the PlanID. They're explained in code
// where such decisions are made.
// TODO(sougou): struct is incomplete.
// TODO(sougou): integrate with the older v3 Plan.
type Route struct {
	// PlanID will be one of the Select IDs from PlanID.
	PlanID PlanID
	Query  string
	// Keypsace represents the keyspace to which
	// the query will be sent.
	Keyspace *Keyspace
	// Vindex represents the vindex that will be used
	// to resolve the route.
	Vindex Vindex
	// Values can be a single value or a list of
	// values that will be used as input to the Vindex
	// to compute the target shard(s) where the query must
	// be sent.
	// TODO(sougou): explain contents of Values.
	Values     interface{}
	UseVars    map[string]struct{}
	SupplyVars map[string]int
}

// MarshalJSON marshals Route into a readable form.
// It's used for testing and diagnostics. The representation
// cannot be used to reconstruct a Route.
func (rt *Route) MarshalJSON() ([]byte, error) {
	var vindexName string
	if rt.Vindex != nil {
		vindexName = rt.Vindex.String()
	}
	marshalRoute := struct {
		PlanID     PlanID              `json:",omitempty"`
		Query      string              `json:",omitempty"`
		Keyspace   *Keyspace           `json:",omitempty"`
		Vindex     string              `json:",omitempty"`
		Values     string              `json:",omitempty"`
		UseVars    map[string]struct{} `json:",omitempty"`
		SupplyVars map[string]int      `json:",omitempty"`
	}{
		PlanID:     rt.PlanID,
		Query:      rt.Query,
		Keyspace:   rt.Keyspace,
		Vindex:     vindexName,
		Values:     prettyValue(rt.Values),
		UseVars:    rt.UseVars,
		SupplyVars: rt.SupplyVars,
	}
	return json.Marshal(marshalRoute)
}

// SetPlan updates the plan info for the route.
func (rt *Route) SetPlan(planID PlanID, vindex Vindex, values interface{}) {
	rt.PlanID = planID
	rt.Vindex = vindex
	rt.Values = values
}

// prettyValue converts the Values to a readable form.
// This function is used for testing and diagnostics.
func prettyValue(value interface{}) string {
	switch value := value.(type) {
	case nil:
		return ""
	case sqlparser.SQLNode:
		return sqlparser.String(value)
	case []byte:
		return string(value)
	}
	return fmt.Sprintf("%v", value)
}

// buildSelectPlan2 is the new function to build a Select plan.
// TODO(sougou): rename after deprecating old one.
func buildSelectPlan2(sel *sqlparser.Select, schema *Schema) (plan interface{}, err error) {
	builder, _, err := processSelect(sel, schema, nil)
	if err != nil {
		return nil, err
	}
	newGenerator().Generate(builder)
	if err != nil {
		return nil, err
	}
	return getUnderlyingPlan(builder), nil
}

// processQuery builds a plan for the given query or subquery.
func processSelect(sel *sqlparser.Select, schema *Schema, outer *symtab) (planBuilder, *symtab, error) {
	plan, syms, err := processTableExprs(sel.From, schema)
	if err != nil {
		return nil, nil, err
	}
	syms.Outer = outer
	if sel.Where != nil {
		err = processBoolExpr(sel.Where.Expr, syms, sqlparser.WhereStr)
		if err != nil {
			return nil, nil, err
		}
	}
	err = processSelectExprs(sel, plan, syms)
	if err != nil {
		return nil, nil, err
	}
	if sel.Having != nil {
		err = processBoolExpr(sel.Having.Expr, syms, sqlparser.HavingStr)
		if err != nil {
			return nil, nil, err
		}
	}
	err = processOrderBy(sel.OrderBy, syms)
	if err != nil {
		return nil, nil, err
	}
	err = processLimit(sel.Limit, plan)
	if err != nil {
		return nil, nil, err
	}
	processMisc(sel, plan)
	return plan, syms, nil
}

// processTableExprs analyzes the FROM clause. It produces a planBuilder
// and the associated symtab with all the routes identified.
func processTableExprs(tableExprs sqlparser.TableExprs, schema *Schema) (planBuilder, *symtab, error) {
	if len(tableExprs) != 1 {
		// TODO(sougou): better error message.
		return nil, nil, errors.New("lists are not supported")
	}
	return processTableExpr(tableExprs[0], schema)
}

// processTableExpr produces a planBuilder subtree and symtab
// for the given TableExpr.
func processTableExpr(tableExpr sqlparser.TableExpr, schema *Schema) (planBuilder, *symtab, error) {
	switch tableExpr := tableExpr.(type) {
	case *sqlparser.AliasedTableExpr:
		return processAliasedTable(tableExpr, schema)
	case *sqlparser.ParenTableExpr:
		plan, syms, err := processTableExprs(tableExpr.Exprs, schema)
		// We want to point to the higher level parenthesis because
		// more routes can be merged with this one. If so, the order
		// should be maintained as dictated by the parenthesis.
		if route, ok := plan.(*routeBuilder); ok {
			route.Select.From = sqlparser.TableExprs{tableExpr}
		}
		return plan, syms, err
	case *sqlparser.JoinTableExpr:
		return processJoin(tableExpr, schema)
	}
	panic("unreachable")
}

// processAliasedTable produces a planBuilder subtree and symtab
// for the given AliasedTableExpr.
func processAliasedTable(tableExpr *sqlparser.AliasedTableExpr, schema *Schema) (planBuilder, *symtab, error) {
	switch expr := tableExpr.Expr.(type) {
	case *sqlparser.TableName:
		route, table, err := getTablePlan(expr, schema)
		if err != nil {
			return nil, nil, err
		}
		plan := &routeBuilder{
			Select: sqlparser.Select{From: sqlparser.TableExprs{tableExpr}},
			order:  1,
			Route:  route,
		}
		alias := expr.Name
		if tableExpr.As != "" {
			alias = tableExpr.As
		}
		syms := newSymtab(alias, table, plan, schema)
		return plan, syms, nil
	case *sqlparser.Subquery:
		// TODO(sougou): implement.
		return nil, nil, errors.New("no subqueries")
	}
	panic("unreachable")
}

// getTablePlan produces the initial Route for the specified TableName.
// It also returns the associated vschema info (*Table) so that
// it can be used to create the symbol table entry.
func getTablePlan(tableName *sqlparser.TableName, schema *Schema) (*Route, *Table, error) {
	if tableName.Qualifier != "" {
		// TODO(sougou): better error message.
		return nil, nil, errors.New("tablename qualifier not allowed")
	}
	table, reason := schema.FindTable(string(tableName.Name))
	if reason != "" {
		return nil, nil, errors.New(reason)
	}
	if table.Keyspace.Sharded {
		return &Route{
			PlanID:     SelectScatter,
			Keyspace:   table.Keyspace,
			UseVars:    make(map[string]struct{}),
			SupplyVars: make(map[string]int),
		}, table, nil
	}
	return &Route{
		PlanID:     SelectUnsharded,
		Keyspace:   table.Keyspace,
		UseVars:    make(map[string]struct{}),
		SupplyVars: make(map[string]int),
	}, table, nil
}

// processJoin produces a planBuilder subtree and symtab
// for the given Join. If the left and right nodes can be part
// of the same route, then it's a routeBuilder. Otherwise,
// it's a joinBuilder.
func processJoin(join *sqlparser.JoinTableExpr, schema *Schema) (planBuilder, *symtab, error) {
	switch join.Join {
	case sqlparser.JoinStr, sqlparser.StraightJoinStr, sqlparser.LeftJoinStr:
	default:
		// TODO(sougou): better error message.
		return nil, nil, errors.New("unsupported join")
	}
	lplan, lsyms, err := processTableExpr(join.LeftExpr, schema)
	if err != nil {
		return nil, nil, err
	}
	rplan, rsyms, err := processTableExpr(join.RightExpr, schema)
	if err != nil {
		return nil, nil, err
	}
	switch lplan := lplan.(type) {
	case *joinBuilder:
		return makejoinBuilder(lplan, lsyms, rplan, rsyms, join)
	case *routeBuilder:
		switch rplan := rplan.(type) {
		case *joinBuilder:
			return makejoinBuilder(lplan, lsyms, rplan, rsyms, join)
		case *routeBuilder:
			return joinRoutes(lplan, lsyms, rplan, rsyms, join)
		}
	}
	panic("unreachable")
}

// makejoinBuilder creates a new joinBuilder node out of the two builders.
// This function is called when the two builders cannot be part of
// the same route.
func makejoinBuilder(lplan planBuilder, lsyms *symtab, rplan planBuilder, rsyms *symtab, join *sqlparser.JoinTableExpr) (planBuilder, *symtab, error) {
	// This function converts ON clauses to WHERE clauses. The WHERE clause
	// scope can see all tables, whereas the ON clause can only see the
	// participants of the JOIN. However, since the ON clause doesn't allow
	// external references, and the FROM clause doesn't allow duplicates,
	// it's safe to perform this conversion and still expect the same behavior.

	// For LEFT JOIN, you have to push the ON clause into the RHS first.
	isLeft := false
	if join.Join == sqlparser.LeftJoinStr {
		isLeft = true
		err := processBoolExpr(join.On, rsyms, sqlparser.WhereStr)
		if err != nil {
			return nil, nil, err
		}
		setRHS(rplan)
	}

	err := lsyms.Add(rsyms)
	if err != nil {
		return nil, nil, err
	}
	assignOrder(rplan, lplan.Order())
	// For normal joins, the ON clause can go to both sides.
	// The push has to happen after the order is assigned.
	if !isLeft {
		err := processBoolExpr(join.On, lsyms, sqlparser.WhereStr)
		if err != nil {
			return nil, nil, err
		}
	}
	return &joinBuilder{
		LeftOrder:  lplan.Order(),
		RightOrder: rplan.Order(),
		Left:       lplan,
		Right:      rplan,
		Join: &Join{
			IsLeft: isLeft,
			Left:   getUnderlyingPlan(lplan),
			Right:  getUnderlyingPlan(rplan),
		},
	}, lsyms, nil
}

func getUnderlyingPlan(plan planBuilder) interface{} {
	switch plan := plan.(type) {
	case *joinBuilder:
		return plan.Join
	case *routeBuilder:
		return plan.Route
	}
	panic("unreachable")
}

// assignOrder sets the order for the nodes of the tree based on the
// starting order.
func assignOrder(plan planBuilder, order int) {
	switch plan := plan.(type) {
	case *joinBuilder:
		assignOrder(plan.Left, order)
		plan.LeftOrder = plan.Left.Order()
		assignOrder(plan.Right, plan.Left.Order())
		plan.RightOrder = plan.Right.Order()
	case *routeBuilder:
		plan.order = order + 1
	}
}

// setRHS marks all routes under the plan as RHS of a left join.
func setRHS(plan planBuilder) {
	switch plan := plan.(type) {
	case *joinBuilder:
		setRHS(plan.Left)
		setRHS(plan.Right)
	case *routeBuilder:
		plan.IsRHS = true
	}
}

// joinRoutes attempts to join two routeBuilder objects into one.
// If it's possible, it produces a joined routeBuilder.
// Otherwise, it's a joinBuilder.
func joinRoutes(lRoute *routeBuilder, lsyms *symtab, rRoute *routeBuilder, rsyms *symtab, join *sqlparser.JoinTableExpr) (planBuilder, *symtab, error) {
	if lRoute.Route.Keyspace.Name != rRoute.Route.Keyspace.Name {
		return makejoinBuilder(lRoute, lsyms, rRoute, rsyms, join)
	}
	if lRoute.Route.PlanID == SelectUnsharded {
		// Two Routes from the same unsharded keyspace can be merged.
		return mergeRoutes(lRoute, lsyms, rRoute, rsyms, join)
	}

	// TODO(sougou): Handle special case for SelectEqual

	// Both routeBuilder are sharded routes. Analyze join condition for merging.
	for _, filter := range splitAndExpression(nil, join.On) {
		if isSameRoute(lRoute, lsyms, rRoute, rsyms, filter) {
			return mergeRoutes(lRoute, lsyms, rRoute, rsyms, join)
		}
	}
	return makejoinBuilder(lRoute, lsyms, rRoute, rsyms, join)
}

// mergeRoutes makes a new routeBuilder by joining the left and right
// nodes of a join. The merged routeBuilder inherits the plan of the
// left Route. This function is called if two routes can be merged.
func mergeRoutes(lRoute *routeBuilder, lsyms *symtab, rRoute *routeBuilder, rsyms *symtab, join *sqlparser.JoinTableExpr) (planBuilder, *symtab, error) {
	lRoute.Select.From = sqlparser.TableExprs{join}
	if join.Join == sqlparser.LeftJoinStr {
		rsyms.SetRHS()
	}
	err := lsyms.Merge(rsyms, lRoute)
	if err != nil {
		return nil, nil, err
	}
	for _, filter := range splitAndExpression(nil, join.On) {
		// If VTGate evolves, this section should be rewritten
		// to use processBoolExpr.
		_, err = findRoute(filter, lsyms)
		if err != nil {
			return nil, nil, err
		}
		updateRoute(lRoute, lsyms, filter)
	}
	return lRoute, lsyms, nil
}

// isSameRoute returns true if the filter constraint causes the
// left and right routes to be part of the same route. For this
// to happen, the constraint has to be an equality like a.id = b.id,
// one should address a table from the left side, the other from the
// right, the referenced columns have to be the same Vindex, and the
// Vindex must be unique.
func isSameRoute(lRoute *routeBuilder, lsyms *symtab, rRoute *routeBuilder, rsyms *symtab, filter sqlparser.BoolExpr) bool {
	comparison, ok := filter.(*sqlparser.ComparisonExpr)
	if !ok {
		return false
	}
	if comparison.Operator != sqlparser.EqualStr {
		return false
	}
	left := comparison.Left
	right := comparison.Right
	lVindex := lsyms.Vindex(left, lRoute, false)
	if lVindex == nil {
		left, right = right, left
		lVindex = lsyms.Vindex(left, lRoute, false)
	}
	if lVindex == nil || !IsUnique(lVindex) {
		return false
	}
	rVindex := rsyms.Vindex(right, rRoute, false)
	if rVindex == nil {
		return false
	}
	if rVindex != lVindex {
		return false
	}
	return true
}
