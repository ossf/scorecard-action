package main

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_filter(t *testing.T) {
	type args[T any] struct {
		slice []T
		f     func(T) bool
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[string]{
		{
			name: "default true",
			args: args[string]{
				slice: []string{"a"},
				f:     func(s string) bool { return s == "a" },
			},
			want: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filter(tt.args.slice, tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetScorecardChecks(t *testing.T) {
	tests := []struct {
		name        string
		want        []string
		fileContent string
		wantErr     bool
	}{
		{
			name:    "default",
			want:    []string{"Dangerous-Workflow", "Binary-Artifacts", "Branch-Protection", "Code-Review", "Dependency-Update-Tool"},
			wantErr: false,
		},
		{
			name: "file with data",
			want: []string{"Binary-Artifacts", "Pinned-Dependencies"},
			fileContent: `[
  "Binary-Artifacts",
  "Pinned-Dependencies"
]`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileContent != "" {
				dir, err := os.MkdirTemp("", "scorecard-checks")
				if err != nil {
					t.Errorf("GetScorecardChecks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				defer os.RemoveAll(dir)

				if err := os.WriteFile(path.Join(dir, "scorecard.txt"), []byte(tt.fileContent), 0644); err != nil {
					t.Errorf("GetScorecardChecks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				t.Setenv("SCORECARD_CHECKS", path.Join(dir, "scorecard.txt"))
			}
			got, err := GetScorecardChecks()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetScorecardChecks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetScorecardChecks() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetScore(t *testing.T) {
	type args struct {
		repo string
	}
	tests := []struct {
		name    string
		args    args
		score   float64
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				repo: "github.com/ossf/scorecard",
			},
			score:   5.0,
			wantErr: false,
		},
		{
			name: "invalid repo",
			args: args{
				repo: "github.com/ossf/invalid",
			},
			score:   0.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetScore(tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetScore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Score < tt.score {
				t.Errorf("GetScore() got = %v, want %v", got, tt.score)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	type args struct {
		token     string
		owner     string
		repo      string
		commitSHA string
		pr        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				token:     "token",
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: "commitSHA",
				pr:        "1",
			},
			wantErr: false,
		},
		{
			name: "invalid token",
			args: args{
				owner:     "ossf",
				repo:      "scorecard",
				commitSHA: "commitSHA",
				pr:        "1",
			},
			wantErr: true,
		},
		{
			name: "invalid owner",
			args: args{
				repo:      "scorecard",
				token:     "token",
				commitSHA: "commitSHA",
				pr:        "1",
			},
			wantErr: true,
		},
		{
			name: "invalid repo",
			args: args{
				owner:     "ossf",
				token:     "token",
				commitSHA: "commitSHA",
				pr:        "1",
			},
			wantErr: true,
		},
		{
			name: "invalid pr",
			args: args{
				owner:     "ossf",
				repo:      "scorecard",
				token:     "token",
				commitSHA: "commitSHA",
			},
			wantErr: true,
		},
		{
			name: "invalid commitSHA",
			args: args{
				owner: "ossf",
				repo:  "scorecard",
				token: "token",
				pr:    "1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.args.token, tt.args.owner, tt.args.repo, tt.args.commitSHA, tt.args.pr); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
