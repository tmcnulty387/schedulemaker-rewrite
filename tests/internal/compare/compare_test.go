package compare

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		name      string
		want, got string
		wantEqual bool
	}{
		{"identical object", `{"a":1,"b":"x"}`, `{"a":1,"b":"x"}`, true},
		{"key order ignored", `{"a":1,"b":2}`, `{"b":2,"a":1}`, true},
		{"whitespace ignored", `{"a":1}`, "{\n  \"a\":   1\n}", true},
		{"array order ignored", `[1,2,3]`, `[3,1,2]`, true},
		{"array of objects reordered", `[{"id":1},{"id":2}]`, `[{"id":2},{"id":1}]`, true},
		{"nested array order ignored", `{"xs":[{"k":[1,2]},{"k":[3]}]}`, `{"xs":[{"k":[3]},{"k":[2,1]}]}`, true},

		{"value mismatch", `{"a":1}`, `{"a":2}`, false},
		{"string vs number", `{"a":"30"}`, `{"a":30}`, false},
		{"bool vs string", `{"a":true}`, `{"a":"true"}`, false},
		{"null vs zero", `{"a":null}`, `{"a":0}`, false},
		{"missing key", `{"a":1,"b":2}`, `{"a":1}`, false},
		{"extra key", `{"a":1}`, `{"a":1,"b":2}`, false},
		{"array length", `[1,2]`, `[1,2,3]`, false},
		{"array element absent", `[1,2,3]`, `[1,2,4]`, false},
		{"object vs array", `{"a":1}`, `[1]`, false},
		{"int precision preserved", `{"id":90071992547409921}`, `{"id":90071992547409922}`, false},
		{"duplicate multiset counts", `[1,1,2]`, `[1,2,2]`, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Compare([]byte(c.want), []byte(c.got))
			if c.wantEqual && err != nil {
				t.Errorf("expected equal, got diff: %v", err)
			}
			if !c.wantEqual && err == nil {
				t.Errorf("expected mismatch, got equal")
			}
		})
	}
}
