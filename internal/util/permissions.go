package util

import "strings"

// PermissionsString joins permission slices into a canonical comma separated value.
func PermissionsString(perms []string) string {
	lowered := make([]string, 0, len(perms))
	for _, p := range perms {
		normalized := NormalizePermission(p)
		if normalized == "" {
			continue
		}
		lowered = append(lowered, normalized)
	}
	return strings.Join(lowered, ",")
}

// SplitPermissions splits a comma separated string into unique lower-case items.
func SplitPermissions(s string) []string {
	if s == "" {
		return []string{}
	}
	seen := map[string]struct{}{}
	parts := strings.Split(s, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.ToLower(strings.TrimSpace(p))
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		res = append(res, v)
	}
	return res
}

// NormalizePermission lowercases and trims a permission string.
func NormalizePermission(p string) string {
	return strings.ToLower(strings.TrimSpace(p))
}
