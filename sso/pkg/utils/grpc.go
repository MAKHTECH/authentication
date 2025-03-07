package utils

import ssov1 "sso/protos/gen/go/sso"

// GetValidRoles возвращает все допустимые значения enum Role из Role_name
func GetValidRoles() []interface{} {
	validRoles := make([]interface{}, 0, len(ssov1.Role_name))
	for roleValue := range ssov1.Role_name {
		validRoles = append(validRoles, ssov1.Role(roleValue))
	}
	return validRoles
}
