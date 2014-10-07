'use strict';

var app = angular.module('MyApp', ['ngResource', 'ngRoute']);

app.config(['$routeProvider', function ($routeProvider){
	$routeProvider.when('/', {
		templateUrl: '/static/user-list.html',
		controller: 'UserList'
	}).when('/users/:id', {
		templateUrl: '/static/user.html',
		controller: 'User'
	}).otherwise({redirect: '/'})
}]);

app.factory('User', ['$resource', function ($resource){
	return $resource('/users/:id', {id: '@id'}, {
		get: {isArray: true}
	});
}]);

app.controller('UserList', ['$scope', 'User', function ($scope, User){
    $scope.users = User.query();
}]);

app.controller('User', ['$scope', '$routeParams', 'User', function ($scope, $routeParams, User) {
	$scope.routeParams = $routeParams;
    $scope.repos = User.get({id: $routeParams.id});
}]);
