package handler

type RoleResponseFormatter interface {
	Format(data any) any
}

type GenericEntity interface {
	ToPublicFormat(domain string) any
}

type roleResponseFormatterFunc[I any] func(data I) interface{}

func responseFormatter[I GenericEntity](data I, roles []string, domain string) any {
	if roles == nil {
		return data.ToPublicFormat(domain)
	}

	for _, role := range roles {
		if role == "admin" {
			return data
		}
	}

	return data.ToPublicFormat(domain)
}

func responseArrFormatter[I GenericEntity](data []I, roles []string, domain string) []any {
	res := []any{}
	for _, v := range data {
		res = append(res, responseFormatter(v, roles, domain))
	}
	return res
}
