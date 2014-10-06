var app = angular.module('MyApp', ['ngResource']);

app.factory('User', ['$resource', function($resource){
	return $resource('/users/:id', {id: '@id'});
}]);

app.controller('Main', ['$scope', 'User', function($scope, User){
    $scope.users = User.query();
}]);
