package vault

import "testing"

func Test_splitPath(t *testing.T) {
	type args struct {
		basepath string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"simple", args{basepath: "kv"}, "kv", "/"},
		{"simple", args{basepath: "/kv"}, "kv", "/"},
		{"simple", args{basepath: "/kv/"}, "kv", "/"},
		{"simple", args{basepath: "/kv/subdir"}, "kv", "subdir"},
		{"simple", args{basepath: "/kv/subdir/"}, "kv", "subdir"},
		{"simple", args{basepath: "/kv/subdir/subsubdir"}, "kv", "subdir/subsubdir"},
		{"simple", args{basepath: "/kv/subdir/subsubdir/"}, "kv", "subdir/subsubdir"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitPath(tt.args.basepath)
			if got != tt.want {
				t.Errorf("splitPath() got = '%v', want '%v'", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("splitPath() got1 = '%v', want '%v'", got1, tt.want1)
			}
		})
	}
}
