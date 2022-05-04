package main

import "testing"

func Test_process(t *testing.T) {
	type args struct {
		item string
	}
	tests := []struct {
		name string
		args args
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			process(tt.args.item)
		})
	}
}
