// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package argsbuilder

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
)

func transformSliceToMap(sets []string, split string) (valuesMap map[string]string) {
	valuesMap = map[string]string{}
	for _, member := range sets {
		splits := strings.SplitN(member, split, 2)
		if len(splits) == 2 {
			valuesMap[splits[0]] = splits[1]
		}
	}

	return valuesMap
}

// parseTolerationString parses a toleration string into TolerationArgs.
// Supported formats:
//   - Simple: "key" (tolerate any taint with this key, operator=Exists)
//   - Without value: "key:effect", "key:effect:operator", "key:effect:operator:tolerationSeconds"
//   - With value (legacy): "key=value:effect,operator" (e.g., "gpu_node=invalid:NoSchedule,Equal")
//   - With value (new): "key=value:effect:operator", "key=value:effect:operator:tolerationSeconds"
//
// Examples:
//   - "gpu_node" -> key=gpu_node, operator=Exists
//   - "node.kubernetes.io/unreachable:NoExecute:Exists:60" -> key, effect, operator, seconds
//   - "dedicated=teamA:NoSchedule:Equal" -> key, value, effect, operator
//   - "dedicated=teamA:NoExecute:Equal:300" -> key, value, effect, operator, seconds
func parseTolerationString(toleration string) (*types.TolerationArgs, error) {
	result := &types.TolerationArgs{}

	// Check if the toleration contains "=" (has value)
	if strings.Contains(toleration, "=") {
		// Split by "=" to get key and the rest
		eqIndex := strings.Index(toleration, "=")
		key := toleration[:eqIndex]
		rest := toleration[eqIndex+1:]

		result.Key = key

		// Check if it's legacy format with comma: key=value:effect,operator
		if strings.Contains(rest, ",") {
			if !strings.Contains(rest, ":") {
				return nil, fmt.Errorf("invalid toleration format: '%s'. Expected 'key=value:effect,operator'", toleration)
			}
			value, rest := split(rest, ":")
			effect, operator := split(rest, ",")

			if err := validateTolerationEffect(effect); err != nil {
				return nil, err
			}
			if err := validateTolerationOperator(operator); err != nil {
				return nil, err
			}
			if err := validateTolerationValueOperator(value, operator); err != nil {
				return nil, err
			}

			result.Value = value
			result.Effect = effect
			result.Operator = operator
			return result, nil
		}

		// New format with value: key=value:effect:operator[:tolerationSeconds]
		parts := strings.Split(rest, ":")
		switch len(parts) {
		case 1:
			// key=value only - invalid, need at least effect
			return nil, fmt.Errorf("invalid toleration format: '%s'. Expected 'key=value:effect:operator' or 'key=value:effect:operator:tolerationSeconds'", toleration)
		case 2:
			// key=value:effect - use Equal operator by default for value-based tolerations
			result.Value = parts[0]
			result.Effect = parts[1]
			result.Operator = "Equal"
			if err := validateTolerationEffect(result.Effect); err != nil {
				return nil, err
			}
			if err := validateTolerationValueOperator(result.Value, result.Operator); err != nil {
				return nil, err
			}
		case 3:
			// key=value:effect:operator
			result.Value = parts[0]
			result.Effect = parts[1]
			result.Operator = parts[2]
			if err := validateTolerationEffect(result.Effect); err != nil {
				return nil, err
			}
			if err := validateTolerationOperator(result.Operator); err != nil {
				return nil, err
			}
			if err := validateTolerationValueOperator(result.Value, result.Operator); err != nil {
				return nil, err
			}
		case 4:
			// key=value:effect:operator:tolerationSeconds
			result.Value = parts[0]
			result.Effect = parts[1]
			result.Operator = parts[2]
			if err := validateTolerationEffect(result.Effect); err != nil {
				return nil, err
			}
			if err := validateTolerationOperator(result.Operator); err != nil {
				return nil, err
			}
			if err := validateTolerationValueOperator(result.Value, result.Operator); err != nil {
				return nil, err
			}
			seconds, err := parseTolerationSeconds(parts[3])
			if err != nil {
				return nil, err
			}
			result.TolerationSeconds = &seconds
		default:
			return nil, fmt.Errorf("invalid toleration format: '%s'. Too many ':' separators after '='", toleration)
		}
		if err := validateTolerationValueOperator(result.Value, result.Operator); err != nil {
			return nil, err
		}
		return result, nil
	}

	// No "=" - format without value: key[:effect[:operator[:tolerationSeconds]]]
	parts := strings.Split(toleration, ":")
	switch len(parts) {
	case 1:
		// Format: key (tolerate any taint with this key)
		result.Key = parts[0]
		result.Operator = "Exists"
	case 2:
		// Format: key:effect
		result.Key = parts[0]
		result.Effect = parts[1]
		result.Operator = "Exists"
		if err := validateTolerationEffect(result.Effect); err != nil {
			return nil, err
		}
	case 3:
		// Format: key:effect:operator
		result.Key = parts[0]
		result.Effect = parts[1]
		result.Operator = parts[2]
		if err := validateTolerationEffect(result.Effect); err != nil {
			return nil, err
		}
		if err := validateTolerationOperator(result.Operator); err != nil {
			return nil, err
		}
	case 4:
		// Format: key:effect:operator:tolerationSeconds
		result.Key = parts[0]
		result.Effect = parts[1]
		result.Operator = parts[2]
		if err := validateTolerationEffect(result.Effect); err != nil {
			return nil, err
		}
		if err := validateTolerationOperator(result.Operator); err != nil {
			return nil, err
		}
		seconds, err := parseTolerationSeconds(parts[3])
		if err != nil {
			return nil, err
		}
		result.TolerationSeconds = &seconds
	default:
		// For keys containing ":", try to parse as key:effect:operator:seconds from the end
		// e.g., "node.kubernetes.io/unreachable:NoExecute:Exists:60"
		if len(parts) >= 4 {
			// Try to parse the last part as tolerationSeconds
			if seconds, err := parseTolerationSeconds(parts[len(parts)-1]); err == nil {
				operator := parts[len(parts)-2]
				effect := parts[len(parts)-3]
				key := strings.Join(parts[:len(parts)-3], ":")

				if err := validateTolerationEffect(effect); err != nil {
					return nil, err
				}
				if err := validateTolerationOperator(operator); err != nil {
					return nil, err
				}

				result.TolerationSeconds = &seconds
				result.Operator = operator
				result.Effect = effect
				result.Key = key
			} else {
				// Last part is not a valid number, treat as key:effect:operator
				operator := parts[len(parts)-1]
				effect := parts[len(parts)-2]
				key := strings.Join(parts[:len(parts)-2], ":")

				if err := validateTolerationEffect(effect); err != nil {
					return nil, err
				}
				if err := validateTolerationOperator(operator); err != nil {
					return nil, err
				}

				result.Operator = operator
				result.Effect = effect
				result.Key = key
			}
		} else {
			return nil, fmt.Errorf("invalid toleration format: '%s'", toleration)
		}
	}

	if err := validateTolerationValueOperator(result.Value, result.Operator); err != nil {
		return nil, err
	}
	return result, nil
}

// parseTolerationSeconds parses and validates tolerationSeconds value
func parseTolerationSeconds(s string) (int64, error) {
	seconds, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid tolerationSeconds '%s': must be a number", s)
	}
	if seconds < 0 {
		return 0, fmt.Errorf("invalid tolerationSeconds '%s': must be >= 0", s)
	}
	return seconds, nil
}

// validateTolerationEffect validates the toleration effect value
func validateTolerationEffect(effect string) error {
	switch effect {
	case "NoSchedule", "PreferNoSchedule", "NoExecute", "":
		return nil
	default:
		return fmt.Errorf("invalid toleration effect '%s': must be NoSchedule, PreferNoSchedule, or NoExecute", effect)
	}
}

// validateTolerationOperator validates the toleration operator value
func validateTolerationOperator(operator string) error {
	switch operator {
	case "Exists", "Equal", "":
		return nil
	default:
		return fmt.Errorf("invalid toleration operator '%s': must be Exists or Equal", operator)
	}
}

// validateTolerationValueOperator enforces value/operator compatibility rules.
func validateTolerationValueOperator(value, operator string) error {
	if value == "" && operator == "Equal" {
		return fmt.Errorf("operator 'Equal' requires a non-empty value. Use 'key=value:effect:Equal'")
	}
	if value != "" && operator != "" && operator != "Equal" {
		return fmt.Errorf("operator '%s' cannot be combined with a value. Use 'Equal' or omit the value", operator)
	}
	return nil
}

func split(value, sep string) (string, string) {
	index := strings.Index(value, sep)
	return value[:index], value[index+1:]
}
