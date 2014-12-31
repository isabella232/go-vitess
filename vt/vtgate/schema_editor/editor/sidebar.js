/**
 * Copyright 2014, Google Inc. All rights reserved. Use of this source code is
 * governed by a BSD-style license that can be found in the LICENSE file.
 */
'use strict';

function SidebarController($scope, $routeParams, curSchema) {
  init();

  function init() {
    $scope.keyspaceEditor = {};
  }
  $scope.addKeyspace = function($keyspaceName, $sharded) {
    if ($keyspaceName in curSchema.keyspaces) {
      $scope.keyspaceEditor.err = $keyspaceName + " already exists";
      return;
    }
    AddKeyspace(curSchema.keyspaces, $keyspaceName, $sharded);
    $scope.clearKeyspaceError();
  };

  $scope.reset = function() {
    curSchema.reset();
    $scope.clearKeyspaceError();
  };

  $scope.clearKeyspaceError = function() {
    $scope.keyspaceEditor.err = "";
  };
}