package handler

type RoleResponseFormatter interface {
	Format(data any) any
}

type GenericEntity interface {
	ToPublicFormat() any
}

type roleResponseFormatterFunc[I any] func(data I) interface{}

func responseFormatter[I GenericEntity](data I, roles []string) any {
	if roles == nil {
		return data.ToPublicFormat()
	}

	for _, role := range roles {
		if role == "admin" {
			return data
		}
	}

	return data.ToPublicFormat()
}

func responseArrFormatter[I GenericEntity](data []I, roles []string) []any {
	res := []any{}
	for _, v := range data {
		res = append(res, responseFormatter(v, roles))
	}
	return res
}
