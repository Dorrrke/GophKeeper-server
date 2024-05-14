package errors

const (
	InvalidAuthError             = "invalid login or password"
	UserExistsError              = "user is alredy exists"
	UserNotExistError            = "user not found"
	CardExistsError              = "card is alredy exists"
	LoginExistsError             = "login is alredy exists"
	TextExistsError              = "text is alredy exists"
	CardNotExistsError           = "card not found"
	LoginNotExistsError          = "login not found"
	TextNotExistsError           = "text not found"
	InvalidPasswordError         = "invalid password"
	DataDecryptError             = "could not decrypt data"
	BinDataExistsError           = "bin data is alredy exists"
	BinDataNotExistsError        = "bin data not found"
	InvalidLoginUserError        = "invalid login/password pair; this user does not exist"
	MetadataError                = "metadata does not exist"
	MissingAuthorizationKeyError = "missing authorization key"
	InvalidTokenError            = "invalid token"
)
