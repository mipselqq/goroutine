package template

type TelegramLinkedNotif struct{}

func (TelegramLinkedNotif) String() string {
	return "Successfully linked your account <3"
}

type TelegramLinkFailedNotif struct{}

func (TelegramLinkFailedNotif) String() string {
	return "Something went wrong. Please try again later."
}

type TelegramLinkTokenExpiredNotif struct{}

func (TelegramLinkTokenExpiredNotif) String() string {
	return "This link has expired or is invalid. Please generate a new link in the app."
}

type TelegramUserNotFoundNotif struct{}

func (TelegramUserNotFoundNotif) String() string {
	return "User account not found."
}
