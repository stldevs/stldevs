package config

import "testing"

var table = []struct{ camel, snake string }{
	{"Forks", "forks"},
	{"ForksCount", "forks_count"},
}

func TestCamelToSnake(t *testing.T) {
	for _, v := range table {
		if actual := CamelToSnake(v.camel); actual != v.snake {
			t.Errorf("passed %v, expected %v, got %v", v.camel, v.snake, actual)
		}
	}
}
