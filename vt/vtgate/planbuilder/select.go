// Copyright 2016, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package planbuilder

import (
	"errors"

	"github.com/youtube/vitess/go/vt/sqlparser"
)

// buildSelectPlan is the new function to build a Select plan.
func buildSelectPlan(sel *sqlparser.Select, vschema *VSchema) (primitive Primitive, err error) {
	bindvars := getBindvars(sel)
	builder, err := processSelect(sel, vschema, nil)
	if err != nil {
		return nil, err
	}
	err = newGenerator(builder, bindvars).Generate()
	if err != nil {
		return nil, err
	}
	return getUnderlyingPlan(builder), nil
}

// getBindvars returns a map of the bind vars referenced in the statement.
func getBindvars(node sqlparser.SQLNode) map[string]struct{} {
	bindvars := make(map[string]struct{})
	_ = sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node := node.(type) {
		case sqlparser.ValArg:
			bindvars[string(node[1:])] = struct{}{}
		case sqlparser.ListArg:
			bindvars[string(node[2:])] = struct{}{}
		}
		return true, nil
	}, node)
	return bindvars
}

// processSelect builds a plan for the given query or subquery.
func processSelect(sel *sqlparser.Select, vschema *VSchema, outer planBuilder) (planBuilder, error) {
	plan, err := processTableExprs(sel.From, vschema)
	if err != nil {
		return nil, err
	}
	if outer != nil {
		plan.Symtab().Outer = outer.Symtab()
	}
	if sel.Where != nil {
		err = pushFilter(sel.Where.Expr, plan, sqlparser.WhereStr)
		if err != nil {
			return nil, err
		}
	}
	err = pushSelectExprs(sel, plan)
	if err != nil {
		return nil, err
	}
	if sel.Having != nil {
		err = pushFilter(sel.Having.Expr, plan, sqlparser.HavingStr)
		if err != nil {
			return nil, err
		}
	}
	err = pushOrderBy(sel.OrderBy, plan)
	if err != nil {
		return nil, err
	}
	err = pushLimit(sel.Limit, plan)
	if err != nil {
		return nil, err
	}
	pushMisc(sel, plan)
	return plan, nil
}

// pushFilter identifies the target route for the specified bool expr,
// pushes it down, and updates the route info if the new constraint improves
// the plan. This function can push to a WHERE or HAVING clause.
func pushFilter(boolExpr sqlparser.BoolExpr, plan planBuilder, whereType string) error {
	filters := splitAndExpression(nil, boolExpr)
	reorderBySubquery(filters)
	for _, filter := range filters {
		route, err := findRoute(filter, plan)
		if err != nil {
			return err
		}
		err = route.PushFilter(filter, whereType)
		if err != nil {
			return err
		}
	}
	return nil
}

// reorderBySubquery reorders the filters by pushing subqueries
// to the end. This allows the non-subquery filters to be
// pushed first because they can potentially improve the routing
// plan, which can later allow a filter containing a subquery
// to successfully merge with the corresponding route.
func reorderBySubquery(filters []sqlparser.BoolExpr) {
	max := len(filters)
	for i := 0; i < max; i++ {
		if !hasSubquery(filters[i]) {
			continue
		}
		saved := filters[i]
		for j := i; j < len(filters)-1; j++ {
			filters[j] = filters[j+1]
		}
		filters[len(filters)-1] = saved
		max--
	}
}

// pushSelectExprs identifies the target route for the
// select expressions and pushes them down.
func pushSelectExprs(sel *sqlparser.Select, plan planBuilder) error {
	err := checkAggregates(sel, plan)
	if err != nil {
		return err
	}
	if sel.Distinct != "" {
		// We know it's a routeBuilder, but this may change
		// in the distant future.
		plan.(*routeBuilder).MakeDistinct()
	}
	colsyms, err := pushSelectRoutes(sel.SelectExprs, plan)
	if err != nil {
		return err
	}
	plan.Symtab().Colsyms = colsyms
	err = pushGroupBy(sel.GroupBy, plan)
	if err != nil {
		return err
	}
	return nil
}

// checkAggregates returns an error if the select statement
// has aggregates that cannot be pushed down due to a complex
// plan.
func checkAggregates(sel *sqlparser.Select, plan planBuilder) error {
	hasAggregates := false
	if sel.Distinct != "" {
		hasAggregates = true
	} else {
		_ = sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
			switch node := node.(type) {
			case *sqlparser.FuncExpr:
				if node.IsAggregate() {
					hasAggregates = true
					return false, errors.New("dummy")
				}
			}
			return true, nil
		}, sel.SelectExprs)
	}
	if !hasAggregates {
		return nil
	}

	// Check if we can allow aggregates.
	route, ok := plan.(*routeBuilder)
	if !ok {
		return errors.New("unsupported: complex join with aggregates")
	}
	if route.IsSingle() {
		return nil
	}
	// It's a scatter route. We can allow aggregates if there is a unique
	// vindex in the select list.
	for _, selectExpr := range sel.SelectExprs {
		switch selectExpr := selectExpr.(type) {
		case *sqlparser.NonStarExpr:
			vindex := plan.Symtab().Vindex(selectExpr.Expr, route, true)
			if vindex != nil && IsUnique(vindex) {
				return nil
			}
		}
	}
	return errors.New("unsupported: scatter with aggregates")
}

// pusheSelectRoutes is a convenience function that pushes all the select
// expressions and returns the list of colsyms generated for it.
func pushSelectRoutes(selectExprs sqlparser.SelectExprs, plan planBuilder) ([]*colsym, error) {
	colsyms := make([]*colsym, len(selectExprs))
	for i, node := range selectExprs {
		switch node := node.(type) {
		case *sqlparser.NonStarExpr:
			route, err := findRoute(node.Expr, plan)
			if err != nil {
				return nil, err
			}
			colsyms[i], _, err = plan.PushSelect(node, route)
			if err != nil {
				return nil, err
			}
		case *sqlparser.StarExpr:
			// We'll allow select * for simple routes.
			route, ok := plan.(*routeBuilder)
			if !ok {
				return nil, errors.New("unsupported: '*' expression in complex join")
			}
			// We can push without validating the reference because
			// MySQL will fail if it's invalid.
			colsyms[i] = route.PushStar(node)
		case *sqlparser.Nextval:
			// For now, this is only supported as an implicit feature
			// for auto_inc in inserts.
			return nil, errors.New("unsupported: NEXTVAL construct")
		}
	}
	return colsyms, nil
}
