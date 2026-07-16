package template

type TelegramLinked struct{}

func (TelegramLinked) String() string {
	return "Successfully linked your account <3"
}

type TelegramLinkFailed struct{}

func (TelegramLinkFailed) String() string {
	return "Something went wrong. Please try again later."
}

type TelegramLinkTokenExpired struct{}

func (TelegramLinkTokenExpired) String() string {
	return "This link has expired or is invalid. Please generate a new link in the app."
}

type TelegramUserNotFound struct{}

func (TelegramUserNotFound) String() string {
	return "User account not found."
}
