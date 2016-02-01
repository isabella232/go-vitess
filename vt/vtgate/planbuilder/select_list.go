// Copyright 2016, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package planbuilder

import (
	"errors"

	"github.com/youtube/vitess/go/vt/sqlparser"
)

func processSelectExprs(sel *sqlparser.Select, planBuilder PlanBuilder, symbolTable *SymbolTable) error {
	allowAggregates := checkAllowAggregates(sel.SelectExprs, planBuilder, symbolTable)
	if sel.Distinct != "" {
		if !allowAggregates {
			return errors.New("query is too complex to allow aggregates")
		}
		// We know it's a RouteBuilder, but this may change
		// in the distant future.
		planBuilder.(*RouteBuilder).Select.Distinct = sel.Distinct
	}
	selectSymbols, err := findSelectRoutes(sel.SelectExprs, allowAggregates, symbolTable)
	if err != nil {
		return err
	}
	for i, selectExpr := range sel.SelectExprs {
		pushSelect(selectExpr, planBuilder, selectSymbols[i].Route.Order())
	}
	symbolTable.SelectSymbols = selectSymbols
	return nil
}

func findSelectRoutes(selectExprs sqlparser.SelectExprs, allowAggregates bool, symbolTable *SymbolTable) ([]SelectSymbol, error) {
	selectSymbols := make([]SelectSymbol, len(selectExprs))
	for colnum, selectExpr := range selectExprs {
		err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
			switch node := node.(type) {
			case *sqlparser.StarExpr:
				return false, errors.New("* expressions not allowed")
			case *sqlparser.Subquery:
				// TODO(sougou): implement.
				return false, errors.New("subqueries not supported yet")
			case *sqlparser.NonStarExpr:
				if node.As != "" {
					selectSymbols[colnum].Alias = node.As
				}
				col, ok := node.Expr.(*sqlparser.ColName)
				if ok {
					if selectSymbols[colnum].Alias == "" {
						selectSymbols[colnum].Alias = sqlparser.SQLName(sqlparser.String(col))
					}
					if _, cv := symbolTable.FindColumn(col, nil, true); cv != nil {
						selectSymbols[colnum].Vindex = cv.Vindex
					}
				}
			case *sqlparser.ColName:
				tableAlias, _ := symbolTable.FindColumn(node, nil, true)
				if tableAlias != nil {
					if selectSymbols[colnum].Route == nil {
						selectSymbols[colnum].Route = tableAlias.Route
					} else if selectSymbols[colnum].Route != tableAlias.Route {
						// TODO(sougou): better error.
						return false, errors.New("select expression is too complex")
					}
				}
			case *sqlparser.FuncExpr:
				if node.IsAggregate() {
					if !allowAggregates {
						// TODO(sougou): better error.
						return false, errors.New("query is too complex to allow aggregates")
					}
				}
			}
			return true, nil
		}, selectExpr)
		if err != nil {
			return nil, err
		}
		if selectSymbols[colnum].Route == nil {
			selectSymbols[colnum].Route = symbolTable.FirstRoute
		}
	}
	return selectSymbols, nil
}

func checkAllowAggregates(selectExprs sqlparser.SelectExprs, planBuilder PlanBuilder, symbolTable *SymbolTable) bool {
	routeBuilder, ok := planBuilder.(*RouteBuilder)
	if !ok {
		return false
	}
	if routeBuilder.Route.PlanID == SelectUnsharded || routeBuilder.Route.PlanID == SelectEqualUnique {
		return true
	}

	// It's a scatter route. We can allow aggregates if there is a unique
	// vindex in the select list.
	for _, selectExpr := range selectExprs {
		switch selectExpr := selectExpr.(type) {
		case *sqlparser.NonStarExpr:
			_, colVindex := symbolTable.FindColumn(selectExpr.Expr, nil, true)
			if colVindex != nil && IsUnique(colVindex.Vindex) {
				return true
			}
		}
	}
	return false
}

func pushSelect(selectExpr sqlparser.SelectExpr, planBuilder PlanBuilder, routeNumber int) {
	switch planBuilder := planBuilder.(type) {
	case *JoinBuilder:
		if routeNumber <= planBuilder.LeftOrder {
			pushSelect(selectExpr, planBuilder.Left, routeNumber)
			planBuilder.Join.LeftCols = append(planBuilder.Join.LeftCols, planBuilder.Join.Len())
			return
		}
		pushSelect(selectExpr, planBuilder.Right, routeNumber)
		planBuilder.Join.RightCols = append(planBuilder.Join.RightCols, planBuilder.Join.Len())
	case *RouteBuilder:
		if routeNumber != planBuilder.Order() {
			// TODO(sougou): remove after testing
			panic("unexpcted values")
		}
		planBuilder.Select.SelectExprs = append(planBuilder.Select.SelectExprs, selectExpr)
	}
}
