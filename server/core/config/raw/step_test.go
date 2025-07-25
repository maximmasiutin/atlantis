package raw_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStepConfig_YAMLMarshalling(t *testing.T) {
	cases := []struct {
		description string
		input       string
		exp         raw.Step
		expErr      string
	}{

		// Single string.
		{
			description: "single string",
			input:       `astring`,
			exp: raw.Step{
				Key: String("astring"),
			},
		},

		// MapType i.e. extra_args style.
		{
			description: "extra_args style",
			input: `
key:
  mapValue: [arg1, arg2]`,
			exp: raw.Step{
				Map: MapType{
					"key": {
						"mapValue": {"arg1", "arg2"},
					},
				},
			},
		},
		{
			description: "extra_args style multiple keys",
			input: `
key:
  mapValue: [arg1, arg2]
  value2: []`,
			exp: raw.Step{
				Map: MapType{
					"key": {
						"mapValue": {"arg1", "arg2"},
						"value2":   {},
					},
				},
			},
		},
		{
			description: "extra_args style multiple top-level keys",
			input: `
key:
  val1: []
key2:
  val2: []`,
			exp: raw.Step{
				Map: MapType{
					"key": {
						"val1": {},
					},
					"key2": {
						"val2": {},
					},
				},
			},
		},
		// Env steps
		{
			description: "env step value",
			input: `
env:
  value: direct_value
  name: test`,
			exp: raw.Step{
				CommandMap: EnvType{
					"env": {
						"value": "direct_value",
						"name":  "test",
					},
				},
			},
		},
		{
			description: "env step command",
			input: `
env:
  command: echo 123
  name: test`,
			exp: raw.Step{
				CommandMap: EnvType{
					"env": {
						"command": "echo 123",
						"name":    "test",
					},
				},
			},
		},

		// Run-step style
		{
			description: "run step",
			input: `
run: my command`,
			exp: raw.Step{
				StringVal: map[string]string{
					"run": "my command",
				},
			},
		},
		{
			description: "run step multiple top-level keys",
			input: `
run: my command
key: value`,
			exp: raw.Step{
				StringVal: map[string]string{
					"run": "my command",
					"key": "value",
				},
			},
		},

		// Empty
		{
			description: "empty",
			input:       "",
			exp: raw.Step{
				Key:        nil,
				Map:        nil,
				StringVal:  nil,
				CommandMap: nil,
			},
		},

		// Errors
		{
			description: "extra args style no map strings",
			input: `
key:
 - value:
     another: map`,
			expErr: `[3:2] sequence was used where mapping is expected
   2 | key:
>  3 |  - value:
        ^
   4 |      another: map`,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var got raw.Step
			err := unmarshalString(c.input, &got)
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
				return
			}
			Ok(t, err)
			Equals(t, c.exp, got)

			_, err = yaml.Marshal(got)
			Ok(t, err)

			var got2 raw.Step
			err = unmarshalString(c.input, &got2)
			Ok(t, err)
			Equals(t, got2, got)
		})
	}
}

func TestStep_Validate(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Step
		expErr      string
	}{
		// Valid inputs.
		{
			description: "init step",
			input: raw.Step{
				Key: String("init"),
			},
			expErr: "",
		},
		{
			description: "plan step",
			input: raw.Step{
				Key: String("plan"),
			},
			expErr: "",
		},
		{
			description: "apply step",
			input: raw.Step{
				Key: String("apply"),
			},
			expErr: "",
		},
		{
			description: "init extra_args",
			input: raw.Step{
				Map: MapType{
					"init": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			expErr: "",
		},
		{
			description: "plan extra_args",
			input: raw.Step{
				Map: MapType{
					"plan": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			expErr: "",
		},
		{
			description: "env",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"name":    "test",
						"command": "echo 123",
					},
				},
			},
			expErr: "",
		},
		{
			description: "env shell",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"name":    "test",
						"command": "echo 123",
						"shell":   "bash",
					},
				},
			},
			expErr: "",
		},
		{
			description: "env shellArgs string",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"name":      "test",
						"command":   "echo 123",
						"shell":     "bash",
						"shellArgs": "-c",
					},
				},
			},
			expErr: "",
		},
		{
			description: "env shellArgs list of strings",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"name":      "test",
						"command":   "echo 123",
						"shell":     "bash",
						"shellArgs": []interface{}{"-c", "--debug"},
					},
				},
			},
			expErr: "",
		},
		{
			description: "apply extra_args",
			input: raw.Step{
				Map: MapType{
					"apply": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			expErr: "",
		},
		{
			description: "run step",
			input: raw.Step{
				StringVal: map[string]string{
					"run": "my command",
				},
			},
			expErr: "",
		},

		// Invalid inputs.
		{
			description: "empty elem",
			input:       raw.Step{},
			expErr:      "step element is empty",
		},
		{
			description: "invalid step name",
			input: raw.Step{
				Key: String("invalid"),
			},
			expErr: "\"invalid\" is not a valid step type, maybe you omitted the 'run' key",
		},
		{
			description: "multiple keys in map",
			input: raw.Step{
				Map: MapType{
					"key1": nil,
					"key2": nil,
				},
			},
			expErr: "step element can only contain a single key, found 2: key1,key2",
		},
		{
			description: "multiple keys in env",
			input: raw.Step{
				CommandMap: EnvType{
					"key1": nil,
					"key2": nil,
				},
			},
			expErr: "step element can only contain a single key, found 2: key1,key2",
		},
		{
			description: "multiple keys in string val",
			input: raw.Step{
				StringVal: map[string]string{
					"key1": "",
					"key2": "",
				},
			},
			expErr: "step element can only contain a single key, found 2: key1,key2",
		},
		{
			description: "invalid key in map",
			input: raw.Step{
				Map: MapType{
					"invalid": nil,
				},
			},
			expErr: "\"invalid\" is not a valid step type",
		},
		{
			description: "invalid key in env",
			input: raw.Step{
				CommandMap: EnvType{
					"invalid": nil,
				},
			},
			expErr: "\"invalid\" is not a valid step type",
		},
		{
			description: "invalid key in string val",
			input: raw.Step{
				StringVal: map[string]string{
					"invalid": "",
				},
			},
			expErr: "\"invalid\" is not a valid step type",
		},
		{
			description: "non extra_arg key",
			input: raw.Step{
				Map: MapType{
					"init": {
						"invalid": nil,
					},
				},
			},
			expErr: "built-in steps only support a single extra_args key, found \"invalid\" in step init",
		},
		{
			description: "non extra_arg key",
			input: raw.Step{
				Map: MapType{
					"init": {
						"invalid": nil,
						"zzzzzzz": nil,
					},
				},
			},
			expErr: "built-in steps only support a single extra_args key, found 2: invalid,zzzzzzz",
		},
		{
			description: "env step with no name key set",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"value": "value",
					},
				},
			},
			expErr: "env steps must have a \"name\" key set",
		},
		{
			description: "env step with invalid key",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"abc":      "",
						"invalid2": "",
					},
				},
			},
			expErr: "env steps only support keys \"name\", \"value\", \"command\", \"shell\" and \"shellArgs\", found key \"abc\"",
		},
		{
			description: "env step with both command and value set",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"name":    "name",
						"command": "command",
						"value":   "value",
					},
				},
			},
			expErr: "env steps only support one of the \"value\" or \"command\" keys, found both",
		},
		{
			description: "env step with shell set but not command",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"name":  "name",
						"shell": "bash",
					},
				},
			},
			expErr: "workflow steps only support \"shell\" key in combination with \"command\" key",
		},
		{
			description: "env step with shellArgs set but not shell",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"name":      "name",
						"shellArgs": "-c",
					},
				},
			},
			expErr: "workflow steps only support \"shellArgs\" key in combination with \"shell\" key",
		},
		{
			description: "run step with shellArgs is not list of strings",
			input: raw.Step{
				CommandMap: EnvType{
					"run": {
						"name":      "name",
						"command":   "echo",
						"shell":     "shell",
						"shellArgs": []int{42, 42},
					},
				},
			},
			expErr: "\"run\" step \"shellArgs\" option must be a string or a list of strings, found [42 42]\n",
		},
		{
			description: "run step with shellArgs contain not strings",
			input: raw.Step{
				CommandMap: EnvType{
					"run": {
						"name":      "name",
						"command":   "echo",
						"shell":     "shell",
						"shellArgs": []interface{}{"-c", 42},
					},
				},
			},
			expErr: "\"run\" step \"shellArgs\" option must contain only strings, found 42\n",
		},
		{
			// For atlantis.yaml v2, this wouldn't parse, but now there should
			// be no error.
			description: "unparsable shell command",
			input: raw.Step{
				StringVal: map[string]string{
					"run": "my 'c",
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := c.input.Validate()
			if c.expErr == "" {
				Ok(t, err)
				return
			}
			ErrEquals(t, c.expErr, err)
		})
	}
}

func TestStep_ToValid(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Step
		exp         valid.Step
	}{
		{
			description: "init step",
			input: raw.Step{
				Key: String("init"),
			},
			exp: valid.Step{
				StepName: "init",
			},
		},
		{
			description: "plan step",
			input: raw.Step{
				Key: String("plan"),
			},
			exp: valid.Step{
				StepName: "plan",
			},
		},
		{
			description: "policy_check step",
			input: raw.Step{
				Key: String("policy_check"),
			},
			exp: valid.Step{
				StepName: "policy_check",
			},
		},
		{
			description: "apply step",
			input: raw.Step{
				Key: String("apply"),
			},
			exp: valid.Step{
				StepName: "apply",
			},
		},
		{
			description: "env step",
			input: raw.Step{
				CommandMap: EnvType{
					"env": {
						"name":    "test",
						"command": "echo 123",
					},
				},
			},
			exp: valid.Step{
				StepName:   "env",
				RunCommand: "echo 123",
				EnvVarName: "test",
			},
		},
		{
			description: "import step",
			input: raw.Step{
				Key: String("import"),
			},
			exp: valid.Step{
				StepName: "import",
			},
		},
		{
			description: "init extra_args",
			input: raw.Step{
				Map: MapType{
					"init": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			exp: valid.Step{
				StepName:  "init",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "plan extra_args",
			input: raw.Step{
				Map: MapType{
					"plan": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			exp: valid.Step{
				StepName:  "plan",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "policy_check extra_args",
			input: raw.Step{
				Map: MapType{
					"policy_check": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			exp: valid.Step{
				StepName:  "policy_check",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "apply extra_args",
			input: raw.Step{
				Map: MapType{
					"apply": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			exp: valid.Step{
				StepName:  "apply",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "import extra_args",
			input: raw.Step{
				Map: MapType{
					"import": {
						"extra_args": []string{"arg1", "arg2"},
					},
				},
			},
			exp: valid.Step{
				StepName:  "import",
				ExtraArgs: []string{"arg1", "arg2"},
			},
		},
		{
			description: "run step",
			input: raw.Step{
				StringVal: map[string]string{
					"run": "my 'run command'",
				},
			},
			exp: valid.Step{
				StepName:   "run",
				RunCommand: "my 'run command'",
			},
		},
		{
			description: "run step with output",
			input: raw.Step{
				CommandMap: RunType{
					"run": {
						"command": "my 'run command'",
						"output":  "hide",
					},
				},
			},
			exp: valid.Step{
				StepName:   "run",
				RunCommand: "my 'run command'",
				Output:     "hide",
			},
		},
		{
			description: "multienv step",
			input: raw.Step{
				StringVal: map[string]string{
					"multienv": "envs.sh",
				},
			},
			exp: valid.Step{
				StepName:   "multienv",
				RunCommand: "envs.sh",
			},
		},
		{
			description: "multienv step with output",
			input: raw.Step{
				CommandMap: MultiEnvType{
					"multienv": {
						"command": "envs.sh",
						"output":  "hide",
					},
				},
			},
			exp: valid.Step{
				StepName:   "multienv",
				RunCommand: "envs.sh",
				Output:     "hide",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.exp, c.input.ToValid())
		})
	}
}

type MapType map[string]map[string][]string
type EnvType map[string]map[string]interface{}
type RunType map[string]map[string]interface{}
type MultiEnvType map[string]map[string]interface{}
