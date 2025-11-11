package responses

// MigrateSuccessResponse is the success payload for POST /migrate.
type MigrateSuccessResponse struct {
	Inserted int `json:"inserted"`
}

// MigrateRowError is the HTTP DTO for row-level validation/conflict details.
type MigrateRowError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// MigrateErrorResponse is the error payload for POST /migrate.
type MigrateErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  []MigrateRowError `json:"errors"`
}
