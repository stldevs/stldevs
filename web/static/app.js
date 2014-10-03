var app = angular.module('MyApp', []);

app.controller('Main', ['$scope', '$http', function($scope, $http){
    $scope.users = $http.get('/users');
}]);
