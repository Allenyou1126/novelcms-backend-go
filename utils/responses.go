package utils

type CommonResponse struct {
	Message string `json:"message"`
}

func BuildCommonResponse(msg string) CommonResponse {
	return CommonResponse{Message: msg}
}

func RequireToken() CommonResponse { return BuildCommonResponse("Require Token.") }
func InvalidToken() CommonResponse { return BuildCommonResponse("Invalid Token.") }

type PermissionErrorResponse struct {
	CommonResponse
	RequiredPermission string `json:"required-permission"`
}

func RequirePermission(target string) PermissionErrorResponse {
	return PermissionErrorResponse{
		CommonResponse:     BuildCommonResponse("Permission denied."),
		RequiredPermission: target,
	}
}
